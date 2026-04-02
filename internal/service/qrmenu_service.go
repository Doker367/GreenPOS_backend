package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/google/uuid"
	"github.com/greenpos/backend/internal/model"
	"github.com/greenpos/backend/internal/repository"
)

// QRMenuService handles QR menu operations
type QRMenuService struct {
	qrCodes    repository.QRCodeRepositoryInterface
	categories repository.CategoryRepositoryInterface
	products   repository.ProductRepositoryInterface
	tables     repository.TableRepositoryInterface
	orders     repository.OrderRepositoryInterface
	taxRate    float64
	baseURL    string
	log        *slog.Logger
}

// NewQRMenuService creates a new QR menu service
func NewQRMenuService(
	qrCodes repository.QRCodeRepositoryInterface,
	categories repository.CategoryRepositoryInterface,
	products repository.ProductRepositoryInterface,
	tables repository.TableRepositoryInterface,
	orders repository.OrderRepositoryInterface,
	taxRate float64,
	baseURL string,
	log *slog.Logger,
) *QRMenuService {
	if log == nil {
		log = slog.Default()
	}
	return &QRMenuService{
		qrCodes:    qrCodes,
		categories: categories,
		products:   products,
		tables:     tables,
		orders:     orders,
		taxRate:    taxRate,
		baseURL:    baseURL,
		log:        log,
	}
}

var (
	ErrQRCodeNotFound      = errors.New("QR code not found")
	ErrQRCodeInactive      = errors.New("QR code is inactive")
	ErrTableNotFoundForQR  = errors.New("table not found for QR code")
	ErrProductNotAvailable = errors.New("product is not available")
	ErrInvalidQuantity     = errors.New("invalid quantity")
	ErrEmptyOrder          = errors.New("order has no items")
)

// generateToken creates a random unique token for QR codes
func generateToken() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// GenerateTableQR generates or retrieves a QR code for a table
func (s *QRMenuService) GenerateTableQR(ctx context.Context, tableID uuid.UUID) (*model.TableQRCode, error) {
	// Check if table exists
	table, err := s.tables.GetByID(ctx, tableID)
	if err != nil {
		return nil, ErrTableNotFound
	}

	// Check if QR code already exists for this table
	existing, err := s.qrCodes.GetByTableID(ctx, tableID)
	if err == nil {
		// QR code exists, reactivate if needed and return
		if !existing.IsActive {
			existing.IsActive = true
			if err := s.qrCodes.Update(ctx, existing); err != nil {
				return nil, fmt.Errorf("failed to reactivate QR code: %w", err)
			}
		}
		return existing, nil
	}
	if !errors.Is(err, repository.ErrNotFound) {
		return nil, fmt.Errorf("failed to check existing QR code: %w", err)
	}

	// Generate new token
	token, err := generateToken()
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	// Build access URL
	accessURL := fmt.Sprintf("%s/menu/%s/%s", s.baseURL, table.BranchID.String(), token)

	qrCode := &model.TableQRCode{
		ID:        uuid.New(),
		TableID:   tableID,
		QRToken:   token,
		AccessURL: accessURL,
		IsActive:  true,
	}

	if err := s.qrCodes.Create(ctx, qrCode); err != nil {
		return nil, fmt.Errorf("failed to create QR code: %w", err)
	}

	s.log.Info("QR code generated for table",
		slog.String("table_id", tableID.String()),
		slog.String("table_number", table.Number),
		slog.String("qr_token", token))

	return qrCode, nil
}

// GetQRByToken retrieves QR code info by token
func (s *QRMenuService) GetQRByToken(ctx context.Context, token string) (*model.TableQRCode, error) {
	qr, err := s.qrCodes.GetByToken(ctx, token)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrQRCodeNotFound
		}
		return nil, err
	}
	if !qr.IsActive {
		return nil, ErrQRCodeInactive
	}
	return qr, nil
}

// GetPublicMenu returns the public menu (categories with available products) for a branch
func (s *QRMenuService) GetPublicMenu(ctx context.Context, branchID uuid.UUID) ([]model.CategoryWithProducts, error) {
	categories, err := s.categories.GetByBranch(ctx, branchID)
	if err != nil {
		return nil, fmt.Errorf("failed to get categories: %w", err)
	}

	products, err := s.products.GetAvailable(ctx, branchID)
	if err != nil {
		return nil, fmt.Errorf("failed to get products: %w", err)
	}

	// Group products by category
	productsByCategory := make(map[uuid.UUID][]model.ProductPublic)
	for _, p := range products {
		var catID uuid.UUID
		if p.CategoryID != nil {
			catID = *p.CategoryID
		}
		productsByCategory[catID] = append(productsByCategory[catID], model.ProductPublic{
			ID:              p.ID,
			Name:            p.Name,
			Description:     p.Description,
			Price:           p.Price,
			ImageURL:        p.ImageURL,
			PreparationTime: p.PreparationTime,
			Allergens:       p.Allergens,
		})
	}

	// Build result with categories that have products
	var result []model.CategoryWithProducts
	for _, cat := range categories {
		if cat.IsActive {
			catProducts := productsByCategory[cat.ID]
			result = append(result, model.CategoryWithProducts{
				ID:          cat.ID,
				Name:        cat.Name,
				Description: cat.Description,
				SortOrder:   cat.SortOrder,
				Products:    catProducts,
			})
		}
	}

	// Add uncategorized products
	if uncategorized, ok := productsByCategory[uuid.Nil]; ok && len(uncategorized) > 0 {
		result = append(result, model.CategoryWithProducts{
			ID:        uuid.Nil,
			Name:      "Otros",
			SortOrder: 9999,
			Products:  uncategorized,
		})
	}

	return result, nil
}

// ClientOrderItemInput represents an item in a client order
type ClientOrderItemInput struct {
	ProductID uuid.UUID `json:"productId"`
	Quantity  int       `json:"quantity"`
	Notes     string    `json:"notes"`
}

// CreateClientOrderInput represents the input for creating a client order
type CreateClientOrderInput struct {
	QRCodeToken   string                `json:"qrCodeToken"`
	CustomerName  string                `json:"customerName"`
	CustomerPhone string                `json:"customerPhone"`
	Items         []ClientOrderItemInput `json:"items"`
	Notes         string                `json:"notes"`
}

// CreateClientOrder creates an order from the QR menu
func (s *QRMenuService) CreateClientOrder(ctx context.Context, input CreateClientOrderInput) (*model.ClientOrderConfirmation, error) {
	if len(input.Items) == 0 {
		return nil, ErrEmptyOrder
	}

	// Validate all items and calculate totals
	var subtotal float64
	orderItems := make([]struct {
		Product    model.Product
		Quantity   int
		Notes      string
		TotalPrice float64
	}, 0, len(input.Items))

	maxPrepTime := 0

	for _, item := range input.Items {
		if item.Quantity <= 0 {
			return nil, ErrInvalidQuantity
		}

		product, err := s.products.GetByID(ctx, item.ProductID)
		if err != nil {
			return nil, fmt.Errorf("product not found: %s", item.ProductID.String())
		}
		if !product.IsAvailable {
			return nil, fmt.Errorf("%w: %s", ErrProductNotAvailable, product.Name)
		}

		itemTotal := product.Price * float64(item.Quantity)
		subtotal += itemTotal

		if product.PreparationTime > maxPrepTime {
			maxPrepTime = product.PreparationTime
		}

		orderItems = append(orderItems, struct {
			Product    model.Product
			Quantity   int
			Notes      string
			TotalPrice float64
		}{
			Product:    *product,
			Quantity:   item.Quantity,
			Notes:      item.Notes,
			TotalPrice: itemTotal,
		})
	}

	// Get QR code and table info
	qr, err := s.qrCodes.GetByToken(ctx, input.QRCodeToken)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrQRCodeNotFound
		}
		return nil, err
	}
	if !qr.IsActive {
		return nil, ErrQRCodeInactive
	}

	table, err := s.tables.GetByID(ctx, qr.TableID)
	if err != nil {
		return nil, ErrTableNotFoundForQR
	}

	tax := subtotal * s.taxRate
	total := subtotal + tax

	// Generate order token
	orderToken, err := generateToken()
	if err != nil {
		return nil, fmt.Errorf("failed to generate order token: %w", err)
	}

	// Create the order
	order := &model.Order{
		ID:            uuid.New(),
		BranchID:      table.BranchID,
		TableID:       &qr.TableID,
		CustomerName:  input.CustomerName,
		CustomerPhone: input.CustomerPhone,
		Status:        model.OrderPending,
		Subtotal:      subtotal,
		Tax:           tax,
		Total:         total,
		Notes:         input.Notes,
	}

	// We need a user ID for the order - use a system user or create anonymous
	// For now, we'll use a placeholder system user ID
	systemUserID := uuid.MustParse("00000000-0000-0000-0000-000000000000")

	// Create order in repository
	if err := s.orders.Create(ctx, order); err != nil {
		return nil, fmt.Errorf("failed to create order: %w", err)
	}

	// Create order items
	for _, item := range orderItems {
		orderItem := &model.OrderItem{
			ID:         uuid.New(),
			OrderID:    order.ID,
			ProductID:  item.Product.ID,
			Quantity:   item.Quantity,
			UnitPrice:  item.Product.Price,
			TotalPrice: item.TotalPrice,
			Notes:      item.Notes,
		}
		if err := s.orders.AddItem(ctx, orderItem); err != nil {
			return nil, fmt.Errorf("failed to add order item: %w", err)
		}
	}

	s.log.Info("client order created via QR menu",
		slog.String("order_id", order.ID.String()),
		slog.String("table_id", qr.TableID.String()),
		slog.String("table_number", table.Number),
		slog.Float64("total", total))

	// Build confirmation message
	message := fmt.Sprintf("¡Pedido #%s recibido! Tu pedido está siendo preparado.", strings.ToUpper(orderToken[:8]))

	return &model.ClientOrderConfirmation{
		OrderID:             order.ID,
		OrderToken:          orderToken,
		Status:              string(model.OrderPending),
		Message:             message,
		EstimatedReadyTime:  &maxPrepTime,
	}, nil
}

// GetTableByQRToken returns the table associated with a QR token
func (s *QRMenuService) GetTableByQRToken(ctx context.Context, token string) (*model.RestaurantTable, error) {
	qr, err := s.qrCodes.GetByToken(ctx, token)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrQRCodeNotFound
		}
		return nil, err
	}
	if !qr.IsActive {
		return nil, ErrQRCodeInactive
	}
	return s.tables.GetByID(ctx, qr.TableID)
}

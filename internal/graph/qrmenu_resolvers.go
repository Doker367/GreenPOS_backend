package graph

import (
	"context"

	"github.com/google/uuid"
	"github.com/greenpos/backend/internal/model"
	"github.com/greenpos/backend/internal/service"
)

// ============ QR Menu Types ============

// TableQR represents a table QR code for the GraphQL schema
type TableQR struct {
	ID        uuid.UUID `json:"id"`
	TableID   uuid.UUID `json:"tableId"`
	QRToken   string    `json:"qrToken"`
	AccessURL string    `json:"accessUrl"`
	IsActive  bool      `json:"isActive"`
	CreatedAt string    `json:"createdAt"`
}

// CategoryWithProductsGQL represents a category with products for the public menu
type CategoryWithProductsGQL struct {
	ID          uuid.UUID        `json:"id"`
	Name        string           `json:"name"`
	Description string           `json:"description"`
	SortOrder   int              `json:"sortOrder"`
	Products    []ProductPublicGQL `json:"products"`
}

// ProductPublicGQL represents a product in the public menu
type ProductPublicGQL struct {
	ID              uuid.UUID `json:"id"`
	Name            string    `json:"name"`
	Description     string    `json:"description"`
	Price           float64   `json:"price"`
	ImageURL        string    `json:"imageUrl"`
	PreparationTime int       `json:"preparationTime"`
	Allergens       []string  `json:"allergens"`
}

// ClientOrderItemInputGQL is the GraphQL input type for client order items
type ClientOrderItemInputGQL struct {
	ProductID uuid.UUID `json:"productId"`
	Quantity  int       `json:"quantity"`
	Notes     string    `json:"notes"`
}

// CreateClientOrderInputGQL is the GraphQL input type for creating a client order
type CreateClientOrderInputGQL struct {
	QRCodeToken   string                    `json:"qrCodeToken"`
	CustomerName  string                    `json:"customerName"`
	CustomerPhone string                    `json:"customerPhone"`
	Items         []ClientOrderItemInputGQL `json:"items"`
	Notes         string                    `json:"notes"`
}

// ClientOrderConfirmationGQL is the GraphQL return type for client order confirmation
type ClientOrderConfirmationGQL struct {
	OrderID             uuid.UUID `json:"orderId"`
	OrderToken          string    `json:"orderToken"`
	Status              string    `json:"status"`
	Message             string    `json:"message"`
	EstimatedReadyTime *int      `json:"estimatedReadyTime"`
}

// ============ QR Menu Resolvers ============

// GenerateTableQR is the resolver for the generateTableQR mutation
func (r *mutationResolver) GenerateTableQR(ctx context.Context, tableID uuid.UUID) (*TableQR, error) {
	qr, err := r.Services.QRMenu.GenerateTableQR(ctx, tableID)
	if err != nil {
		return nil, err
	}

	return &TableQR{
		ID:        qr.ID,
		TableID:   qr.TableID,
		QRToken:   qr.QRToken,
		AccessURL: qr.AccessURL,
		IsActive:  qr.IsActive,
		CreatedAt: qr.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}, nil
}

// PublicMenu is the resolver for the publicMenu query
func (r *queryResolver) PublicMenu(ctx context.Context, branchID uuid.UUID) ([]CategoryWithProductsGQL, error) {
	categories, err := r.Services.QRMenu.GetPublicMenu(ctx, branchID)
	if err != nil {
		return nil, err
	}

	result := make([]CategoryWithProductsGQL, 0, len(categories))
	for _, cat := range categories {
		products := make([]ProductPublicGQL, 0, len(cat.Products))
		for _, p := range cat.Products {
			products = append(products, ProductPublicGQL{
				ID:              p.ID,
				Name:            p.Name,
				Description:     p.Description,
				Price:           p.Price,
				ImageURL:        p.ImageURL,
				PreparationTime: p.PreparationTime,
				Allergens:       p.Allergens,
			})
		}
		result = append(result, CategoryWithProductsGQL{
			ID:          cat.ID,
			Name:        cat.Name,
			Description: cat.Description,
			SortOrder:   cat.SortOrder,
			Products:    products,
		})
	}

	return result, nil
}

// CreateClientOrder is the resolver for the createClientOrder mutation
func (r *mutationResolver) CreateClientOrder(ctx context.Context, input CreateClientOrderInputGQL) (*ClientOrderConfirmationGQL, error) {
	items := make([]service.ClientOrderItemInput, 0, len(input.Items))
	for _, item := range input.Items {
		items = append(items, service.ClientOrderItemInput{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
			Notes:     item.Notes,
		})
	}

	serviceInput := service.CreateClientOrderInput{
		QRCodeToken:   input.QRCodeToken,
		CustomerName: input.CustomerName,
		CustomerPhone: input.CustomerPhone,
		Items:         items,
		Notes:         input.Notes,
	}

	confirmation, err := r.Services.QRMenu.CreateClientOrder(ctx, serviceInput)
	if err != nil {
		return nil, err
	}

	return &ClientOrderConfirmationGQL{
		OrderID:             confirmation.OrderID,
		OrderToken:          confirmation.OrderToken,
		Status:              confirmation.Status,
		Message:             confirmation.Message,
		EstimatedReadyTime:  confirmation.EstimatedReadyTime,
	}, nil
}

// TableByQRToken resolves table info from a QR token (for the public menu)
func (r *queryResolver) TableByQRToken(ctx context.Context, token string) (*Table, error) {
	table, err := r.Services.QRMenu.GetTableByQRToken(ctx, token)
	if err != nil {
		return nil, err
	}

	qr, _ := r.Services.QRMenu.GetQRByToken(ctx, token)

	return &Table{
		ID:        table.ID,
		BranchID:  table.BranchID,
		Number:    table.Number,
		Capacity:  table.Capacity,
		Status:    table.Status,
		CreatedAt: table.CreatedAt,
		UpdatedAt: table.UpdatedAt,
		QRCode:    qr.AccessURL,
	}, nil
}

// GetQRByToken retrieves QR code info by token
func (r *queryResolver) GetQRByToken(ctx context.Context, token string) (*TableQR, error) {
	qr, err := r.Services.QRMenu.GetQRByToken(ctx, token)
	if err != nil {
		return nil, err
	}

	return &TableQR{
		ID:        qr.ID,
		TableID:   qr.TableID,
		QRToken:   qr.QRToken,
		AccessURL: qr.AccessURL,
		IsActive:  qr.IsActive,
		CreatedAt: qr.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}, nil
}

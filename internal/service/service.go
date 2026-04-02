package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/greenpos/backend/internal/model"
	"github.com/greenpos/backend/internal/repository"
)

// DefaultTaxRate is the default tax rate (16%)
const DefaultTaxRate = 0.16

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserInactive       = errors.New("user is inactive")
	ErrInvalidToken       = errors.New("invalid token")
	ErrOrderNotCancellable = errors.New("order cannot be cancelled in current state")
	ErrTableNotAvailable  = errors.New("table is not available")
	ErrProductNotFound    = errors.New("product not found")
	ErrCategoryNotFound   = errors.New("category not found")
	ErrTableNotFound      = errors.New("table not found")
	ErrOrderNotFound      = errors.New("order not found")
	ErrReservationNotFound = errors.New("reservation not found")
	ErrUnauthorized       = errors.New("unauthorized action")
	ErrInvalidStatusTransition = errors.New("invalid status transition")
)

// Services aggregates all service instances
type Services struct {
	Auth        *AuthService
	Tenant      *TenantService
	Branch      *BranchService
	User        *UserService
	Category    *CategoryService
	Product     *ProductService
	Table       *TableService
	Order       *OrderService
	Reservation *ReservationService
}

// NewServices creates a new Services instance
func NewServices(repos *repository.Repositories, jwtSecret string) *Services {
	return &Services{
		Auth:        NewAuthService(repos.Users, repos.Branches, jwtSecret),
		Tenant:      NewTenantService(repos.Tenants),
		Branch:      NewBranchService(repos.Branches),
		User:        NewUserService(repos.Users),
		Category:    NewCategoryService(repos.Categories),
		Product:     NewProductService(repos.Products),
		Table:       NewTableService(repos.Tables),
		Order:       NewOrderService(repos.Orders, repos.Products, repos.Tables),
		Reservation: NewReservationService(repos.Reservations, repos.Tables),
	}
}

// ============ AUTH SERVICE ============

type AuthService struct {
	Users    *repository.UserRepository
	Branches *repository.BranchRepository
	JwtSecret []byte
}

func NewAuthService(users *repository.UserRepository, branches *repository.BranchRepository, jwtSecret string) *AuthService {
	return &AuthService{
		Users:    users,
		Branches: branches,
		JwtSecret: []byte(jwtSecret),
	}
}

// Login authenticates a user and returns the user with JWT token
func (s *AuthService) Login(ctx context.Context, branchID uuid.UUID, email, password string) (*model.User, error) {
	user, err := s.Users.GetByEmail(ctx, branchID, email)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	if !user.IsActive {
		return nil, ErrUserInactive
	}

	return user, nil
}

// ValidateToken validates a JWT token and returns the claims
func (s *AuthService) ValidateToken(ctx context.Context, tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return s.JwtSecret, nil
	})
	if err != nil {
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

// GetUserFromToken extracts user from a valid JWT token
func (s *AuthService) GetUserFromToken(ctx context.Context, tokenString string) (*model.User, error) {
	claims, err := s.ValidateToken(ctx, tokenString)
	if err != nil {
		return nil, err
	}

	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		return nil, ErrInvalidToken
	}

	user, err := s.Users.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if !user.IsActive {
		return nil, ErrUserInactive
	}

	return user, nil
}

func (s *AuthService) generateToken(user *model.User, tenantID uuid.UUID) (string, error) {
	claims := &Claims{
		UserID:    user.ID.String(),
		TenantID:  tenantID.String(),
		BranchID:  user.BranchID.String(),
		Email:     user.Email,
		Role:      string(user.Role),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.JwtSecret)
}

// Claims represents JWT token claims
type Claims struct {
	UserID   string `json:"userId"`
	TenantID string `json:"tenantId"`
	BranchID string `json:"branchId"`
	Email    string `json:"email"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

// ============ TENANT SERVICE ============

type TenantService struct {
	Tenants *repository.TenantRepository
}

func NewTenantService(tenants *repository.TenantRepository) *TenantService {
	return &TenantService{Tenants: tenants}
}

// Create creates a new tenant
func (s *TenantService) Create(ctx context.Context, name, slug string) (*model.Tenant, error) {
	tenant := &model.Tenant{
		ID:       uuid.New(),
		Name:     name,
		Slug:     slug,
		Settings: model.JSON{},
		IsActive: true,
	}
	if err := s.Tenants.Create(ctx, tenant); err != nil {
		return nil, err
	}
	return tenant, nil
}

// GetByID retrieves a tenant by ID
func (s *TenantService) GetByID(ctx context.Context, id uuid.UUID) (*model.Tenant, error) {
	return s.Tenants.GetByID(ctx, id)
}

// GetBySlug retrieves a tenant by slug
func (s *TenantService) GetBySlug(ctx context.Context, slug string) (*model.Tenant, error) {
	return s.Tenants.GetBySlug(ctx, slug)
}

// Update updates a tenant
func (s *TenantService) Update(ctx context.Context, id uuid.UUID, name *string, slug *string, settings *model.JSON, isActive *bool) (*model.Tenant, error) {
	tenant, err := s.Tenants.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if name != nil {
		tenant.Name = *name
	}
	if slug != nil {
		tenant.Slug = *slug
	}
	if settings != nil {
		tenant.Settings = *settings
	}
	if isActive != nil {
		tenant.IsActive = *isActive
	}

	if err := s.Tenants.Update(ctx, tenant); err != nil {
		return nil, err
	}
	return tenant, nil
}

// List returns all tenants
func (s *TenantService) List(ctx context.Context) ([]model.Tenant, error) {
	return s.Tenants.List(ctx)
}

// ============ BRANCH SERVICE ============

type BranchService struct {
	Branches *repository.BranchRepository
}

func NewBranchService(branches *repository.BranchRepository) *BranchService {
	return &BranchService{Branches: branches}
}

// Create creates a new branch
func (s *BranchService) Create(ctx context.Context, tenantID uuid.UUID, name, address, phone string) (*model.Branch, error) {
	branch := &model.Branch{
		ID:       uuid.New(),
		TenantID: tenantID,
		Name:     name,
		Address:  address,
		Phone:    phone,
		IsActive: true,
	}
	if err := s.Branches.Create(ctx, branch); err != nil {
		return nil, err
	}
	return branch, nil
}

// GetByID retrieves a branch by ID
func (s *BranchService) GetByID(ctx context.Context, id uuid.UUID) (*model.Branch, error) {
	return s.Branches.GetByID(ctx, id)
}

// GetByTenant retrieves all branches for a tenant
func (s *BranchService) GetByTenant(ctx context.Context, tenantID uuid.UUID) ([]model.Branch, error) {
	return s.Branches.GetByTenant(ctx, tenantID)
}

// Update updates a branch
func (s *BranchService) Update(ctx context.Context, id uuid.UUID, name, address, phone string, isActive bool) (*model.Branch, error) {
	branch, err := s.Branches.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	branch.Name = name
	branch.Address = address
	branch.Phone = phone
	branch.IsActive = isActive

	if err := s.Branches.Update(ctx, branch); err != nil {
		return nil, err
	}
	return branch, nil
}

// ============ USER SERVICE ============

type UserService struct {
	Users *repository.UserRepository
}

func NewUserService(users *repository.UserRepository) *UserService {
	return &UserService{Users: users}
}

// Create creates a new user
func (s *UserService) Create(ctx context.Context, branchID uuid.UUID, email, password, name string, role model.UserRole) (*model.User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &model.User{
		ID:           uuid.New(),
		BranchID:     branchID,
		Email:        email,
		PasswordHash: string(hash),
		Name:         name,
		Role:         role,
		IsActive:     true,
	}
	if err := s.Users.Create(ctx, user); err != nil {
		return nil, err
	}
	return user, nil
}

// GetByID retrieves a user by ID
func (s *UserService) GetByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	return s.Users.GetByID(ctx, id)
}

// GetByBranch retrieves all active users for a branch
func (s *UserService) GetByBranch(ctx context.Context, branchID uuid.UUID) ([]model.User, error) {
	return s.Users.GetByBranch(ctx, branchID)
}

// Update updates a user
func (s *UserService) Update(ctx context.Context, id uuid.UUID, name string, role *model.UserRole, isActive *bool) (*model.User, error) {
	user, err := s.Users.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if name != "" {
		user.Name = name
	}
	if role != nil {
		user.Role = *role
	}
	if isActive != nil {
		user.IsActive = *isActive
	}

	if err := s.Users.Update(ctx, user); err != nil {
		return nil, err
	}
	return user, nil
}

// Delete soft-deletes a user (sets is_active to false)
func (s *UserService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.Users.Delete(ctx, id)
}

// ChangePassword updates a user's password
func (s *UserService) ChangePassword(ctx context.Context, id uuid.UUID, newPassword string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	return s.Users.UpdatePassword(ctx, id, string(hash))
}

// ============ CATEGORY SERVICE ============

type CategoryService struct {
	Categories *repository.CategoryRepository
}

func NewCategoryService(categories *repository.CategoryRepository) *CategoryService {
	return &CategoryService{Categories: categories}
}

// Create creates a new category
func (s *CategoryService) Create(ctx context.Context, branchID uuid.UUID, name, description string, sortOrder int) (*model.Category, error) {
	// If sortOrder not provided, get the next available
	if sortOrder == 0 {
		maxOrder, err := s.Categories.GetMaxSortOrder(ctx, branchID)
		if err != nil {
			return nil, err
		}
		sortOrder = maxOrder + 1
	}

	category := &model.Category{
		ID:          uuid.New(),
		BranchID:    branchID,
		Name:        name,
		Description: description,
		SortOrder:   sortOrder,
		IsActive:    true,
	}
	if err := s.Categories.Create(ctx, category); err != nil {
		return nil, err
	}
	return category, nil
}

// GetByID retrieves a category by ID
func (s *CategoryService) GetByID(ctx context.Context, id uuid.UUID) (*model.Category, error) {
	return s.Categories.GetByID(ctx, id)
}

// GetByBranch retrieves all active categories for a branch
func (s *CategoryService) GetByBranch(ctx context.Context, branchID uuid.UUID) ([]model.Category, error) {
	return s.Categories.GetByBranch(ctx, branchID)
}

// Update updates a category (handles nil pointers for partial updates)
func (s *CategoryService) Update(ctx context.Context, id uuid.UUID, name *string, description *string, sortOrder *int, isActive *bool) (*model.Category, error) {
	category, err := s.Categories.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if name != nil {
		category.Name = *name
	}
	if description != nil {
		category.Description = *description
	}
	if sortOrder != nil {
		category.SortOrder = *sortOrder
	}
	if isActive != nil {
		category.IsActive = *isActive
	}

	if err := s.Categories.Update(ctx, category); err != nil {
		return nil, err
	}
	return category, nil
}

// Delete soft-deletes a category
func (s *CategoryService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.Categories.Delete(ctx, id)
}

// Reorder updates the sort order of multiple categories
// categoryIDs are ordered as they should appear; sort order is assigned by index
func (s *CategoryService) Reorder(ctx context.Context, branchID uuid.UUID, categoryIDs []uuid.UUID) ([]*model.Category, error) {
	orders := make([]struct {
		ID        uuid.UUID
		SortOrder int
	}, len(categoryIDs))
	for i, id := range categoryIDs {
		orders[i] = struct {
			ID        uuid.UUID
			SortOrder int
		}{ID: id, SortOrder: i + 1}
	}
	if err := s.Categories.Reorder(ctx, branchID, orders); err != nil {
		return nil, err
	}
	// Return categories in new order
	cats, err := s.Categories.GetByBranch(ctx, branchID)
	if err != nil {
		return nil, err
	}
	result := make([]*model.Category, len(cats))
	for i := range cats {
		result[i] = &cats[i]
	}
	return result, nil
}

// ============ PRODUCT SERVICE ============

type ProductService struct {
	Products *repository.ProductRepository
}

func NewProductService(products *repository.ProductRepository) *ProductService {
	return &ProductService{Products: products}
}

// CreateProductInput is the input type for creating a product (matches GraphQL resolver input)
type CreateProductInput struct {
	BranchID        uuid.UUID
	CategoryID      *uuid.UUID
	Name            string
	Description     *string
	Price           float64
	ImageURL        *string
	IsAvailable     *bool
	IsFeatured      *bool
	PreparationTime *int
	Allergens       []string
}

// Create creates a new product
func (s *ProductService) Create(ctx context.Context, input CreateProductInput) (*model.Product, error) {
	isAvailable := true // default
	if input.IsAvailable != nil {
		isAvailable = *input.IsAvailable
	}

	isFeatured := false
	if input.IsFeatured != nil {
		isFeatured = *input.IsFeatured
	}

	preparationTime := 0
	if input.PreparationTime != nil {
		preparationTime = *input.PreparationTime
	}

	var description, imageURL string
	if input.Description != nil {
		description = *input.Description
	}
	if input.ImageURL != nil {
		imageURL = *input.ImageURL
	}

	product := &model.Product{
		ID:              uuid.New(),
		BranchID:        input.BranchID,
		CategoryID:      input.CategoryID,
		Name:            input.Name,
		Description:     description,
		Price:           input.Price,
		ImageURL:        imageURL,
		IsAvailable:     isAvailable,
		IsFeatured:      isFeatured,
		PreparationTime: preparationTime,
		Allergens:       input.Allergens,
		Rating:          0,
		ReviewCount:     0,
	}
	if err := s.Products.Create(ctx, product); err != nil {
		return nil, err
	}
	return product, nil
}

// GetByID retrieves a product by ID
func (s *ProductService) GetByID(ctx context.Context, id uuid.UUID) (*model.Product, error) {
	return s.Products.GetByID(ctx, id)
}

// GetByBranch retrieves all available products for a branch
func (s *ProductService) GetByBranch(ctx context.Context, branchID uuid.UUID) ([]model.Product, error) {
	return s.Products.GetAvailable(ctx, branchID)
}

// GetByBranchAll retrieves all products for a branch (including unavailable)
func (s *ProductService) GetByBranchAll(ctx context.Context, branchID uuid.UUID) ([]model.Product, error) {
	return s.Products.GetByBranch(ctx, branchID)
}

// GetByCategory retrieves all available products for a category
func (s *ProductService) GetByCategory(ctx context.Context, branchID, categoryID uuid.UUID) ([]model.Product, error) {
	return s.Products.GetByCategory(ctx, branchID, categoryID)
}

// GetFeatured retrieves featured products for a branch
func (s *ProductService) GetFeatured(ctx context.Context, branchID uuid.UUID) ([]model.Product, error) {
	return s.Products.GetFeatured(ctx, branchID)
}

// Update updates a product
func (s *ProductService) Update(ctx context.Context, id uuid.UUID, input CreateProductInput) (*model.Product, error) {
	product, err := s.Products.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	product.CategoryID = input.CategoryID
	product.Name = input.Name
	if input.Description != nil {
		product.Description = *input.Description
	}
	product.Price = input.Price
	if input.ImageURL != nil {
		product.ImageURL = *input.ImageURL
	}
	if input.IsAvailable != nil {
		product.IsAvailable = *input.IsAvailable
	}
	if input.IsFeatured != nil {
		product.IsFeatured = *input.IsFeatured
	}
	if input.PreparationTime != nil {
		product.PreparationTime = *input.PreparationTime
	}
	product.Allergens = input.Allergens

	if err := s.Products.Update(ctx, product); err != nil {
		return nil, err
	}
	return product, nil
}

// Delete soft-deletes a product (sets is_available to false)
func (s *ProductService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.Products.Delete(ctx, id)
}

// ToggleAvailability toggles product availability
func (s *ProductService) ToggleAvailability(ctx context.Context, id uuid.UUID) (*model.Product, error) {
	product, err := s.Products.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	newStatus := !product.IsAvailable
	if err := s.Products.ToggleAvailability(ctx, id, newStatus); err != nil {
		return nil, err
	}
	product.IsAvailable = newStatus
	return product, nil
}

// ============ TABLE SERVICE ============

type TableService struct {
	Tables *repository.TableRepository
}

func NewTableService(tables *repository.TableRepository) *TableService {
	return &TableService{Tables: tables}
}

// Create creates a new table
func (s *TableService) Create(ctx context.Context, branchID uuid.UUID, number string, capacity int) (*model.RestaurantTable, error) {
	table := &model.RestaurantTable{
		ID:       uuid.New(),
		BranchID: branchID,
		Number:   number,
		Capacity: capacity,
		Status:   model.TableAvailable,
	}
	if err := s.Tables.Create(ctx, table); err != nil {
		return nil, err
	}
	return table, nil
}

// GetByID retrieves a table by ID
func (s *TableService) GetByID(ctx context.Context, id uuid.UUID) (*model.RestaurantTable, error) {
	return s.Tables.GetByID(ctx, id)
}

// GetByBranch retrieves all tables for a branch
func (s *TableService) GetByBranch(ctx context.Context, branchID uuid.UUID) ([]model.RestaurantTable, error) {
	return s.Tables.GetByBranch(ctx, branchID)
}

// Update updates a table
func (s *TableService) Update(ctx context.Context, id uuid.UUID, number string, capacity int) (*model.RestaurantTable, error) {
	table, err := s.Tables.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	table.Number = number
	table.Capacity = capacity

	if err := s.Tables.Update(ctx, table); err != nil {
		return nil, err
	}
	return table, nil
}

// Delete removes a table
func (s *TableService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.Tables.Delete(ctx, id)
}

// UpdateStatus updates a table's status
func (s *TableService) UpdateStatus(ctx context.Context, id uuid.UUID, status model.TableStatus) error {
	return s.Tables.UpdateStatus(ctx, id, status)
}

// GetAvailableTables retrieves all available tables for a branch
func (s *TableService) GetAvailableTables(ctx context.Context, branchID uuid.UUID) ([]model.RestaurantTable, error) {
	return s.Tables.GetAvailable(ctx, branchID)
}

// ============ ORDER SERVICE ============

type OrderService struct {
	Orders   *repository.OrderRepository
	Products *repository.ProductRepository
	Tables   *repository.TableRepository
	TaxRate  float64
}

func NewOrderService(orders *repository.OrderRepository, products *repository.ProductRepository, tables *repository.TableRepository) *OrderService {
	return &OrderService{
		Orders:   orders,
		Products: products,
		Tables:   tables,
		TaxRate:  DefaultTaxRate,
	}
}

// SetTaxRate sets a custom tax rate
func (s *OrderService) SetTaxRate(rate float64) {
	s.TaxRate = rate
}

// CreateOrderInput represents input for creating an order
type CreateOrderInput struct {
	BranchID      uuid.UUID
	TableID       *uuid.UUID
	UserID        uuid.UUID
	CustomerName  string
	CustomerPhone string
	Notes         string
	Discount      float64
	Items         []OrderItemInput
}

// OrderItemInput represents an order item in the input
type OrderItemInput struct {
	ProductID uuid.UUID
	Quantity  int
	Notes     string
}

// CreateOrderInputFromGraphQL converts GraphQL input to service input
func CreateOrderInputFromGraphQL(branchID uuid.UUID, tableID *uuid.UUID, userID uuid.UUID,
	customerName, customerPhone, notes string, items []OrderItemInput) *CreateOrderInput {
	return &CreateOrderInput{
		BranchID:      branchID,
		TableID:       tableID,
		UserID:        userID,
		CustomerName:  customerName,
		CustomerPhone: customerPhone,
		Notes:         notes,
		Discount:      0,
		Items:         items,
	}
}

// Create creates a new order with items (accepts GraphQL input struct directly)
func (s *OrderService) Create(ctx context.Context, input struct {
	BranchID      uuid.UUID
	TableID       *uuid.UUID
	UserID        uuid.UUID
	CustomerName  *string
	CustomerPhone *string
	Notes         *string
	Items         []struct {
		ProductID uuid.UUID
		Quantity  int
		Notes     *string
	}
}) (*model.Order, error) {
	// Convert input to internal format
	var customerName, customerPhone, notes string
	if input.CustomerName != nil {
		customerName = *input.CustomerName
	}
	if input.CustomerPhone != nil {
		customerPhone = *input.CustomerPhone
	}
	if input.Notes != nil {
		notes = *input.Notes
	}

	orderItems := make([]OrderItemInput, len(input.Items))
	for i, item := range input.Items {
		var itemNotes string
		if item.Notes != nil {
			itemNotes = *item.Notes
		}
		orderItems[i] = OrderItemInput{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
			Notes:     itemNotes,
		}
	}

	internalInput := &CreateOrderInput{
		BranchID:      input.BranchID,
		TableID:       input.TableID,
		UserID:        input.UserID,
		CustomerName:  customerName,
		CustomerPhone: customerPhone,
		Notes:         notes,
		Discount:      0,
		Items:         orderItems,
	}

	// If table is specified, check availability and update status
	if internalInput.TableID != nil {
		table, err := s.Tables.GetByID(ctx, *internalInput.TableID)
		if err != nil {
			return nil, ErrTableNotFound
		}
		if table.Status != model.TableAvailable {
			return nil, ErrTableNotAvailable
		}

		// Mark table as occupied
		if err := s.Tables.UpdateStatus(ctx, *internalInput.TableID, model.TableOccupied); err != nil {
			return nil, err
		}
	}

	// Calculate subtotal and create order items
	var subtotal float64
	var orderItemsModel []model.OrderItem

	for _, item := range internalInput.Items {
		product, err := s.Products.GetByID(ctx, item.ProductID)
		if err != nil {
			return nil, ErrProductNotFound
		}
		if !product.IsAvailable {
			return nil, errors.New("product is not available: " + product.Name)
		}

		itemTotal := product.Price * float64(item.Quantity)
		subtotal += itemTotal

		orderItemsModel = append(orderItemsModel, model.OrderItem{
			ID:         uuid.New(),
			ProductID:  item.ProductID,
			Quantity:   item.Quantity,
			UnitPrice:  product.Price,
			TotalPrice: itemTotal,
			Notes:      item.Notes,
		})
	}

	// Calculate tax and total
	tax := subtotal * s.TaxRate
	total := subtotal + tax - internalInput.Discount

	order := &model.Order{
		ID:            uuid.New(),
		BranchID:      internalInput.BranchID,
		TableID:       internalInput.TableID,
		UserID:        internalInput.UserID,
		CustomerName:  internalInput.CustomerName,
		CustomerPhone: internalInput.CustomerPhone,
		Status:        model.OrderPending,
		Subtotal:      subtotal,
		Tax:           tax,
		Discount:      internalInput.Discount,
		Total:         total,
		Notes:         internalInput.Notes,
	}

	if err := s.Orders.Create(ctx, order); err != nil {
		// Rollback table status if order creation fails
		if internalInput.TableID != nil {
			s.Tables.UpdateStatus(ctx, *internalInput.TableID, model.TableAvailable)
		}
		return nil, err
	}

	// Save order items
	for i := range orderItemsModel {
		orderItemsModel[i].OrderID = order.ID
		if err := s.Orders.AddItem(ctx, &orderItemsModel[i]); err != nil {
			return nil, err
		}
	}

	return order, nil
}

// CreateOrder creates a new order with items (legacy method)
func (s *OrderService) CreateOrder(ctx context.Context, input *CreateOrderInput) (*model.Order, error) {
	// If table is specified, check availability and update status
	if input.TableID != nil {
		table, err := s.Tables.GetByID(ctx, *input.TableID)
		if err != nil {
			return nil, ErrTableNotFound
		}
		if table.Status != model.TableAvailable {
			return nil, ErrTableNotAvailable
		}

		// Mark table as occupied
		if err := s.Tables.UpdateStatus(ctx, *input.TableID, model.TableOccupied); err != nil {
			return nil, err
		}
	}

	// Calculate subtotal and create order items
	var subtotal float64
	var orderItems []model.OrderItem

	for _, item := range input.Items {
		product, err := s.Products.GetByID(ctx, item.ProductID)
		if err != nil {
			return nil, ErrProductNotFound
		}
		if !product.IsAvailable {
			return nil, errors.New("product is not available: " + product.Name)
		}

		itemTotal := product.Price * float64(item.Quantity)
		subtotal += itemTotal

		orderItems = append(orderItems, model.OrderItem{
			ID:         uuid.New(),
			ProductID:  item.ProductID,
			Quantity:   item.Quantity,
			UnitPrice:  product.Price,
			TotalPrice: itemTotal,
			Notes:      item.Notes,
		})
	}

	// Calculate tax and total
	tax := subtotal * s.TaxRate
	total := subtotal + tax - input.Discount

	order := &model.Order{
		ID:            uuid.New(),
		BranchID:      input.BranchID,
		TableID:       input.TableID,
		UserID:        input.UserID,
		CustomerName:  input.CustomerName,
		CustomerPhone: input.CustomerPhone,
		Status:        model.OrderPending,
		Subtotal:      subtotal,
		Tax:           tax,
		Discount:      input.Discount,
		Total:         total,
		Notes:         input.Notes,
	}

	if err := s.Orders.Create(ctx, order); err != nil {
		// Rollback table status if order creation fails
		if input.TableID != nil {
			s.Tables.UpdateStatus(ctx, *input.TableID, model.TableAvailable)
		}
		return nil, err
	}

	// Save order items
	for i := range orderItems {
		orderItems[i].OrderID = order.ID
		if err := s.Orders.AddItem(ctx, &orderItems[i]); err != nil {
			return nil, err
		}
	}

	return order, nil
}

// GetByID retrieves an order by ID
func (s *OrderService) GetByID(ctx context.Context, id uuid.UUID) (*model.Order, error) {
	return s.Orders.GetByID(ctx, id)
}

// GetByBranch retrieves all orders for a branch
func (s *OrderService) GetByBranch(ctx context.Context, branchID uuid.UUID) ([]model.Order, error) {
	return s.Orders.GetByBranch(ctx, branchID)
}

// GetItems retrieves order items for an order
func (s *OrderService) GetItems(ctx context.Context, orderID uuid.UUID) ([]model.OrderItem, error) {
	return s.Orders.GetItems(ctx, orderID)
}

// GetItem retrieves a single order item by ID
func (s *OrderService) GetItem(ctx context.Context, itemID uuid.UUID) (*model.OrderItem, error) {
	return s.Orders.GetItem(ctx, itemID)
}

// UpdateStatus updates order status with business rule validation
func (s *OrderService) UpdateStatus(ctx context.Context, id uuid.UUID, newStatus model.OrderStatus) error {
	order, err := s.Orders.GetByID(ctx, id)
	if err != nil {
		return ErrOrderNotFound
	}

	// Validate status transition
	if !s.isValidStatusTransition(order.Status, newStatus) {
		return ErrInvalidStatusTransition
	}

	// If order is being delivered, mark table as available
	if newStatus == model.OrderDelivered && order.TableID != nil {
		if err := s.Tables.UpdateStatus(ctx, *order.TableID, model.TableAvailable); err != nil {
			return err
		}
	}

	// If order is being cancelled and had a table, free the table
	if newStatus == model.OrderCancelled && order.TableID != nil {
		if err := s.Tables.UpdateStatus(ctx, *order.TableID, model.TableAvailable); err != nil {
			return err
		}
	}

	return s.Orders.UpdateStatus(ctx, id, newStatus)
}

// isValidStatusTransition checks if a status transition is valid
// Valid transitions: PENDING → ACCEPTED → PREPARING → READY → DELIVERED → PAID
// Also: PENDING/ACCEPTED → CANCELLED
func (s *OrderService) isValidStatusTransition(current, new model.OrderStatus) bool {
	validTransitions := map[model.OrderStatus][]model.OrderStatus{
		model.OrderPending:   {model.OrderAccepted, model.OrderCancelled},
		model.OrderAccepted:  {model.OrderPreparing, model.OrderCancelled},
		model.OrderPreparing: {model.OrderReady},
		model.OrderReady:      {model.OrderDelivered},
		model.OrderDelivered: {model.OrderPaid},
		model.OrderPaid:      {},
		model.OrderCancelled: {},
	}

	allowed, ok := validTransitions[current]
	if !ok {
		return false
	}

	for _, status := range allowed {
		if status == new {
			return true
		}
	}
	return false
}

// Cancel cancels an order (only PENDING or ACCEPTED)
func (s *OrderService) Cancel(ctx context.Context, id uuid.UUID) error {
	order, err := s.Orders.GetByID(ctx, id)
	if err != nil {
		return ErrOrderNotFound
	}

	if order.Status != model.OrderPending && order.Status != model.OrderAccepted {
		return ErrOrderNotCancellable
	}

	// Free the table if order had one
	if order.TableID != nil {
		if err := s.Tables.UpdateStatus(ctx, *order.TableID, model.TableAvailable); err != nil {
			return err
		}
	}

	return s.Orders.UpdateStatus(ctx, id, model.OrderCancelled)
}

// AddItemInput represents input for adding an item to an order
type AddItemInput struct {
	ProductID uuid.UUID
	Quantity  int
	Notes     string
}

// AddItem adds an item to an existing order
func (s *OrderService) AddItem(ctx context.Context, orderID uuid.UUID, input AddItemInput) (*model.OrderItem, error) {
	order, err := s.Orders.GetByID(ctx, orderID)
	if err != nil {
		return nil, ErrOrderNotFound
	}

	// Can only add items to PENDING or ACCEPTED orders
	if order.Status != model.OrderPending && order.Status != model.OrderAccepted {
		return nil, errors.New("can only add items to pending or accepted orders")
	}

	product, err := s.Products.GetByID(ctx, input.ProductID)
	if err != nil {
		return nil, ErrProductNotFound
	}
	if !product.IsAvailable {
		return nil, errors.New("product is not available")
	}

	itemTotal := product.Price * float64(input.Quantity)

	orderItem := &model.OrderItem{
		ID:         uuid.New(),
		OrderID:    orderID,
		ProductID:  input.ProductID,
		Quantity:   input.Quantity,
		UnitPrice:  product.Price,
		TotalPrice: itemTotal,
		Notes:      input.Notes,
	}

	if err := s.Orders.AddItem(ctx, orderItem); err != nil {
		return nil, err
	}

	// Recalculate totals
	if _, err := s.CalculateTotals(ctx, orderID); err != nil {
		return nil, err
	}

	return orderItem, nil
}

// RemoveItem removes an item from an order
func (s *OrderService) RemoveItem(ctx context.Context, orderID, itemID uuid.UUID) (*model.Order, error) {
	order, err := s.Orders.GetByID(ctx, orderID)
	if err != nil {
		return nil, ErrOrderNotFound
	}

	// Can only remove items from PENDING or ACCEPTED orders
	if order.Status != model.OrderPending && order.Status != model.OrderAccepted {
		return nil, errors.New("can only remove items from pending or accepted orders")
	}

	if err := s.Orders.RemoveItem(ctx, itemID); err != nil {
		return nil, err
	}

	// Recalculate totals
	return s.CalculateTotals(ctx, orderID)
}

// CalculateTotals recalculates order subtotal, tax, and total
func (s *OrderService) CalculateTotals(ctx context.Context, orderID uuid.UUID) (*model.Order, error) {
	order, err := s.Orders.GetByID(ctx, orderID)
	if err != nil {
		return nil, ErrOrderNotFound
	}

	items, err := s.Orders.GetItems(ctx, orderID)
	if err != nil {
		return nil, err
	}

	var subtotal float64
	for _, item := range items {
		subtotal += item.TotalPrice
	}

	order.Subtotal = subtotal
	order.Tax = subtotal * s.TaxRate
	order.Total = subtotal + order.Tax - order.Discount

	if err := s.Orders.Update(ctx, order); err != nil {
		return nil, err
	}

	return order, nil
}

// ApplyDiscount applies a discount to an order
func (s *OrderService) ApplyDiscount(ctx context.Context, orderID uuid.UUID, discount float64) (*model.Order, error) {
	order, err := s.Orders.GetByID(ctx, orderID)
	if err != nil {
		return nil, ErrOrderNotFound
	}

	if discount < 0 || discount > order.Subtotal {
		return nil, errors.New("invalid discount amount")
	}

	order.Discount = discount
	order.Total = order.Subtotal + order.Tax - discount

	if err := s.Orders.Update(ctx, order); err != nil {
		return nil, err
	}

	return order, nil
}

// ============ RESERVATION SERVICE ============

type ReservationService struct {
	Reservations *repository.ReservationRepository
	Tables       *repository.TableRepository
}

func NewReservationService(reservations *repository.ReservationRepository, tables *repository.TableRepository) *ReservationService {
	return &ReservationService{
		Reservations: reservations,
		Tables:       tables,
	}
}

// CreateReservationInput represents input for creating a reservation (GraphQL schema)
type CreateReservationInput struct {
	BranchID         uuid.UUID
	TableID          *uuid.UUID
	CustomerName     string
	CustomerPhone    *string
	CustomerEmail    *string
	GuestCount       int
	ReservationDate  string
	ReservationTime  string
	Notes           *string
}

// Create creates a new reservation (accepts GraphQL input struct directly)
func (s *ReservationService) Create(ctx context.Context, input CreateReservationInput) (*model.Reservation, error) {
	// Parse reservation date
	reservationDate, err := time.Parse("2006-01-02", input.ReservationDate)
	if err != nil {
		return nil, fmt.Errorf("invalid reservation date format: %w", err)
	}

	var customerPhone, customerEmail, notes string
	if input.CustomerPhone != nil {
		customerPhone = *input.CustomerPhone
	}
	if input.CustomerEmail != nil {
		customerEmail = *input.CustomerEmail
	}
	if input.Notes != nil {
		notes = *input.Notes
	}

	// If table is specified, mark it as reserved
	if input.TableID != nil {
		table, err := s.Tables.GetByID(ctx, *input.TableID)
		if err != nil {
			return nil, ErrTableNotFound
		}
		if table.Status != model.TableAvailable {
			return nil, ErrTableNotAvailable
		}

		if err := s.Tables.UpdateStatus(ctx, *input.TableID, model.TableReserved); err != nil {
			return nil, err
		}
	}

	reservation := &model.Reservation{
		ID:              uuid.New(),
		BranchID:        input.BranchID,
		TableID:         input.TableID,
		CustomerName:    input.CustomerName,
		CustomerPhone:   customerPhone,
		CustomerEmail:   customerEmail,
		GuestCount:      input.GuestCount,
		ReservationDate: reservationDate,
		ReservationTime: input.ReservationTime,
		Status:          model.ReservationPending,
		Notes:           notes,
	}

	if err := s.Reservations.Create(ctx, reservation); err != nil {
		// Rollback table status if reservation creation fails
		if input.TableID != nil {
			s.Tables.UpdateStatus(ctx, *input.TableID, model.TableAvailable)
		}
		return nil, err
	}

	return reservation, nil
}

// GetByID retrieves a reservation by ID
func (s *ReservationService) GetByID(ctx context.Context, id uuid.UUID) (*model.Reservation, error) {
	res, err := s.Reservations.GetByID(ctx, id)
	if err != nil {
		return nil, ErrReservationNotFound
	}
	return res, nil
}

// GetByBranch retrieves all reservations for a branch
func (s *ReservationService) GetByBranch(ctx context.Context, branchID uuid.UUID) ([]model.Reservation, error) {
	return s.Reservations.GetByBranch(ctx, branchID)
}

// GetByDate retrieves all reservations for a branch on a specific date
func (s *ReservationService) GetByDate(ctx context.Context, branchID uuid.UUID, date time.Time) ([]model.Reservation, error) {
	return s.Reservations.GetByDate(ctx, branchID, date.Format("2006-01-02"))
}

// Update updates a reservation (accepts GraphQL input struct directly)
func (s *ReservationService) Update(ctx context.Context, id uuid.UUID, input CreateReservationInput) (*model.Reservation, error) {
	reservation, err := s.Reservations.GetByID(ctx, id)
	if err != nil {
		return nil, ErrReservationNotFound
	}

	// Parse reservation date
	reservationDate, err := time.Parse("2006-01-02", input.ReservationDate)
	if err != nil {
		return nil, fmt.Errorf("invalid reservation date format: %w", err)
	}

	var customerPhone, customerEmail, notes string
	if input.CustomerPhone != nil {
		customerPhone = *input.CustomerPhone
	}
	if input.CustomerEmail != nil {
		customerEmail = *input.CustomerEmail
	}
	if input.Notes != nil {
		notes = *input.Notes
	}

	// If table is being changed, update old and new table statuses
	if input.TableID != nil && (reservation.TableID == nil || *reservation.TableID != *input.TableID) {
		// Free old table
		if reservation.TableID != nil {
			if err := s.Tables.UpdateStatus(ctx, *reservation.TableID, model.TableAvailable); err != nil {
				return nil, err
			}
		}
		// Reserve new table
		newTable, err := s.Tables.GetByID(ctx, *input.TableID)
		if err != nil {
			return nil, ErrTableNotFound
		}
		if newTable.Status != model.TableAvailable {
			return nil, ErrTableNotAvailable
		}
		if err := s.Tables.UpdateStatus(ctx, *input.TableID, model.TableReserved); err != nil {
			return nil, err
		}
	}

	reservation.TableID = input.TableID
	reservation.CustomerName = input.CustomerName
	reservation.CustomerPhone = customerPhone
	reservation.CustomerEmail = customerEmail
	reservation.GuestCount = input.GuestCount
	reservation.ReservationDate = reservationDate
	reservation.ReservationTime = input.ReservationTime
	reservation.Notes = notes

	if err := s.Reservations.Update(ctx, reservation); err != nil {
		return nil, err
	}

	return reservation, nil
}

// Cancel cancels a reservation
func (s *ReservationService) Cancel(ctx context.Context, id uuid.UUID) error {
	reservation, err := s.Reservations.GetByID(ctx, id)
	if err != nil {
		return ErrReservationNotFound
	}

	if reservation.Status == model.ReservationCancelled || reservation.Status == model.ReservationCompleted {
		return errors.New("reservation cannot be cancelled in current state")
	}

	// Free the table if reservation had one
	if reservation.TableID != nil {
		if err := s.Tables.UpdateStatus(ctx, *reservation.TableID, model.TableAvailable); err != nil {
			return err
		}
	}

	return s.Reservations.UpdateStatus(ctx, id, model.ReservationCancelled)
}

// Confirm confirms a reservation
func (s *ReservationService) Confirm(ctx context.Context, id uuid.UUID) error {
	reservation, err := s.Reservations.GetByID(ctx, id)
	if err != nil {
		return ErrReservationNotFound
	}

	if reservation.Status != model.ReservationPending {
		return errors.New("only pending reservations can be confirmed")
	}

	return s.Reservations.UpdateStatus(ctx, id, model.ReservationConfirmed)
}

// Complete marks a reservation as completed
func (s *ReservationService) Complete(ctx context.Context, id uuid.UUID) error {
	reservation, err := s.Reservations.GetByID(ctx, id)
	if err != nil {
		return ErrReservationNotFound
	}

	if reservation.Status != model.ReservationConfirmed {
		return errors.New("only confirmed reservations can be completed")
	}

	// Free the table if reservation had one
	if reservation.TableID != nil {
		if err := s.Tables.UpdateStatus(ctx, *reservation.TableID, model.TableAvailable); err != nil {
			return err
		}
	}

	return s.Reservations.UpdateStatus(ctx, id, model.ReservationCompleted)
}

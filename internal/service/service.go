package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/greenpos/backend/internal/middleware"
	"github.com/greenpos/backend/internal/model"
	"github.com/greenpos/backend/internal/repository"
)

// DefaultTaxRate is the default tax rate (16%)
const DefaultTaxRate = 0.16

var (
	ErrInvalidCredentials       = errors.New("invalid credentials")
	ErrUserInactive             = errors.New("user is inactive")
	ErrInvalidToken             = errors.New("invalid token")
	ErrOrderNotCancellable      = errors.New("order cannot be cancelled in current state")
	ErrTableNotAvailable        = errors.New("table is not available")
	ErrProductNotFound           = errors.New("product not found")
	ErrCategoryNotFound         = errors.New("category not found")
	ErrTableNotFound            = errors.New("table not found")
	ErrOrderNotFound            = errors.New("order not found")
	ErrReservationNotFound      = errors.New("reservation not found")
	ErrUnauthorized             = errors.New("unauthorized action")
	ErrInvalidStatusTransition  = errors.New("invalid status transition")
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
	Invoice     *InvoiceService
	TenantFiscal *TenantFiscalService
	Analytics   *AnalyticsService
	Inventory   *InventoryService
	FCM         *FCMService
	Stripe      *StripeService
	PayPal      *PayPalService
	QRMenu      *QRMenuService
	Logger      *slog.Logger
}

// ServiceConfig holds configuration for services
type ServiceConfig struct {
	JWTSecret            string
	JWTExpiryHours       int
	RefreshTokenExpiryDays int
	BCryptCost           int
	TaxRate              float64
}

// NewServices creates a new Services instance with dependency injection
func NewServices(repos *repository.Repositories, cfg *ServiceConfig, log *slog.Logger) *Services {
	if cfg == nil {
		cfg = &ServiceConfig{
			JWTSecret:            "default-secret-change-in-production",
			JWTExpiryHours:       24,
			RefreshTokenExpiryDays: 7,
			BCryptCost:           bcrypt.DefaultCost,
			TaxRate:              DefaultTaxRate,
		}
	}

	if log == nil {
		log = slog.Default()
	}

	// Create auth service with JWT config
	authConfig := &middleware.JWTConfig{
		Secret:               []byte(cfg.JWTSecret),
		ExpiryHours:         cfg.JWTExpiryHours,
		RefreshTokenExpiryDays: cfg.RefreshTokenExpiryDays,
	}

	return &Services{
		Auth:        NewAuthService(repos.Users, repos.Branches, authConfig, log),
		Tenant:      NewTenantService(repos.Tenants, log),
		Branch:      NewBranchService(repos.Branches, log),
		User:        NewUserService(repos.Users, cfg.BCryptCost, log),
		Category:    NewCategoryService(repos.Categories, log),
		Product:     NewProductService(repos.Products, log),
		Table:      NewTableService(repos.Tables, log),
		Order:       NewOrderService(repos.Orders, repos.Products, repos.Tables, cfg.TaxRate, log),
		Reservation: NewReservationService(repos.Reservations, repos.Tables, log),
		Invoice: NewInvoiceService(
			repos.Invoices.(*repository.InvoiceRepository),
			repos.InvoiceItems.(*repository.InvoiceItemRepository),
			repos.TenantFiscal.(*repository.TenantFiscalRepository),
			NewFacturapiService(os.Getenv("FACTURAPI_API_KEY")),
		),
		TenantFiscal: NewTenantFiscalService(repos.TenantFiscal.(*repository.TenantFiscalRepository)),
		Analytics:   NewAnalyticsService(repos.Orders, repos.Products, repos.Tables, log),
		Inventory:   NewInventoryService(repos.InventoryItems, repos.InventoryMovements, log),
		FCM:         NewFCMService(repos.FCMTokens, repos.Users, log),
		Stripe:      NewStripeService(repos.Payments, repos.Orders, log),
		PayPal:      NewPayPalService(repos.Payments, repos.Orders, log),
		QRMenu: NewQRMenuService(
			repos.QRCodes,
			repos.Categories,
			repos.Products,
			repos.Tables,
			repos.Orders,
			cfg.TaxRate,
			os.Getenv("PUBLIC_MENU_BASE_URL"),
			log,
		),
		Logger:      log,
	}
}

// ============ AUTH SERVICE ============

// AuthService handles authentication operations
type AuthService struct {
	users    repository.UserRepositoryInterface
	branches repository.BranchRepositoryInterface
	config   *middleware.JWTConfig
	log      *slog.Logger
}

// NewAuthService creates a new auth service
func NewAuthService(users repository.UserRepositoryInterface, branches repository.BranchRepositoryInterface, config *middleware.JWTConfig, log *slog.Logger) *AuthService {
	return &AuthService{
		users:    users,
		branches: branches,
		config:   config,
		log:      log,
	}
}

// Login authenticates a user and returns tokens
func (s *AuthService) Login(ctx context.Context, branchID uuid.UUID, email, password string) (*model.User, string, string, error) {
	user, err := s.users.GetByEmail(ctx, branchID, email)
	if err != nil {
		s.log.Warn("login failed - user not found", slog.String("email", email))
		return nil, "", "", ErrInvalidCredentials
	}

	if !middleware.VerifyPassword(password, user.PasswordHash) {
		s.log.Warn("login failed - invalid password", slog.String("email", email))
		return nil, "", "", ErrInvalidCredentials
	}

	if !user.IsActive {
		s.log.Warn("login failed - user inactive", slog.String("email", email))
		return nil, "", "", ErrUserInactive
	}

	// Generate token pair
	tokenUser := &middleware.TokenUser{
		ID:       user.ID,
		TenantID: user.BranchID, // Would need proper tenant lookup
		BranchID: user.BranchID,
		Email:    user.Email,
		Role:     string(user.Role),
	}

	authSvc := middleware.NewAuthService(s.config, s.log)
	accessToken, refreshToken, err := authSvc.GenerateTokenPair(tokenUser)
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to generate tokens: %w", err)
	}

	s.log.Info("user logged in", slog.String("user_id", user.ID.String()), slog.String("email", email))

	return user, accessToken, refreshToken, nil
}

// ValidateToken validates a JWT token
func (s *AuthService) ValidateToken(ctx context.Context, tokenString string) (*middleware.JWTClaims, error) {
	authSvc := middleware.NewAuthService(s.config, s.log)
	return authSvc.ValidateAccessToken(ctx, tokenString)
}

// RefreshToken refreshes an access token using a refresh token
func (s *AuthService) RefreshToken(ctx context.Context, refreshToken string) (string, error) {
	// Parse the refresh token to get user ID
	authSvc := middleware.NewAuthService(s.config, s.log)
	claims, err := authSvc.ValidateRefreshToken(ctx, refreshToken)
	if err != nil {
		return "", ErrInvalidToken
	}

	// Get user
	user, err := s.users.GetByID(ctx, claims.UserID)
	if err != nil {
		return "", ErrInvalidToken
	}

	if !user.IsActive {
		return "", ErrUserInactive
	}

	// Generate new access token
	tokenUser := &middleware.TokenUser{
		ID:       user.ID,
		TenantID: claims.TenantID,
		BranchID: claims.BranchID,
		Email:    user.Email,
		Role:     string(user.Role),
	}

	return authSvc.RefreshAccessToken(ctx, refreshToken, tokenUser)
}

// GetUserFromToken extracts user from a valid JWT token
func (s *AuthService) GetUserFromToken(ctx context.Context, tokenString string) (*model.User, error) {
	authSvc := middleware.NewAuthService(s.config, s.log)
	claims, err := authSvc.ValidateAccessToken(ctx, tokenString)
	if err != nil {
		return nil, err
	}

	user, err := s.users.GetByID(ctx, claims.UserID)
	if err != nil {
		return nil, err
	}

	if !user.IsActive {
		return nil, ErrUserInactive
	}

	return user, nil
}

// ============ TENANT SERVICE ============

type TenantService struct {
	tenants repository.TenantRepositoryInterface
	log     *slog.Logger
}

func NewTenantService(tenants repository.TenantRepositoryInterface, log *slog.Logger) *TenantService {
	return &TenantService{
		tenants: tenants,
		log:     log,
	}
}

func (s *TenantService) Create(ctx context.Context, name, slug string) (*model.Tenant, error) {
	tenant := &model.Tenant{
		ID:       uuid.New(),
		Name:     name,
		Slug:     slug,
		Settings: model.JSON{},
		IsActive: true,
	}
	if err := s.tenants.Create(ctx, tenant); err != nil {
		return nil, fmt.Errorf("failed to create tenant: %w", err)
	}
	s.log.Info("tenant created", slog.String("id", tenant.ID.String()), slog.String("name", name))
	return tenant, nil
}

func (s *TenantService) GetByID(ctx context.Context, id uuid.UUID) (*model.Tenant, error) {
	tenant, err := s.tenants.GetByID(ctx, id)
	if err != nil {
		return nil, model.NewNotFoundError("tenant", id.String())
	}
	return tenant, nil
}

func (s *TenantService) GetBySlug(ctx context.Context, slug string) (*model.Tenant, error) {
	tenant, err := s.tenants.GetBySlug(ctx, slug)
	if err != nil {
		return nil, model.NewNotFoundError("tenant", slug)
	}
	return tenant, nil
}

func (s *TenantService) Update(ctx context.Context, id uuid.UUID, name *string, slug *string, settings *model.JSON, isActive *bool) (*model.Tenant, error) {
	tenant, err := s.tenants.GetByID(ctx, id)
	if err != nil {
		return nil, model.NewNotFoundError("tenant", id.String())
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

	if err := s.tenants.Update(ctx, tenant); err != nil {
		return nil, fmt.Errorf("failed to update tenant: %w", err)
	}
	return tenant, nil
}

func (s *TenantService) List(ctx context.Context) ([]model.Tenant, error) {
	return s.tenants.List(ctx)
}

// ============ BRANCH SERVICE ============

type BranchService struct {
	branches repository.BranchRepositoryInterface
	log      *slog.Logger
}

func NewBranchService(branches repository.BranchRepositoryInterface, log *slog.Logger) *BranchService {
	return &BranchService{branches: branches, log: log}
}

func (s *BranchService) Create(ctx context.Context, tenantID uuid.UUID, name, address, phone string) (*model.Branch, error) {
	branch := &model.Branch{
		ID:       uuid.New(),
		TenantID: tenantID,
		Name:     name,
		Address:  address,
		Phone:    phone,
		IsActive: true,
	}
	if err := s.branches.Create(ctx, branch); err != nil {
		return nil, fmt.Errorf("failed to create branch: %w", err)
	}
	s.log.Info("branch created", slog.String("id", branch.ID.String()), slog.String("name", name))
	return branch, nil
}

func (s *BranchService) GetByID(ctx context.Context, id uuid.UUID) (*model.Branch, error) {
	branch, err := s.branches.GetByID(ctx, id)
	if err != nil {
		return nil, model.NewNotFoundError("branch", id.String())
	}
	return branch, nil
}

func (s *BranchService) GetByTenant(ctx context.Context, tenantID uuid.UUID) ([]model.Branch, error) {
	return s.branches.GetByTenant(ctx, tenantID)
}

func (s *BranchService) Update(ctx context.Context, id uuid.UUID, name, address, phone string, isActive bool) (*model.Branch, error) {
	branch, err := s.branches.GetByID(ctx, id)
	if err != nil {
		return nil, model.NewNotFoundError("branch", id.String())
	}

	branch.Name = name
	branch.Address = address
	branch.Phone = phone
	branch.IsActive = isActive

	if err := s.branches.Update(ctx, branch); err != nil {
		return nil, fmt.Errorf("failed to update branch: %w", err)
	}
	return branch, nil
}

// ============ USER SERVICE ============

type UserService struct {
	users     repository.UserRepositoryInterface
	bcryptCost int
	log       *slog.Logger
}

func NewUserService(users repository.UserRepositoryInterface, bcryptCost int, log *slog.Logger) *UserService {
	return &UserService{users: users, bcryptCost: bcryptCost, log: log}
}

func (s *UserService) Create(ctx context.Context, branchID uuid.UUID, email, password, name string, role model.UserRole) (*model.User, error) {
	hash, err := middleware.HashPassword(password, s.bcryptCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	user := &model.User{
		ID:           uuid.New(),
		BranchID:     branchID,
		Email:        email,
		PasswordHash: hash,
		Name:         name,
		Role:         role,
		IsActive:     true,
	}
	if err := s.users.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}
	s.log.Info("user created", slog.String("id", user.ID.String()), slog.String("email", email))
	return user, nil
}

func (s *UserService) GetByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	user, err := s.users.GetByID(ctx, id)
	if err != nil {
		return nil, model.NewNotFoundError("user", id.String())
	}
	return user, nil
}

func (s *UserService) GetByBranch(ctx context.Context, branchID uuid.UUID) ([]model.User, error) {
	return s.users.GetByBranch(ctx, branchID)
}

func (s *UserService) Update(ctx context.Context, id uuid.UUID, name string, role *model.UserRole, isActive *bool) (*model.User, error) {
	user, err := s.users.GetByID(ctx, id)
	if err != nil {
		return nil, model.NewNotFoundError("user", id.String())
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

	if err := s.users.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}
	return user, nil
}

func (s *UserService) Delete(ctx context.Context, id uuid.UUID) error {
	if err := s.users.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}
	s.log.Info("user deleted", slog.String("id", id.String()))
	return nil
}

func (s *UserService) ChangePassword(ctx context.Context, id uuid.UUID, newPassword string) error {
	hash, err := middleware.HashPassword(newPassword, s.bcryptCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}
	if err := s.users.UpdatePassword(ctx, id, hash); err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}
	s.log.Info("user password changed", slog.String("id", id.String()))
	return nil
}

// ============ CATEGORY SERVICE ============

type CategoryService struct {
	categories repository.CategoryRepositoryInterface
	log        *slog.Logger
}

func NewCategoryService(categories repository.CategoryRepositoryInterface, log *slog.Logger) *CategoryService {
	return &CategoryService{categories: categories, log: log}
}

func (s *CategoryService) Create(ctx context.Context, branchID uuid.UUID, name, description string, sortOrder int) (*model.Category, error) {
	if sortOrder == 0 {
		maxOrder, err := s.categories.GetMaxSortOrder(ctx, branchID)
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
	if err := s.categories.Create(ctx, category); err != nil {
		return nil, fmt.Errorf("failed to create category: %w", err)
	}
	s.log.Info("category created", slog.String("id", category.ID.String()), slog.String("name", name))
	return category, nil
}

func (s *CategoryService) GetByID(ctx context.Context, id uuid.UUID) (*model.Category, error) {
	category, err := s.categories.GetByID(ctx, id)
	if err != nil {
		return nil, model.NewNotFoundError("category", id.String())
	}
	return category, nil
}

func (s *CategoryService) GetByBranch(ctx context.Context, branchID uuid.UUID) ([]model.Category, error) {
	return s.categories.GetByBranch(ctx, branchID)
}

func (s *CategoryService) Update(ctx context.Context, id uuid.UUID, name *string, description *string, sortOrder *int, isActive *bool) (*model.Category, error) {
	category, err := s.categories.GetByID(ctx, id)
	if err != nil {
		return nil, model.NewNotFoundError("category", id.String())
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

	if err := s.categories.Update(ctx, category); err != nil {
		return nil, fmt.Errorf("failed to update category: %w", err)
	}
	return category, nil
}

func (s *CategoryService) Delete(ctx context.Context, id uuid.UUID) error {
	if err := s.categories.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete category: %w", err)
	}
	s.log.Info("category deleted", slog.String("id", id.String()))
	return nil
}

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
	if err := s.categories.Reorder(ctx, branchID, orders); err != nil {
		return nil, fmt.Errorf("failed to reorder categories: %w", err)
	}
	cats, err := s.categories.GetByBranch(ctx, branchID)
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
	products repository.ProductRepositoryInterface
	log      *slog.Logger
}

func NewProductService(products repository.ProductRepositoryInterface, log *slog.Logger) *ProductService {
	return &ProductService{products: products, log: log}
}

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

func (s *ProductService) Create(ctx context.Context, input CreateProductInput) (*model.Product, error) {
	isAvailable := true
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
	if err := s.products.Create(ctx, product); err != nil {
		return nil, fmt.Errorf("failed to create product: %w", err)
	}
	s.log.Info("product created", slog.String("id", product.ID.String()), slog.String("name", product.Name))
	return product, nil
}

func (s *ProductService) GetByID(ctx context.Context, id uuid.UUID) (*model.Product, error) {
	product, err := s.products.GetByID(ctx, id)
	if err != nil {
		return nil, model.NewNotFoundError("product", id.String())
	}
	return product, nil
}

func (s *ProductService) GetByBranch(ctx context.Context, branchID uuid.UUID) ([]model.Product, error) {
	return s.products.GetAvailable(ctx, branchID)
}

func (s *ProductService) GetByBranchAll(ctx context.Context, branchID uuid.UUID) ([]model.Product, error) {
	return s.products.GetByBranch(ctx, branchID)
}

func (s *ProductService) GetByCategory(ctx context.Context, branchID, categoryID uuid.UUID) ([]model.Product, error) {
	return s.products.GetByCategory(ctx, branchID, categoryID)
}

func (s *ProductService) GetFeatured(ctx context.Context, branchID uuid.UUID) ([]model.Product, error) {
	return s.products.GetFeatured(ctx, branchID)
}

func (s *ProductService) Update(ctx context.Context, id uuid.UUID, input CreateProductInput) (*model.Product, error) {
	product, err := s.products.GetByID(ctx, id)
	if err != nil {
		return nil, model.NewNotFoundError("product", id.String())
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

	if err := s.products.Update(ctx, product); err != nil {
		return nil, fmt.Errorf("failed to update product: %w", err)
	}
	return product, nil
}

func (s *ProductService) Delete(ctx context.Context, id uuid.UUID) error {
	if err := s.products.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete product: %w", err)
	}
	s.log.Info("product deleted", slog.String("id", id.String()))
	return nil
}

func (s *ProductService) ToggleAvailability(ctx context.Context, id uuid.UUID) (*model.Product, error) {
	product, err := s.products.GetByID(ctx, id)
	if err != nil {
		return nil, model.NewNotFoundError("product", id.String())
	}
	newStatus := !product.IsAvailable
	if err := s.products.ToggleAvailability(ctx, id, newStatus); err != nil {
		return nil, fmt.Errorf("failed to toggle availability: %w", err)
	}
	product.IsAvailable = newStatus
	return product, nil
}

// ============ TABLE SERVICE ============

type TableService struct {
	tables repository.TableRepositoryInterface
	log    *slog.Logger
}

func NewTableService(tables repository.TableRepositoryInterface, log *slog.Logger) *TableService {
	return &TableService{tables: tables, log: log}
}

func (s *TableService) Create(ctx context.Context, branchID uuid.UUID, number string, capacity int) (*model.RestaurantTable, error) {
	table := &model.RestaurantTable{
		ID:       uuid.New(),
		BranchID: branchID,
		Number:   number,
		Capacity: capacity,
		Status:   model.TableAvailable,
	}
	if err := s.tables.Create(ctx, table); err != nil {
		return nil, fmt.Errorf("failed to create table: %w", err)
	}
	s.log.Info("table created", slog.String("id", table.ID.String()), slog.String("number", number))
	return table, nil
}

func (s *TableService) GetByID(ctx context.Context, id uuid.UUID) (*model.RestaurantTable, error) {
	table, err := s.tables.GetByID(ctx, id)
	if err != nil {
		return nil, model.NewNotFoundError("table", id.String())
	}
	return table, nil
}

func (s *TableService) GetByBranch(ctx context.Context, branchID uuid.UUID) ([]model.RestaurantTable, error) {
	return s.tables.GetByBranch(ctx, branchID)
}

func (s *TableService) Update(ctx context.Context, id uuid.UUID, number string, capacity int) (*model.RestaurantTable, error) {
	table, err := s.tables.GetByID(ctx, id)
	if err != nil {
		return nil, model.NewNotFoundError("table", id.String())
	}

	table.Number = number
	table.Capacity = capacity

	if err := s.tables.Update(ctx, table); err != nil {
		return nil, fmt.Errorf("failed to update table: %w", err)
	}
	return table, nil
}

func (s *TableService) Delete(ctx context.Context, id uuid.UUID) error {
	if err := s.tables.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete table: %w", err)
	}
	s.log.Info("table deleted", slog.String("id", id.String()))
	return nil
}

func (s *TableService) UpdateStatus(ctx context.Context, id uuid.UUID, status model.TableStatus) error {
	if err := s.tables.UpdateStatus(ctx, id, status); err != nil {
		return fmt.Errorf("failed to update table status: %w", err)
	}
	s.log.Info("table status updated", slog.String("id", id.String()), slog.String("status", string(status)))
	return nil
}

func (s *TableService) GetAvailableTables(ctx context.Context, branchID uuid.UUID) ([]model.RestaurantTable, error) {
	return s.tables.GetAvailable(ctx, branchID)
}

// ============ ORDER SERVICE ============

type OrderService struct {
	orders   repository.OrderRepositoryInterface
	products repository.ProductRepositoryInterface
	tables   repository.TableRepositoryInterface
	taxRate  float64
	log      *slog.Logger
}

func NewOrderService(orders repository.OrderRepositoryInterface, products repository.ProductRepositoryInterface, tables repository.TableRepositoryInterface, taxRate float64, log *slog.Logger) *OrderService {
	return &OrderService{
		orders:   orders,
		products: products,
		tables:   tables,
		taxRate:  taxRate,
		log:      log,
	}
}

func (s *OrderService) SetTaxRate(rate float64) {
	s.taxRate = rate
}

type OrderItemInput struct {
	ProductID uuid.UUID
	Quantity  int
	Notes     string
}

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

	// If table is specified, check availability
	if input.TableID != nil {
		table, err := s.tables.GetByID(ctx, *input.TableID)
		if err != nil {
			return nil, ErrTableNotFound
		}
		if table.Status != model.TableAvailable {
			return nil, ErrTableNotAvailable
		}
		if err := s.tables.UpdateStatus(ctx, *input.TableID, model.TableOccupied); err != nil {
			return nil, err
		}
	}

	// Calculate subtotal
	var subtotal float64
	var orderItemsModel []model.OrderItem

	for _, item := range orderItems {
		product, err := s.products.GetByID(ctx, item.ProductID)
		if err != nil {
			return nil, ErrProductNotFound
		}
		if !product.IsAvailable {
			return nil, fmt.Errorf("product is not available: %s", product.Name)
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

	tax := subtotal * s.taxRate
	total := subtotal + tax

	order := &model.Order{
		ID:            uuid.New(),
		BranchID:      input.BranchID,
		TableID:       input.TableID,
		UserID:        input.UserID,
		CustomerName:  customerName,
		CustomerPhone: customerPhone,
		Status:        model.OrderPending,
		Subtotal:      subtotal,
		Tax:           tax,
		Discount:      0,
		Total:         total,
		Notes:         notes,
	}

	if err := s.orders.Create(ctx, order); err != nil {
		if input.TableID != nil {
			s.tables.UpdateStatus(ctx, *input.TableID, model.TableAvailable)
		}
		return nil, fmt.Errorf("failed to create order: %w", err)
	}

	for i := range orderItemsModel {
		orderItemsModel[i].OrderID = order.ID
		if err := s.orders.AddItem(ctx, &orderItemsModel[i]); err != nil {
			return nil, err
		}
	}

	s.log.Info("order created", slog.String("id", order.ID.String()), slog.Float64("total", order.Total))
	return order, nil
}

func (s *OrderService) GetByID(ctx context.Context, id uuid.UUID) (*model.Order, error) {
	order, err := s.orders.GetByID(ctx, id)
	if err != nil {
		return nil, model.NewNotFoundError("order", id.String())
	}
	return order, nil
}

func (s *OrderService) GetByBranch(ctx context.Context, branchID uuid.UUID) ([]model.Order, error) {
	return s.orders.GetByBranch(ctx, branchID)
}

func (s *OrderService) GetItems(ctx context.Context, orderID uuid.UUID) ([]model.OrderItem, error) {
	return s.orders.GetItems(ctx, orderID)
}

func (s *OrderService) GetItem(ctx context.Context, itemID uuid.UUID) (*model.OrderItem, error) {
	item, err := s.orders.GetItem(ctx, itemID)
	if err != nil {
		return nil, model.NewNotFoundError("order item", itemID.String())
	}
	return item, nil
}

func (s *OrderService) UpdateStatus(ctx context.Context, id uuid.UUID, newStatus model.OrderStatus) error {
	order, err := s.orders.GetByID(ctx, id)
	if err != nil {
		return ErrOrderNotFound
	}

	if !s.isValidStatusTransition(order.Status, newStatus) {
		return ErrInvalidStatusTransition
	}

	if newStatus == model.OrderDelivered && order.TableID != nil {
		if err := s.tables.UpdateStatus(ctx, *order.TableID, model.TableAvailable); err != nil {
			return err
		}
	}

	if newStatus == model.OrderCancelled && order.TableID != nil {
		if err := s.tables.UpdateStatus(ctx, *order.TableID, model.TableAvailable); err != nil {
			return err
		}
	}

	if err := s.orders.UpdateStatus(ctx, id, newStatus); err != nil {
		return fmt.Errorf("failed to update order status: %w", err)
	}
	s.log.Info("order status updated", slog.String("id", id.String()), slog.String("status", string(newStatus)))
	return nil
}

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

func (s *OrderService) Cancel(ctx context.Context, id uuid.UUID) error {
	order, err := s.orders.GetByID(ctx, id)
	if err != nil {
		return ErrOrderNotFound
	}

	if order.Status != model.OrderPending && order.Status != model.OrderAccepted {
		return ErrOrderNotCancellable
	}

	if order.TableID != nil {
		if err := s.tables.UpdateStatus(ctx, *order.TableID, model.TableAvailable); err != nil {
			return err
		}
	}

	return s.orders.UpdateStatus(ctx, id, model.OrderCancelled)
}

func (s *OrderService) AddItem(ctx context.Context, orderID uuid.UUID, input OrderItemInput) (*model.OrderItem, error) {
	order, err := s.orders.GetByID(ctx, orderID)
	if err != nil {
		return nil, ErrOrderNotFound
	}

	if order.Status != model.OrderPending && order.Status != model.OrderAccepted {
		return nil, errors.New("can only add items to pending or accepted orders")
	}

	product, err := s.products.GetByID(ctx, input.ProductID)
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

	if err := s.orders.AddItem(ctx, orderItem); err != nil {
		return nil, err
	}

	if _, err := s.CalculateTotals(ctx, orderID); err != nil {
		return nil, err
	}

	return orderItem, nil
}

func (s *OrderService) RemoveItem(ctx context.Context, orderID, itemID uuid.UUID) (*model.Order, error) {
	order, err := s.orders.GetByID(ctx, orderID)
	if err != nil {
		return nil, ErrOrderNotFound
	}

	if order.Status != model.OrderPending && order.Status != model.OrderAccepted {
		return nil, errors.New("can only remove items from pending or accepted orders")
	}

	if err := s.orders.RemoveItem(ctx, itemID); err != nil {
		return nil, err
	}

	return s.CalculateTotals(ctx, orderID)
}

func (s *OrderService) CalculateTotals(ctx context.Context, orderID uuid.UUID) (*model.Order, error) {
	order, err := s.orders.GetByID(ctx, orderID)
	if err != nil {
		return nil, ErrOrderNotFound
	}

	items, err := s.orders.GetItems(ctx, orderID)
	if err != nil {
		return nil, err
	}

	var subtotal float64
	for _, item := range items {
		subtotal += item.TotalPrice
	}

	order.Subtotal = subtotal
	order.Tax = subtotal * s.taxRate
	order.Total = subtotal + order.Tax - order.Discount

	if err := s.orders.Update(ctx, order); err != nil {
		return nil, err
	}

	return order, nil
}

func (s *OrderService) ApplyDiscount(ctx context.Context, orderID uuid.UUID, discount float64) (*model.Order, error) {
	order, err := s.orders.GetByID(ctx, orderID)
	if err != nil {
		return nil, ErrOrderNotFound
	}

	if discount < 0 || discount > order.Subtotal {
		return nil, errors.New("invalid discount amount")
	}

	order.Discount = discount
	order.Total = order.Subtotal + order.Tax - discount

	if err := s.orders.Update(ctx, order); err != nil {
		return nil, err
	}

	return order, nil
}

// ============ RESERVATION SERVICE ============

type ReservationService struct {
	reservations repository.ReservationRepositoryInterface
	tables       repository.TableRepositoryInterface
	log          *slog.Logger
}

func NewReservationService(reservations repository.ReservationRepositoryInterface, tables repository.TableRepositoryInterface, log *slog.Logger) *ReservationService {
	return &ReservationService{
		reservations: reservations,
		tables:       tables,
		log:          log,
	}
}

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

func (s *ReservationService) Create(ctx context.Context, input CreateReservationInput) (*model.Reservation, error) {
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

	if input.TableID != nil {
		table, err := s.tables.GetByID(ctx, *input.TableID)
		if err != nil {
			return nil, ErrTableNotFound
		}
		if table.Status != model.TableAvailable {
			return nil, ErrTableNotAvailable
		}
		if err := s.tables.UpdateStatus(ctx, *input.TableID, model.TableReserved); err != nil {
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

	if err := s.reservations.Create(ctx, reservation); err != nil {
		if input.TableID != nil {
			s.tables.UpdateStatus(ctx, *input.TableID, model.TableAvailable)
		}
		return nil, fmt.Errorf("failed to create reservation: %w", err)
	}

	s.log.Info("reservation created", slog.String("id", reservation.ID.String()))
	return reservation, nil
}

func (s *ReservationService) GetByID(ctx context.Context, id uuid.UUID) (*model.Reservation, error) {
	res, err := s.reservations.GetByID(ctx, id)
	if err != nil {
		return nil, model.NewNotFoundError("reservation", id.String())
	}
	return res, nil
}

func (s *ReservationService) GetByBranch(ctx context.Context, branchID uuid.UUID) ([]model.Reservation, error) {
	return s.reservations.GetByBranch(ctx, branchID)
}

func (s *ReservationService) GetByDate(ctx context.Context, branchID uuid.UUID, date time.Time) ([]model.Reservation, error) {
	return s.reservations.GetByDate(ctx, branchID, date.Format("2006-01-02"))
}

func (s *ReservationService) Update(ctx context.Context, id uuid.UUID, input CreateReservationInput) (*model.Reservation, error) {
	reservation, err := s.reservations.GetByID(ctx, id)
	if err != nil {
		return nil, model.NewNotFoundError("reservation", id.String())
	}

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

	if input.TableID != nil && (reservation.TableID == nil || *reservation.TableID != *input.TableID) {
		if reservation.TableID != nil {
			if err := s.tables.UpdateStatus(ctx, *reservation.TableID, model.TableAvailable); err != nil {
				return nil, err
			}
		}
		newTable, err := s.tables.GetByID(ctx, *input.TableID)
		if err != nil {
			return nil, ErrTableNotFound
		}
		if newTable.Status != model.TableAvailable {
			return nil, ErrTableNotAvailable
		}
		if err := s.tables.UpdateStatus(ctx, *input.TableID, model.TableReserved); err != nil {
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

	if err := s.reservations.Update(ctx, reservation); err != nil {
		return nil, fmt.Errorf("failed to update reservation: %w", err)
	}

	return reservation, nil
}

func (s *ReservationService) Cancel(ctx context.Context, id uuid.UUID) error {
	reservation, err := s.reservations.GetByID(ctx, id)
	if err != nil {
		return ErrReservationNotFound
	}

	if reservation.Status == model.ReservationCancelled || reservation.Status == model.ReservationCompleted {
		return errors.New("reservation cannot be cancelled in current state")
	}

	if reservation.TableID != nil {
		if err := s.tables.UpdateStatus(ctx, *reservation.TableID, model.TableAvailable); err != nil {
			return err
		}
	}

	if err := s.reservations.UpdateStatus(ctx, id, model.ReservationCancelled); err != nil {
		return fmt.Errorf("failed to cancel reservation: %w", err)
	}
	s.log.Info("reservation cancelled", slog.String("id", id.String()))
	return nil
}

func (s *ReservationService) Confirm(ctx context.Context, id uuid.UUID) error {
	reservation, err := s.reservations.GetByID(ctx, id)
	if err != nil {
		return ErrReservationNotFound
	}

	if reservation.Status != model.ReservationPending {
		return errors.New("only pending reservations can be confirmed")
	}

	if err := s.reservations.UpdateStatus(ctx, id, model.ReservationConfirmed); err != nil {
		return fmt.Errorf("failed to confirm reservation: %w", err)
	}
	s.log.Info("reservation confirmed", slog.String("id", id.String()))
	return nil
}

func (s *ReservationService) Complete(ctx context.Context, id uuid.UUID) error {
	reservation, err := s.reservations.GetByID(ctx, id)
	if err != nil {
		return ErrReservationNotFound
	}

	if reservation.Status != model.ReservationConfirmed {
		return errors.New("only confirmed reservations can be completed")
	}

	if reservation.TableID != nil {
		if err := s.tables.UpdateStatus(ctx, *reservation.TableID, model.TableAvailable); err != nil {
			return err
		}
	}

	if err := s.reservations.UpdateStatus(ctx, id, model.ReservationCompleted); err != nil {
		return fmt.Errorf("failed to complete reservation: %w", err)
	}
	s.log.Info("reservation completed", slog.String("id", id.String()))
	return nil
}

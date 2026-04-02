package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/greenpos/backend/internal/model"
)

// ErrNotFound is returned when an entity is not found
var ErrNotFound = model.ErrNotFound

// ErrInvoiceNotFound is returned when an invoice is not found
var ErrInvoiceNotFound = model.NewNotFoundError("invoice", "")

// ============ Repository Interfaces ============
// These interfaces define the contract for data access layer
// Implementations should be in separate files (e.g., postgres_*.go, memory_*.go)

// TenantRepositoryInterface defines operations for tenants
type TenantRepositoryInterface interface {
	// Create creates a new tenant
	Create(ctx context.Context, tenant *model.Tenant) error

	// GetByID retrieves a tenant by ID
	GetByID(ctx context.Context, id uuid.UUID) (*model.Tenant, error)

	// GetBySlug retrieves a tenant by slug
	GetBySlug(ctx context.Context, slug string) (*model.Tenant, error)

	// Update updates an existing tenant
	Update(ctx context.Context, tenant *model.Tenant) error

	// List returns all tenants
	List(ctx context.Context) ([]model.Tenant, error)

	// Seed populates initial data (for development/testing)
	Seed(ctx context.Context) error
}

// BranchRepositoryInterface defines operations for branches
type BranchRepositoryInterface interface {
	// Create creates a new branch
	Create(ctx context.Context, branch *model.Branch) error

	// GetByID retrieves a branch by ID
	GetByID(ctx context.Context, id uuid.UUID) (*model.Branch, error)

	// GetByTenant retrieves all branches for a tenant
	GetByTenant(ctx context.Context, tenantID uuid.UUID) ([]model.Branch, error)

	// Update updates an existing branch
	Update(ctx context.Context, branch *model.Branch) error

	// Delete removes a branch (soft delete)
	Delete(ctx context.Context, id uuid.UUID) error

	// Seed populates initial data
	Seed(ctx context.Context) error
}

// UserRepositoryInterface defines operations for users
type UserRepositoryInterface interface {
	// Create creates a new user
	Create(ctx context.Context, user *model.User) error

	// GetByID retrieves a user by ID
	GetByID(ctx context.Context, id uuid.UUID) (*model.User, error)

	// GetByEmail retrieves a user by email within a branch
	GetByEmail(ctx context.Context, branchID uuid.UUID, email string) (*model.User, error)

	// GetByBranch retrieves all users for a branch
	GetByBranch(ctx context.Context, branchID uuid.UUID) ([]model.User, error)

	// ListByBranch is an alias for GetByBranch
	ListByBranch(ctx context.Context, branchID uuid.UUID) ([]model.User, error)

	// Update updates an existing user
	Update(ctx context.Context, user *model.User) error

	// UpdatePassword updates a user's password hash
	UpdatePassword(ctx context.Context, id uuid.UUID, passwordHash string) error

	// Delete soft-deletes a user (sets is_active to false)
	Delete(ctx context.Context, id uuid.UUID) error

	// Seed populates initial data
	Seed(ctx context.Context) error
}

// CategoryRepositoryInterface defines operations for categories
type CategoryRepositoryInterface interface {
	// Create creates a new category
	Create(ctx context.Context, category *model.Category) error

	// GetByID retrieves a category by ID
	GetByID(ctx context.Context, id uuid.UUID) (*model.Category, error)

	// GetByBranch retrieves all categories for a branch
	GetByBranch(ctx context.Context, branchID uuid.UUID) ([]model.Category, error)

	// Update updates an existing category
	Update(ctx context.Context, category *model.Category) error

	// Delete soft-deletes a category
	Delete(ctx context.Context, id uuid.UUID) error

	// GetMaxSortOrder returns the maximum sort order for a branch
	GetMaxSortOrder(ctx context.Context, branchID uuid.UUID) (int, error)

	// Reorder updates sort orders for multiple categories
	Reorder(ctx context.Context, branchID uuid.UUID, orders []struct{ ID uuid.UUID; SortOrder int }) error

	// Seed populates initial data
	Seed(ctx context.Context) error
}

// ProductRepositoryInterface defines operations for products
type ProductRepositoryInterface interface {
	// Create creates a new product
	Create(ctx context.Context, product *model.Product) error

	// GetByID retrieves a product by ID
	GetByID(ctx context.Context, id uuid.UUID) (*model.Product, error)

	// GetByBranch retrieves all products for a branch
	GetByBranch(ctx context.Context, branchID uuid.UUID) ([]model.Product, error)

	// GetByCategory retrieves all products for a category
	GetByCategory(ctx context.Context, branchID, categoryID uuid.UUID) ([]model.Product, error)

	// GetAvailable retrieves available products for a branch
	GetAvailable(ctx context.Context, branchID uuid.UUID) ([]model.Product, error)

	// GetFeatured retrieves featured products for a branch
	GetFeatured(ctx context.Context, branchID uuid.UUID) ([]model.Product, error)

	// Update updates an existing product
	Update(ctx context.Context, product *model.Product) error

	// Delete soft-deletes a product
	Delete(ctx context.Context, id uuid.UUID) error

	// ToggleAvailability toggles product availability
	ToggleAvailability(ctx context.Context, id uuid.UUID, isAvailable bool) error

	// Seed populates initial data
	Seed(ctx context.Context) error
}

// TableRepositoryInterface defines operations for restaurant tables
type TableRepositoryInterface interface {
	// Create creates a new table
	Create(ctx context.Context, table *model.RestaurantTable) error

	// GetByID retrieves a table by ID
	GetByID(ctx context.Context, id uuid.UUID) (*model.RestaurantTable, error)

	// GetByBranch retrieves all tables for a branch
	GetByBranch(ctx context.Context, branchID uuid.UUID) ([]model.RestaurantTable, error)

	// ListByBranch is an alias for GetByBranch
	ListByBranch(ctx context.Context, branchID uuid.UUID) ([]model.RestaurantTable, error)

	// Update updates an existing table
	Update(ctx context.Context, table *model.RestaurantTable) error

	// UpdateStatus updates a table's status
	UpdateStatus(ctx context.Context, id uuid.UUID, status model.TableStatus) error

	// Delete removes a table
	Delete(ctx context.Context, id uuid.UUID) error

	// GetAvailable retrieves available tables for a branch
	GetAvailable(ctx context.Context, branchID uuid.UUID) ([]model.RestaurantTable, error)

	// Seed populates initial data
	Seed(ctx context.Context) error
}

// OrderRepositoryInterface defines operations for orders
type OrderRepositoryInterface interface {
	// Create creates a new order
	Create(ctx context.Context, order *model.Order) error

	// GetByID retrieves an order by ID
	GetByID(ctx context.Context, id uuid.UUID) (*model.Order, error)

	// GetByBranch retrieves all orders for a branch
	GetByBranch(ctx context.Context, branchID uuid.UUID) ([]model.Order, error)

	// ListByBranch retrieves orders with optional filtering
	ListByBranch(ctx context.Context, branchID uuid.UUID, status *model.OrderStatus, limit *int) ([]model.Order, error)

	// Update updates an existing order
	Update(ctx context.Context, order *model.Order) error

	// UpdateStatus updates an order's status
	UpdateStatus(ctx context.Context, id uuid.UUID, status model.OrderStatus) error

	// AddItem adds an item to an order
	AddItem(ctx context.Context, item *model.OrderItem) error

	// GetItems retrieves all items for an order
	GetItems(ctx context.Context, orderID uuid.UUID) ([]model.OrderItem, error)

	// GetItem retrieves a single order item
	GetItem(ctx context.Context, itemID uuid.UUID) (*model.OrderItem, error)

	// RemoveItem removes an item from an order
	RemoveItem(ctx context.Context, itemID uuid.UUID) error

	// GetActiveByTable retrieves the active order for a table
	GetActiveByTable(ctx context.Context, tableID uuid.UUID) (*model.Order, error)

	// GetActive retrieves all active orders for a branch
	GetActive(ctx context.Context, branchID uuid.UUID) ([]*model.Order, error)

	// Seed populates initial data
	Seed(ctx context.Context) error
}

// ReservationRepositoryInterface defines operations for reservations
type ReservationRepositoryInterface interface {
	// Create creates a new reservation
	Create(ctx context.Context, reservation *model.Reservation) error

	// GetByID retrieves a reservation by ID
	GetByID(ctx context.Context, id uuid.UUID) (*model.Reservation, error)

	// GetByBranch retrieves all reservations for a branch
	GetByBranch(ctx context.Context, branchID uuid.UUID) ([]model.Reservation, error)

	// ListByBranch retrieves reservations with optional date filter
	ListByBranch(ctx context.Context, branchID uuid.UUID, date *time.Time) ([]model.Reservation, error)

	// GetByDate retrieves reservations for a specific date
	GetByDate(ctx context.Context, branchID uuid.UUID, date string) ([]model.Reservation, error)

	// Update updates an existing reservation
	Update(ctx context.Context, reservation *model.Reservation) error

	// UpdateStatus updates a reservation's status
	UpdateStatus(ctx context.Context, id uuid.UUID, status model.ReservationStatus) error

	// Seed populates initial data
	Seed(ctx context.Context) error
}

// RefreshTokenRepositoryInterface defines operations for refresh tokens
type RefreshTokenRepositoryInterface interface {
	// Create creates a new refresh token
	Create(ctx context.Context, token *RefreshToken) error

	// GetByTokenHash retrieves a refresh token by its hash
	GetByTokenHash(ctx context.Context, tokenHash string) (*RefreshToken, error)

	// Revoke revokes a refresh token
	Revoke(ctx context.Context, tokenHash string) error

	// RevokeAllForUser revokes all refresh tokens for a user
	RevokeAllForUser(ctx context.Context, userID uuid.UUID) error

	// DeleteExpired removes all expired tokens
	DeleteExpired(ctx context.Context) error
}

// InvoiceRepositoryInterface defines operations for invoices
type InvoiceRepositoryInterface interface {
	Create(ctx context.Context, invoice *model.Invoice) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.Invoice, error)
	GetByTenant(ctx context.Context, tenantID uuid.UUID) ([]model.Invoice, error)
	GetByBranch(ctx context.Context, branchID uuid.UUID) ([]model.Invoice, error)
	GetByOrder(ctx context.Context, orderID uuid.UUID) (*model.Invoice, error)
	Update(ctx context.Context, invoice *model.Invoice) error
	UpdateStatus(ctx context.Context, id uuid.UUID, status model.InvoiceStatus) error
	SetTimbradaData(ctx context.Context, id uuid.UUID, invoiceUUID, pdfURL, xmlURL, facturapiID string) error
	ListByDateRange(ctx context.Context, branchID uuid.UUID, start, end time.Time) ([]model.Invoice, error)
	CountByBranch(ctx context.Context, branchID uuid.UUID) (int, error)
	Seed(ctx context.Context) error
}

// InvoiceItemRepositoryInterface defines operations for invoice items
type InvoiceItemRepositoryInterface interface {
	Create(ctx context.Context, item *model.InvoiceItem) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.InvoiceItem, error)
	GetByInvoice(ctx context.Context, invoiceID uuid.UUID) ([]model.InvoiceItem, error)
	Update(ctx context.Context, item *model.InvoiceItem) error
	Delete(ctx context.Context, id uuid.UUID) error
	Seed(ctx context.Context) error
}

// TenantFiscalRepositoryInterface defines operations for tenant fiscal data
type TenantFiscalRepositoryInterface interface {
	Create(ctx context.Context, fiscal *model.TenantFiscal) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.TenantFiscal, error)
	GetByTenant(ctx context.Context, tenantID uuid.UUID) (*model.TenantFiscal, error)
	Update(ctx context.Context, fiscal *model.TenantFiscal) error
	Seed(ctx context.Context) error
}

// ============ Type Aliases for Concrete Types ============
// These aliases allow existing code to compile while we transition to interfaces

type (
	TenantRepository      = TenantMemoryRepository
	BranchRepository      = BranchMemoryRepository
	UserRepository        = UserMemoryRepository
	CategoryRepository    = CategoryMemoryRepository
	ProductRepository     = ProductMemoryRepository
	TableRepository       = TableMemoryRepository
	OrderRepository       = OrderMemoryRepository
	ReservationRepository = ReservationMemoryRepository
)

// RefreshToken represents a refresh token for JWT
type RefreshToken struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	TokenHash string
	ExpiresAt time.Time
	CreatedAt time.Time
	RevokedAt *time.Time
}

// IsExpired returns true if the token is expired
func (rt *RefreshToken) IsExpired() bool {
	return time.Now().After(rt.ExpiresAt)
}

// IsRevoked returns true if the token has been revoked
func (rt *RefreshToken) IsRevoked() bool {
	return rt.RevokedAt != nil
}

// IsValid returns true if the token is valid (not expired and not revoked)
func (rt *RefreshToken) IsValid() bool {
	return !rt.IsExpired() && !rt.IsRevoked()
}

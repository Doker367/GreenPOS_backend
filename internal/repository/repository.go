package repository

import (
	"context"
	"database/sql"

	"gorm.io/gorm"
)

// Repositories holds all repository instances
type Repositories struct {
	Tenants       TenantRepositoryInterface
	Branches      BranchRepositoryInterface
	Users         UserRepositoryInterface
	Categories    CategoryRepositoryInterface
	Products      ProductRepositoryInterface
	Tables        TableRepositoryInterface
	Orders        OrderRepositoryInterface
	Reservations  ReservationRepositoryInterface
	Invoices      InvoiceRepositoryInterface
	InvoiceItems  InvoiceItemRepositoryInterface
	TenantFiscal  TenantFiscalRepositoryInterface
	InventoryItems    InventoryItemRepositoryInterface
	InventoryMovements InventoryMovementRepositoryInterface
	FCMTokens         FCMTokenRepositoryInterface
	Payments      PaymentRepositoryInterface
	QRCodes       QRCodeRepositoryInterface
}

// NewMemoryRepositories creates repositories with in-memory storage
func NewMemoryRepositories() *Repositories {
	return &Repositories{
		Tenants:      NewTenantMemoryRepository(),
		Branches:     NewBranchMemoryRepository(),
		Users:        NewUserMemoryRepository(),
		Categories:   NewCategoryMemoryRepository(),
		Products:     NewProductMemoryRepository(),
		Tables:       NewTableMemoryRepository(),
		Orders:       NewOrderMemoryRepository(),
		Reservations: NewReservationMemoryRepository(),
		Invoices:     NewInvoiceMemoryRepository(),
		InvoiceItems: NewInvoiceItemMemoryRepository(),
		TenantFiscal: NewTenantFiscalMemoryRepository(),
		InventoryItems:    NewInventoryItemMemoryRepository(),
		InventoryMovements: NewInventoryMovementMemoryRepository(),
		FCMTokens:    NewMemoryFCMTokenRepository(),
		// Note: Memory repository for payments would need to be implemented
		// For now, payments require database backend
	}
}

// NewDBRepositories creates database-backed repositories using GORM.
// Each repository wraps GORM operations for the corresponding domain model.
func NewDBRepositories(db *gorm.DB) *Repositories {
	return &Repositories{
		Tenants:      NewGormTenantRepository(db),
		Branches:     NewGormBranchRepository(db),
		Users:        NewGormUserRepository(db),
		Categories:   NewGormCategoryRepository(db),
		Products:     NewGormProductRepository(db),
		Tables:       NewGormTableRepository(db),
		Orders:       NewGormOrderRepository(db),
		Reservations: NewGormReservationRepository(db),
		Invoices:     NewGormInvoiceRepository(db),
		InvoiceItems: NewGormInvoiceItemRepository(db),
		TenantFiscal: NewGormTenantFiscalRepository(db),
		InventoryItems:    NewGormInventoryItemRepository(db),
		InventoryMovements: NewGormInventoryMovementRepository(db),
		Payments:     NewPaymentRepository(db.Client().(*sql.DB)),
		QRCodes:     NewGormQRCodeRepository(db),
	}
}

// Seed populates all repositories with sample data
func (r *Repositories) Seed(ctx context.Context) error {
	if err := r.Tenants.Seed(ctx); err != nil {
		return err
	}
	if err := r.Branches.Seed(ctx); err != nil {
		return err
	}
	if err := r.Users.Seed(ctx); err != nil {
		return err
	}
	if err := r.Categories.Seed(ctx); err != nil {
		return err
	}
	if err := r.Products.Seed(ctx); err != nil {
		return err
	}
	if err := r.Tables.Seed(ctx); err != nil {
		return err
	}
	if err := r.Orders.Seed(ctx); err != nil {
		return err
	}
	if err := r.Reservations.Seed(ctx); err != nil {
		return err
	}
	// Note: Invoices and TenantFiscal are not seeded automatically
	// as they contain sensitive fiscal data
	if err := r.InventoryItems.Seed(ctx); err != nil {
		return err
	}
	return nil
}

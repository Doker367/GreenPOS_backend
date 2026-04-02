package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/greenpos/backend/internal/model"
	"gorm.io/gorm"
)

// ============ GORM Invoice Repository ============

type GormInvoiceRepository struct {
	db *gorm.DB
}

func NewGormInvoiceRepository(db *gorm.DB) *GormInvoiceRepository {
	return &GormInvoiceRepository{db: db}
}

func (r *GormInvoiceRepository) Create(ctx context.Context, invoice *model.Invoice) error {
	if err := r.db.WithContext(ctx).Create(invoice).Error; err != nil {
		return fmt.Errorf("failed to create invoice: %w", err)
	}
	return nil
}

func (r *GormInvoiceRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Invoice, error) {
	var invoice model.Invoice
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&invoice).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInvoiceNotFound
		}
		return nil, fmt.Errorf("failed to get invoice: %w", err)
	}
	return &invoice, nil
}

func (r *GormInvoiceRepository) GetByTenant(ctx context.Context, tenantID uuid.UUID) ([]model.Invoice, error) {
	var invoices []model.Invoice
	if err := r.db.WithContext(ctx).Where("tenant_id = ?", tenantID).Order("created_at DESC").Find(&invoices).Error; err != nil {
		return nil, fmt.Errorf("failed to get invoices: %w", err)
	}
	return invoices, nil
}

func (r *GormInvoiceRepository) GetByBranch(ctx context.Context, branchID uuid.UUID) ([]model.Invoice, error) {
	var invoices []model.Invoice
	if err := r.db.WithContext(ctx).Where("branch_id = ?", branchID).Order("created_at DESC").Find(&invoices).Error; err != nil {
		return nil, fmt.Errorf("failed to get invoices: %w", err)
	}
	return invoices, nil
}

func (r *GormInvoiceRepository) GetByOrder(ctx context.Context, orderID uuid.UUID) (*model.Invoice, error) {
	var invoice model.Invoice
	if err := r.db.WithContext(ctx).Where("order_id = ?", orderID).First(&invoice).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInvoiceNotFound
		}
		return nil, fmt.Errorf("failed to get invoice: %w", err)
	}
	return &invoice, nil
}

func (r *GormInvoiceRepository) Update(ctx context.Context, invoice *model.Invoice) error {
	if err := r.db.WithContext(ctx).Save(invoice).Error; err != nil {
		return fmt.Errorf("failed to update invoice: %w", err)
	}
	return nil
}

func (r *GormInvoiceRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status model.InvoiceStatus) error {
	if err := r.db.WithContext(ctx).Model(&model.Invoice{}).Where("id = ?", id).Update("status", status).Error; err != nil {
		return fmt.Errorf("failed to update invoice status: %w", err)
	}
	return nil
}

func (r *GormInvoiceRepository) SetTimbradaData(ctx context.Context, id uuid.UUID, invoiceUUID, pdfURL, xmlURL, facturapiID string) error {
	updates := map[string]interface{}{
		"uuid":        invoiceUUID,
		"pdf_url":     pdfURL,
		"xml_url":     xmlURL,
		"facturapi_id": facturapiID,
		"status":      model.InvoiceTimbrada,
	}
	if err := r.db.WithContext(ctx).Model(&model.Invoice{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		return fmt.Errorf("failed to set timbrada data: %w", err)
	}
	return nil
}

func (r *GormInvoiceRepository) ListByDateRange(ctx context.Context, branchID uuid.UUID, start, end time.Time) ([]model.Invoice, error) {
	var invoices []model.Invoice
	if err := r.db.WithContext(ctx).
		Where("branch_id = ? AND created_at >= ? AND created_at <= ?", branchID, start, end).
		Order("created_at DESC").
		Find(&invoices).Error; err != nil {
		return nil, fmt.Errorf("failed to list invoices by date: %w", err)
	}
	return invoices, nil
}

func (r *GormInvoiceRepository) CountByBranch(ctx context.Context, branchID uuid.UUID) (int, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&model.Invoice{}).Where("branch_id = ?", branchID).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count invoices: %w", err)
	}
	return int(count), nil
}

func (r *GormInvoiceRepository) Seed(ctx context.Context) error { return nil }

// ============ GORM InvoiceItem Repository ============

type GormInvoiceItemRepository struct {
	db *gorm.DB
}

func NewGormInvoiceItemRepository(db *gorm.DB) *GormInvoiceItemRepository {
	return &GormInvoiceItemRepository{db: db}
}

func (r *GormInvoiceItemRepository) Create(ctx context.Context, item *model.InvoiceItem) error {
	if err := r.db.WithContext(ctx).Create(item).Error; err != nil {
		return fmt.Errorf("failed to create invoice item: %w", err)
	}
	return nil
}

func (r *GormInvoiceItemRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.InvoiceItem, error) {
	var item model.InvoiceItem
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&item).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInvoiceNotFound
		}
		return nil, fmt.Errorf("failed to get invoice item: %w", err)
	}
	return &item, nil
}

func (r *GormInvoiceItemRepository) GetByInvoice(ctx context.Context, invoiceID uuid.UUID) ([]model.InvoiceItem, error) {
	var items []model.InvoiceItem
	if err := r.db.WithContext(ctx).Where("invoice_id = ?", invoiceID).Find(&items).Error; err != nil {
		return nil, fmt.Errorf("failed to get invoice items: %w", err)
	}
	return items, nil
}

func (r *GormInvoiceItemRepository) Update(ctx context.Context, item *model.InvoiceItem) error {
	if err := r.db.WithContext(ctx).Save(item).Error; err != nil {
		return fmt.Errorf("failed to update invoice item: %w", err)
	}
	return nil
}

func (r *GormInvoiceItemRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if err := r.db.WithContext(ctx).Delete(&model.InvoiceItem{}, "id = ?", id).Error; err != nil {
		return fmt.Errorf("failed to delete invoice item: %w", err)
	}
	return nil
}

func (r *GormInvoiceItemRepository) Seed(ctx context.Context) error { return nil }

// ============ GORM TenantFiscal Repository ============

type GormTenantFiscalRepository struct {
	db *gorm.DB
}

func NewGormTenantFiscalRepository(db *gorm.DB) *GormTenantFiscalRepository {
	return &GormTenantFiscalRepository{db: db}
}

func (r *GormTenantFiscalRepository) Create(ctx context.Context, fiscal *model.TenantFiscal) error {
	if err := r.db.WithContext(ctx).Create(fiscal).Error; err != nil {
		return fmt.Errorf("failed to create tenant fiscal: %w", err)
	}
	return nil
}

func (r *GormTenantFiscalRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.TenantFiscal, error) {
	var fiscal model.TenantFiscal
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&fiscal).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInvoiceNotFound
		}
		return nil, fmt.Errorf("failed to get tenant fiscal: %w", err)
	}
	return &fiscal, nil
}

func (r *GormTenantFiscalRepository) GetByTenant(ctx context.Context, tenantID uuid.UUID) (*model.TenantFiscal, error) {
	var fiscal model.TenantFiscal
	if err := r.db.WithContext(ctx).Where("tenant_id = ?", tenantID).First(&fiscal).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInvoiceNotFound
		}
		return nil, fmt.Errorf("failed to get tenant fiscal: %w", err)
	}
	return &fiscal, nil
}

func (r *GormTenantFiscalRepository) Update(ctx context.Context, fiscal *model.TenantFiscal) error {
	if err := r.db.WithContext(ctx).Save(fiscal).Error; err != nil {
		return fmt.Errorf("failed to update tenant fiscal: %w", err)
	}
	return nil
}

func (r *GormTenantFiscalRepository) Seed(ctx context.Context) error { return nil }

// ============ Wrapper types for InvoiceService (adapts interface to concrete type) ============

// InvoiceRepository wraps an InvoiceRepositoryInterface to satisfy concrete type expectations
type InvoiceRepository struct {
	Repo InvoiceRepositoryInterface
}

func (r *InvoiceRepository) Create(ctx context.Context, invoice *model.Invoice) error {
	return r.Repo.Create(ctx, invoice)
}
func (r *InvoiceRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Invoice, error) {
	return r.Repo.GetByID(ctx, id)
}
func (r *InvoiceRepository) GetByTenant(ctx context.Context, tenantID uuid.UUID) ([]model.Invoice, error) {
	return r.Repo.GetByTenant(ctx, tenantID)
}
func (r *InvoiceRepository) GetByBranch(ctx context.Context, branchID uuid.UUID) ([]model.Invoice, error) {
	return r.Repo.GetByBranch(ctx, branchID)
}
func (r *InvoiceRepository) GetByOrder(ctx context.Context, orderID uuid.UUID) (*model.Invoice, error) {
	return r.Repo.GetByOrder(ctx, orderID)
}
func (r *InvoiceRepository) Update(ctx context.Context, invoice *model.Invoice) error {
	return r.Repo.Update(ctx, invoice)
}
func (r *InvoiceRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status model.InvoiceStatus) error {
	return r.Repo.UpdateStatus(ctx, id, status)
}
func (r *InvoiceRepository) SetTimbradaData(ctx context.Context, id uuid.UUID, invoiceUUID, pdfURL, xmlURL, facturapiID string) error {
	return r.Repo.SetTimbradaData(ctx, id, invoiceUUID, pdfURL, xmlURL, facturapiID)
}
func (r *InvoiceRepository) ListByDateRange(ctx context.Context, branchID uuid.UUID, start, end time.Time) ([]model.Invoice, error) {
	return r.Repo.ListByDateRange(ctx, branchID, start, end)
}
func (r *InvoiceRepository) CountByBranch(ctx context.Context, branchID uuid.UUID) (int, error) {
	return r.Repo.CountByBranch(ctx, branchID)
}

func (r *InvoiceRepository) Seed(ctx context.Context) error {
	return nil // invoices are sensitive; seed manually or via migrations
}

// InvoiceItemRepository wraps InvoiceItemRepositoryInterface
type InvoiceItemRepository struct {
	Repo InvoiceItemRepositoryInterface
}

func (r *InvoiceItemRepository) Create(ctx context.Context, item *model.InvoiceItem) error {
	return r.Repo.Create(ctx, item)
}
func (r *InvoiceItemRepository) GetByInvoice(ctx context.Context, invoiceID uuid.UUID) ([]model.InvoiceItem, error) {
	return r.Repo.GetByInvoice(ctx, invoiceID)
}
func (r *InvoiceItemRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.Repo.Delete(ctx, id)
}

// TenantFiscalRepository wraps TenantFiscalRepositoryInterface
type TenantFiscalRepository struct {
	Repo TenantFiscalRepositoryInterface
}

func (r *TenantFiscalRepository) Create(ctx context.Context, fiscal *model.TenantFiscal) error {
	return r.Repo.Create(ctx, fiscal)
}
func (r *TenantFiscalRepository) GetByTenant(ctx context.Context, tenantID uuid.UUID) (*model.TenantFiscal, error) {
	return r.Repo.GetByTenant(ctx, tenantID)
}
func (r *TenantFiscalRepository) Update(ctx context.Context, fiscal *model.TenantFiscal) error {
	return r.Repo.Update(ctx, fiscal)
}

func (r *TenantFiscalRepository) Seed(ctx context.Context) error {
	return nil // fiscal info must be configured by the tenant
}

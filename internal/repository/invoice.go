package repository

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/greenpos/backend/internal/model"
)

// ErrInvoiceNotFound is returned when an invoice is not found
var ErrInvoiceNotFound = errors.New("invoice not found")

// InvoiceRepositoryInterface defines operations for invoices
type InvoiceRepositoryInterface interface {
	Create(ctx context.Context, invoice *model.Invoice) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.Invoice, error)
	GetByTenant(ctx context.Context, tenantID uuid.UUID) ([]model.Invoice, error)
	GetByBranch(ctx context.Context, branchID uuid.UUID) ([]model.Invoice, error)
	GetByOrder(ctx context.Context, orderID uuid.UUID) (*model.Invoice, error)
	Update(ctx context.Context, invoice *model.Invoice) error
	UpdateStatus(ctx context.Context, id uuid.UUID, status model.InvoiceStatus) error
	SetTimbradaData(ctx context.Context, id uuid.UUID, uuid, pdfURL, xmlURL, facturapiID string) error
	ListByDateRange(ctx context.Context, branchID uuid.UUID, start, end time.Time) ([]model.Invoice, error)
	CountByBranch(ctx context.Context, branchID uuid.UUID) (int, error)
	Seed(ctx context.Context) error
}

// InvoiceMemoryRepository is an in-memory implementation of InvoiceRepository
type InvoiceMemoryRepository struct {
	mu        sync.RWMutex
	invoices  map[uuid.UUID]*model.Invoice
	byTenant  map[uuid.UUID][]*model.Invoice
	byBranch  map[uuid.UUID][]*model.Invoice
	byOrder   map[uuid.UUID]*model.Invoice
}

// NewInvoiceMemoryRepository creates a new in-memory invoice repository
func NewInvoiceMemoryRepository() *InvoiceMemoryRepository {
	return &InvoiceMemoryRepository{
		invoices: make(map[uuid.UUID]*model.Invoice),
		byTenant: make(map[uuid.UUID][]*model.Invoice),
		byBranch: make(map[uuid.UUID][]*model.Invoice),
		byOrder:  make(map[uuid.UUID]*model.Invoice),
	}
}

func (r *InvoiceMemoryRepository) Create(ctx context.Context, invoice *model.Invoice) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	invoice.CreatedAt = now
	invoice.UpdatedAt = now
	if invoice.ID == uuid.Nil {
		invoice.ID = uuid.New()
	}

	r.invoices[invoice.ID] = invoice
	r.byTenant[invoice.TenantID] = append(r.byTenant[invoice.TenantID], invoice)
	r.byBranch[invoice.BranchID] = append(r.byBranch[invoice.BranchID], invoice)
	r.byOrder[invoice.OrderID] = invoice
	return nil
}

func (r *InvoiceMemoryRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Invoice, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if inv, ok := r.invoices[id]; ok {
		return inv, nil
	}
	return nil, ErrInvoiceNotFound
}

func (r *InvoiceMemoryRepository) GetByTenant(ctx context.Context, tenantID uuid.UUID) ([]model.Invoice, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	invoices := r.byTenant[tenantID]
	result := make([]model.Invoice, len(invoices))
	for i, inv := range invoices {
		result[i] = *inv
	}
	return result, nil
}

func (r *InvoiceMemoryRepository) GetByBranch(ctx context.Context, branchID uuid.UUID) ([]model.Invoice, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	invoices := r.byBranch[branchID]
	result := make([]model.Invoice, len(invoices))
	for i, inv := range invoices {
		result[i] = *inv
	}
	return result, nil
}

func (r *InvoiceMemoryRepository) GetByOrder(ctx context.Context, orderID uuid.UUID) (*model.Invoice, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if inv, ok := r.byOrder[orderID]; ok {
		return inv, nil
	}
	return nil, ErrInvoiceNotFound
}

func (r *InvoiceMemoryRepository) Update(ctx context.Context, invoice *model.Invoice) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	invoice.UpdatedAt = time.Now()
	if _, ok := r.invoices[invoice.ID]; !ok {
		return ErrInvoiceNotFound
	}
	r.invoices[invoice.ID] = invoice
	return nil
}

func (r *InvoiceMemoryRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status model.InvoiceStatus) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	invoice, ok := r.invoices[id]
	if !ok {
		return ErrInvoiceNotFound
	}
	invoice.Status = status
	invoice.UpdatedAt = time.Now()
	if status == model.InvoiceCancelled {
		now := time.Now()
		invoice.CancelledAt = &now
	}
	return nil
}

func (r *InvoiceMemoryRepository) SetTimbradaData(ctx context.Context, id uuid.UUID, uuid, pdfURL, xmlURL, facturapiID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	invoice, ok := r.invoices[id]
	if !ok {
		return ErrInvoiceNotFound
	}
	invoice.UUID = uuid
	invoice.PDFURL = pdfURL
	invoice.XMLURL = xmlURL
	invoice.FacturapiID = facturapiID
	invoice.Status = model.InvoiceTimbrada
	invoice.UpdatedAt = time.Now()
	return nil
}

func (r *InvoiceMemoryRepository) ListByDateRange(ctx context.Context, branchID uuid.UUID, start, end time.Time) ([]model.Invoice, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	invoices := r.byBranch[branchID]
	result := make([]model.Invoice, 0)
	for _, inv := range invoices {
		if inv.CreatedAt.After(start) && inv.CreatedAt.Before(end) {
			result = append(result, *inv)
		}
	}
	return result, nil
}

func (r *InvoiceMemoryRepository) CountByBranch(ctx context.Context, branchID uuid.UUID) (int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return len(r.byBranch[branchID]), nil
}

func (r *InvoiceMemoryRepository) Seed(ctx context.Context) error {
	// No seed data for invoices by default - they are created on demand
	return nil
}

// TenantFiscalRepositoryInterface defines operations for tenant fiscal configuration
type TenantFiscalRepositoryInterface interface {
	Create(ctx context.Context, fiscal *model.TenantFiscal) error
	GetByTenant(ctx context.Context, tenantID uuid.UUID) (*model.TenantFiscal, error)
	Update(ctx context.Context, fiscal *model.TenantFiscal) error
	Seed(ctx context.Context) error
}

// TenantFiscalMemoryRepository is an in-memory implementation of TenantFiscalRepository
type TenantFiscalMemoryRepository struct {
	mu      sync.RWMutex
	fiscals map[uuid.UUID]*model.TenantFiscal
}

// NewTenantFiscalMemoryRepository creates a new in-memory tenant fiscal repository
func NewTenantFiscalMemoryRepository() *TenantFiscalMemoryRepository {
	return &TenantFiscalMemoryRepository{
		fiscals: make(map[uuid.UUID]*model.TenantFiscal),
	}
}

func (r *TenantFiscalMemoryRepository) Create(ctx context.Context, fiscal *model.TenantFiscal) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	fiscal.CreatedAt = now
	fiscal.UpdatedAt = now
	if fiscal.ID == uuid.Nil {
		fiscal.ID = uuid.New()
	}

	r.fiscals[fiscal.TenantID] = fiscal
	return nil
}

func (r *TenantFiscalMemoryRepository) GetByTenant(ctx context.Context, tenantID uuid.UUID) (*model.TenantFiscal, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if fiscal, ok := r.fiscals[tenantID]; ok {
		return fiscal, nil
	}
	return nil, ErrNotFound
}

func (r *TenantFiscalMemoryRepository) Update(ctx context.Context, fiscal *model.TenantFiscal) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	fiscal.UpdatedAt = time.Now()
	if _, ok := r.fiscals[fiscal.TenantID]; !ok {
		return ErrNotFound
	}
	r.fiscals[fiscal.TenantID] = fiscal
	return nil
}

func (r *TenantFiscalMemoryRepository) Seed(ctx context.Context) error {
	// No seed data - fiscal info is sensitive and must be configured by the tenant
	return nil
}

// InvoiceItemRepositoryInterface defines operations for invoice items
type InvoiceItemRepositoryInterface interface {
	Create(ctx context.Context, item *model.InvoiceItem) error
	GetByInvoice(ctx context.Context, invoiceID uuid.UUID) ([]model.InvoiceItem, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

// InvoiceItemMemoryRepository is an in-memory implementation of InvoiceItemRepository
type InvoiceItemMemoryRepository struct {
	mu    sync.RWMutex
	items map[uuid.UUID][]*model.InvoiceItem
}

// NewInvoiceItemMemoryRepository creates a new in-memory invoice item repository
func NewInvoiceItemMemoryRepository() *InvoiceItemMemoryRepository {
	return &InvoiceItemMemoryRepository{
		items: make(map[uuid.UUID][]*model.InvoiceItem),
	}
}

func (r *InvoiceItemMemoryRepository) Create(ctx context.Context, item *model.InvoiceItem) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if item.ID == uuid.Nil {
		item.ID = uuid.New()
	}
	r.items[item.InvoiceID] = append(r.items[item.InvoiceID], item)
	return nil
}

func (r *InvoiceItemMemoryRepository) GetByInvoice(ctx context.Context, invoiceID uuid.UUID) ([]model.InvoiceItem, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	items := r.items[invoiceID]
	result := make([]model.InvoiceItem, len(items))
	for i, item := range items {
		result[i] = *item
	}
	return result, nil
}

func (r *InvoiceItemMemoryRepository) Delete(ctx context.Context, id uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Find and remove the item
	for invoiceID, items := range r.items {
		for i, item := range items {
			if item.ID == id {
				// Remove item from slice
				r.items[invoiceID] = append(items[:i], items[i+1:]...)
				return nil
			}
		}
	}
	return ErrNotFound
}

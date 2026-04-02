package graph

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/greenpos/backend/internal/model"
)

// invoices is the resolver for the invoices query
func (r *queryResolver) Invoices(ctx context.Context, branchID uuid.UUID, status *model.InvoiceStatus, startDate *time.Time, endDate *time.Time) ([]model.Invoice, error) {
	if status != nil && startDate != nil && endDate != nil {
		return r.Services.Invoice.ListByDateRange(ctx, branchID, *startDate, *endDate)
	}
	if status != nil {
		all, err := r.Services.Invoice.GetInvoicesByBranch(ctx, branchID)
		if err != nil {
			return nil, err
		}
		var filtered []model.Invoice
		for _, inv := range all {
			if inv.Status == *status {
				filtered = append(filtered, inv)
			}
		}
		return filtered, nil
	}
	return r.Services.Invoice.GetInvoicesByBranch(ctx, branchID)
}

// invoice is the resolver for the invoice query
func (r *queryResolver) Invoice(ctx context.Context, id uuid.UUID) (*model.Invoice, error) {
	return r.Services.Invoice.GetInvoice(ctx, id)
}

// tenantFiscal is the resolver for the tenantFiscal query
func (r *queryResolver) TenantFiscal(ctx context.Context) (*model.TenantFiscal, error) {
	tenantID, err := GetTenantIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	return r.Services.TenantFiscal.GetTenantFiscal(ctx, tenantID)
}

// createInvoice is the resolver for the createInvoice mutation
func (r *mutationResolver) CreateInvoice(ctx context.Context, input CreateInvoiceInput) (*model.Invoice, error) {
	tenantID, err := GetTenantIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	branchID, err := GetBranchIDFromContext(ctx)
	if err != nil {
		// Try to get branch from the order
		order, orderErr := r.Services.Order.GetByID(ctx, input.OrderID)
		if orderErr != nil {
			return nil, orderErr
		}
		branchID = order.BranchID
	}

	// Build service input — enrich items with product data from catalog
	var svcItems []CreateInvoiceItemInput
	for _, item := range input.Items {
		product, err := r.Services.Product.GetByID(ctx, item.ProductID)
		if err != nil {
			// Product not found — use minimal data from input
			svcItems = append(svcItems, CreateInvoiceItemInput{
				ProductID:   item.ProductID,
				ProductName: "Producto",
				Quantity:    item.Quantity,
				UnitPrice:   0,
				TaxRate:     0.16,
				TaxAmount:   0,
				Total:       0,
				ClaveProdServ: item.ClaveProdServ,
				ClaveUnidad:   item.ClaveUnidad,
				Discount:      item.Discount,
			})
			continue
		}
		taxAmount := (product.Price * float64(item.Quantity)) * 0.16
		if item.Discount > 0 {
			taxAmount = (product.Price*float64(item.Quantity) - item.Discount) * 0.16
		}
		total := product.Price*float64(item.Quantity) - item.Discount + taxAmount
		svcItems = append(svcItems, CreateInvoiceItemInput{
			ProductID:     item.ProductID,
			ProductName:   product.Name,
			Quantity:      item.Quantity,
			UnitPrice:     product.Price,
			TaxRate:       0.16,
			TaxAmount:     taxAmount,
			Total:         total,
			ClaveProdServ: item.ClaveProdServ,
			ClaveUnidad:   item.ClaveUnidad,
			Discount:      item.Discount,
		})
	}

	svcInput := CreateInvoiceInput{
		TenantID:     tenantID,
		BranchID:    branchID,
		OrderID:     input.OrderID,
		Serie:       input.Serie,
		ReceptorRfc:       input.ReceptorRfc,
		ReceptorNombre:    input.ReceptorNombre,
		ReceptorUsoCfdi:   input.ReceptorUsoCfdi,
		ReceptorDomicilio: input.ReceptorDomicilio,
		FormaPago:   input.FormaPago,
		MetodoPago:  input.MetodoPago,
		Descuento:   input.Descuento,
		Items:       svcItems,
	}
	return r.Services.Invoice.CreateInvoice(ctx, svcInput)
}

// stampInvoice is the resolver for the stampInvoice mutation
func (r *mutationResolver) StampInvoice(ctx context.Context, id uuid.UUID) (*model.Invoice, error) {
	return r.Services.Invoice.StampInvoice(ctx, id)
}

// cancelInvoice is the resolver for the cancelInvoice mutation
func (r *mutationResolver) CancelInvoice(ctx context.Context, id uuid.UUID, motivo string) (*model.Invoice, error) {
	return r.Services.Invoice.CancelInvoice(ctx, id, motivo)
}

// updateTenantFiscal is the resolver for the updateTenantFiscal mutation
func (r *mutationResolver) UpdateTenantFiscal(ctx context.Context, input GQLUpdateTenantFiscalInput) (*model.TenantFiscal, error) {
	tenantID, err := GetTenantIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	svcInput := service.UpdateTenantFiscalInput{
		RFC:           input.RFC,
		RazonSocial:   input.RazonSocial,
		RegimenFiscal: input.RegimenFiscal,
		Calle:         input.Calle,
		Numero:        input.Numero,
		Colonia:       input.Colonia,
		CP:            input.CP,
		Ciudad:        input.Ciudad,
		Estado:        input.Estado,
		Pais:          input.Pais,
	}
	return r.Services.TenantFiscal.UpdateTenantFiscal(ctx, tenantID, svcInput)
}

// createTenantFiscal is the resolver for the createTenantFiscal mutation
func (r *mutationResolver) CreateTenantFiscal(ctx context.Context, input GQLUpdateTenantFiscalInput) (*model.TenantFiscal, error) {
	tenantID, err := GetTenantIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	svcInput := service.UpdateTenantFiscalInput{
		RFC:           input.RFC,
		RazonSocial:   input.RazonSocial,
		RegimenFiscal: input.RegimenFiscal,
		Calle:         input.Calle,
		Numero:        input.Numero,
		Colonia:       input.Colonia,
		CP:            input.CP,
		Ciudad:        input.Ciudad,
		Estado:        input.Estado,
		Pais:          input.Pais,
	}
	return r.Services.TenantFiscal.CreateTenantFiscal(ctx, tenantID, svcInput)
}

// Invoices field resolver on Branch
func (r *branchResolver) Invoices(ctx context.Context, obj *model.Branch) ([]model.Invoice, error) {
	return r.Services.Invoice.GetInvoicesByBranch(ctx, obj.ID)
}

// --- GraphQL input types (generated by gqlgen from schema) ---
// These are defined here manually since gqlgen generate hasn't run yet.
// When gqlgen runs, these will be in models_gen.go.

// CreateInvoiceInput matches the GraphQL schema input
type CreateInvoiceInput struct {
	OrderID      uuid.UUID
	ReceptorRfc       string
	ReceptorNombre    string
	ReceptorUsoCfdi   string
	ReceptorDomicilio *string
	FormaPago   string
	MetodoPago  string
	Serie       *string
	Descuento   *float64
	Items       []*CreateInvoiceItemInput
}

// CreateInvoiceItemInput matches the GraphQL schema input
type CreateInvoiceItemInput struct {
	ProductID     uuid.UUID
	ProductName   string
	Quantity      int
	UnitPrice     float64
	TaxRate       float64
	TaxAmount     float64
	Total         float64
	ClaveProdServ string
	ClaveUnidad   string
	Discount      float64
}

// UpdateTenantFiscalInput matches the GraphQL schema input
type UpdateTenantFiscalInput struct {
	RFC           *string
	RazonSocial   *string
	RegimenFiscal *string
	Calle         *string
	Numero        *string
	Colonia       *string
	CP            *string
	Ciudad        *string
	Estado        *string
	Pais          *string
}

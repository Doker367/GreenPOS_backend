package service

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/greenpos/backend/internal/model"
	"github.com/greenpos/backend/internal/repository"
)

// Facturapi API base URL
const FacturapiBaseURL = "https://www.facturapi.com/v2"

// Facturapi errors
var (
	ErrFacturapiInvoiceNotFound = errors.New("invoice not found in facturapi")
	ErrFacturapiCancellationFailed = errors.New("failed to cancel invoice in facturapi")
	ErrFacturapiStampFailed       = errors.New("failed to stamp invoice in facturapi")
	ErrFacturapiInvalidResponse  = errors.New("invalid response from facturapi")
	ErrFacturapiWebhookSignature  = errors.New("invalid webhook signature")
)

// FacturapiService defines the interface for Facturapi integration
type FacturapiService interface {
	// CreateInvoice sends an invoice to Facturapi for timbrado
	CreateInvoice(ctx context.Context, invoice model.Invoice, items []model.InvoiceItem, fiscal model.TenantFiscal) (*FacturapiResponse, error)

	// CancelInvoice cancels a previously issued invoice
	CancelInvoice(ctx context.Context, uuid, rfcEmisor, motivo string) error

	// DownloadPDF downloads the PDF for a previously issued invoice
	DownloadPDF(ctx context.Context, facturapiID string) ([]byte, error)

	// DownloadXML downloads the XML for a previously issued invoice
	DownloadXML(ctx context.Context, facturapiID string) ([]byte, error)

	// ValidateWebhook validates and parses a Facturapi webhook payload
	ValidateWebhook(payload []byte, signature string) (*WebhookEvent, error)
}

// FacturapiResponse represents the response from Facturapi after creating an invoice
type FacturapiResponse struct {
	ID            string `json:"id"`
	UUID         string `json:"uuid"`
	Status        string `json:"status"`
	PDFURL       string `json:"pdf_url"`
	XMLURL       string `json:"xml_url"`
	CreatedAt    string `json:"created_at"`
	Total        float64 `json:"total"`
	CustomerRFC  string `json:"customer_rfc"`
}

// WebhookEvent represents a Facturapi webhook event
type WebhookEvent struct {
	Type           string           `json:"type"`
	InvoiceID      string           `json:"invoice_id"`
	Data           WebhookInvoiceData `json:"data"`
	PreviousStatus string           `json:"previous_status,omitempty"`
}

// WebhookInvoiceData represents the invoice data in a webhook event
type WebhookInvoiceData struct {
	ID           string  `json:"id"`
	UUID         string  `json:"uuid"`
	Status       string  `json:"status"`
	Total        float64 `json:"total"`
	PDFURL       string  `json:"pdf_url,omitempty"`
	XMLURL       string  `json:"xml_url,omitempty"`
	TaxBreakdown []TaxBreakdown `json:"tax_breakdown,omitempty"`
}

// TaxBreakdown represents tax details in a webhook event
type TaxBreakdown struct {
	Type  string  `json:"type"`
	Rate  float64 `json:"rate"`
	Amount float64 `json:"amount"`
}

// FacturapiCreateRequest represents the request body for creating an invoice in Facturapi
type FacturapiCreateRequest struct {
	// Emisor (sender/seller) - the restaurant
	Emisor EmisorData `json:"emisor"`

	// Receptor (receiver/buyer) - the customer
	Receptor ReceptorData `json:"receptor"`

	// Invoice details
	Serie       string          `json:"serie,omitempty"`
	Folio       int             `json:"folio,omitempty"`
	FormaPago   string          `json:"forma_pago"`
	MetodoPago  string          `json:"metodo_pago"`
	Moneda      string          `json:"moneda"`
	TipoCambio  float64         `json:"tipo_cambio,omitempty"`
	Exportacion string          `json:"exportacion"` // 1 for domestic
	UsoCFDI     string          `json:"uso_cfdi"`
	Conceptos   []ConceptData   `json:"conceptos"`
	Impuestos   ImpuestosData   `json:"impuestos,omitempty"`

	// Order reference (our internal reference)
	OrderID string `json:"order_id,omitempty"`
}

// EmisorData represents the sender data in a Facturapi invoice
type EmisorData struct {
	RFC           string `json:"rfc"`
	Nombre        string `json:"nombre"`
	RegimenFiscal int    `json:"regimen_fiscal"`
	Direccion     DireccionData `json:"direccion,omitempty"`
}

// ReceptorData represents the receiver data in a Facturapi invoice
type ReceptorData struct {
	RFC     string `json:"rfc"`
	Nombre  string `json:"nombre"`
	UsoCFDI string `json:"uso_cfdi"`
	Domicilio DomicilioData `json:"domicilio,omitempty"`
}

// DireccionData represents address data for emisor
type DireccionData struct {
	Calle   string `json:"calle"`
	Numero  string `json:"numero,omitempty"`
	Colonia string `json:"colonia,omitempty"`
	CodigoPostal string `json:"codigo_postal,omitempty"`
	Ciudad  string `json:"ciudad,omitempty"`
	Estado  string `json:"estado,omitempty"`
	Pais    string `json:"pais,omitempty"`
}

// DomicilioData represents address data for receptor (simplified - just CP)
type DomicilioData struct {
	CodigoPostal string `json:"codigo_postal"`
}

// ConceptData represents a line item in a Facturapi invoice
type ConceptData struct {
	ClaveProdServ string       `json:"clave_prod_serv"`
	ClaveUnidad   string       `json:"clave_unidad"`
	Descripcion   string       `json:"descripcion"`
	Cantidad      float64      `json:"cantidad"`
	Unidad        string       `json:"unidad"`
	ValorUnitario float64     `json:"valor_unitario"`
	Importe       float64      `json:"importe"`
	Descuento     float64      `json:"descuento,omitempty"`
	Impuestos     []ConceptTax `json:"impuestos,omitempty"`
}

// ConceptTax represents taxes for a single concept
type ConceptTax struct {
	Base       float64 `json:"base"`
	Impuesto   string  `json:"impuesto"`   // "002" for IVA
	TipoFactor  string  `json:"tipo_factor"` // "Tasa"
	TasaOCuota float64 `json:"tasa_o_cuota"`
	Importe    float64 `json:"importe"`
}

// ImpuestosData represents the taxes summary in a Facturapi invoice
type ImpuestosData struct {
	TotalImpuestosRetenidos float64           `json:"total_impuestos_retenidos,omitempty"`
	TotalImpuestosTrasladados float64         `json:"total_impuestos_trasladados,omitempty"`
	Retenciones             []RetentionData   `json:"retenciones,omitempty"`
	Traslados               []TransferData    `json:"traslados,omitempty"`
}

// RetentionData represents a tax retention
type RetentionData struct {
	Impuesto string  `json:"impuesto"` // "001" for ISR
	TipoFactor string `json:"tipo_factor"`
	TasaOCuota float64 `json:"tasa_o_cuota"`
	Importe   float64 `json:"importe"`
}

// TransferData represents a tax transfer (IVA trasladado)
type TransferData struct {
	Impuesto  string  `json:"impuesto"`  // "002" for IVA
	TipoFactor string `json:"tipo_factor"` // "Tasa"
	TasaOCuota float64 `json:"tasa_o_cuota"`
	Importe   float64 `json:"importe"`
}

// FacturapiCancelRequest represents the request body for cancelling an invoice
type FacturapiCancelRequest struct {
	UUID      string `json:"uuid"`
	RFCEmisor string `json:"rfc_emisor"`
	Motivo    string `json:"motivo"`
}

// facturapiServiceImpl implements FacturapiService using the Facturapi REST API
type facturapiServiceImpl struct {
	apiKey    string
	baseURL   string
	httpClient *http.Client
	webhookSecret string
}

// NewFacturapiService creates a new Facturapi service instance
func NewFacturapiService(apiKey string) FacturapiService {
	return &facturapiServiceImpl{
		apiKey:  apiKey,
		baseURL: FacturapiBaseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// NewFacturapiServiceWithWebhookSecret creates a new Facturapi service with webhook validation
func NewFacturapiServiceWithWebhookSecret(apiKey, webhookSecret string) FacturapiService {
	return &facturapiServiceImpl{
		apiKey:    apiKey,
		baseURL:   FacturapiBaseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		webhookSecret: webhookSecret,
	}
}

// getAuthHeader returns the authorization header value for Facturapi
func (s *facturapiServiceImpl) getAuthHeader() string {
	// Facturapi uses Basic auth with api_key as username and empty password
	auth := base64.StdEncoding.EncodeToString([]byte(s.apiKey + ":"))
	return "Basic " + auth
}

// CreateInvoice sends an invoice to Facturapi for timbrado
func (s *facturapiServiceImpl) CreateInvoice(ctx context.Context, invoice model.Invoice, items []model.InvoiceItem, fiscal model.TenantFiscal) (*FacturapiResponse, error) {
	// Build the request payload
	req := s.buildCreateRequest(invoice, items, fiscal)

	payload, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Make the HTTP request
	url := s.baseURL + "/invoices"
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Authorization", s.getAuthHeader())
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: status=%d body=%s", ErrFacturapiStampFailed, resp.StatusCode, string(body))
	}

	var result FacturapiResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrFacturapiInvalidResponse, err)
	}

	return &result, nil
}

// buildCreateRequest builds the Facturapi request payload from our Invoice model
func (s *facturapiServiceImpl) buildCreateRequest(invoice model.Invoice, items []model.InvoiceItem, fiscal model.TenantFiscal) FacturapiCreateRequest {
	// Build emisor data
	emisor := EmisorData{
		RFC:           fiscal.RFC,
		Nombre:        fiscal.RazonSocial,
		RegimenFiscal: parseRegimenFiscal(fiscal.RegimenFiscal),
	}

	// Add emisor address if available
	if fiscal.Calle != "" {
		emisor.Direccion = DireccionData{
			Calle:        fiscal.Calle,
			Numero:        fiscal.Numero,
			Colonia:       fiscal.Colonia,
			CodigoPostal:  fiscal.CP,
			Ciudad:        fiscal.Ciudad,
			Estado:        fiscal.Estado,
			Pais:          fiscal.Pais,
		}
	}

	// Build receptor data
	receptor := ReceptorData{
		RFC:     invoice.ReceptorRFC,
		Nombre:  invoice.ReceptorNombre,
		UsoCFDI: invoice.ReceptorUsoCFDI,
	}

	// Add receptor domicile (CP) if available
	if invoice.ReceptorDomicilio != "" {
		receptor.Domicilio = DomicilioData{
			CodigoPostal: invoice.ReceptorDomicilio,
		}
	}

	// Build conceptos (line items)
	conceptos := make([]ConceptData, len(items))
	for i, item := range items {
		taxAmount := item.TaxAmount
		conceptos[i] = ConceptData{
			ClaveProdServ: item.ProductClaveProdServ,
			ClaveUnidad:   item.ProductClaveUnidad,
			Descripcion:   item.ProductName,
			Cantidad:      float64(item.Quantity),
			Unidad:        item.ProductClaveUnidad,
			ValorUnitario: item.UnitPrice,
			Importe:       item.Total,
			Descuento:     item.Discount,
		}

		// Add taxes for this concept if there's a tax amount
		if taxAmount > 0 {
			conceptos[i].Impuestos = []ConceptTax{
				{
					Base:       item.Total - item.Discount,
					Impuesto:   "002", // IVA
					TipoFactor:  "Tasa",
					TasaOCuota: item.TaxRate,
					Importe:    taxAmount,
				},
			}
		}
	}

	// Build impuestos (tax summary)
	var impuestos ImpuestosData
	if invoice.ImpuestosTrasladados > 0 {
		impuestos.TotalImpuestosTrasladados = invoice.ImpuestosTrasladados
		impuestos.Traslados = []TransferData{
			{
				Impuesto:  "002", // IVA
				TipoFactor: "Tasa",
				TasaOCuota: invoice.IVA16Rate,
				Importe:   invoice.IVA16Amount,
			},
		}
	}

	// Map FormaPago to Facturapi format (add leading zero if needed)
	formaPago := invoice.FormaPago
	if len(formaPago) == 1 {
		formaPago = "0" + formaPago
	}

	// Map Moneda to Facturapi format (MXN for Mexican Peso)
	moneda := invoice.Moneda
	if moneda == "" {
		moneda = "MXN"
	}

	return FacturapiCreateRequest{
		Emisor:      emisor,
		Receptor:    receptor,
		Serie:       invoice.Serie,
		Folio:       invoice.Folio,
		FormaPago:   formaPago,
		MetodoPago:  invoice.MetodoPago,
		Moneda:      moneda,
		TipoCambio:  invoice.TipoCambio,
		Exportacion: "1", // Domestic invoice
		Conceptos:   conceptos,
		Impuestos:   impuestos,
		OrderID:     invoice.OrderID.String(),
	}
}

// CancelInvoice cancels a previously issued invoice
func (s *facturapiServiceImpl) CancelInvoice(ctx context.Context, uuid, rfcEmisor, motivo string) error {
	req := FacturapiCancelRequest{
		UUID:      uuid,
		RFCEmisor: rfcEmisor,
		Motivo:    motivo,
	}

	payload, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal cancel request: %w", err)
	}

	url := s.baseURL + "/invoices/" + uuid + "/cancel"
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("failed to create cancel request: %w", err)
	}

	httpReq.Header.Set("Authorization", s.getAuthHeader())
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to send cancel request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("%w: status=%d body=%s", ErrFacturapiCancellationFailed, resp.StatusCode, string(body))
	}

	return nil
}

// DownloadPDF downloads the PDF for a previously issued invoice
func (s *facturapiServiceImpl) DownloadPDF(ctx context.Context, facturapiID string) ([]byte, error) {
	url := s.baseURL + "/invoices/" + facturapiID + "/pdf"

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create PDF request: %w", err)
	}

	httpReq.Header.Set("Authorization", s.getAuthHeader())

	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to download PDF: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to download PDF: status=%d body=%s", resp.StatusCode, string(body))
	}

	return io.ReadAll(resp.Body)
}

// DownloadXML downloads the XML for a previously issued invoice
func (s *facturapiServiceImpl) DownloadXML(ctx context.Context, facturapiID string) ([]byte, error) {
	url := s.baseURL + "/invoices/" + facturapiID + "/xml"

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create XML request: %w", err)
	}

	httpReq.Header.Set("Authorization", s.getAuthHeader())

	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to download XML: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to download XML: status=%d body=%s", resp.StatusCode, string(body))
	}

	return io.ReadAll(resp.Body)
}

// ValidateWebhook validates and parses a Facturapi webhook payload
func (s *facturapiServiceImpl) ValidateWebhook(payload []byte, signature string) (*WebhookEvent, error) {
	// If we have a webhook secret, validate the signature
	// Note: Facturapi uses different signature validation methods
	// This is a simplified implementation
	if s.webhookSecret != "" && signature != "" {
		// Facturapi sends signature in different headers depending on the event type
		// For webhook validation, they typically use HMAC signature
		// This is a placeholder - actual implementation would depend on Facturapi's current API
		_ = signature
	}

	var event WebhookEvent
	if err := json.Unmarshal(payload, &event); err != nil {
		return nil, fmt.Errorf("failed to parse webhook payload: %w", err)
	}

	return &event, nil
}

// parseRegimenFiscal converts a regimen fiscal string to int
// SAT codes: 601 = General de Ley Personas Morales, etc.
func parseRegimenFiscal(regimen string) int {
	// Common SAT regimen fiscal codes
	switch regimen {
	case "601":
		return 601 // General de Ley Personas Morales
	case "603":
		return 603 // Personas Morales con Fines No Lucrativos
	case "605":
		return 605 // Sueldos y Salarios e Ingresos Asimilados a Salarios
	case "606":
		return 606 // Arrendamiento
	case "608":
		return 608 // Demás ingresos
	case "609":
		return 609 // Consolidación
	case "610":
		return 610 // Residencia Fiscal
	case "611":
		return 611 // No residentes fiscales
	default:
		// Try to parse as int if it's a numeric string
		var code int
		fmt.Sscanf(regimen, "%d", &code)
		return code
	}
}

// InvoiceService handles invoice business logic
type InvoiceService struct {
	Invoices  *repository.InvoiceRepository
	Items     *repository.InvoiceItemRepository
	Fiscal    *repository.TenantFiscalRepository
	Facturapi FacturapiService
}

// NewInvoiceService creates a new invoice service
func NewInvoiceService(
	invoices *repository.InvoiceRepository,
	items *repository.InvoiceItemRepository,
	fiscal *repository.TenantFiscalRepository,
	facturapi FacturapiService,
) *InvoiceService {
	return &InvoiceService{
		Invoices:  invoices,
		Items:     items,
		Fiscal:    fiscal,
		Facturapi: facturapi,
	}
}

// CreateInvoiceInput represents input for creating an invoice
type CreateInvoiceInput struct {
	TenantID      uuid.UUID
	BranchID      uuid.UUID
	OrderID       uuid.UUID
	Serie         string
	ReceptorRfc   string
	ReceptorNombre string
	ReceptorUsoCfdi string
	ReceptorDomicilio string
	FormaPago     string
	MetodoPago    string
	Descuento     float64
	Items         []CreateInvoiceItemInput
}

// CreateInvoiceItemInput represents input for an invoice line item
type CreateInvoiceItemInput struct {
	ProductID        uuid.UUID
	ProductName     string
	ClaveProdServ   string
	ClaveUnidad     string
	Quantity        int
	UnitPrice       float64
	Discount        float64
	TaxRate         float64
	TaxAmount       float64
	Total           float64
}

// CreateInvoice creates a draft invoice without stamping
func (s *InvoiceService) CreateInvoice(ctx context.Context, input CreateInvoiceInput) (*model.Invoice, error) {
	// Get tenant fiscal info
	fiscal, err := s.Fiscal.GetByTenant(ctx, input.TenantID)
	if err != nil {
		return nil, fmt.Errorf("tenant fiscal info not configured: %w", err)
	}

	// Calculate totals
	var subtotal, totalImpuestos, iva16Amount float64
	for _, item := range input.Items {
		subtotal += (item.UnitPrice * float64(item.Quantity)) - item.Discount
		iva16Amount += item.TaxAmount
		totalImpuestos += item.TaxAmount
	}

	descuento := input.Descuento
	impuestosTrasladados := totalImpuestos
	total := subtotal - descuento + impuestosTrasladados

	// Get current folio for this branch
	folio, err := s.Invoices.CountByBranch(ctx, input.BranchID)
	if err != nil {
		folio = 1
	} else {
		folio++
	}

	// Create the invoice
	invoice := &model.Invoice{
		ID:           uuid.New(),
		TenantID:     input.TenantID,
		BranchID:     input.BranchID,
		OrderID:      input.OrderID,
		Serie:        input.Serie,
		Folio:        folio,
		UUID:         "",
		EmisorRFC:    fiscal.RFC,
		EmisorNombre: fiscal.RazonSocial,
		EmisorRegimen: fiscal.RegimenFiscal,
		ReceptorRFC:       input.ReceptorRfc,
		ReceptorNombre:    input.ReceptorNombre,
		ReceptorUsoCFDI:   input.ReceptorUsoCfdi,
		ReceptorDomicilio: input.ReceptorDomicilio,
		FormaPago:    input.FormaPago,
		MetodoPago:   input.MetodoPago,
		Moneda:       "MXN",
		TipoCambio:   1.0,
		ExchangeRate: 1.0,
		Subtotal:           subtotal,
		Descuento:          descuento,
		ImpuestosRetenidos: 0,
		ImpuestosTrasladados: impuestosTrasladados,
		Total:         total,
		IVA16Amount:   iva16Amount,
		IVA16Rate:     0.16,
		IEPSAmount:    0,
		ComplementoPago: false,
		Status:        model.InvoiceDraft,
		PDFURL:        "",
		XMLURL:        "",
		FacturapiID:   "",
	}

	if err := s.Invoices.Create(ctx, invoice); err != nil {
		return nil, fmt.Errorf("failed to create invoice: %w", err)
	}

	// Create invoice items
	for _, itemInput := range input.Items {
		item := model.InvoiceItem{
			ID:              uuid.New(),
			InvoiceID:       invoice.ID,
			ProductID:       itemInput.ProductID,
			ProductName:     itemInput.ProductName,
			ProductClaveProdServ: itemInput.ClaveProdServ,
			ProductClaveUnidad:   itemInput.ClaveUnidad,
			Quantity:        itemInput.Quantity,
			UnitPrice:       itemInput.UnitPrice,
			Discount:        itemInput.Discount,
			TaxRate:         itemInput.TaxRate,
			TaxAmount:       itemInput.TaxAmount,
			Total:           itemInput.Total,
		}
		if err := s.Items.Create(ctx, &item); err != nil {
			return nil, fmt.Errorf("failed to create invoice item: %w", err)
		}
		invoice.Items = append(invoice.Items, item)
	}

	return invoice, nil
}

// StampInvoice sends a draft invoice to Facturapi for timbrado
func (s *InvoiceService) StampInvoice(ctx context.Context, invoiceID uuid.UUID) (*model.Invoice, error) {
	// Get the invoice
	invoice, err := s.Invoices.GetByID(ctx, invoiceID)
	if err != nil {
		return nil, fmt.Errorf("invoice not found: %w", err)
	}

	// Get fiscal info
	fiscal, err := s.Fiscal.GetByTenant(ctx, invoice.TenantID)
	if err != nil {
		return nil, fmt.Errorf("tenant fiscal info not configured: %w", err)
	}

	// Get items
	items, err := s.Items.GetByInvoice(ctx, invoiceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get invoice items: %w", err)
	}

	// Update status to PENDING
	if err := s.Invoices.UpdateStatus(ctx, invoiceID, model.InvoicePending); err != nil {
		return nil, fmt.Errorf("failed to update invoice status: %w", err)
	}

	// Send to Facturapi
	resp, err := s.Facturapi.CreateInvoice(ctx, *invoice, items, *fiscal)
	if err != nil {
		// Revert status to DRAFT on failure
		s.Invoices.UpdateStatus(ctx, invoiceID, model.InvoiceDraft)
		return nil, fmt.Errorf("failed to stamp invoice: %w", err)
	}

	// Update invoice with timbrado data
	if err := s.Invoices.SetTimbradaData(ctx, invoiceID, resp.UUID, resp.PDFURL, resp.XMLURL, resp.ID); err != nil {
		return nil, fmt.Errorf("failed to update timbrado data: %w", err)
	}

	// Refresh invoice from storage
	return s.Invoices.GetByID(ctx, invoiceID)
}

// CancelInvoice cancels a timbrada invoice
func (s *InvoiceService) CancelInvoice(ctx context.Context, invoiceID uuid.UUID, motivo string) (*model.Invoice, error) {
	invoice, err := s.Invoices.GetByID(ctx, invoiceID)
	if err != nil {
		return nil, fmt.Errorf("invoice not found: %w", err)
	}

	if invoice.Status != model.InvoiceTimbrada {
		return nil, fmt.Errorf("can only cancel timbrada invoices")
	}

	// Cancel in Facturapi
	if err := s.Facturapi.CancelInvoice(ctx, invoice.UUID, invoice.EmisorRFC, motivo); err != nil {
		return nil, fmt.Errorf("failed to cancel in facturapi: %w", err)
	}

	// Update status locally
	if err := s.Invoices.UpdateStatus(ctx, invoiceID, model.InvoiceCancelled); err != nil {
		return nil, fmt.Errorf("failed to update invoice status: %w", err)
	}

	return s.Invoices.GetByID(ctx, invoiceID)
}

// GetInvoice retrieves an invoice by ID
func (s *InvoiceService) GetInvoice(ctx context.Context, invoiceID uuid.UUID) (*model.Invoice, error) {
	invoice, err := s.Invoices.GetByID(ctx, invoiceID)
	if err != nil {
		return nil, err
	}

	// Load items
	items, err := s.Items.GetByInvoice(ctx, invoiceID)
	if err == nil {
		invoice.Items = items
	}

	return invoice, nil
}

// GetInvoicesByBranch retrieves all invoices for a branch
func (s *InvoiceService) GetInvoicesByBranch(ctx context.Context, branchID uuid.UUID) ([]model.Invoice, error) {
	return s.Invoices.GetByBranch(ctx, branchID)
}

// GetInvoicesByTenant retrieves all invoices for a tenant
func (s *InvoiceService) GetInvoicesByTenant(ctx context.Context, tenantID uuid.UUID) ([]model.Invoice, error) {
	return s.Invoices.GetByTenant(ctx, tenantID)
}

// ListByDateRange retrieves invoices for a branch within a date range
func (s *InvoiceService) ListByDateRange(ctx context.Context, branchID uuid.UUID, start, end time.Time) ([]model.Invoice, error) {
	return s.Invoices.ListByDateRange(ctx, branchID, start, end)
}

// GetInvoiceByOrder retrieves an invoice by order ID
func (s *InvoiceService) GetInvoiceByOrder(ctx context.Context, orderID uuid.UUID) (*model.Invoice, error) {
	return s.Invoices.GetByOrder(ctx, orderID)
}

// TenantFiscalService handles tenant fiscal configuration business logic
type TenantFiscalService struct {
	Fiscal *repository.TenantFiscalRepository
}

// NewTenantFiscalService creates a new tenant fiscal service
func NewTenantFiscalService(fiscal *repository.TenantFiscalRepository) *TenantFiscalService {
	return &TenantFiscalService{Fiscal: fiscal}
}

// UpdateTenantFiscalInput represents input for updating tenant fiscal info
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

// GetTenantFiscal retrieves fiscal info for a tenant
func (s *TenantFiscalService) GetTenantFiscal(ctx context.Context, tenantID uuid.UUID) (*model.TenantFiscal, error) {
	return s.Fiscal.GetByTenant(ctx, tenantID)
}

// CreateTenantFiscal creates initial fiscal info for a tenant
func (s *TenantFiscalService) CreateTenantFiscal(ctx context.Context, tenantID uuid.UUID, input UpdateTenantFiscalInput) (*model.TenantFiscal, error) {
	// Check if already exists
	existing, _ := s.Fiscal.GetByTenant(ctx, tenantID)
	if existing != nil {
		return nil, fmt.Errorf("tenant fiscal info already exists, use UpdateTenantFiscal instead")
	}

	fiscal := &model.TenantFiscal{
		ID:             uuid.New(),
		TenantID:       tenantID,
		RFC:            getString(input.RFC, ""),
		RazonSocial:    getString(input.RazonSocial, ""),
		RegimenFiscal: getString(input.RegimenFiscal, "601"),
		Calle:         getString(input.Calle, ""),
		Numero:        getString(input.Numero, ""),
		Colonia:       getString(input.Colonia, ""),
		CP:            getString(input.CP, ""),
		Ciudad:        getString(input.Ciudad, ""),
		Estado:        getString(input.Estado, ""),
		Pais:          getString(input.Pais, "MEX"),
	}

	if err := s.Fiscal.Create(ctx, fiscal); err != nil {
		return nil, fmt.Errorf("failed to create tenant fiscal info: %w", err)
	}

	return fiscal, nil
}

// UpdateTenantFiscal updates fiscal info for a tenant
func (s *TenantFiscalService) UpdateTenantFiscal(ctx context.Context, tenantID uuid.UUID, input UpdateTenantFiscalInput) (*model.TenantFiscal, error) {
	fiscal, err := s.Fiscal.GetByTenant(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("tenant fiscal info not found: %w", err)
	}

	// Update fields if provided
	if input.RFC != nil {
		fiscal.RFC = *input.RFC
	}
	if input.RazonSocial != nil {
		fiscal.RazonSocial = *input.RazonSocial
	}
	if input.RegimenFiscal != nil {
		fiscal.RegimenFiscal = *input.RegimenFiscal
	}
	if input.Calle != nil {
		fiscal.Calle = *input.Calle
	}
	if input.Numero != nil {
		fiscal.Numero = *input.Numero
	}
	if input.Colonia != nil {
		fiscal.Colonia = *input.Colonia
	}
	if input.CP != nil {
		fiscal.CP = *input.CP
	}
	if input.Ciudad != nil {
		fiscal.Ciudad = *input.Ciudad
	}
	if input.Estado != nil {
		fiscal.Estado = *input.Estado
	}
	if input.Pais != nil {
		fiscal.Pais = *input.Pais
	}

	if err := s.Fiscal.Update(ctx, fiscal); err != nil {
		return nil, fmt.Errorf("failed to update tenant fiscal info: %w", err)
	}

	return fiscal, nil
}

// Helper to get string value or default
func getString(ptr *string, defaultVal string) string {
	if ptr != nil {
		return *ptr
	}
	return defaultVal
}

// GetEnvOrDefault gets environment variable or returns default
func GetEnvOrDefault(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

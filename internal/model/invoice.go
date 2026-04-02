package model

import (
	"time"

	"github.com/google/uuid"
)

// InvoiceStatus represents the status of an invoice in its lifecycle
type InvoiceStatus string

const (
	InvoiceDraft     InvoiceStatus = "DRAFT"      // guardada sin timbrar
	InvoicePending  InvoiceStatus = "PENDING"    // enviada a timbrar
	InvoiceTimbrada InvoiceStatus = "TIMBRADA"    // timbrada exitosamente
	InvoiceCancelled InvoiceStatus = "CANCELLED"  // cancelada
)

// Invoice represents a CFDI 4.0 invoice (factura electrónica)
type Invoice struct {
	ID        uuid.UUID `json:"id"`
	TenantID  uuid.UUID `json:"tenantId"`
	BranchID  uuid.UUID `json:"branchId"`
	OrderID   uuid.UUID `json:"orderId"` // Order relacionada

	// Serie y Folio (para numeración consecutica)
	Serie string `json:"serie"` // ej: "A", "FAC"
	Folio int    `json:"folio"` // número consecutico

	// UUID del SAT (asignado por el PAC)
	UUID string `json:"uuid"` // UUID v4 del timbre fiscal

	// Datos del emisor (el restaurant/tenant)
	EmisorRFC     string `json:"emisorRfc"`
	EmisorNombre  string `json:"emisorNombre"`
	EmisorRegimen string `json:"emisorRegimen"` // RÉGIMEN FISCAL SAT: 601, 603, 605...

	// Datos del receptor (el cliente)
	ReceptorRFC       string `json:"receptorRfc"`
	ReceptorNombre    string `json:"receptorNombre"`
	ReceptorUsoCFDI   string `json:"receptorUsoCfdi"`   // USO_CFDI: G01, G02, G03...
	ReceptorDomicilio string `json:"receptorDomicilio"` // CP del receptor

	// Datos de la transacción
	FormaPago   string  `json:"formaPago"`   // 01, 03, 04... (Efectivo, Transferencia, etc.)
	MetodoPago  string  `json:"metodoPago"`  // PUE (una exhibición) o PPD (parcialidades)
	Moneda      string  `json:"moneda"`     // MXN por default
	TipoCambio  float64 `json:"tipoCambio"` // 1.0 para MXN
	ExchangeRate float64 `json:"exchangeRate"`

	// Subtotales y totales
	Subtotal           float64 `json:"subtotal"`
	Descuento          float64 `json:"descuento"`
	ImpuestosRetenidos float64 `json:"impuestosRetenidos"`
	ImpuestosTrasladados float64 `json:"impuestosTrasladados"`
	Total              float64 `json:"total"`

	// Impuestos detallados
	IVA16Amount float64 `json:"iva16Amount"`
	IVA16Rate   float64 `json:"iva16Rate"` // 0.16
	IEPSAmount  float64 `json:"iepsAmount"`

	// Complemento de pago (si aplica)
	ComplementoPago bool `json:"complementoPago"`

	// Estado y metadata
	Status      InvoiceStatus `json:"status"` // DRAFT, PENDING, TIMBRADA, CANCELLED
	PDFURL      string        `json:"pdfUrl"`  // URL del PDF generado
	XMLURL      string        `json:"xmlUrl"`  // URL del XML timbrado
	FacturapiID string        `json:"facturapiId"` // ID en Facturapi para webhooks

	CreatedAt  time.Time  `json:"createdAt"`
	UpdatedAt  time.Time  `json:"updatedAt"`
	CancelledAt *time.Time `json:"cancelledAt"`

	// Items de la factura (no se persiste en este modelo, se maneja separado)
	Items []InvoiceItem `json:"items,omitempty"`
}

// InvoiceItem represents a line item in an invoice
type InvoiceItem struct {
	ID              uuid.UUID `json:"id"`
	InvoiceID       uuid.UUID `json:"invoiceId"`
	ProductID       uuid.UUID `json:"productId"`
	ProductName     string    `json:"productName"`
	ProductClaveProdServ string `json:"productClaveProdServ"` // SAT product classification
	ProductClaveUnidad string   `json:"productClaveUnidad"`    // SAT unit classification
	Quantity        int       `json:"quantity"`
	UnitPrice       float64   `json:"unitPrice"`
	Discount        float64   `json:"discount"`
	TaxRate         float64   `json:"taxRate"`
	TaxAmount       float64   `json:"taxAmount"`
	Total           float64   `json:"total"`
}

// TenantFiscal represents the fiscal configuration for a tenant (RFC, certificates, etc.)
type TenantFiscal struct {
	ID uuid.UUID `json:"id"`
	// TenantID is the tenant this fiscal info belongs to
	TenantID uuid.UUID `json:"tenantId"`

	// RFC (Registro Federal de Contribuyentes) - 12-13 characters
	RFC string `json:"rfc"`

	// Razón Social (legal business name)
	RazonSocial string `json:"razonSocial"`

	// Régimen Fiscal (SAT tax regime code): 601, 603, 605, etc.
	RegimenFiscal string `json:"regimenFiscal"`

	// Address fields
	Calle    string `json:"calle"`
	Numero   string `json:"numero"`
	Colonia  string `json:"colonia"`
	CP       string `json:"cp"` // Código Postal
	Ciudad   string `json:"ciudad"`
	Estado   string `json:"estado"`
	Pais     string `json:"pais"`

	// Certificate data (stored separately for security - not in this model for direct access)
	// CertificateFileName string `json:"certificateFileName"`

	// CreatedAt and UpdatedAt
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

package graph

import (
	"time"

	"github.com/google/uuid"
	"github.com/greenpos/backend/internal/model"
)

// Tenant is a restaurant group or company
type Tenant struct {
	ID        uuid.UUID `json:"id"`
	Name      string   `json:"name"`
	Slug      string   `json:"slug"`
	Logo      *string  `json:"logo"`
	Plan      string   `json:"plan"`
	IsActive  bool     `json:"isActive"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// Branch is a restaurant location
type Branch struct {
	ID        uuid.UUID `json:"id"`
	TenantID  uuid.UUID `json:"tenantId"`
	Name      string   `json:"name"`
	Address   *string  `json:"address"`
	Phone     *string  `json:"phone"`
	IsActive  bool     `json:"isActive"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// User is a staff member
type User struct {
	ID        uuid.UUID       `json:"id"`
	TenantID  uuid.UUID       `json:"tenantId"`
	BranchID  *uuid.UUID     `json:"branchId"`
	Name      string         `json:"name"`
	Email     string         `json:"email"`
	Password  string         `json:"-"`
	Role      model.UserRole `json:"role"`
	IsActive  bool           `json:"isActive"`
	CreatedAt time.Time     `json:"createdAt"`
	UpdatedAt time.Time     `json:"updatedAt"`
}

// Category for menu items
type Category struct {
	ID        uuid.UUID `json:"id"`
	BranchID  uuid.UUID `json:"branchId"`
	Name      string   `json:"name"`
	SortOrder int      `json:"sortOrder"`
	IsActive  bool     `json:"isActive"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// Product is a menu item
type Product struct {
	ID          uuid.UUID  `json:"id"`
	BranchID   uuid.UUID  `json:"branchId"`
	CategoryID *uuid.UUID `json:"categoryId"`
	Name       string     `json:"name"`
	Description *string   `json:"description"`
	Price      float64    `json:"price"`
	Image      *string    `json:"image"`
	IsAvailable bool      `json:"isAvailable"`
	CreatedAt  time.Time  `json:"createdAt"`
	UpdatedAt  time.Time  `json:"updatedAt"`
}

// RestaurantTable is a table in the restaurant
type RestaurantTable struct {
	ID        uuid.UUID         `json:"id"`
	BranchID  uuid.UUID        `json:"branchId"`
	Name      string           `json:"name"`
	Capacity  int              `json:"capacity"`
	Status    model.TableStatus `json:"status"`
	CreatedAt time.Time       `json:"createdAt"`
	UpdatedAt time.Time       `json:"updatedAt"`
	QRCode    string           `json:"qrCode"`
}

// Table is an alias for RestaurantTable (used by resolvers)
type Table = RestaurantTable

// Order is a customer order
type Order struct {
	ID           uuid.UUID         `json:"id"`
	BranchID    uuid.UUID        `json:"branchId"`
	TableID     *uuid.UUID       `json:"tableId"`
	UserID      *uuid.UUID       `json:"userId"`
	CustomerName *string         `json:"customerName"`
	Status      model.OrderStatus `json:"status"`
	Subtotal    float64         `json:"subtotal"`
	Tax         float64         `json:"tax"`
	Total       float64         `json:"total"`
	Notes       *string         `json:"notes"`
	CreatedAt   time.Time       `json:"createdAt"`
	UpdatedAt   time.Time       `json:"updatedAt"`
}

// OrderItem is a line item in an order
type OrderItem struct {
	ID         uuid.UUID `json:"id"`
	OrderID   uuid.UUID `json:"orderId"`
	ProductID  uuid.UUID `json:"productId"`
	Quantity   int       `json:"quantity"`
	UnitPrice  float64   `json:"unitPrice"`
	Total      float64   `json:"total"`
	Notes      *string   `json:"notes"`
	CreatedAt  time.Time `json:"createdAt"`
}

// Reservation for table bookings
type Reservation struct {
	ID            uuid.UUID               `json:"id"`
	BranchID     uuid.UUID              `json:"branchId"`
	TableID      *uuid.UUID            `json:"tableId"`
	CustomerName  string               `json:"customerName"`
	CustomerPhone *string             `json:"customerPhone"`
	DateTime     time.Time            `json:"dateTime"`
	Guests       int                  `json:"guests"`
	Status       model.ReservationStatus `json:"status"`
	Notes        *string              `json:"notes"`
	CreatedAt    time.Time            `json:"createdAt"`
	UpdatedAt    time.Time            `json:"updatedAt"`
}

// AuthPayload returned on login
type AuthPayload struct {
	Token        string `json:"token"`
	RefreshToken string `json:"refreshToken"`
	User         *User  `json:"user"`
}

// RevenueData for analytics
type RevenueData struct {
	Date     time.Time `json:"date"`
	Revenue  float64  `json:"revenue"`
	Orders   int      `json:"orders"`
	AvgOrder float64  `json:"avgOrder"`
}

// TopProduct for analytics
type TopProduct struct {
	ProductID    string  `json:"productId"`
	ProductName string    `json:"productName"`
	QuantitySold int    `json:"quantitySold"`
	Revenue     float64   `json:"revenue"`
}

// CategoryOrder for reordering
type CategoryOrder struct {
	ID       uuid.UUID `json:"id"`
	SortOrder int      `json:"sortOrder"`
}

// Invoice for CFDI 4.0
type Invoice struct {
	ID             uuid.UUID            `json:"id"`
	BranchID      uuid.UUID           `json:"branchId"`
	OrderID       uuid.UUID           `json:"orderId"`
	Serie         string             `json:"serie"`
	Folio         int                `json:"folio"`
	UUID_         *string            `json:"uuid"`
	Status        model.InvoiceStatus `json:"status"`
	EmisorRFC     string             `json:"emisorRfc"`
	EmisorNombre  string             `json:"emisorNombre"`
	EmisorRegimen string             `json:"emisorRegimen"`
	ReceptorRFC   string             `json:"receptorRfc"`
	ReceptorNombre string             `json:"receptorNombre"`
	ReceptorUsoCFDI string          `json:"receptorUsoCfdi"`
	ReceptorDomicilio *string        `json:"receptorDomicilio"`
	FormaPago     string             `json:"formaPago"`
	MetodoPago    string             `json:"metodoPago"`
	Subtotal      float64            `json:"subtotal"`
	Descuento     float64            `json:"descuento"`
	ImpuestosTrasladados float64     `json:"impuestosTrasladados"`
	Total         float64            `json:"total"`
	IVA16Amount   float64            `json:"iva16Amount"`
	IEPSAmount   *float64           `json:"iepsAmount"`
	PDFURL        *string            `json:"pdfUrl"`
	XMLURL        *string            `json:"xmlUrl"`
	CreatedAt    time.Time          `json:"createdAt"`
}

// InvoiceItem is a line item in an invoice
type InvoiceItem struct {
	ID             uuid.UUID `json:"id"`
	ProductName    string    `json:"productName"`
	ClaveProdServ  string    `json:"claveProdServ"`
	ClaveUnidad    string    `json:"claveUnidad"`
	Quantity       int       `json:"quantity"`
	UnitPrice      float64   `json:"unitPrice"`
	Discount       float64   `json:"discount"`
	TaxRate        float64   `json:"taxRate"`
	TaxAmount      float64   `json:"taxAmount"`
	Total          float64   `json:"total"`
}

// TenantFiscal holds fiscal data for a tenant
type TenantFiscal struct {
	ID            uuid.UUID `json:"id"`
	RFC           string    `json:"rfc"`
	RazonSocial   string    `json:"razonSocial"`
	RegimenFiscal string   `json:"regimenFiscal"`
	Calle         *string   `json:"calle"`
	Numero        *string   `json:"numero"`
	Colonia       *string   `json:"colonia"`
	CP            *string   `json:"cp"`
	Ciudad        *string   `json:"ciudad"`
	Estado        *string   `json:"estado"`
	Pais          *string   `json:"pais"`
}

// DashboardMetrics for analytics
type DashboardMetrics struct {
	TotalRevenue     float64       `json:"totalRevenue"`
	TodayRevenue     float64       `json:"todayRevenue"`
	WeekRevenue      float64       `json:"weekRevenue"`
	MonthRevenue     float64       `json:"monthRevenue"`
	TotalOrders      int           `json:"totalOrders"`
	OrdersToday      int           `json:"ordersToday"`
	OrdersThisWeek   int           `json:"ordersThisWeek"`
	OrdersThisMonth  int           `json:"ordersThisMonth"`
	AverageTicket    float64       `json:"averageTicket"`
	ActiveOrders     int           `json:"activeOrders"`
	TotalTables      int           `json:"totalTables"`
	AvailableTables  int           `json:"availableTables"`
	TopProducts      []TopProduct  `json:"topProducts"`
	DailySales       []DailySales  `json:"dailySales"`
}
// DailySales for daily revenue
type DailySales struct {
	Date     string  `json:"date"`
	Revenue  float64 `json:"revenue"`
	Orders   int     `json:"orders"`
}

// StatusCount for order status breakdown
type StatusCount struct {
	Status model.OrderStatus `json:"status"`
	Count  int              `json:"count"`
}

// RevenueReport for revenue analytics
type RevenueReport struct {
	TotalRevenue    float64       `json:"totalRevenue"`
	TotalOrders     int            `json:"totalOrders"`
	AverageTicket   float64        `json:"averageTicket"`
	ByStatus      []StatusCount  `json:"byStatus"`
	ByPaymentMethod []PaymentMethodRevenue `json:"byPaymentMethod"`
	DailySales    []DailySales   `json:"dailySales"`
}

// PaymentMethodRevenue for payment method breakdown
type PaymentMethodRevenue struct {
	Method  string  `json:"method"`
	Revenue float64 `json:"revenue"`
	Count   int     `json:"count"`
}

// ============================================================
// INPUTS (GraphQL input types)
// ============================================================

type CreateTenantInput struct {
	Name string `json:"name"`
	Slug string `json:"slug"`
	Logo *string `json:"logo"`
}

type UpdateTenantInput struct {
	Name *string `json:"name"`
	Logo *string `json:"logo"`
}

type CreateBranchInput struct {
	TenantID uuid.UUID `json:"tenantId"`
	Name     string   `json:"name"`
	Address  *string  `json:"address"`
	Phone    *string  `json:"phone"`
}

type UpdateBranchInput struct {
	Name    *string `json:"name"`
	Address *string `json:"address"`
	Phone   *string `json:"phone"`
}

type CreateUserInput struct {
	TenantID uuid.UUID `json:"tenantId"`
	BranchID *uuid.UUID `json:"branchId"`
	Name     string     `json:"name"`
	Email    string     `json:"email"`
	Password string     `json:"password"`
	Role     model.UserRole `json:"role"`
}

type UpdateUserInput struct {
	Name     *string          `json:"name"`
	Email    *string          `json:"email"`
	Password *string          `json:"password"`
	Role     *model.UserRole `json:"role"`
}

type CreateCategoryInput struct {
	BranchID  uuid.UUID `json:"branchId"`
	Name      string   `json:"name"`
	SortOrder int      `json:"sortOrder"`
}

type UpdateCategoryInput struct {
	Name      *string `json:"name"`
	SortOrder *int    `json:"sortOrder"`
	IsActive  *bool   `json:"isActive"`
}

type CreateProductInput struct {
	BranchID   uuid.UUID  `json:"branchId"`
	CategoryID *uuid.UUID `json:"categoryId"`
	Name       string     `json:"name"`
	Description *string   `json:"description"`
	Price      float64    `json:"price"`
	Image      *string    `json:"image"`
}

type UpdateProductInput struct {
	CategoryID  *uuid.UUID `json:"categoryId"`
	Name        *string   `json:"name"`
	Description *string   `json:"description"`
	Price       *float64  `json:"price"`
	Image       *string   `json:"image"`
	IsAvailable *bool     `json:"isAvailable"`
}

type CreateTableInput struct {
	BranchID uuid.UUID `json:"branchId"`
	Name     string   `json:"name"`
	Capacity int      `json:"capacity"`
}

type UpdateTableInput struct {
	Name     *string            `json:"name"`
	Capacity *int              `json:"capacity"`
	Status   *model.TableStatus `json:"status"`
}

type CreateOrderInput struct {
	BranchID     uuid.UUID `json:"branchId"`
	TableID     *uuid.UUID `json:"tableId"`
	UserID      *uuid.UUID `json:"userId"`
	CustomerName *string   `json:"customerName"`
	Notes       *string   `json:"notes"`
	Items       []CreateOrderItemInput `json:"items"`
}

type CreateOrderItemInput struct {
	ProductID uuid.UUID `json:"productId"`
	Quantity  int       `json:"quantity"`
	Notes     *string   `json:"notes"`
}

type AddOrderItemInput struct {
	ProductID uuid.UUID `json:"productId"`
	Quantity  int       `json:"quantity"`
	Notes     *string   `json:"notes"`
}

type CreateReservationInput struct {
	BranchID      uuid.UUID `json:"branchId"`
	TableID       *uuid.UUID `json:"tableId"`
	CustomerName   string    `json:"customerName"`
	CustomerPhone  *string   `json:"customerPhone"`
	DateTime      time.Time `json:"dateTime"`
	Guests        int       `json:"guests"`
	Notes         *string   `json:"notes"`
}

// Invoice inputs
// GQLCreateInvoiceInput is the GraphQL-layer input type for creating an invoice
type GQLCreateInvoiceInput struct {
	OrderID      uuid.UUID  `json:"orderId"`
	ReceptorRfc  string     `json:"receptorRfc"`
	ReceptorNombre string   `json:"receptorNombre"`
	ReceptorUsoCfdi string  `json:"receptorUsoCfdi"`
	ReceptorDomicilio *string `json:"receptorDomicilio"`
	FormaPago    string     `json:"formaPago"`
	MetodoPago   string     `json:"metodoPago"`
	Serie        *string   `json:"serie"`
	Descuento    *float64  `json:"descuento"`
	Items        []*GQLCreateInvoiceItemInput `json:"items"`
}

type GQLCreateInvoiceItemInput struct {
	ProductID     uuid.UUID `json:"productId"`
	Quantity      int       `json:"quantity"`
	ClaveProdServ string    `json:"claveProdServ"`
	ClaveUnidad   string    `json:"claveUnidad"`
	Discount      *float64  `json:"discount"`
}

type GQLUpdateTenantFiscalInput struct {
	RFC           *string `json:"rfc"`
	RazonSocial   *string `json:"razonSocial"`
	RegimenFiscal *string `json:"regimenFiscal"`
	Calle         *string `json:"calle"`
	Numero        *string `json:"numero"`
	Colonia       *string `json:"colonia"`
	CP            *string `json:"cp"`
	Ciudad        *string `json:"ciudad"`
	Estado        *string `json:"estado"`
	Pais          *string `json:"pais"`
}

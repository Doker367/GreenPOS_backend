// generated.go - GraphQL resolver interfaces
// This file is manually maintained to provide the resolver interface types
// that gqlgen normally generates from the schema.

package graph

import (
	"context"

	"github.com/google/uuid"
	"github.com/greenpos/backend/internal/model"
)

// QueryResolver handles Query type field resolution
type QueryResolver interface {
	Tenants(ctx context.Context) ([]model.Tenant, error)
	Tenant(ctx context.Context, id uuid.UUID) (*model.Tenant, error)
	Branches(ctx context.Context, tenantID uuid.UUID) ([]model.Branch, error)
	Branch(ctx context.Context, id uuid.UUID) (*model.Branch, error)
	Users(ctx context.Context, branchID uuid.UUID) ([]model.User, error)
	User(ctx context.Context, id uuid.UUID) (*model.User, error)
	Categories(ctx context.Context, branchID uuid.UUID) ([]model.Category, error)
	Category(ctx context.Context, id uuid.UUID) (*model.Category, error)
	Products(ctx context.Context, branchID uuid.UUID, categoryID *uuid.UUID) ([]model.Product, error)
	Product(ctx context.Context, id uuid.UUID) (*model.Product, error)
	Tables(ctx context.Context, branchID uuid.UUID) ([]model.RestaurantTable, error)
	Table(ctx context.Context, id uuid.UUID) (*model.RestaurantTable, error)
	Orders(ctx context.Context, branchID uuid.UUID, status *model.OrderStatus) ([]model.Order, error)
	ActiveOrders(ctx context.Context, branchID uuid.UUID) ([]*model.Order, error)
	Order(ctx context.Context, id uuid.UUID) (*model.Order, error)
	Reservations(ctx context.Context, branchID uuid.UUID, date *string) ([]model.Reservation, error)
	Reservation(ctx context.Context, id uuid.UUID) (*model.Reservation, error)
	Revenue(ctx context.Context, branchID uuid.UUID, startDate string, endDate string) ([]RevenueData, error)
	TopProducts(ctx context.Context, branchID uuid.UUID, startDate string, endDate string, limit int) ([]TopProduct, error)
	Invoices(ctx context.Context, branchID uuid.UUID, status *model.InvoiceStatus, startDate, endDate *string) ([]model.Invoice, error)
	Invoice(ctx context.Context, id uuid.UUID) (*model.Invoice, error)
	TenantFiscal(ctx context.Context) (*model.TenantFiscal, error)
	KitchenOrders(ctx context.Context, branchID uuid.UUID) ([]model.Order, error)
	DashboardMetrics(ctx context.Context, branchID uuid.UUID, period string) (*DashboardMetrics, error)
	SalesByDay(ctx context.Context, branchID uuid.UUID, days int) ([]DailySales, error)
	OrdersByStatus(ctx context.Context, branchID uuid.UUID) ([]StatusCount, error)
	RevenueByPeriod(ctx context.Context, branchID uuid.UUID, period string) (*RevenueReport, error)
	// QR Menu queries
	PublicMenu(ctx context.Context, branchID uuid.UUID) ([]CategoryWithProductsGQL, error)
	TableByQRToken(ctx context.Context, token string) (*Table, error)
	GetQRByToken(ctx context.Context, token string) (*TableQR, error)
}

// MutationResolver handles Mutation type field resolution
type MutationResolver interface {
	CreateTenant(ctx context.Context, input CreateTenantInput) (*model.Tenant, error)
	UpdateTenant(ctx context.Context, id uuid.UUID, input UpdateTenantInput) (*model.Tenant, error)
	DeleteTenant(ctx context.Context, id uuid.UUID) (*model.Tenant, error)
	CreateBranch(ctx context.Context, input CreateBranchInput) (*model.Branch, error)
	UpdateBranch(ctx context.Context, id uuid.UUID, input UpdateBranchInput) (*model.Branch, error)
	DeleteBranch(ctx context.Context, id uuid.UUID) (*model.Branch, error)
	Login(ctx context.Context, email, password string) (*AuthPayload, error)
	RefreshToken(ctx context.Context, token string) (*AuthPayload, error)
	CreateUser(ctx context.Context, input CreateUserInput) (*model.User, error)
	UpdateUser(ctx context.Context, id uuid.UUID, input UpdateUserInput) (*model.User, error)
	DeleteUser(ctx context.Context, id uuid.UUID) (*model.User, error)
	CreateCategory(ctx context.Context, input CreateCategoryInput) (*model.Category, error)
	UpdateCategory(ctx context.Context, id uuid.UUID, input UpdateCategoryInput) (*model.Category, error)
	DeleteCategory(ctx context.Context, id uuid.UUID) (*model.Category, error)
	ReorderCategories(ctx context.Context, branchID uuid.UUID, orders []CategoryOrder) ([]model.Category, error)
	CreateProduct(ctx context.Context, input CreateProductInput) (*model.Product, error)
	UpdateProduct(ctx context.Context, id uuid.UUID, input UpdateProductInput) (*model.Product, error)
	DeleteProduct(ctx context.Context, id uuid.UUID) (*model.Product, error)
	ToggleProductAvailability(ctx context.Context, id uuid.UUID, isAvailable bool) (*model.Product, error)
	CreateTable(ctx context.Context, input CreateTableInput) (*model.RestaurantTable, error)
	UpdateTable(ctx context.Context, id uuid.UUID, input UpdateTableInput) (*model.RestaurantTable, error)
	DeleteTable(ctx context.Context, id uuid.UUID) (*model.RestaurantTable, error)
	CreateOrder(ctx context.Context, input CreateOrderInput) (*model.Order, error)
	UpdateOrderStatus(ctx context.Context, id uuid.UUID, status model.OrderStatus) (*model.Order, error)
	AddOrderItem(ctx context.Context, orderID uuid.UUID, input AddOrderItemInput) (*model.OrderItem, error)
	UpdateOrderItem(ctx context.Context, itemID uuid.UUID, quantity int) (*model.OrderItem, error)
	RemoveOrderItem(ctx context.Context, itemID uuid.UUID) (*model.Order, error)
	CreateReservation(ctx context.Context, input CreateReservationInput) (*model.Reservation, error)
	UpdateReservation(ctx context.Context, id uuid.UUID, input CreateReservationInput) (*model.Reservation, error)
	CancelReservation(ctx context.Context, id uuid.UUID) (*model.Reservation, error)
	CreateInvoice(ctx context.Context, input GQLCreateInvoiceInput) (*model.Invoice, error)
	StampInvoice(ctx context.Context, id uuid.UUID) (*model.Invoice, error)
	CancelInvoice(ctx context.Context, id uuid.UUID, motivo string) (*model.Invoice, error)
	UpdateTenantFiscal(ctx context.Context, input GQLUpdateTenantFiscalInput) (*model.TenantFiscal, error)
	CreateTenantFiscal(ctx context.Context, input GQLUpdateTenantFiscalInput) (*model.TenantFiscal, error)
	MarkOrderReady(ctx context.Context, id uuid.UUID) (*model.Order, error)
	// FCM Token management
	SaveFCMToken(ctx context.Context, token string, device *string) (*SaveFCMTokenResult, error)
	RemoveFCMToken(ctx context.Context, token string) (*SaveFCMTokenResult, error)
	SendPushNotification(ctx context.Context, userID uuid.UUID, title string, body string) (*SaveFCMTokenResult, error)
	// QR Menu mutations
	GenerateTableQR(ctx context.Context, tableID uuid.UUID) (*TableQR, error)
	CreateClientOrder(ctx context.Context, input CreateClientOrderInputGQL) (*ClientOrderConfirmationGQL, error)
}

// BranchResolver handles Branch type field resolution
type BranchResolver interface {
	Invoices(ctx context.Context, obj *model.Branch) ([]model.Invoice, error)
}

// OrderResolver handles Order type field resolution
type OrderResolver interface {
	Items(ctx context.Context, obj *model.Order) ([]model.OrderItem, error)
	TableInfo(ctx context.Context, obj *model.Order) (*model.RestaurantTable, error)
	Waiter(ctx context.Context, obj *model.Order) (*model.User, error)
}

// OrderItemResolver handles OrderItem type field resolution
type OrderItemResolver interface{}

// CategoryResolver handles Category type field resolution
type CategoryResolver interface {
	Products(ctx context.Context, obj *model.Category) ([]model.Product, error)
}

// ProductResolver handles Product type field resolution
type ProductResolver interface {
	CategoryInfo(ctx context.Context, obj *model.Product) (*model.Category, error)
}

// TableResolver handles RestaurantTable type field resolution
type TableResolver interface {
	BranchInfo(ctx context.Context, obj *model.RestaurantTable) (*model.Branch, error)
}

// UserResolver handles User type field resolution
type UserResolver interface {
	BranchInfo(ctx context.Context, obj *model.User) (*model.Branch, error)
}

// ReservationResolver handles Reservation type field resolution
type ReservationResolver interface {
	TableInfo(ctx context.Context, obj *model.Reservation) (*model.RestaurantTable, error)
}

// InvoiceResolver handles Invoice type field resolution
type InvoiceResolver interface {
	Items(ctx context.Context, obj *model.Invoice) ([]model.InvoiceItem, error)
	Branch(ctx context.Context, obj *model.Invoice) (*model.Branch, error)
}

// TenantResolver handles Tenant type field resolution
type TenantResolver interface {
	Branches(ctx context.Context, obj *model.Tenant) ([]model.Branch, error)
}

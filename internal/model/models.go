package model

import (
	"time"

	"github.com/google/uuid"
)

type Tenant struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Slug      string    `json:"slug"`
	Settings  JSON      `json:"settings"`
	IsActive  bool      `json:"isActive"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type Branch struct {
	ID        uuid.UUID `json:"id"`
	TenantID  uuid.UUID `json:"tenantId"`
	Name      string    `json:"name"`
	Address   string    `json:"address"`
	Phone     string    `json:"phone"`
	IsActive  bool      `json:"isActive"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type User struct {
	ID           uuid.UUID `json:"id"`
	BranchID     uuid.UUID `json:"branchId"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	Name         string    `json:"name"`
	Role         UserRole  `json:"role"`
	IsActive     bool      `json:"isActive"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

type UserRole string

const (
	RoleOwner   UserRole = "OWNER"
	RoleAdmin   UserRole = "ADMIN"
	RoleManager UserRole = "MANAGER"
	RoleWaiter  UserRole = "WAITER"
	RoleKitchen UserRole = "KITCHEN"
)

type Category struct {
	ID          uuid.UUID  `json:"id"`
	BranchID    uuid.UUID  `json:"branchId"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	SortOrder   int        `json:"sortOrder"`
	IsActive    bool       `json:"isActive"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
}

type Product struct {
	ID              uuid.UUID  `json:"id"`
	BranchID        uuid.UUID  `json:"branchId"`
	CategoryID      *uuid.UUID `json:"categoryId"`
	Name            string     `json:"name"`
	Description     string     `json:"description"`
	Price           float64    `json:"price"`
	ImageURL        string     `json:"imageUrl"`
	IsAvailable     bool       `json:"isAvailable"`
	IsFeatured      bool       `json:"isFeatured"`
	PreparationTime int        `json:"preparationTime"`
	Allergens       []string   `json:"allergens"`
	Rating          float64    `json:"rating"`
	ReviewCount     int        `json:"reviewCount"`
	CreatedAt       time.Time  `json:"createdAt"`
	UpdatedAt       time.Time  `json:"updatedAt"`
}

type RestaurantTable struct {
	ID        uuid.UUID `json:"id"`
	BranchID  uuid.UUID `json:"branchId"`
	Number    string    `json:"number"`
	Capacity  int       `json:"capacity"`
	Status    TableStatus `json:"status"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type TableStatus string

const (
	TableAvailable TableStatus = "AVAILABLE"
	TableOccupied  TableStatus = "OCCUPIED"
	TableReserved  TableStatus = "RESERVED"
)

type Order struct {
	ID            uuid.UUID   `json:"id"`
	BranchID      uuid.UUID   `json:"branchId"`
	TableID       *uuid.UUID  `json:"tableId"`
	UserID        uuid.UUID   `json:"userId"`
	CustomerName  string      `json:"customerName"`
	CustomerPhone string      `json:"customerPhone"`
	Status        OrderStatus `json:"status"`
	Subtotal      float64     `json:"subtotal"`
	Tax           float64     `json:"tax"`
	Discount      float64     `json:"discount"`
	Total         float64     `json:"total"`
	Notes         string      `json:"notes"`
	CreatedAt     time.Time   `json:"createdAt"`
	UpdatedAt     time.Time   `json:"updatedAt"`
}

type OrderStatus string

const (
	OrderPending   OrderStatus = "PENDING"
	OrderAccepted  OrderStatus = "ACCEPTED"
	OrderPreparing OrderStatus = "PREPARING"
	OrderReady     OrderStatus = "READY"
	OrderDelivered OrderStatus = "DELIVERED"
	OrderCancelled OrderStatus = "CANCELLED"
	OrderPaid      OrderStatus = "PAID"
)

type OrderItem struct {
	ID         uuid.UUID `json:"id"`
	OrderID    uuid.UUID `json:"orderId"`
	ProductID  uuid.UUID `json:"productId"`
	Quantity   int       `json:"quantity"`
	UnitPrice  float64   `json:"unitPrice"`
	TotalPrice float64   `json:"totalPrice"`
	Notes      string    `json:"notes"`
	CreatedAt  time.Time `json:"createdAt"`
}

type Reservation struct {
	ID               uuid.UUID         `json:"id"`
	BranchID         uuid.UUID         `json:"branchId"`
	TableID          *uuid.UUID        `json:"tableId"`
	CustomerName     string            `json:"customerName"`
	CustomerPhone    string            `json:"customerPhone"`
	CustomerEmail    string            `json:"customerEmail"`
	GuestCount       int               `json:"guestCount"`
	ReservationDate  time.Time         `json:"reservationDate"`
	ReservationTime  string            `json:"reservationTime"`
	Status           ReservationStatus `json:"status"`
	Notes            string            `json:"notes"`
	CreatedAt        time.Time         `json:"createdAt"`
	UpdatedAt        time.Time         `json:"updatedAt"`
}

type ReservationStatus string

const (
	ReservationPending   ReservationStatus = "PENDING"
	ReservationConfirmed ReservationStatus = "CONFIRMED"
	ReservationCancelled ReservationStatus = "CANCELLED"
	ReservationCompleted ReservationStatus = "COMPLETED"
)

type JSON map[string]interface{}

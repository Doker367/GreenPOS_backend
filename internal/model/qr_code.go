package model

import (
	"time"

	"github.com/google/uuid"
)

// TableQRCode represents a QR code for a restaurant table
type TableQRCode struct {
	ID         uuid.UUID `json:"id"`
	TableID    uuid.UUID `json:"tableId"`
	QRToken    string    `json:"qrToken"`
	AccessURL  string    `json:"accessUrl"`
	IsActive   bool      `json:"isActive"`
	CreatedAt  time.Time `json:"createdAt"`
}

// ClientOrder represents an order placed by a client via QR menu
type ClientOrder struct {
	ID            uuid.UUID        `json:"id"`
	BranchID      uuid.UUID        `json:"branchId"`
	TableID       uuid.UUID        `json:"tableId"`
	QRCodeID      uuid.UUID        `json:"qrCodeId"`
	CustomerName  string           `json:"customerName"`
	CustomerPhone string           `json:"customerPhone"`
	Status        ClientOrderStatus `json:"status"`
	Subtotal      float64          `json:"subtotal"`
	Tax           float64          `json:"tax"`
	Total         float64          `json:"total"`
	Notes         string           `json:"notes"`
	CreatedAt     time.Time        `json:"createdAt"`
}

// ClientOrderStatus represents the status of a client order
type ClientOrderStatus string

const (
	ClientOrderPending   ClientOrderStatus = "PENDING"
	ClientOrderAccepted  ClientOrderStatus = "ACCEPTED"
	ClientOrderPreparing ClientOrderStatus = "PREPARING"
	ClientOrderReady    ClientOrderStatus = "READY"
	ClientOrderDelivered ClientOrderStatus = "DELIVERED"
	ClientOrderCancelled ClientOrderStatus = "CANCELLED"
)

// ClientOrderItem represents an item in a client order
type ClientOrderItem struct {
	ID           uuid.UUID `json:"id"`
	OrderID      uuid.UUID `json:"orderId"`
	ProductID    uuid.UUID `json:"productId"`
	ProductName  string    `json:"productName"`
	Quantity     int       `json:"quantity"`
	UnitPrice    float64   `json:"unitPrice"`
	TotalPrice   float64   `json:"totalPrice"`
	Notes        string    `json:"notes"`
	CreatedAt    time.Time `json:"createdAt"`
}

// CategoryWithProducts represents a category with its available products for public menu
type CategoryWithProducts struct {
	ID          uuid.UUID       `json:"id"`
	Name        string          `json:"name"`
	Description string          `json:"description"`
	SortOrder   int             `json:"sortOrder"`
	Products    []ProductPublic `json:"products"`
}

// ProductPublic represents a product in the public menu (limited info)
type ProductPublic struct {
	ID              uuid.UUID `json:"id"`
	Name            string    `json:"name"`
	Description     string    `json:"description"`
	Price           float64   `json:"price"`
	ImageURL        string    `json:"imageUrl"`
	PreparationTime int       `json:"preparationTime"`
	Allergens       []string  `json:"allergens"`
}

// ClientOrderConfirmation represents the confirmation returned after creating a client order
type ClientOrderConfirmation struct {
	OrderID    uuid.UUID `json:"orderId"`
	OrderToken string    `json:"orderToken"`
	Status     string    `json:"status"`
	Message    string    `json:"message"`
	EstimatedReadyTime *int `json:"estimatedReadyTime"`
}

package model

import (
	"time"

	"github.com/google/uuid"
)

// InventoryItem represents a stock item in inventory
type InventoryItem struct {
	ID          uuid.UUID `json:"id"`
	BranchID    uuid.UUID `json:"branchId"`
	Name        string    `json:"name"`
	Unit        string    `json:"unit"`
	CurrentStock float64  `json:"currentStock"`
	MinStock    float64  `json:"minStock"`
	Cost        float64  `json:"cost"`
	IsActive    bool     `json:"isActive"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// IsLowStock returns true if current stock is below minimum
func (i *InventoryItem) IsLowStock() bool {
	return i.CurrentStock < i.MinStock
}

// InventoryMovementType represents the type of stock movement
type InventoryMovementType string

const (
	MovementIn        InventoryMovementType = "IN"
	MovementOut       InventoryMovementType = "OUT"
	MovementAdjustment InventoryMovementType = "ADJUSTMENT"
)

// InventoryMovement records a stock change
type InventoryMovement struct {
	ID              uuid.UUID               `json:"id"`
	InventoryItemID uuid.UUID               `json:"inventoryItemId"`
	Type            InventoryMovementType   `json:"type"`
	Quantity        float64                 `json:"quantity"`
	Reason          string                  `json:"reason"`
	UserID          *uuid.UUID              `json:"userId"`
	CreatedAt       time.Time               `json:"createdAt"`
}

// NewNotFoundError creates a not found error for inventory items
func NewNotFoundError(entity string, id string) *NotFoundError {
	return &NotFoundError{Message: entity + " not found: " + id}
}

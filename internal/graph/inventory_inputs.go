package graph

import (
	"github.com/google/uuid"
)

// CreateInventoryItemInput is the input for creating an inventory item
type CreateInventoryItemInput struct {
	BranchID     uuid.UUID
	Name         string
	Unit         string
	CurrentStock float64
	MinStock     float64
	Cost         float64
}

// UpdateInventoryItemInput is the input for updating an inventory item
type UpdateInventoryItemInput struct {
	Name     *string
	Unit     *string
	MinStock *float64
	Cost     *float64
	IsActive *bool
}

// AddInventoryInput is the input for adding inventory
type AddInventoryInput struct {
	ItemID   uuid.UUID
	Quantity float64
	Reason   string
}

// RemoveInventoryInput is the input for removing inventory
type RemoveInventoryInput struct {
	ItemID   uuid.UUID
	Quantity float64
	Reason   string
}

// AdjustInventoryInput is the input for adjusting inventory stock
type AdjustInventoryInput struct {
	ItemID   uuid.UUID
	NewStock float64
	Reason   string
}

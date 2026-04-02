package graph

import (
	"context"

	"github.com/google/uuid"
	"github.com/greenpos/backend/internal/model"
	"github.com/greenpos/backend/internal/service"
)

// ============ Inventory Item Resolvers ============

// InventoryItem is the resolver for the inventoryItem field.
func (r *inventoryItemResolver) InventoryItem(ctx context.Context, obj *InventoryItem) (*model.InventoryItem, error) {
	return r.Services.Inventory.GetItem(ctx, obj.ID)
}

// Branch is the resolver for the branch field.
func (r *inventoryItemResolver) Branch(ctx context.Context, obj *model.InventoryItem) (*model.Branch, error) {
	return r.Services.Branch.GetByID(ctx, obj.BranchID)
}

// InventoryItems returns items for a branch.
func (r *inventoryItemResolver) InventoryItems(ctx context.Context, branchID uuid.UUID) ([]model.InventoryItem, error) {
	return r.Services.Inventory.GetItemsByBranch(ctx, branchID)
}

// LowStockItems returns items below minimum stock.
func (r *inventoryItemResolver) LowStockItems(ctx context.Context, branchID uuid.UUID) ([]model.InventoryItem, error) {
	return r.Services.Inventory.GetLowStockItems(ctx, branchID)
}

// Movements returns movement history for an item.
func (r *inventoryItemResolver) Movements(ctx context.Context, obj *model.InventoryItem, limit *int) ([]model.InventoryMovement, error) {
	l := 50
	if limit != nil {
		l = *limit
	}
	return r.Services.Inventory.GetMovementsByItem(ctx, obj.ID, l)
}

// ============ Query Resolvers ============

// InventoryItems is the resolver for the inventoryItems field.
func (r *queryResolver) InventoryItems(ctx context.Context, branchID uuid.UUID) ([]model.InventoryItem, error) {
	return r.Services.Inventory.GetItemsByBranch(ctx, branchID)
}

// LowStockItems is the resolver for the lowStockItems field.
func (r *queryResolver) LowStockItems(ctx context.Context, branchID uuid.UUID) ([]model.InventoryItem, error) {
	return r.Services.Inventory.GetLowStockItems(ctx, branchID)
}

// InventoryItem is the resolver for the inventoryItem field.
func (r *queryResolver) InventoryItem(ctx context.Context, id uuid.UUID) (*model.InventoryItem, error) {
	return r.Services.Inventory.GetItem(ctx, id)
}

// InventoryMovements is the resolver for the inventoryMovements field.
func (r *queryResolver) InventoryMovements(ctx context.Context, itemID uuid.UUID, limit *int) ([]model.InventoryMovement, error) {
	l := 50
	if limit != nil {
		l = *limit
	}
	return r.Services.Inventory.GetMovementsByItem(ctx, itemID, l)
}

// ============ Mutation Resolvers ============

// CreateInventoryItem is the resolver for the createInventoryItem field.
func (r *mutationResolver) CreateInventoryItem(ctx context.Context, input CreateInventoryItemInput) (*model.InventoryItem, error) {
	return r.Services.Inventory.CreateItem(ctx, service.CreateItemInput{
		BranchID:     input.BranchID,
		Name:         input.Name,
		Unit:         input.Unit,
		CurrentStock: input.CurrentStock,
		MinStock:     input.MinStock,
		Cost:         input.Cost,
	})
}

// UpdateInventoryItem is the resolver for the updateInventoryItem field.
func (r *mutationResolver) UpdateInventoryItem(ctx context.Context, id uuid.UUID, input UpdateInventoryItemInput) (*model.InventoryItem, error) {
	svcInput := service.UpdateItemInput{}
	if input.Name != nil {
		svcInput.Name = input.Name
	}
	if input.Unit != nil {
		svcInput.Unit = input.Unit
	}
	if input.MinStock != nil {
		svcInput.MinStock = input.MinStock
	}
	if input.Cost != nil {
		svcInput.Cost = input.Cost
	}
	if input.IsActive != nil {
		svcInput.IsActive = input.IsActive
	}
	return r.Services.Inventory.UpdateItem(ctx, id, svcInput)
}

// DeleteInventoryItem is the resolver for the deleteInventoryItem field.
func (r *mutationResolver) DeleteInventoryItem(ctx context.Context, id uuid.UUID) (bool, error) {
	err := r.Services.Inventory.DeleteItem(ctx, id)
	return err == nil, err
}

// AddInventory is the resolver for the addInventory field.
func (r *mutationResolver) AddInventory(ctx context.Context, input AddInventoryInput) (*model.InventoryItem, error) {
	userID, _ := GetUserIDFromContext(ctx)
	var uid *uuid.UUID
	if userID != uuid.Nil {
		uid = &userID
	}
	return r.Services.Inventory.AddStock(ctx, service.AddStockInput{
		ItemID:   input.ItemID,
		Quantity: input.Quantity,
		Reason:   input.Reason,
		UserID:   uid,
	})
}

// RemoveInventory is the resolver for the removeInventory field.
func (r *mutationResolver) RemoveInventory(ctx context.Context, input RemoveInventoryInput) (*model.InventoryItem, error) {
	userID, _ := GetUserIDFromContext(ctx)
	var uid *uuid.UUID
	if userID != uuid.Nil {
		uid = &userID
	}
	return r.Services.Inventory.RemoveStock(ctx, service.RemoveStockInput{
		ItemID:   input.ItemID,
		Quantity: input.Quantity,
		Reason:   input.Reason,
		UserID:   uid,
	})
}

// AdjustInventory is the resolver for the adjustInventory field.
func (r *mutationResolver) AdjustInventory(ctx context.Context, input AdjustInventoryInput) (*model.InventoryItem, error) {
	userID, _ := GetUserIDFromContext(ctx)
	var uid *uuid.UUID
	if userID != uuid.Nil {
		uid = &userID
	}
	return r.Services.Inventory.AdjustStock(ctx, service.AdjustStockInput{
		ItemID:   input.ItemID,
		NewStock: input.NewStock,
		Reason:   input.Reason,
		UserID:   uid,
	})
}

// InventoryItem returns InventoryItemResolver implementation.
func (r *Resolver) InventoryItem() InventoryItemResolver { return &inventoryItemResolver{r} }

// InventoryItemResolver has methods to resolve inventory item fields.
type inventoryItemResolver struct{ *Resolver }

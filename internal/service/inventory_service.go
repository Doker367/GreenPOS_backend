package service

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/greenpos/backend/internal/model"
	"github.com/greenpos/backend/internal/repository"
)

// InventoryService handles inventory operations
type InventoryService struct {
	items     repository.InventoryItemRepositoryInterface
	movements repository.InventoryMovementRepositoryInterface
	log       *slog.Logger
}

// NewInventoryService creates a new inventory service
func NewInventoryService(
	items repository.InventoryItemRepositoryInterface,
	movements repository.InventoryMovementRepositoryInterface,
	log *slog.Logger,
) *InventoryService {
	return &InventoryService{
		items:     items,
		movements: movements,
		log:       log,
	}
}

// ============ Inventory Item Operations ============

// CreateItemInput holds data for creating an inventory item
type CreateItemInput struct {
	BranchID    uuid.UUID
	Name        string
	Unit        string
	CurrentStock float64
	MinStock    float64
	Cost        float64
}

// CreateItem creates a new inventory item
func (s *InventoryService) CreateItem(ctx context.Context, input CreateItemInput) (*model.InventoryItem, error) {
	item := &model.InventoryItem{
		ID:           uuid.New(),
		BranchID:    input.BranchID,
		Name:        input.Name,
		Unit:        input.Unit,
		CurrentStock: input.CurrentStock,
		MinStock:    input.MinStock,
		Cost:        input.Cost,
		IsActive:    true,
	}
	if err := s.items.Create(ctx, item); err != nil {
		return nil, fmt.Errorf("failed to create inventory item: %w", err)
	}
	s.log.Info("inventory item created",
		slog.String("id", item.ID.String()),
		slog.String("name", item.Name),
		slog.Float64("stock", item.CurrentStock))
	return item, nil
}

// GetItem retrieves an inventory item by ID
func (s *InventoryService) GetItem(ctx context.Context, id uuid.UUID) (*model.InventoryItem, error) {
	item, err := s.items.GetByID(ctx, id)
	if err != nil {
		return nil, model.NewNotFoundError("inventory item", id.String())
	}
	return item, nil
}

// GetItemsByBranch retrieves all inventory items for a branch
func (s *InventoryService) GetItemsByBranch(ctx context.Context, branchID uuid.UUID) ([]model.InventoryItem, error) {
	return s.items.GetByBranch(ctx, branchID)
}

// GetLowStockItems retrieves items where current_stock < min_stock
func (s *InventoryService) GetLowStockItems(ctx context.Context, branchID uuid.UUID) ([]model.InventoryItem, error) {
	return s.items.GetLowStock(ctx, branchID)
}

// UpdateItemInput holds data for updating an inventory item
type UpdateItemInput struct {
	Name        *string
	Unit        *string
	MinStock    *float64
	Cost        *float64
	IsActive    *bool
}

// UpdateItem updates an inventory item
func (s *InventoryService) UpdateItem(ctx context.Context, id uuid.UUID, input UpdateItemInput) (*model.InventoryItem, error) {
	item, err := s.items.GetByID(ctx, id)
	if err != nil {
		return nil, model.NewNotFoundError("inventory item", id.String())
	}

	if input.Name != nil {
		item.Name = *input.Name
	}
	if input.Unit != nil {
		item.Unit = *input.Unit
	}
	if input.MinStock != nil {
		item.MinStock = *input.MinStock
	}
	if input.Cost != nil {
		item.Cost = *input.Cost
	}
	if input.IsActive != nil {
		item.IsActive = *input.IsActive
	}

	if err := s.items.Update(ctx, item); err != nil {
		return nil, fmt.Errorf("failed to update inventory item: %w", err)
	}
	s.log.Info("inventory item updated", slog.String("id", item.ID.String()))
	return item, nil
}

// DeleteItem soft-deletes an inventory item
func (s *InventoryService) DeleteItem(ctx context.Context, id uuid.UUID) error {
	if err := s.items.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete inventory item: %w", err)
	}
	s.log.Info("inventory item deleted", slog.String("id", id.String()))
	return nil
}

// ============ Inventory Movement Operations ============

// AddStockInput holds data for adding stock
type AddStockInput struct {
	ItemID   uuid.UUID
	Quantity float64
	Reason   string
	UserID   *uuid.UUID
}

// AddStock adds inventory (IN movement)
func (s *InventoryService) AddStock(ctx context.Context, input AddStockInput) (*model.InventoryItem, error) {
	item, err := s.items.GetByID(ctx, input.ItemID)
	if err != nil {
		return nil, model.NewNotFoundError("inventory item", input.ItemID.String())
	}

	if input.Quantity <= 0 {
		return nil, fmt.Errorf("quantity must be positive")
	}

	// Create movement record
	movement := &model.InventoryMovement{
		ID:              uuid.New(),
		InventoryItemID: input.ItemID,
		Type:            model.MovementIn,
		Quantity:        input.Quantity,
		Reason:          input.Reason,
		UserID:          input.UserID,
	}
	if err := s.movements.Create(ctx, movement); err != nil {
		return nil, fmt.Errorf("failed to create movement: %w", err)
	}

	// Update stock
	newStock := item.CurrentStock + input.Quantity
	if err := s.items.UpdateStock(ctx, item.ID, newStock); err != nil {
		return nil, fmt.Errorf("failed to update stock: %w", err)
	}

	item.CurrentStock = newStock
	s.log.Info("stock added",
		slog.String("item_id", item.ID.String()),
		slog.Float64("quantity", input.Quantity),
		slog.Float64("new_stock", newStock))
	return item, nil
}

// RemoveStockInput holds data for removing stock
type RemoveStockInput struct {
	ItemID   uuid.UUID
	Quantity float64
	Reason   string
	UserID   *uuid.UUID
}

// RemoveStock removes inventory (OUT movement)
func (s *InventoryService) RemoveStock(ctx context.Context, input RemoveStockInput) (*model.InventoryItem, error) {
	item, err := s.items.GetByID(ctx, input.ItemID)
	if err != nil {
		return nil, model.NewNotFoundError("inventory item", input.ItemID.String())
	}

	if input.Quantity <= 0 {
		return nil, fmt.Errorf("quantity must be positive")
	}

	if item.CurrentStock < input.Quantity {
		return nil, fmt.Errorf("insufficient stock: available %.2f, requested %.2f", item.CurrentStock, input.Quantity)
	}

	// Create movement record
	movement := &model.InventoryMovement{
		ID:              uuid.New(),
		InventoryItemID: input.ItemID,
		Type:            model.MovementOut,
		Quantity:        input.Quantity,
		Reason:          input.Reason,
		UserID:          input.UserID,
	}
	if err := s.movements.Create(ctx, movement); err != nil {
		return nil, fmt.Errorf("failed to create movement: %w", err)
	}

	// Update stock
	newStock := item.CurrentStock - input.Quantity
	if err := s.items.UpdateStock(ctx, item.ID, newStock); err != nil {
		return nil, fmt.Errorf("failed to update stock: %w", err)
	}

	item.CurrentStock = newStock
	s.log.Info("stock removed",
		slog.String("item_id", item.ID.String()),
		slog.Float64("quantity", input.Quantity),
		slog.Float64("new_stock", newStock))
	return item, nil
}

// AdjustStockInput holds data for adjusting stock
type AdjustStockInput struct {
	ItemID      uuid.UUID
	NewStock    float64
	Reason      string
	UserID      *uuid.UUID
}

// AdjustStock adjusts stock to a specific value (ADJUSTMENT movement)
func (s *InventoryService) AdjustStock(ctx context.Context, input AdjustStockInput) (*model.InventoryItem, error) {
	item, err := s.items.GetByID(ctx, input.ItemID)
	if err != nil {
		return nil, model.NewNotFoundError("inventory item", input.ItemID.String())
	}

	if input.NewStock < 0 {
		return nil, fmt.Errorf("stock cannot be negative")
	}

	difference := input.NewStock - item.CurrentStock

	// Create movement record
	movType := model.MovementAdjustment
	if difference > 0 {
		movType = model.MovementIn
	} else if difference < 0 {
		movType = model.MovementOut
	}

	movement := &model.InventoryMovement{
		ID:              uuid.New(),
		InventoryItemID: input.ItemID,
		Type:            movType,
		Quantity:        difference,
		Reason:          input.Reason,
		UserID:          input.UserID,
	}
	if err := s.movements.Create(ctx, movement); err != nil {
		return nil, fmt.Errorf("failed to create movement: %w", err)
	}

	// Update stock
	if err := s.items.UpdateStock(ctx, item.ID, input.NewStock); err != nil {
		return nil, fmt.Errorf("failed to update stock: %w", err)
	}

	item.CurrentStock = input.NewStock
	s.log.Info("stock adjusted",
		slog.String("item_id", item.ID.String()),
		slog.Float64("old_stock", item.CurrentStock),
		slog.Float64("new_stock", input.NewStock))
	return item, nil
}

// GetMovementsByItem retrieves movement history for an item
func (s *InventoryService) GetMovementsByItem(ctx context.Context, itemID uuid.UUID, limit int) ([]model.InventoryMovement, error) {
	// Verify item exists
	if _, err := s.items.GetByID(ctx, itemID); err != nil {
		return nil, model.NewNotFoundError("inventory item", itemID.String())
	}
	return s.movements.GetByItem(ctx, itemID, limit)
}

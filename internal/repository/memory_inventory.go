package repository

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/greenpos/backend/internal/model"
)

// InventoryItemMemoryRepository is an in-memory implementation of InventoryItemRepositoryInterface
type InventoryItemMemoryRepository struct {
	mu    sync.RWMutex
	items map[uuid.UUID]*model.InventoryItem
}

// NewInventoryItemMemoryRepository creates a new in-memory inventory item repository
func NewInventoryItemMemoryRepository() *InventoryItemMemoryRepository {
	return &InventoryItemMemoryRepository{
		items: make(map[uuid.UUID]*model.InventoryItem),
	}
}

func (r *InventoryItemMemoryRepository) Create(ctx context.Context, item *model.InventoryItem) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	item.CreatedAt = time.Now()
	item.UpdatedAt = time.Now()
	if item.ID == uuid.Nil {
		item.ID = uuid.New()
	}

	r.items[item.ID] = item
	return nil
}

func (r *InventoryItemMemoryRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.InventoryItem, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if item, ok := r.items[id]; ok {
		return item, nil
	}
	return nil, ErrNotFound
}

func (r *InventoryItemMemoryRepository) GetByBranch(ctx context.Context, branchID uuid.UUID) ([]model.InventoryItem, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []model.InventoryItem
	for _, item := range r.items {
		if item.BranchID == branchID && item.IsActive {
			result = append(result, *item)
		}
	}
	return result, nil
}

func (r *InventoryItemMemoryRepository) GetLowStock(ctx context.Context, branchID uuid.UUID) ([]model.InventoryItem, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []model.InventoryItem
	for _, item := range r.items {
		if item.BranchID == branchID && item.IsActive && item.CurrentStock < item.MinStock {
			result = append(result, *item)
		}
	}
	return result, nil
}

func (r *InventoryItemMemoryRepository) Update(ctx context.Context, item *model.InventoryItem) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	item.UpdatedAt = time.Now()
	if _, ok := r.items[item.ID]; ok {
		r.items[item.ID] = item
		return nil
	}
	return ErrNotFound
}

func (r *InventoryItemMemoryRepository) UpdateStock(ctx context.Context, id uuid.UUID, newStock float64) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if item, ok := r.items[id]; ok {
		item.CurrentStock = newStock
		item.UpdatedAt = time.Now()
		return nil
	}
	return ErrNotFound
}

func (r *InventoryItemMemoryRepository) Delete(ctx context.Context, id uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if item, ok := r.items[id]; ok {
		item.IsActive = false
		item.UpdatedAt = time.Now()
		return nil
	}
	return ErrNotFound
}

func (r *InventoryItemMemoryRepository) Seed(ctx context.Context) error {
	// No seed for memory repository
	return nil
}

// InventoryMovementMemoryRepository is an in-memory implementation of InventoryMovementRepositoryInterface
type InventoryMovementMemoryRepository struct {
	mu         sync.RWMutex
	movements  map[uuid.UUID]*model.InventoryMovement
	byItem     map[uuid.UUID][]*model.InventoryMovement
}

// NewInventoryMovementMemoryRepository creates a new in-memory inventory movement repository
func NewInventoryMovementMemoryRepository() *InventoryMovementMemoryRepository {
	return &InventoryMovementMemoryRepository{
		movements: make(map[uuid.UUID]*model.InventoryMovement),
		byItem:    make(map[uuid.UUID][]*model.InventoryMovement),
	}
}

func (r *InventoryMovementMemoryRepository) Create(ctx context.Context, movement *model.InventoryMovement) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	movement.CreatedAt = time.Now()
	if movement.ID == uuid.Nil {
		movement.ID = uuid.New()
	}

	r.movements[movement.ID] = movement
	r.byItem[movement.InventoryItemID] = append(r.byItem[movement.InventoryItemID], movement)
	return nil
}

func (r *InventoryMovementMemoryRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.InventoryMovement, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if m, ok := r.movements[id]; ok {
		return m, nil
	}
	return nil, ErrNotFound
}

func (r *InventoryMovementMemoryRepository) GetByItem(ctx context.Context, itemID uuid.UUID, limit int) ([]model.InventoryMovement, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	movements := r.byItem[itemID]
	if limit > 0 && len(movements) > limit {
		movements = movements[len(movements)-limit:]
	}

	result := make([]model.InventoryMovement, len(movements))
	for i, m := range movements {
		result[i] = *m
	}
	return result, nil
}

func (r *InventoryMovementMemoryRepository) Seed(ctx context.Context) error {
	// No seed for memory repository
	return nil
}

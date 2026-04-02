package repository

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/greenpos/backend/internal/model"
)

// TableMemoryRepository is an in-memory implementation of TableRepository
type TableMemoryRepository struct {
	mu       sync.RWMutex
	tables   map[uuid.UUID]*model.RestaurantTable
	byBranch map[uuid.UUID][]*model.RestaurantTable
}

// NewTableMemoryRepository creates a new in-memory table repository
func NewTableMemoryRepository() *TableMemoryRepository {
	return &TableMemoryRepository{
		tables:   make(map[uuid.UUID]*model.RestaurantTable),
		byBranch: make(map[uuid.UUID][]*model.RestaurantTable),
	}
}

func (r *TableMemoryRepository) Create(ctx context.Context, table *model.RestaurantTable) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	table.CreatedAt = time.Now()
	table.UpdatedAt = time.Now()
	if table.ID == uuid.Nil {
		table.ID = uuid.New()
	}
	if table.Status == "" {
		table.Status = model.TableAvailable
	}

	r.tables[table.ID] = table
	r.byBranch[table.BranchID] = append(r.byBranch[table.BranchID], table)
	return nil
}

func (r *TableMemoryRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.RestaurantTable, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if t, ok := r.tables[id]; ok {
		return t, nil
	}
	return nil, ErrNotFound
}

func (r *TableMemoryRepository) GetByBranch(ctx context.Context, branchID uuid.UUID) ([]model.RestaurantTable, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	tables := r.byBranch[branchID]
	result := make([]model.RestaurantTable, len(tables))
	for i, t := range tables {
		result[i] = *t
	}
	return result, nil
}

func (r *TableMemoryRepository) ListByBranch(ctx context.Context, branchID uuid.UUID) ([]model.RestaurantTable, error) {
	return r.GetByBranch(ctx, branchID)
}

func (r *TableMemoryRepository) Update(ctx context.Context, table *model.RestaurantTable) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	table.UpdatedAt = time.Now()
	r.tables[table.ID] = table
	return nil
}

func (r *TableMemoryRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status model.TableStatus) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if t, ok := r.tables[id]; ok {
		t.Status = status
		t.UpdatedAt = time.Now()
		return nil
	}
	return ErrNotFound
}

func (r *TableMemoryRepository) Delete(ctx context.Context, id uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if t, ok := r.tables[id]; ok {
		t.Status = model.TableStatus("UNAVAILABLE")
		t.UpdatedAt = time.Now()
		return nil
	}
	return ErrNotFound
}

func (r *TableMemoryRepository) GetAvailable(ctx context.Context, branchID uuid.UUID) ([]model.RestaurantTable, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	tables := r.byBranch[branchID]
	result := make([]model.RestaurantTable, 0)
	for _, t := range tables {
		if t.Status == model.TableAvailable {
			result = append(result, *t)
		}
	}
	return result, nil
}

func (r *TableMemoryRepository) Seed(ctx context.Context) error {
	now := time.Now()
	tables := []*model.RestaurantTable{
		{
			ID:        uuid.MustParse("66666666-6666-6666-6666-666666666661"),
			BranchID:  uuid.MustParse("22222222-2222-2222-2222-222222222222"),
			Number:    "1",
			Capacity:  2,
			Status:    model.TableAvailable,
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			ID:        uuid.MustParse("66666666-6666-6666-6666-666666666662"),
			BranchID:  uuid.MustParse("22222222-2222-2222-2222-222222222222"),
			Number:    "2",
			Capacity:  4,
			Status:    model.TableAvailable,
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			ID:        uuid.MustParse("66666666-6666-6666-6666-666666666663"),
			BranchID:  uuid.MustParse("22222222-2222-2222-2222-222222222222"),
			Number:    "3",
			Capacity:  6,
			Status:    model.TableAvailable,
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			ID:        uuid.MustParse("66666666-6666-6666-6666-666666666664"),
			BranchID:  uuid.MustParse("22222222-2222-2222-2222-222222222222"),
			Number:    "4",
			Capacity:  8,
			Status:    model.TableAvailable,
			CreatedAt: now,
			UpdatedAt: now,
		},
	}

	for _, t := range tables {
		if err := r.Create(ctx, t); err != nil {
			return err
		}
	}
	return nil
}

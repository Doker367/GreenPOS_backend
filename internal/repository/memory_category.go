package repository

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/greenpos/backend/internal/model"
)

// CategoryMemoryRepository is an in-memory implementation of CategoryRepository
type CategoryMemoryRepository struct {
	mu        sync.RWMutex
	categories map[uuid.UUID]*model.Category
	byBranch  map[uuid.UUID][]*model.Category
}

// NewCategoryMemoryRepository creates a new in-memory category repository
func NewCategoryMemoryRepository() *CategoryMemoryRepository {
	return &CategoryMemoryRepository{
		categories: make(map[uuid.UUID]*model.Category),
		byBranch:   make(map[uuid.UUID][]*model.Category),
	}
}

func (r *CategoryMemoryRepository) Create(ctx context.Context, category *model.Category) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	category.CreatedAt = time.Now()
	category.UpdatedAt = time.Now()
	if category.ID == uuid.Nil {
		category.ID = uuid.New()
	}

	r.categories[category.ID] = category
	r.byBranch[category.BranchID] = append(r.byBranch[category.BranchID], category)
	return nil
}

func (r *CategoryMemoryRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Category, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if cat, ok := r.categories[id]; ok {
		return cat, nil
	}
	return nil, ErrNotFound
}

func (r *CategoryMemoryRepository) GetByBranch(ctx context.Context, branchID uuid.UUID) ([]model.Category, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	cats := r.byBranch[branchID]
	result := make([]model.Category, 0, len(cats))
	for _, c := range cats {
		if c.IsActive {
			result = append(result, *c)
		}
	}
	return result, nil
}

func (r *CategoryMemoryRepository) ListByBranch(ctx context.Context, branchID uuid.UUID) ([]model.Category, error) {
	return r.GetByBranch(ctx, branchID)
}

func (r *CategoryMemoryRepository) Update(ctx context.Context, category *model.Category) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	category.UpdatedAt = time.Now()
	r.categories[category.ID] = category
	return nil
}

func (r *CategoryMemoryRepository) Delete(ctx context.Context, id uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if cat, ok := r.categories[id]; ok {
		cat.IsActive = false
		cat.UpdatedAt = time.Now()
		return nil
	}
	return ErrNotFound
}

func (r *CategoryMemoryRepository) GetMaxSortOrder(ctx context.Context, branchID uuid.UUID) (int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	max := 0
	for _, c := range r.byBranch[branchID] {
		if c.SortOrder > max {
			max = c.SortOrder
		}
	}
	return max, nil
}

func (r *CategoryMemoryRepository) Reorder(ctx context.Context, branchID uuid.UUID, orders []struct {
	ID        uuid.UUID
	SortOrder int
}) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, o := range orders {
		if cat, ok := r.categories[o.ID]; ok {
			cat.SortOrder = o.SortOrder
			cat.UpdatedAt = time.Now()
		}
	}
	return nil
}

func (r *CategoryMemoryRepository) Seed(ctx context.Context) error {
	now := time.Now()
	categories := []*model.Category{
		{
			ID:          uuid.MustParse("44444444-4444-4444-4444-444444444441"),
			BranchID:    uuid.MustParse("22222222-2222-2222-2222-222222222222"),
			Name:        "Appetizers",
			Description: "Start your meal right",
			SortOrder:   1,
			IsActive:    true,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			ID:          uuid.MustParse("44444444-4444-4444-4444-444444444442"),
			BranchID:    uuid.MustParse("22222222-2222-2222-2222-222222222222"),
			Name:        "Main Courses",
			Description: "Hearty entrees",
			SortOrder:   2,
			IsActive:    true,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			ID:          uuid.MustParse("44444444-4444-4444-4444-444444444443"),
			BranchID:    uuid.MustParse("22222222-2222-2222-2222-222222222222"),
			Name:        "Desserts",
			Description: "Sweet endings",
			SortOrder:   3,
			IsActive:    true,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			ID:          uuid.MustParse("44444444-4444-4444-4444-444444444444"),
			BranchID:    uuid.MustParse("22222222-2222-2222-2222-222222222222"),
			Name:        "Beverages",
			Description: "Refreshing drinks",
			SortOrder:   4,
			IsActive:    true,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
	}

	for _, c := range categories {
		if err := r.Create(ctx, c); err != nil {
			return err
		}
	}
	return nil
}

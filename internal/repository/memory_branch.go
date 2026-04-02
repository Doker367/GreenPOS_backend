package repository

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/greenpos/backend/internal/model"
)

// BranchMemoryRepository is an in-memory implementation of BranchRepository
type BranchMemoryRepository struct {
	mu      sync.RWMutex
	branches map[uuid.UUID]*model.Branch
	byTenant map[uuid.UUID][]*model.Branch
}

// NewBranchMemoryRepository creates a new in-memory branch repository
func NewBranchMemoryRepository() *BranchMemoryRepository {
	return &BranchMemoryRepository{
		branches: make(map[uuid.UUID]*model.Branch),
		byTenant: make(map[uuid.UUID][]*model.Branch),
	}
}

func (r *BranchMemoryRepository) Create(ctx context.Context, branch *model.Branch) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	branch.CreatedAt = time.Now()
	branch.UpdatedAt = time.Now()
	if branch.ID == uuid.Nil {
		branch.ID = uuid.New()
	}

	r.branches[branch.ID] = branch
	r.byTenant[branch.TenantID] = append(r.byTenant[branch.TenantID], branch)
	return nil
}

func (r *BranchMemoryRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Branch, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if branch, ok := r.branches[id]; ok {
		return branch, nil
	}
	return nil, ErrNotFound
}

func (r *BranchMemoryRepository) GetByTenant(ctx context.Context, tenantID uuid.UUID) ([]model.Branch, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	branches := r.byTenant[tenantID]
	result := make([]model.Branch, len(branches))
	for i, b := range branches {
		result[i] = *b
	}
	return result, nil
}

func (r *BranchMemoryRepository) ListByTenant(ctx context.Context, tenantID uuid.UUID) ([]model.Branch, error) {
	return r.GetByTenant(ctx, tenantID)
}

func (r *BranchMemoryRepository) Update(ctx context.Context, branch *model.Branch) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	branch.UpdatedAt = time.Now()
	r.branches[branch.ID] = branch
	return nil
}

func (r *BranchMemoryRepository) Delete(ctx context.Context, id uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.branches[id]; !ok {
		return ErrNotFound
	}
	delete(r.branches, id)
	return nil
}

func (r *BranchMemoryRepository) Seed(ctx context.Context) error {
	now := time.Now()
	branch := &model.Branch{
		ID:        uuid.MustParse("22222222-2222-2222-2222-222222222222"),
		TenantID:  uuid.MustParse("11111111-1111-1111-1111-111111111111"),
		Name:      "Downtown Branch",
		Address:   "123 Main Street",
		Phone:     "+1-555-0100",
		IsActive:  true,
		CreatedAt: now,
		UpdatedAt: now,
	}
	return r.Create(ctx, branch)
}

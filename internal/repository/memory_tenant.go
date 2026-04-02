package repository

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/greenpos/backend/internal/model"
)

// TenantMemoryRepository is an in-memory implementation of TenantRepository
type TenantMemoryRepository struct {
	mu      sync.RWMutex
	tenants map[uuid.UUID]*model.Tenant
	slugs   map[string]*model.Tenant
}

// NewTenantMemoryRepository creates a new in-memory tenant repository
func NewTenantMemoryRepository() *TenantMemoryRepository {
	return &TenantMemoryRepository{
		tenants: make(map[uuid.UUID]*model.Tenant),
		slugs:   make(map[string]*model.Tenant),
	}
}

func (r *TenantMemoryRepository) Create(ctx context.Context, tenant *model.Tenant) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	tenant.CreatedAt = time.Now()
	tenant.UpdatedAt = time.Now()
	if tenant.ID == uuid.Nil {
		tenant.ID = uuid.New()
	}
	if tenant.Settings == nil {
		tenant.Settings = model.JSON{}
	}

	r.tenants[tenant.ID] = tenant
	r.slugs[tenant.Slug] = tenant
	return nil
}

func (r *TenantMemoryRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Tenant, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if tenant, ok := r.tenants[id]; ok {
		return tenant, nil
	}
	return nil, ErrNotFound
}

func (r *TenantMemoryRepository) GetBySlug(ctx context.Context, slug string) (*model.Tenant, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if tenant, ok := r.slugs[slug]; ok {
		return tenant, nil
	}
	return nil, ErrNotFound
}

func (r *TenantMemoryRepository) Update(ctx context.Context, tenant *model.Tenant) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	tenant.UpdatedAt = time.Now()
	if existing, ok := r.tenants[tenant.ID]; ok {
		delete(r.slugs, existing.Slug)
	}
	r.tenants[tenant.ID] = tenant
	r.slugs[tenant.Slug] = tenant
	return nil
}

func (r *TenantMemoryRepository) List(ctx context.Context) ([]model.Tenant, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]model.Tenant, 0, len(r.tenants))
	for _, t := range r.tenants {
		result = append(result, *t)
	}
	return result, nil
}

func (r *TenantMemoryRepository) Seed(ctx context.Context) error {
	now := time.Now()
	tenant := &model.Tenant{
		ID:        uuid.MustParse("11111111-1111-1111-1111-111111111111"),
		Name:      "GreenPOS Demo",
		Slug:      "greenpos",
		Settings:  model.JSON{"theme": "light", "currency": "USD"},
		IsActive:  true,
		CreatedAt: now,
		UpdatedAt: now,
	}
	return r.Create(ctx, tenant)
}

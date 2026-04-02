package repository

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/greenpos/backend/internal/model"
)

// ProductMemoryRepository is an in-memory implementation of ProductRepository
type ProductMemoryRepository struct {
	mu        sync.RWMutex
	products  map[uuid.UUID]*model.Product
	byBranch  map[uuid.UUID][]*model.Product
	byCategory map[uuid.UUID][]*model.Product
}

// NewProductMemoryRepository creates a new in-memory product repository
func NewProductMemoryRepository() *ProductMemoryRepository {
	return &ProductMemoryRepository{
		products:    make(map[uuid.UUID]*model.Product),
		byBranch:    make(map[uuid.UUID][]*model.Product),
		byCategory:  make(map[uuid.UUID][]*model.Product),
	}
}

func (r *ProductMemoryRepository) Create(ctx context.Context, product *model.Product) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	product.CreatedAt = time.Now()
	product.UpdatedAt = time.Now()
	if product.ID == uuid.Nil {
		product.ID = uuid.New()
	}
	if product.Allergens == nil {
		product.Allergens = []string{}
	}

	r.products[product.ID] = product
	r.byBranch[product.BranchID] = append(r.byBranch[product.BranchID], product)
	if product.CategoryID != nil {
		r.byCategory[*product.CategoryID] = append(r.byCategory[*product.CategoryID], product)
	}
	return nil
}

func (r *ProductMemoryRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Product, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if p, ok := r.products[id]; ok {
		return p, nil
	}
	return nil, ErrNotFound
}

func (r *ProductMemoryRepository) GetByBranch(ctx context.Context, branchID uuid.UUID) ([]model.Product, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	products := r.byBranch[branchID]
	result := make([]model.Product, 0, len(products))
	for _, p := range products {
		result = append(result, *p)
	}
	return result, nil
}

func (r *ProductMemoryRepository) ListByBranch(ctx context.Context, branchID uuid.UUID) ([]model.Product, error) {
	return r.GetByBranch(ctx, branchID)
}

func (r *ProductMemoryRepository) GetByCategory(ctx context.Context, branchID, categoryID uuid.UUID) ([]model.Product, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	products := r.byCategory[categoryID]
	result := make([]model.Product, 0, len(products))
	for _, p := range products {
		if p.BranchID == branchID && p.IsAvailable {
			result = append(result, *p)
		}
	}
	return result, nil
}

func (r *ProductMemoryRepository) ListByCategory(ctx context.Context, branchID, categoryID uuid.UUID) ([]model.Product, error) {
	return r.GetByCategory(ctx, branchID, categoryID)
}

func (r *ProductMemoryRepository) GetAvailable(ctx context.Context, branchID uuid.UUID) ([]model.Product, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	products := r.byBranch[branchID]
	result := make([]model.Product, 0, len(products))
	for _, p := range products {
		if p.IsAvailable {
			result = append(result, *p)
		}
	}
	return result, nil
}

func (r *ProductMemoryRepository) GetFeatured(ctx context.Context, branchID uuid.UUID) ([]model.Product, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	products := r.byBranch[branchID]
	result := make([]model.Product, 0)
	for _, p := range products {
		if p.IsFeatured && p.IsAvailable {
			result = append(result, *p)
		}
	}
	return result, nil
}

func (r *ProductMemoryRepository) Update(ctx context.Context, product *model.Product) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	product.UpdatedAt = time.Now()
	r.products[product.ID] = product
	return nil
}

func (r *ProductMemoryRepository) Delete(ctx context.Context, id uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if p, ok := r.products[id]; ok {
		p.IsAvailable = false
		p.UpdatedAt = time.Now()
		return nil
	}
	return ErrNotFound
}

func (r *ProductMemoryRepository) ToggleAvailability(ctx context.Context, id uuid.UUID, isAvailable bool) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if p, ok := r.products[id]; ok {
		p.IsAvailable = isAvailable
		p.UpdatedAt = time.Now()
		return nil
	}
	return ErrNotFound
}

func (r *ProductMemoryRepository) Seed(ctx context.Context) error {
	products := []*model.Product{
		{
			ID:              uuid.MustParse("55555555-5555-5555-5555-555555555551"),
			BranchID:        uuid.MustParse("22222222-2222-2222-2222-222222222222"),
			CategoryID:      uuidPtr(uuid.MustParse("44444444-4444-4444-4444-444444444441")),
			Name:            "Spring Rolls",
			Description:     "Crispy vegetable spring rolls",
			Price:           8.99,
			ImageURL:        "https://example.com/spring-rolls.jpg",
			IsAvailable:     true,
			IsFeatured:      true,
			PreparationTime: 10,
			Allergens:       []string{"wheat", "soy"},
			Rating:          4.5,
			ReviewCount:     120,
		},
		{
			ID:              uuid.MustParse("55555555-5555-5555-5555-555555555552"),
			BranchID:        uuid.MustParse("22222222-2222-2222-2222-222222222222"),
			CategoryID:      uuidPtr(uuid.MustParse("44444444-4444-4444-4444-444444444441")),
			Name:            "Garlic Bread",
			Description:     "Toasted with garlic butter",
			Price:           5.99,
			ImageURL:        "https://example.com/garlic-bread.jpg",
			IsAvailable:     true,
			IsFeatured:      false,
			PreparationTime: 5,
			Allergens:       []string{"wheat", "dairy"},
			Rating:          4.2,
			ReviewCount:     85,
		},
		{
			ID:              uuid.MustParse("55555555-5555-5555-5555-555555555553"),
			BranchID:        uuid.MustParse("22222222-2222-2222-2222-222222222222"),
			CategoryID:      uuidPtr(uuid.MustParse("44444444-4444-4444-4444-444444444442")),
			Name:            "Grilled Salmon",
			Description:     "Fresh Atlantic salmon with herbs",
			Price:           24.99,
			ImageURL:        "https://example.com/salmon.jpg",
			IsAvailable:     true,
			IsFeatured:      true,
			PreparationTime: 20,
			Allergens:       []string{"fish"},
			Rating:          4.8,
			ReviewCount:     200,
		},
		{
			ID:              uuid.MustParse("55555555-5555-5555-5555-555555555554"),
			BranchID:        uuid.MustParse("22222222-2222-2222-2222-222222222222"),
			CategoryID:      uuidPtr(uuid.MustParse("44444444-4444-4444-4444-444444444442")),
			Name:            "Ribeye Steak",
			Description:     "12oz prime ribeye, cooked to perfection",
			Price:           34.99,
			ImageURL:        "https://example.com/ribeye.jpg",
			IsAvailable:     true,
			IsFeatured:      true,
			PreparationTime: 25,
			Allergens:       []string{},
			Rating:          4.9,
			ReviewCount:     310,
		},
		{
			ID:              uuid.MustParse("55555555-5555-5555-5555-555555555555"),
			BranchID:        uuid.MustParse("22222222-2222-2222-2222-222222222222"),
			CategoryID:      uuidPtr(uuid.MustParse("44444444-4444-4444-4444-444444444443")),
			Name:            "Chocolate Cake",
			Description:     "Rich dark chocolate layer cake",
			Price:           9.99,
			ImageURL:        "https://example.com/chocolate-cake.jpg",
			IsAvailable:     true,
			IsFeatured:      false,
			PreparationTime: 5,
			Allergens:       []string{"wheat", "dairy", "eggs"},
			Rating:          4.6,
			ReviewCount:     150,
		},
		{
			ID:              uuid.MustParse("55555555-5555-5555-5555-555555555556"),
			BranchID:        uuid.MustParse("22222222-2222-2222-2222-222222222222"),
			CategoryID:      uuidPtr(uuid.MustParse("44444444-4444-4444-4444-444444444444")),
			Name:            "Fresh Lemonade",
			Description:     "Homemade with fresh lemons",
			Price:           4.99,
			ImageURL:        "https://example.com/lemonade.jpg",
			IsAvailable:     true,
			IsFeatured:      false,
			PreparationTime: 2,
			Allergens:       []string{},
			Rating:          4.3,
			ReviewCount:     90,
		},
	}

	now := time.Now()
	for _, p := range products {
		p.CreatedAt = now
		p.UpdatedAt = now
		if err := r.Create(ctx, p); err != nil {
			return err
		}
	}
	return nil
}

func uuidPtr(u uuid.UUID) *uuid.UUID {
	return &u
}

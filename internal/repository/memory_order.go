package repository

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/greenpos/backend/internal/model"
)

// OrderMemoryRepository is an in-memory implementation of OrderRepository
type OrderMemoryRepository struct {
	mu       sync.RWMutex
	orders   map[uuid.UUID]*model.Order
	items    map[uuid.UUID][]*model.OrderItem
	byBranch map[uuid.UUID][]*model.Order
	byTable  map[uuid.UUID]*model.Order
}

// NewOrderMemoryRepository creates a new in-memory order repository
func NewOrderMemoryRepository() *OrderMemoryRepository {
	return &OrderMemoryRepository{
		orders:   make(map[uuid.UUID]*model.Order),
		items:    make(map[uuid.UUID][]*model.OrderItem),
		byBranch: make(map[uuid.UUID][]*model.Order),
		byTable:  make(map[uuid.UUID]*model.Order),
	}
}

func (r *OrderMemoryRepository) Create(ctx context.Context, order *model.Order) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	order.CreatedAt = time.Now()
	order.UpdatedAt = time.Now()
	if order.ID == uuid.Nil {
		order.ID = uuid.New()
	}

	r.orders[order.ID] = order
	r.byBranch[order.BranchID] = append(r.byBranch[order.BranchID], order)
	if order.TableID != nil {
		r.byTable[*order.TableID] = order
	}
	return nil
}

func (r *OrderMemoryRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Order, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if o, ok := r.orders[id]; ok {
		return o, nil
	}
	return nil, ErrNotFound
}

func (r *OrderMemoryRepository) GetByBranch(ctx context.Context, branchID uuid.UUID) ([]model.Order, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	orders := r.byBranch[branchID]
	result := make([]model.Order, len(orders))
	for i, o := range orders {
		result[i] = *o
	}
	return result, nil
}

func (r *OrderMemoryRepository) ListByBranch(ctx context.Context, branchID uuid.UUID, status *model.OrderStatus, limit *int) ([]model.Order, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	orders := r.byBranch[branchID]
	result := make([]model.Order, 0, len(orders))
	for _, o := range orders {
		if status != nil && o.Status != *status {
			continue
		}
		result = append(result, *o)
	}
	
	// Apply limit
	if limit != nil && len(result) > *limit {
		result = result[:*limit]
	}
	return result, nil
}

func (r *OrderMemoryRepository) Update(ctx context.Context, order *model.Order) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	order.UpdatedAt = time.Now()
	r.orders[order.ID] = order
	return nil
}

func (r *OrderMemoryRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status model.OrderStatus) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if o, ok := r.orders[id]; ok {
		o.Status = status
		o.UpdatedAt = time.Now()
		return nil
	}
	return ErrNotFound
}

func (r *OrderMemoryRepository) AddItem(ctx context.Context, item *model.OrderItem) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	item.CreatedAt = time.Now()
	if item.ID == uuid.Nil {
		item.ID = uuid.New()
	}

	r.items[item.OrderID] = append(r.items[item.OrderID], item)
	return nil
}

func (r *OrderMemoryRepository) GetItems(ctx context.Context, orderID uuid.UUID) ([]model.OrderItem, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	items := r.items[orderID]
	result := make([]model.OrderItem, len(items))
	for i, item := range items {
		result[i] = *item
	}
	return result, nil
}

func (r *OrderMemoryRepository) GetItem(ctx context.Context, itemID uuid.UUID) (*model.OrderItem, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, items := range r.items {
		for _, item := range items {
			if item.ID == itemID {
				return item, nil
			}
		}
	}
	return nil, ErrNotFound
}

func (r *OrderMemoryRepository) RemoveItem(ctx context.Context, itemID uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for orderID, items := range r.items {
		for i, item := range items {
			if item.ID == itemID {
				// Remove item
				r.items[orderID] = append(items[:i], items[i+1:]...)
				return nil
			}
		}
	}
	return ErrNotFound
}

func (r *OrderMemoryRepository) GetActiveByTable(ctx context.Context, tableID uuid.UUID) (*model.Order, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if o, ok := r.byTable[tableID]; ok {
		if o.Status != model.OrderPaid && o.Status != model.OrderCancelled {
			return o, nil
		}
	}
	return nil, ErrNotFound
}

func (r *OrderMemoryRepository) GetActive(ctx context.Context, branchID uuid.UUID) ([]*model.Order, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	orders := r.byBranch[branchID]
	result := make([]*model.Order, 0, len(orders))
	for _, o := range orders {
		if o.Status != model.OrderPaid && o.Status != model.OrderCancelled {
			result = append(result, o)
		}
	}
	return result, nil
}

func (r *OrderMemoryRepository) Seed(ctx context.Context) error {
	now := time.Now()
	orders := []*model.Order{
		{
			ID:            uuid.MustParse("77777777-7777-7777-7777-777777777771"),
			BranchID:      uuid.MustParse("22222222-2222-2222-2222-222222222222"),
			TableID:       uuidPtr(uuid.MustParse("66666666-6666-6666-6666-666666666662")),
			UserID:        uuid.MustParse("33333333-3333-3333-3333-333333333333"),
			CustomerName:  "Alice Johnson",
			CustomerPhone: "+1-555-0101",
			Status:        model.OrderPreparing,
			Subtotal:      45.98,
			Tax:           7.36,
			Discount:      0,
			Total:         53.34,
			Notes:         "No onions please",
			CreatedAt:     now.Add(-30 * time.Minute),
			UpdatedAt:     now,
		},
		{
			ID:            uuid.MustParse("77777777-7777-7777-7777-777777777772"),
			BranchID:      uuid.MustParse("22222222-2222-2222-2222-222222222222"),
			TableID:       uuidPtr(uuid.MustParse("66666666-6666-6666-6666-666666666663")),
			UserID:        uuid.MustParse("33333333-3333-3333-3333-333333333335"),
			CustomerName:  "Bob Smith",
			CustomerPhone: "+1-555-0102",
			Status:        model.OrderPending,
			Subtotal:      29.98,
			Tax:           4.80,
			Discount:      0,
			Total:         34.78,
			Notes:         "",
			CreatedAt:     now.Add(-10 * time.Minute),
			UpdatedAt:     now,
		},
	}

	for _, o := range orders {
		if err := r.Create(ctx, o); err != nil {
			return err
		}
	}

	// Add items for the first order
	items := []*model.OrderItem{
		{
			ID:         uuid.MustParse("88888888-8888-8888-8888-888888888881"),
			OrderID:    uuid.MustParse("77777777-7777-7777-7777-777777777771"),
			ProductID:  uuid.MustParse("55555555-5555-5555-5555-555555555553"),
			Quantity:   1,
			UnitPrice:  24.99,
			TotalPrice: 24.99,
			Notes:      "",
			CreatedAt:  now.Add(-30 * time.Minute),
		},
		{
			ID:         uuid.MustParse("88888888-8888-8888-8888-888888888882"),
			OrderID:    uuid.MustParse("77777777-7777-7777-7777-777777777771"),
			ProductID:  uuid.MustParse("55555555-5555-5555-5555-555555555552"),
			Quantity:   1,
			UnitPrice:  5.99,
			TotalPrice: 5.99,
			Notes:      "Extra garlic butter",
			CreatedAt:  now.Add(-30 * time.Minute),
		},
		{
			ID:         uuid.MustParse("88888888-8888-8888-8888-888888888883"),
			OrderID:    uuid.MustParse("77777777-7777-7777-7777-777777777771"),
			ProductID:  uuid.MustParse("55555555-5555-5555-5555-555555555556"),
			Quantity:   3,
			UnitPrice:  4.99,
			TotalPrice: 14.97,
			Notes:      "",
			CreatedAt:  now.Add(-30 * time.Minute),
		},
	}

	for _, item := range items {
		if err := r.AddItem(ctx, item); err != nil {
			return err
		}
	}

	return nil
}

package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/greenpos/backend/internal/model"
	"gorm.io/gorm"
)

// GormOrderRepository is a GORM-backed OrderRepository
type GormOrderRepository struct {
	db *gorm.DB
}

type GormOrderModel struct {
	ID            uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	BranchID      uuid.UUID  `gorm:"type:uuid;not null;index" json:"branchId"`
	TableID       *uuid.UUID `gorm:"type:uuid;index" json:"tableId"`
	UserID        uuid.UUID  `gorm:"type:uuid;not null;index" json:"userId"`
	CustomerName  string     `gorm:"type:varchar(255)" json:"customerName"`
	CustomerPhone string     `gorm:"type:varchar(50)" json:"customerPhone"`
	Status        string     `gorm:"type:varchar(50);default:'PENDING';index" json:"status"`
	Subtotal      float64    `gorm:"type:decimal(10,2);default:0" json:"subtotal"`
	Tax           float64    `gorm:"type:decimal(10,2);default:0" json:"tax"`
	Discount      float64    `gorm:"type:decimal(10,2);default:0" json:"discount"`
	Total         float64    `gorm:"type:decimal(10,2);default:0" json:"total"`
	Notes         string     `gorm:"type:text" json:"notes"`
	CreatedAt     time.Time  `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt     time.Time  `gorm:"autoUpdateTime" json:"updatedAt"`
}

func (GormOrderModel) TableName() string { return "orders" }

type GormOrderItemModel struct {
	ID         uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	OrderID    uuid.UUID `gorm:"type:uuid;not null;index" json:"orderId"`
	ProductID  uuid.UUID `gorm:"type:uuid;not null" json:"productId"`
	Quantity   int       `gorm:"not null;default:1" json:"quantity"`
	UnitPrice  float64   `gorm:"type:decimal(10,2);not null" json:"unitPrice"`
	TotalPrice float64   `gorm:"type:decimal(10,2);not null" json:"totalPrice"`
	Notes      string    `gorm:"type:text" json:"notes"`
	CreatedAt  time.Time `gorm:"autoCreateTime" json:"createdAt"`
}

func (GormOrderItemModel) TableName() string { return "order_items" }

func toOrderModel(m *GormOrderModel) *model.Order {
	return &model.Order{
		ID:            m.ID,
		BranchID:      m.BranchID,
		TableID:       m.TableID,
		UserID:        m.UserID,
		CustomerName:  m.CustomerName,
		CustomerPhone: m.CustomerPhone,
		Status:        model.OrderStatus(m.Status),
		Subtotal:      m.Subtotal,
		Tax:           m.Tax,
		Discount:      m.Discount,
		Total:         m.Total,
		Notes:         m.Notes,
		CreatedAt:     m.CreatedAt,
		UpdatedAt:     m.UpdatedAt,
	}
}

func toOrderGorm(m *model.Order) *GormOrderModel {
	return &GormOrderModel{
		ID:            m.ID,
		BranchID:      m.BranchID,
		TableID:       m.TableID,
		UserID:        m.UserID,
		CustomerName:  m.CustomerName,
		CustomerPhone: m.CustomerPhone,
		Status:        string(m.Status),
		Subtotal:      m.Subtotal,
		Tax:           m.Tax,
		Discount:      m.Discount,
		Total:         m.Total,
		Notes:         m.Notes,
		CreatedAt:     m.CreatedAt,
		UpdatedAt:     m.UpdatedAt,
	}
}

func toOrderItemModel(m *GormOrderItemModel) *model.OrderItem {
	return &model.OrderItem{
		ID:         m.ID,
		OrderID:    m.OrderID,
		ProductID:  m.ProductID,
		Quantity:   m.Quantity,
		UnitPrice:  m.UnitPrice,
		TotalPrice: m.TotalPrice,
		Notes:      m.Notes,
		CreatedAt:  m.CreatedAt,
	}
}

func toOrderItemGorm(m *model.OrderItem) *GormOrderItemModel {
	return &GormOrderItemModel{
		ID:         m.ID,
		OrderID:    m.OrderID,
		ProductID:  m.ProductID,
		Quantity:   m.Quantity,
		UnitPrice:  m.UnitPrice,
		TotalPrice: m.TotalPrice,
		Notes:      m.Notes,
		CreatedAt:  m.CreatedAt,
	}
}

// NewGormOrderRepository creates a GORM-backed order repository
func NewGormOrderRepository(db *gorm.DB) *GormOrderRepository {
	return &GormOrderRepository{db: db}
}

func (r *GormOrderRepository) Create(ctx context.Context, order *model.Order) error {
	if order.ID == uuid.Nil {
		order.ID = uuid.New()
	}
	order.CreatedAt = time.Now()
	order.UpdatedAt = time.Now()
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(toOrderGorm(order)).Error; err != nil {
			return err
		}
		return nil
	})
}

func (r *GormOrderRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Order, error) {
	var m GormOrderModel
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&m).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return toOrderModel(&m), nil
}

func (r *GormOrderRepository) GetByBranch(ctx context.Context, branchID uuid.UUID) ([]model.Order, error) {
	var models []GormOrderModel
	if err := r.db.WithContext(ctx).
		Where("branch_id = ?", branchID).
		Order("created_at DESC").
		Find(&models).Error; err != nil {
		return nil, err
	}
	result := make([]model.Order, len(models))
	for i := range models {
		result[i] = *toOrderModel(&models[i])
	}
	return result, nil
}

func (r *GormOrderRepository) ListByBranch(ctx context.Context, branchID uuid.UUID, status *model.OrderStatus, limit *int) ([]model.Order, error) {
	var models []GormOrderModel
	query := r.db.WithContext(ctx).Where("branch_id = ?", branchID)
	if status != nil {
		query = query.Where("status = ?", string(*status))
	}
	if limit != nil {
		query = query.Limit(*limit)
	}
	if err := query.Order("created_at DESC").Find(&models).Error; err != nil {
		return nil, err
	}
	result := make([]model.Order, len(models))
	for i := range models {
		result[i] = *toOrderModel(&models[i])
	}
	return result, nil
}

func (r *GormOrderRepository) Update(ctx context.Context, order *model.Order) error {
	order.UpdatedAt = time.Now()
	return r.db.WithContext(ctx).Save(toOrderGorm(order)).Error
}

func (r *GormOrderRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status model.OrderStatus) error {
	return r.db.WithContext(ctx).
		Model(&GormOrderModel{}).
		Where("id = ?", id).
		Update("status", string(status)).Error
}

func (r *GormOrderRepository) AddItem(ctx context.Context, item *model.OrderItem) error {
	if item.ID == uuid.Nil {
		item.ID = uuid.New()
	}
	item.CreatedAt = time.Now()
	return r.db.WithContext(ctx).Create(toOrderItemGorm(item)).Error
}

func (r *GormOrderRepository) GetItems(ctx context.Context, orderID uuid.UUID) ([]model.OrderItem, error) {
	var models []GormOrderItemModel
	if err := r.db.WithContext(ctx).
		Where("order_id = ?", orderID).
		Find(&models).Error; err != nil {
		return nil, err
	}
	result := make([]model.OrderItem, len(models))
	for i := range models {
		result[i] = *toOrderItemModel(&models[i])
	}
	return result, nil
}

func (r *GormOrderRepository) GetItem(ctx context.Context, itemID uuid.UUID) (*model.OrderItem, error) {
	var m GormOrderItemModel
	if err := r.db.WithContext(ctx).Where("id = ?", itemID).First(&m).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return toOrderItemModel(&m), nil
}

func (r *GormOrderRepository) RemoveItem(ctx context.Context, itemID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("id = ?", itemID).
		Delete(&GormOrderItemModel{}).Error
}

func (r *GormOrderRepository) GetActiveByTable(ctx context.Context, tableID uuid.UUID) (*model.Order, error) {
	var m GormOrderModel
	if err := r.db.WithContext(ctx).
		Where("table_id = ? AND status NOT IN (?, ?)", tableID, "PAID", "CANCELLED").
		First(&m).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return toOrderModel(&m), nil
}

func (r *GormOrderRepository) GetActive(ctx context.Context, branchID uuid.UUID) ([]*model.Order, error) {
	var models []GormOrderModel
	if err := r.db.WithContext(ctx).
		Where("branch_id = ? AND status NOT IN (?, ?)", branchID, "PAID", "CANCELLED").
		Order("created_at DESC").
		Find(&models).Error; err != nil {
		return nil, err
	}
	result := make([]*model.Order, len(models))
	for i := range models {
		result[i] = toOrderModel(&models[i])
	}
	return result, nil
}

func (r *GormOrderRepository) Seed(ctx context.Context) error {
	var count int64
	r.db.WithContext(ctx).Model(&GormOrderModel{}).Count(&count)
	if count > 0 {
		return nil
	}

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
		if err := r.db.WithContext(ctx).Create(toOrderGorm(o)).Error; err != nil {
			return err
		}
	}

	// Add items
	items := []GormOrderItemModel{
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
		if err := r.db.WithContext(ctx).Create(&item).Error; err != nil {
			return err
		}
	}

	return nil
}

// Helper to create a uuid.Pointer
func uuidPtr(u uuid.UUID) *uuid.UUID {
	return &u
}

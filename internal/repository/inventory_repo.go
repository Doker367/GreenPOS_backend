package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/greenpos/backend/internal/model"
	"gorm.io/gorm"
)

// ============ GORM Models ============

// GormInventoryItemModel is the GORM model for inventory_items
type GormInventoryItemModel struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	BranchID     uuid.UUID `gorm:"type:uuid;not null;index" json:"branchId"`
	Name         string    `gorm:"type:varchar(255);not null" json:"name"`
	Unit         string    `gorm:"type:varchar(50);not null" json:"unit"`
	CurrentStock float64   `gorm:"type:decimal(10,2);default:0" json:"currentStock"`
	MinStock     float64   `gorm:"type:decimal(10,2);default:0" json:"minStock"`
	Cost         float64   `gorm:"type:decimal(10,2);default:0" json:"cost"`
	IsActive     bool      `gorm:"default:true" json:"isActive"`
	CreatedAt    time.Time `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime" json:"updatedAt"`
}

func (GormInventoryItemModel) TableName() string { return "inventory_items" }

// GormInventoryMovementModel is the GORM model for inventory_movements
type GormInventoryMovementModel struct {
	ID              uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	InventoryItemID uuid.UUID `gorm:"type:uuid;not null;index" json:"inventoryItemId"`
	Type            string    `gorm:"type:varchar(20);not null" json:"type"`
	Quantity        float64   `gorm:"type:decimal(10,2);not null" json:"quantity"`
	Reason          string    `gorm:"type:varchar(255)" json:"reason"`
	UserID          *uuid.UUID `gorm:"type:uuid" json:"userId"`
	CreatedAt       time.Time `gorm:"autoCreateTime" json:"createdAt"`
}

func (GormInventoryMovementModel) TableName() string { return "inventory_movements" }

// ============ Repository Interface ============

// InventoryItemRepositoryInterface defines operations for inventory items
type InventoryItemRepositoryInterface interface {
	Create(ctx context.Context, item *model.InventoryItem) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.InventoryItem, error)
	GetByBranch(ctx context.Context, branchID uuid.UUID) ([]model.InventoryItem, error)
	GetLowStock(ctx context.Context, branchID uuid.UUID) ([]model.InventoryItem, error)
	Update(ctx context.Context, item *model.InventoryItem) error
	UpdateStock(ctx context.Context, id uuid.UUID, newStock float64) error
	Delete(ctx context.Context, id uuid.UUID) error
	Seed(ctx context.Context) error
}

// InventoryMovementRepositoryInterface defines operations for inventory movements
type InventoryMovementRepositoryInterface interface {
	Create(ctx context.Context, movement *model.InventoryMovement) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.InventoryMovement, error)
	GetByItem(ctx context.Context, itemID uuid.UUID, limit int) ([]model.InventoryMovement, error)
	Seed(ctx context.Context) error
}

// ============ GORM Repository Implementation ============

// GormInventoryItemRepository implements InventoryItemRepositoryInterface using GORM
type GormInventoryItemRepository struct {
	db *gorm.DB
}

// NewGormInventoryItemRepository creates a new GORM-backed inventory item repository
func NewGormInventoryItemRepository(db *gorm.DB) *GormInventoryItemRepository {
	return &GormInventoryItemRepository{db: db}
}

func (r *GormInventoryItemRepository) toModel(m *GormInventoryItemModel) *model.InventoryItem {
	return &model.InventoryItem{
		ID:           m.ID,
		BranchID:     m.BranchID,
		Name:         m.Name,
		Unit:         m.Unit,
		CurrentStock: m.CurrentStock,
		MinStock:     m.MinStock,
		Cost:         m.Cost,
		IsActive:     m.IsActive,
		CreatedAt:    m.CreatedAt,
		UpdatedAt:    m.UpdatedAt,
	}
}

func (r *GormInventoryItemRepository) toGorm(m *model.InventoryItem) *GormInventoryItemModel {
	return &GormInventoryItemModel{
		ID:           m.ID,
		BranchID:     m.BranchID,
		Name:         m.Name,
		Unit:         m.Unit,
		CurrentStock: m.CurrentStock,
		MinStock:     m.MinStock,
		Cost:         m.Cost,
		IsActive:     m.IsActive,
		CreatedAt:    m.CreatedAt,
		UpdatedAt:    m.UpdatedAt,
	}
}

func (r *GormInventoryItemRepository) Create(ctx context.Context, item *model.InventoryItem) error {
	if item.ID == uuid.Nil {
		item.ID = uuid.New()
	}
	item.CreatedAt = time.Now()
	item.UpdatedAt = time.Now()
	return r.db.WithContext(ctx).Create(r.toGorm(item)).Error
}

func (r *GormInventoryItemRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.InventoryItem, error) {
	var m GormInventoryItemModel
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&m).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return r.toModel(&m), nil
}

func (r *GormInventoryItemRepository) GetByBranch(ctx context.Context, branchID uuid.UUID) ([]model.InventoryItem, error) {
	var models []GormInventoryItemModel
	if err := r.db.WithContext(ctx).
		Where("branch_id = ? AND is_active = ?", branchID, true).
		Order("name ASC").
		Find(&models).Error; err != nil {
		return nil, err
	}
	result := make([]model.InventoryItem, len(models))
	for i := range models {
		result[i] = *r.toModel(&models[i])
	}
	return result, nil
}

func (r *GormInventoryItemRepository) GetLowStock(ctx context.Context, branchID uuid.UUID) ([]model.InventoryItem, error) {
	var models []GormInventoryItemModel
	if err := r.db.WithContext(ctx).
		Where("branch_id = ? AND is_active = ? AND current_stock < min_stock", branchID, true).
		Order("current_stock ASC").
		Find(&models).Error; err != nil {
		return nil, err
	}
	result := make([]model.InventoryItem, len(models))
	for i := range models {
		result[i] = *r.toModel(&models[i])
	}
	return result, nil
}

func (r *GormInventoryItemRepository) Update(ctx context.Context, item *model.InventoryItem) error {
	item.UpdatedAt = time.Now()
	return r.db.WithContext(ctx).Save(r.toGorm(item)).Error
}

func (r *GormInventoryItemRepository) UpdateStock(ctx context.Context, id uuid.UUID, newStock float64) error {
	return r.db.WithContext(ctx).
		Model(&GormInventoryItemModel{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"current_stock": newStock,
			"updated_at":    time.Now(),
		}).Error
}

func (r *GormInventoryItemRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).
		Model(&GormInventoryItemModel{}).
		Where("id = ?", id).
		Update("is_active", false).Error
}

func (r *GormInventoryItemRepository) Seed(ctx context.Context) error {
	var count int64
	r.db.WithContext(ctx).Model(&GormInventoryItemModel{}).Count(&count)
	if count > 0 {
		return nil
	}

	branchID := uuid.MustParse("22222222-2222-2222-2222-222222222222")

	items := []*model.InventoryItem{
		{ID: uuid.MustParse("66666666-6666-6666-6666-666666666661"), BranchID: branchID, Name: "Tomates", Unit: "kg", CurrentStock: 25.0, MinStock: 10.0, Cost: 45.0, IsActive: true},
		{ID: uuid.MustParse("66666666-6666-6666-6666-666666666662"), BranchID: branchID, Name: "Cebolla", Unit: "kg", CurrentStock: 8.0, MinStock: 10.0, Cost: 35.0, IsActive: true},
		{ID: uuid.MustParse("66666666-6666-6666-6666-666666666663"), BranchID: branchID, Name: "Ajo", Unit: "kg", CurrentStock: 3.0, MinStock: 2.0, Cost: 120.0, IsActive: true},
		{ID: uuid.MustParse("66666666-6666-6666-6666-666666666664"), BranchID: branchID, Name: "Aceite de oliva", Unit: "litros", CurrentStock: 5.0, MinStock: 3.0, Cost: 180.0, IsActive: true},
		{ID: uuid.MustParse("66666666-6666-6666-6666-666666666665"), BranchID: branchID, Name: "Sal", Unit: "kg", CurrentStock: 15.0, MinStock: 5.0, Cost: 15.0, IsActive: true},
		{ID: uuid.MustParse("66666666-6666-6666-6666-666666666666"), BranchID: branchID, Name: "Pimienta", Unit: "kg", CurrentStock: 2.0, MinStock: 1.0, Cost: 250.0, IsActive: true},
		{ID: uuid.MustParse("66666666-6666-6666-6666-666666666667"), BranchID: branchID, Name: "Harina de trigo", Unit: "kg", CurrentStock: 30.0, MinStock: 15.0, Cost: 25.0, IsActive: true},
		{ID: uuid.MustParse("66666666-6666-6666-6666-666666666668"), BranchID: branchID, Name: "Huevos", Unit: "piezas", CurrentStock: 48.0, MinStock: 50.0, Cost: 1.5, IsActive: true},
		{ID: uuid.MustParse("66666666-6666-6666-6666-666666666669"), BranchID: branchID, Name: "Queso parmesano", Unit: "kg", CurrentStock: 4.0, MinStock: 2.0, Cost: 350.0, IsActive: true},
		{ID: uuid.MustParse("66666666-6666-6666-6666-66666666666a"), BranchID: branchID, Name: "Albahaca fresca", Unit: "kg", CurrentStock: 0.5, MinStock: 1.0, Cost: 200.0, IsActive: true},
	}

	for _, item := range items {
		item.CreatedAt = time.Now()
		item.UpdatedAt = time.Now()
		if err := r.Create(ctx, item); err != nil {
			return err
		}
	}
	return nil
}

// ============ GORM Inventory Movement Repository ============

// GormInventoryMovementRepository implements InventoryMovementRepositoryInterface using GORM
type GormInventoryMovementRepository struct {
	db *gorm.DB
}

// NewGormInventoryMovementRepository creates a new GORM-backed inventory movement repository
func NewGormInventoryMovementRepository(db *gorm.DB) *GormInventoryMovementRepository {
	return &GormInventoryMovementRepository{db: db}
}

func (r *GormInventoryMovementRepository) toModel(m *GormInventoryMovementModel) *model.InventoryMovement {
	return &model.InventoryMovement{
		ID:              m.ID,
		InventoryItemID: m.InventoryItemID,
		Type:            model.InventoryMovementType(m.Type),
		Quantity:        m.Quantity,
		Reason:          m.Reason,
		UserID:          m.UserID,
		CreatedAt:       m.CreatedAt,
	}
}

func (r *GormInventoryMovementRepository) toGorm(m *model.InventoryMovement) *GormInventoryMovementModel {
	return &GormInventoryMovementModel{
		ID:              m.ID,
		InventoryItemID: m.InventoryItemID,
		Type:            string(m.Type),
		Quantity:        m.Quantity,
		Reason:          m.Reason,
		UserID:          m.UserID,
		CreatedAt:       m.CreatedAt,
	}
}

func (r *GormInventoryMovementRepository) Create(ctx context.Context, movement *model.InventoryMovement) error {
	if movement.ID == uuid.Nil {
		movement.ID = uuid.New()
	}
	movement.CreatedAt = time.Now()
	return r.db.WithContext(ctx).Create(r.toGorm(movement)).Error
}

func (r *GormInventoryMovementRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.InventoryMovement, error) {
	var m GormInventoryMovementModel
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&m).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return r.toModel(&m), nil
}

func (r *GormInventoryMovementRepository) GetByItem(ctx context.Context, itemID uuid.UUID, limit int) ([]model.InventoryMovement, error) {
	var models []GormInventoryMovementModel
	query := r.db.WithContext(ctx).
		Where("inventory_item_id = ?", itemID).
		Order("created_at DESC")
	
	if limit > 0 {
		query = query.Limit(limit)
	}
	
	if err := query.Find(&models).Error; err != nil {
		return nil, err
	}
	result := make([]model.InventoryMovement, len(models))
	for i := range models {
		result[i] = *r.toModel(&models[i])
	}
	return result, nil
}

func (r *GormInventoryMovementRepository) Seed(ctx context.Context) error {
	// No seed data needed for movements - they are created dynamically
	return nil
}

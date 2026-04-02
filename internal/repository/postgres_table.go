package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/greenpos/backend/internal/model"
	"gorm.io/gorm"
)

// GormTableRepository is a GORM-backed TableRepository
// Note: SQL table is named "pos_tables" to avoid reserved word conflict
type GormTableRepository struct {
	db *gorm.DB
}

type GormTableModel struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	BranchID  uuid.UUID `gorm:"type:uuid;not null;index" json:"branchId"`
	Number    string    `gorm:"type:varchar(20);not null" json:"number"`
	Capacity  int       `gorm:"default:4" json:"capacity"`
	Status    string    `gorm:"type:varchar(50);default:'AVAILABLE'" json:"status"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updatedAt"`
}

func (GormTableModel) TableName() string { return "pos_tables" }

func toTableModel(m *GormTableModel) *model.RestaurantTable {
	return &model.RestaurantTable{
		ID:        m.ID,
		BranchID:  m.BranchID,
		Number:    m.Number,
		Capacity:  m.Capacity,
		Status:    model.TableStatus(m.Status),
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
}

func toTableGorm(m *model.RestaurantTable) *GormTableModel {
	return &GormTableModel{
		ID:        m.ID,
		BranchID:  m.BranchID,
		Number:    m.Number,
		Capacity:  m.Capacity,
		Status:    string(m.Status),
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
}

// NewGormTableRepository creates a GORM-backed table repository
func NewGormTableRepository(db *gorm.DB) *GormTableRepository {
	return &GormTableRepository{db: db}
}

func (r *GormTableRepository) Create(ctx context.Context, table *model.RestaurantTable) error {
	if table.ID == uuid.Nil {
		table.ID = uuid.New()
	}
	if table.Status == "" {
		table.Status = model.TableAvailable
	}
	table.CreatedAt = time.Now()
	table.UpdatedAt = time.Now()
	return r.db.WithContext(ctx).Create(toTableGorm(table)).Error
}

func (r *GormTableRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.RestaurantTable, error) {
	var m GormTableModel
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&m).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return toTableModel(&m), nil
}

func (r *GormTableRepository) GetByBranch(ctx context.Context, branchID uuid.UUID) ([]model.RestaurantTable, error) {
	var models []GormTableModel
	if err := r.db.WithContext(ctx).
		Where("branch_id = ?", branchID).
		Order("number ASC").
		Find(&models).Error; err != nil {
		return nil, err
	}
	result := make([]model.RestaurantTable, len(models))
	for i := range models {
		result[i] = *toTableModel(&models[i])
	}
	return result, nil
}

func (r *GormTableRepository) ListByBranch(ctx context.Context, branchID uuid.UUID) ([]model.RestaurantTable, error) {
	return r.GetByBranch(ctx, branchID)
}

func (r *GormTableRepository) Update(ctx context.Context, table *model.RestaurantTable) error {
	table.UpdatedAt = time.Now()
	return r.db.WithContext(ctx).Save(toTableGorm(table)).Error
}

func (r *GormTableRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status model.TableStatus) error {
	return r.db.WithContext(ctx).
		Model(&GormTableModel{}).
		Where("id = ?", id).
		Update("status", string(status)).Error
}

func (r *GormTableRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&GormTableModel{}, "id = ?", id).Error
}

func (r *GormTableRepository) GetAvailable(ctx context.Context, branchID uuid.UUID) ([]model.RestaurantTable, error) {
	var models []GormTableModel
	if err := r.db.WithContext(ctx).
		Where("branch_id = ? AND status = ?", branchID, "AVAILABLE").
		Order("number ASC").
		Find(&models).Error; err != nil {
		return nil, err
	}
	result := make([]model.RestaurantTable, len(models))
	for i := range models {
		result[i] = *toTableModel(&models[i])
	}
	return result, nil
}

func (r *GormTableRepository) Seed(ctx context.Context) error {
	var count int64
	r.db.WithContext(ctx).Model(&GormTableModel{}).Count(&count)
	if count > 0 {
		return nil
	}

	tables := []*model.RestaurantTable{
		{
			ID:        uuid.MustParse("66666666-6666-6666-6666-666666666661"),
			BranchID:  uuid.MustParse("22222222-2222-2222-2222-222222222222"),
			Number:    "1",
			Capacity:  2,
			Status:    model.TableAvailable,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:        uuid.MustParse("66666666-6666-6666-6666-666666666662"),
			BranchID:  uuid.MustParse("22222222-2222-2222-2222-222222222222"),
			Number:    "2",
			Capacity:  4,
			Status:    model.TableAvailable,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:        uuid.MustParse("66666666-6666-6666-6666-666666666663"),
			BranchID:  uuid.MustParse("22222222-2222-2222-2222-222222222222"),
			Number:    "3",
			Capacity:  6,
			Status:    model.TableAvailable,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:        uuid.MustParse("66666666-6666-6666-6666-666666666664"),
			BranchID:  uuid.MustParse("22222222-2222-2222-2222-222222222222"),
			Number:    "4",
			Capacity:  8,
			Status:    model.TableAvailable,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}

	for _, t := range tables {
		if err := r.Create(ctx, t); err != nil {
			return err
		}
	}
	return nil
}

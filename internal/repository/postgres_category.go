package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/greenpos/backend/internal/model"
	"gorm.io/gorm"
)

// GormCategoryRepository is a GORM-backed CategoryRepository
type GormCategoryRepository struct {
	db *gorm.DB
}

type GormCategoryModel struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	BranchID    uuid.UUID `gorm:"type:uuid;not null;index" json:"branchId"`
	Name        string    `gorm:"type:varchar(255);not null" json:"name"`
	Description string    `gorm:"type:text" json:"description"`
	SortOrder   int       `gorm:"default:0" json:"sortOrder"`
	IsActive    bool      `gorm:"default:true" json:"isActive"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime" json:"updatedAt"`
}

func (GormCategoryModel) TableName() string { return "categories" }

func toCategoryModel(m *GormCategoryModel) *model.Category {
	return &model.Category{
		ID:          m.ID,
		BranchID:    m.BranchID,
		Name:        m.Name,
		Description: m.Description,
		SortOrder:   m.SortOrder,
		IsActive:    m.IsActive,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}
}

func toCategoryGorm(m *model.Category) *GormCategoryModel {
	return &GormCategoryModel{
		ID:          m.ID,
		BranchID:    m.BranchID,
		Name:        m.Name,
		Description: m.Description,
		SortOrder:   m.SortOrder,
		IsActive:    m.IsActive,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}
}

// NewGormCategoryRepository creates a GORM-backed category repository
func NewGormCategoryRepository(db *gorm.DB) *GormCategoryRepository {
	return &GormCategoryRepository{db: db}
}

func (r *GormCategoryRepository) Create(ctx context.Context, category *model.Category) error {
	if category.ID == uuid.Nil {
		category.ID = uuid.New()
	}
	category.CreatedAt = time.Now()
	category.UpdatedAt = time.Now()
	return r.db.WithContext(ctx).Create(toCategoryGorm(category)).Error
}

func (r *GormCategoryRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Category, error) {
	var m GormCategoryModel
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&m).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return toCategoryModel(&m), nil
}

func (r *GormCategoryRepository) GetByBranch(ctx context.Context, branchID uuid.UUID) ([]model.Category, error) {
	var models []GormCategoryModel
	if err := r.db.WithContext(ctx).
		Where("branch_id = ? AND is_active = ?", branchID, true).
		Order("sort_order ASC").
		Find(&models).Error; err != nil {
		return nil, err
	}
	result := make([]model.Category, len(models))
	for i := range models {
		result[i] = *toCategoryModel(&models[i])
	}
	return result, nil
}

func (r *GormCategoryRepository) Update(ctx context.Context, category *model.Category) error {
	category.UpdatedAt = time.Now()
	return r.db.WithContext(ctx).Save(toCategoryGorm(category)).Error
}

func (r *GormCategoryRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).
		Model(&GormCategoryModel{}).
		Where("id = ?", id).
		Update("is_active", false).Error
}

func (r *GormCategoryRepository) GetMaxSortOrder(ctx context.Context, branchID uuid.UUID) (int, error) {
	var max int
	if err := r.db.WithContext(ctx).
		Model(&GormCategoryModel{}).
		Where("branch_id = ?", branchID).
		Select("COALESCE(MAX(sort_order), 0)").
		Scan(&max).Error; err != nil {
		return 0, err
	}
	return max, nil
}

func (r *GormCategoryRepository) Reorder(ctx context.Context, branchID uuid.UUID, orders []struct {
	ID        uuid.UUID
	SortOrder int
}) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, o := range orders {
			if err := tx.Model(&GormCategoryModel{}).
				Where("id = ? AND branch_id = ?", o.ID, branchID).
				Update("sort_order", o.SortOrder).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *GormCategoryRepository) Seed(ctx context.Context) error {
	var count int64
	r.db.WithContext(ctx).Model(&GormCategoryModel{}).Count(&count)
	if count > 0 {
		return nil
	}

	categories := []*model.Category{
		{
			ID:          uuid.MustParse("44444444-4444-4444-4444-444444444441"),
			BranchID:    uuid.MustParse("22222222-2222-2222-2222-222222222222"),
			Name:        "Appetizers",
			Description: "Start your meal right",
			SortOrder:   1,
			IsActive:    true,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          uuid.MustParse("44444444-4444-4444-4444-444444444442"),
			BranchID:    uuid.MustParse("22222222-2222-2222-2222-222222222222"),
			Name:        "Main Courses",
			Description: "Hearty entrees",
			SortOrder:   2,
			IsActive:    true,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          uuid.MustParse("44444444-4444-4444-4444-444444444443"),
			BranchID:    uuid.MustParse("22222222-2222-2222-2222-222222222222"),
			Name:        "Desserts",
			Description: "Sweet endings",
			SortOrder:   3,
			IsActive:    true,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          uuid.MustParse("44444444-4444-4444-4444-444444444444"),
			BranchID:    uuid.MustParse("22222222-2222-2222-2222-222222222222"),
			Name:        "Beverages",
			Description: "Refreshing drinks",
			SortOrder:   4,
			IsActive:    true,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
	}

	for _, c := range categories {
		if err := r.Create(ctx, c); err != nil {
			return err
		}
	}
	return nil
}

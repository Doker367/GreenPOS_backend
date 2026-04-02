package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/greenpos/backend/internal/model"
	"gorm.io/gorm"
)

// GormBranchRepository is a GORM-backed BranchRepository
type GormBranchRepository struct {
	db *gorm.DB
}

type GormBranchModel struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TenantID  uuid.UUID `gorm:"type:uuid;not null;index" json:"tenantId"`
	Name      string    `gorm:"type:varchar(255);not null" json:"name"`
	Address   string    `gorm:"type:text" json:"address"`
	Phone     string    `gorm:"type:varchar(50)" json:"phone"`
	IsActive  bool      `gorm:"default:true" json:"isActive"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updatedAt"`
}

func (GormBranchModel) TableName() string { return "branches" }

func toBranchModel(m *GormBranchModel) *model.Branch {
	return &model.Branch{
		ID:        m.ID,
		TenantID:  m.TenantID,
		Name:      m.Name,
		Address:   m.Address,
		Phone:     m.Phone,
		IsActive:  m.IsActive,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
}

func toBranchGorm(m *model.Branch) *GormBranchModel {
	return &GormBranchModel{
		ID:        m.ID,
		TenantID:  m.TenantID,
		Name:      m.Name,
		Address:   m.Address,
		Phone:     m.Phone,
		IsActive:  m.IsActive,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
}

// NewGormBranchRepository creates a GORM-backed branch repository
func NewGormBranchRepository(db *gorm.DB) *GormBranchRepository {
	return &GormBranchRepository{db: db}
}

func (r *GormBranchRepository) Create(ctx context.Context, branch *model.Branch) error {
	if branch.ID == uuid.Nil {
		branch.ID = uuid.New()
	}
	branch.CreatedAt = time.Now()
	branch.UpdatedAt = time.Now()
	return r.db.WithContext(ctx).Create(toBranchGorm(branch)).Error
}

func (r *GormBranchRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Branch, error) {
	var m GormBranchModel
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&m).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return toBranchModel(&m), nil
}

func (r *GormBranchRepository) GetByTenant(ctx context.Context, tenantID uuid.UUID) ([]model.Branch, error) {
	var models []GormBranchModel
	if err := r.db.WithContext(ctx).Where("tenant_id = ?", tenantID).Find(&models).Error; err != nil {
		return nil, err
	}
	result := make([]model.Branch, len(models))
	for i := range models {
		result[i] = *toBranchModel(&models[i])
	}
	return result, nil
}

func (r *GormBranchRepository) Update(ctx context.Context, branch *model.Branch) error {
	branch.UpdatedAt = time.Now()
	return r.db.WithContext(ctx).Save(toBranchGorm(branch)).Error
}

func (r *GormBranchRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&GormBranchModel{}, "id = ?", id).Error
}

func (r *GormBranchRepository) Seed(ctx context.Context) error {
	var count int64
	r.db.WithContext(ctx).Model(&GormBranchModel{}).Count(&count)
	if count > 0 {
		return nil
	}

	branch := &model.Branch{
		ID:        uuid.MustParse("22222222-2222-2222-2222-222222222222"),
		TenantID:  uuid.MustParse("11111111-1111-1111-1111-111111111111"),
		Name:      "Downtown Branch",
		Address:   "123 Main Street",
		Phone:     "+1-555-0100",
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	return r.Create(ctx, branch)
}

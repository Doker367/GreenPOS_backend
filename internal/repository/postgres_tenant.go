package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/greenpos/backend/internal/model"
	"gorm.io/gorm"
)

// GormTenantRepository is a GORM-backed TenantRepository
type GormTenantRepository struct {
	db *gorm.DB
}

// GormTenantModel maps to the tenants table with snake_case columns
type GormTenantModel struct {
	ID        uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Name      string         `gorm:"type:varchar(255);not null" json:"name"`
	Slug      string         `gorm:"type:varchar(100);uniqueIndex;not null" json:"slug"`
	Settings  model.JSON     `gorm:"type:jsonb;default:'{}'" json:"settings"`
	IsActive  bool           `gorm:"default:true" json:"isActive"`
	CreatedAt time.Time      `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"updatedAt"`
}

func (GormTenantModel) TableName() string { return "tenants" }

func toTenantModel(m *GormTenantModel) *model.Tenant {
	return &model.Tenant{
		ID:        m.ID,
		Name:      m.Name,
		Slug:      m.Slug,
		Settings:  m.Settings,
		IsActive:  m.IsActive,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
}

func toTenantGorm(m *model.Tenant) *GormTenantModel {
	return &GormTenantModel{
		ID:        m.ID,
		Name:      m.Name,
		Slug:      m.Slug,
		Settings:  m.Settings,
		IsActive:  m.IsActive,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
}

// NewGormTenantRepository creates a GORM-backed tenant repository
func NewGormTenantRepository(db *gorm.DB) *GormTenantRepository {
	return &GormTenantRepository{db: db}
}

func (r *GormTenantRepository) Create(ctx context.Context, tenant *model.Tenant) error {
	if tenant.ID == uuid.Nil {
		tenant.ID = uuid.New()
	}
	tenant.CreatedAt = time.Now()
	tenant.UpdatedAt = time.Now()
	return r.db.WithContext(ctx).Create(toTenantGorm(tenant)).Error
}

func (r *GormTenantRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Tenant, error) {
	var m GormTenantModel
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&m).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return toTenantModel(&m), nil
}

func (r *GormTenantRepository) GetBySlug(ctx context.Context, slug string) (*model.Tenant, error) {
	var m GormTenantModel
	if err := r.db.WithContext(ctx).Where("slug = ?", slug).First(&m).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return toTenantModel(&m), nil
}

func (r *GormTenantRepository) Update(ctx context.Context, tenant *model.Tenant) error {
	tenant.UpdatedAt = time.Now()
	return r.db.WithContext(ctx).Save(toTenantGorm(tenant)).Error
}

func (r *GormTenantRepository) List(ctx context.Context) ([]model.Tenant, error) {
	var models []GormTenantModel
	if err := r.db.WithContext(ctx).Find(&models).Error; err != nil {
		return nil, err
	}
	result := make([]model.Tenant, len(models))
	for i := range models {
		result[i] = *toTenantModel(&models[i])
	}
	return result, nil
}

func (r *GormTenantRepository) Seed(ctx context.Context) error {
	// Check if tenant already exists
	var count int64
	r.db.WithContext(ctx).Model(&GormTenantModel{}).Count(&count)
	if count > 0 {
		return nil
	}

	tenant := &model.Tenant{
		ID:        uuid.MustParse("11111111-1111-1111-1111-111111111111"),
		Name:      "GreenPOS Demo",
		Slug:      "greenpos",
		Settings:  model.JSON{"theme": "light", "currency": "MXN"},
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	return r.Create(ctx, tenant)
}

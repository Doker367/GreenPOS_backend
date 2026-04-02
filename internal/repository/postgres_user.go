package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/greenpos/backend/internal/model"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// GormUserRepository is a GORM-backed UserRepository
type GormUserRepository struct {
	db *gorm.DB
}

type GormUserModel struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	BranchID     uuid.UUID `gorm:"type:uuid;not null;index" json:"branchId"`
	Email        string    `gorm:"type:varchar(255);not null;index:idx_users_branch_email" json:"email"`
	PasswordHash string    `gorm:"type:varchar(255);not null" json:"-"`
	Name         string    `gorm:"type:varchar(255);not null" json:"name"`
	Role         string    `gorm:"type:varchar(50);not null" json:"role"`
	IsActive     bool      `gorm:"default:true" json:"isActive"`
	CreatedAt    time.Time `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime" json:"updatedAt"`
}

func (GormUserModel) TableName() string { return "users" }

func (GormUserModel) BeforeCreate(tx *gorm.DB) error {
	if tx.Model.ID == uuid.Nil {
		tx.Model.ID = uuid.New()
	}
	return nil
}

func toUserModel(m *GormUserModel) *model.User {
	return &model.User{
		ID:           m.ID,
		BranchID:     m.BranchID,
		Email:        m.Email,
		PasswordHash: m.PasswordHash,
		Name:         m.Name,
		Role:         model.UserRole(m.Role),
		IsActive:     m.IsActive,
		CreatedAt:    m.CreatedAt,
		UpdatedAt:    m.UpdatedAt,
	}
}

func toUserGorm(m *model.User) *GormUserModel {
	return &GormUserModel{
		ID:           m.ID,
		BranchID:     m.BranchID,
		Email:        m.Email,
		PasswordHash: m.PasswordHash,
		Name:         m.Name,
		Role:         string(m.Role),
		IsActive:     m.IsActive,
		CreatedAt:    m.CreatedAt,
		UpdatedAt:    m.UpdatedAt,
	}
}

// NewGormUserRepository creates a GORM-backed user repository
func NewGormUserRepository(db *gorm.DB) *GormUserRepository {
	return &GormUserRepository{db: db}
}

func (r *GormUserRepository) Create(ctx context.Context, user *model.User) error {
	if user.ID == uuid.Nil {
		user.ID = uuid.New()
	}
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()
	return r.db.WithContext(ctx).Create(toUserGorm(user)).Error
}

func (r *GormUserRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	var m GormUserModel
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&m).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return toUserModel(&m), nil
}

func (r *GormUserRepository) GetByEmail(ctx context.Context, branchID uuid.UUID, email string) (*model.User, error) {
	var m GormUserModel
	if err := r.db.WithContext(ctx).
		Where("branch_id = ? AND email = ?", branchID, email).
		First(&m).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return toUserModel(&m), nil
}

func (r *GormUserRepository) GetByBranch(ctx context.Context, branchID uuid.UUID) ([]model.User, error) {
	var models []GormUserModel
	if err := r.db.WithContext(ctx).
		Where("branch_id = ? AND is_active = ?", branchID, true).
		Find(&models).Error; err != nil {
		return nil, err
	}
	result := make([]model.User, len(models))
	for i := range models {
		result[i] = *toUserModel(&models[i])
	}
	return result, nil
}

func (r *GormUserRepository) ListByBranch(ctx context.Context, branchID uuid.UUID) ([]model.User, error) {
	return r.GetByBranch(ctx, branchID)
}

func (r *GormUserRepository) Update(ctx context.Context, user *model.User) error {
	user.UpdatedAt = time.Now()
	return r.db.WithContext(ctx).Save(toUserGorm(user)).Error
}

func (r *GormUserRepository) UpdatePassword(ctx context.Context, id uuid.UUID, passwordHash string) error {
	return r.db.WithContext(ctx).
		Model(&GormUserModel{}).
		Where("id = ?", id).
		Update("password_hash", passwordHash).Error
}

func (r *GormUserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).
		Model(&GormUserModel{}).
		Where("id = ?", id).
		Update("is_active", false).Error
}

func (r *GormUserRepository) Seed(ctx context.Context) error {
	var count int64
	r.db.WithContext(ctx).Model(&GormUserModel{}).Count(&count)
	if count > 0 {
		return nil
	}

	password, _ := bcrypt.GenerateFromPassword([]byte("demo123"), bcrypt.DefaultCost)

	users := []*model.User{
		{
			ID:           uuid.MustParse("33333333-3333-3333-3333-333333333333"),
			BranchID:     uuid.MustParse("22222222-2222-2222-2222-222222222222"),
			Email:        "owner@greenpos.com",
			PasswordHash: string(password),
			Name:         "John Owner",
			Role:         model.RoleOwner,
			IsActive:     true,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		},
		{
			ID:           uuid.MustParse("33333333-3333-3333-3333-333333333334"),
			BranchID:     uuid.MustParse("22222222-2222-2222-2222-222222222222"),
			Email:        "admin@greenpos.com",
			PasswordHash: string(password),
			Name:         "Jane Admin",
			Role:         model.RoleAdmin,
			IsActive:     true,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		},
		{
			ID:           uuid.MustParse("33333333-3333-3333-3333-333333333335"),
			BranchID:     uuid.MustParse("22222222-2222-2222-2222-222222222222"),
			Email:        "waiter@greenpos.com",
			PasswordHash: string(password),
			Name:         "Bob Waiter",
			Role:         model.RoleWaiter,
			IsActive:     true,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		},
	}

	for _, u := range users {
		if err := r.Create(ctx, u); err != nil {
			return err
		}
	}
	return nil
}

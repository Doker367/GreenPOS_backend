package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/greenpos/backend/internal/model"
	"gorm.io/gorm"
)

// FCMTokenRepositoryInterface defines operations for FCM tokens
type FCMTokenRepositoryInterface interface {
	Create(ctx context.Context, token *model.FCMToken) error
	Upsert(ctx context.Context, token *model.FCMToken) error
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]model.FCMToken, error)
	GetActiveByUserID(ctx context.Context, userID uuid.UUID) ([]model.FCMToken, error)
	GetByBranchAndRole(ctx context.Context, branchID uuid.UUID, role model.UserRole) ([]model.FCMToken, error)
	Deactivate(ctx context.Context, id uuid.UUID) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// GormFCMTokenRepository is a GORM-backed FCMTokenRepository
type GormFCMTokenRepository struct {
	db *gorm.DB
}

type GormFCMTokenModel struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID    uuid.UUID `gorm:"type:uuid;not null;index:idx_fcm_user" json:"userId"`
	FCMToken  string    `gorm:"type:text;not null" json:"fcmToken"`
	Device    string    `gorm:"type:varchar(100)" json:"device"`
	IsActive  bool      `gorm:"default:true" json:"isActive"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updatedAt"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"createdAt"`
}

func (GormFCMTokenModel) TableName() string { return "user_fcm_tokens" }

func toFCMTokenModel(m *GormFCMTokenModel) *model.FCMToken {
	return &model.FCMToken{
		ID:        m.ID,
		UserID:    m.UserID,
		FCMToken:  m.FCMToken,
		Device:    m.Device,
		IsActive:  m.IsActive,
		UpdatedAt: m.UpdatedAt,
		CreatedAt: m.CreatedAt,
	}
}

func toFCMTokenGorm(m *model.FCMToken) *GormFCMTokenModel {
	return &GormFCMTokenModel{
		ID:       m.ID,
		UserID:   m.UserID,
		FCMToken: m.FCMToken,
		Device:   m.Device,
		IsActive: m.IsActive,
	}
}

func NewGormFCMTokenRepository(db *gorm.DB) *GormFCMTokenRepository {
	return &GormFCMTokenRepository{db: db}
}

func (r *GormFCMTokenRepository) Create(ctx context.Context, token *model.FCMToken) error {
	gormModel := toFCMTokenGorm(token)
	gormModel.ID = uuid.New()
	return r.db.WithContext(ctx).Create(gormModel).Error
}

func (r *GormFCMTokenRepository) Upsert(ctx context.Context, token *model.FCMToken) error {
	gormModel := toFCMTokenGorm(token)
	gormModel.UpdatedAt = time.Now()

	// Upsert: update if exists (same user + token), otherwise insert
	result := r.db.WithContext(ctx).
		Where("user_id = ? AND fcm_token = ?", token.UserID, token.FCMToken).
		Assign(GormFCMTokenModel{Device: token.Device, IsActive: true, UpdatedAt: gormModel.UpdatedAt}).
		FirstOrCreate(gormModel)

	return result.Error
}

func (r *GormFCMTokenRepository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]model.FCMToken, error) {
	var models []GormFCMTokenModel
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&models).Error
	if err != nil {
		return nil, err
	}

	tokens := make([]model.FCMToken, len(models))
	for i, m := range models {
		tokens[i] = *toFCMTokenModel(&m)
	}
	return tokens, nil
}

func (r *GormFCMTokenRepository) GetActiveByUserID(ctx context.Context, userID uuid.UUID) ([]model.FCMToken, error) {
	var models []GormFCMTokenModel
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND is_active = ?", userID, true).
		Find(&models).Error
	if err != nil {
		return nil, err
	}

	tokens := make([]model.FCMToken, len(models))
	for i, m := range models {
		tokens[i] = *toFCMTokenModel(&m)
	}
	return tokens, nil
}

func (r *GormFCMTokenRepository) GetByBranchAndRole(ctx context.Context, branchID uuid.UUID, role model.UserRole) ([]model.FCMToken, error) {
	var models []GormFCMTokenModel
	err := r.db.WithContext(ctx).
		Joins("JOIN users ON users.id = user_fcm_tokens.user_id").
		Where("users.branch_id = ? AND users.role = ? AND users.is_active = ? AND user_fcm_tokens.is_active = ?",
			branchID, role, true, true).
		Find(&models).Error
	if err != nil {
		return nil, err
	}

	tokens := make([]model.FCMToken, len(models))
	for i, m := range models {
		tokens[i] = *toFCMTokenModel(&m)
	}
	return tokens, nil
}

func (r *GormFCMTokenRepository) Deactivate(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).
		Model(&GormFCMTokenModel{}).
		Where("id = ?", id).
		Update("is_active", false).Error
}

func (r *GormFCMTokenRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&GormFCMTokenModel{}, "id = ?", id).Error
}

// MemoryFCMTokenRepository is an in-memory implementation for development/testing
type MemoryFCMTokenRepository struct {
	tokens map[uuid.UUID]*model.FCMToken
}

func NewMemoryFCMTokenRepository() *MemoryFCMTokenRepository {
	return &MemoryFCMTokenRepository{
		tokens: make(map[uuid.UUID]*model.FCMToken),
	}
}

func (r *MemoryFCMTokenRepository) Create(ctx context.Context, token *model.FCMToken) error {
	token.ID = uuid.New()
	token.CreatedAt = time.Now()
	token.UpdatedAt = time.Now()
	r.tokens[token.ID] = token
	return nil
}

func (r *MemoryFCMTokenRepository) Upsert(ctx context.Context, token *model.FCMToken) error {
	// Find existing token for this user with same FCM token
	for _, t := range r.tokens {
		if t.UserID == token.UserID && t.FCMToken == token.FCMToken {
			t.Device = token.Device
			t.IsActive = true
			t.UpdatedAt = time.Now()
			return nil
		}
	}
	// Not found, create new
	token.ID = uuid.New()
	token.CreatedAt = time.Now()
	token.UpdatedAt = time.Now()
	r.tokens[token.ID] = token
	return nil
}

func (r *MemoryFCMTokenRepository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]model.FCMToken, error) {
	var tokens []model.FCMToken
	for _, t := range r.tokens {
		if t.UserID == userID {
			tokens = append(tokens, *t)
		}
	}
	return tokens, nil
}

func (r *MemoryFCMTokenRepository) GetActiveByUserID(ctx context.Context, userID uuid.UUID) ([]model.FCMToken, error) {
	var tokens []model.FCMToken
	for _, t := range r.tokens {
		if t.UserID == userID && t.IsActive {
			tokens = append(tokens, *t)
		}
	}
	return tokens, nil
}

func (r *MemoryFCMTokenRepository) GetByBranchAndRole(ctx context.Context, branchID uuid.UUID, role model.UserRole) ([]model.FCMToken, error) {
	// This would need access to users - for now return empty
	// In real implementation, we'd join with user repository
	return []model.FCMToken{}, nil
}

func (r *MemoryFCMTokenRepository) Deactivate(ctx context.Context, id uuid.UUID) error {
	if t, ok := r.tokens[id]; ok {
		t.IsActive = false
		t.UpdatedAt = time.Now()
	}
	return nil
}

func (r *MemoryFCMTokenRepository) Delete(ctx context.Context, id uuid.UUID) error {
	delete(r.tokens, id)
	return nil
}

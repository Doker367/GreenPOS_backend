package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/greenpos/backend/internal/model"
	"gorm.io/gorm"
)

// GormQRCodeRepository is a GORM-backed QR code repository
type GormQRCodeRepository struct {
	db *gorm.DB
}

// GormQRCodeModel is the GORM model for table_qr_codes
type GormQRCodeModel struct {
	ID         uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TableID    uuid.UUID `gorm:"type:uuid;not null;uniqueIndex" json:"tableId"`
	QRToken    string    `gorm:"type:varchar(255);uniqueIndex;not null" json:"qrToken"`
	AccessURL  string    `gorm:"type:varchar(500);not null" json:"accessUrl"`
	IsActive   bool      `gorm:"default:true" json:"isActive"`
	CreatedAt  time.Time `gorm:"autoCreateTime" json:"createdAt"`
}

func (GormQRCodeModel) TableName() string { return "table_qr_codes" }

func toQRCodeModel(m *GormQRCodeModel) *model.TableQRCode {
	return &model.TableQRCode{
		ID:        m.ID,
		TableID:   m.TableID,
		QRToken:   m.QRToken,
		AccessURL: m.AccessURL,
		IsActive:  m.IsActive,
		CreatedAt: m.CreatedAt,
	}
}

func toQRCodeGorm(m *model.TableQRCode) *GormQRCodeModel {
	return &GormQRCodeModel{
		ID:        m.ID,
		TableID:   m.TableID,
		QRToken:   m.QRToken,
		AccessURL: m.AccessURL,
		IsActive:  m.IsActive,
		CreatedAt: m.CreatedAt,
	}
}

// NewGormQRCodeRepository creates a GORM-backed QR code repository
func NewGormQRCodeRepository(db *gorm.DB) *GormQRCodeRepository {
	return &GormQRCodeRepository{db: db}
}

func (r *GormQRCodeRepository) Create(ctx context.Context, qr *model.TableQRCode) error {
	if qr.ID == uuid.Nil {
		qr.ID = uuid.New()
	}
	qr.CreatedAt = time.Now()
	return r.db.WithContext(ctx).Create(toQRCodeGorm(qr)).Error
}

func (r *GormQRCodeRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.TableQRCode, error) {
	var m GormQRCodeModel
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&m).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return toQRCodeModel(&m), nil
}

func (r *GormQRCodeRepository) GetByTableID(ctx context.Context, tableID uuid.UUID) (*model.TableQRCode, error) {
	var m GormQRCodeModel
	if err := r.db.WithContext(ctx).Where("table_id = ?", tableID).First(&m).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return toQRCodeModel(&m), nil
}

func (r *GormQRCodeRepository) GetByToken(ctx context.Context, token string) (*model.TableQRCode, error) {
	var m GormQRCodeModel
	if err := r.db.WithContext(ctx).Where("qr_token = ? AND is_active = true", token).First(&m).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return toQRCodeModel(&m), nil
}

func (r *GormQRCodeRepository) Update(ctx context.Context, qr *model.TableQRCode) error {
	return r.db.WithContext(ctx).Save(toQRCodeGorm(qr)).Error
}

func (r *GormQRCodeRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&GormQRCodeModel{}, "id = ?", id).Error
}

func (r *GormQRCodeRepository) DeactivateByTableID(ctx context.Context, tableID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Model(&GormQRCodeModel{}).
		Where("table_id = ?", tableID).
		Update("is_active", false).Error
}

// QRCodeRepositoryInterface defines operations for QR codes
type QRCodeRepositoryInterface interface {
	Create(ctx context.Context, qr *model.TableQRCode) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.TableQRCode, error)
	GetByTableID(ctx context.Context, tableID uuid.UUID) (*model.TableQRCode, error)
	GetByToken(ctx context.Context, token string) (*model.TableQRCode, error)
	Update(ctx context.Context, qr *model.TableQRCode) error
	Delete(ctx context.Context, id uuid.UUID) error
	DeactivateByTableID(ctx context.Context, tableID uuid.UUID) error
}

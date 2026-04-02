package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/greenpos/backend/internal/model"
	"gorm.io/gorm"
)

// GormReservationRepository is a GORM-backed ReservationRepository
type GormReservationRepository struct {
	db *gorm.DB
}

type GormReservationModel struct {
	ID              uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	BranchID        uuid.UUID  `gorm:"type:uuid;not null;index" json:"branchId"`
	TableID         *uuid.UUID `gorm:"type:uuid;index" json:"tableId"`
	CustomerName    string     `gorm:"type:varchar(255);not null" json:"customerName"`
	CustomerPhone   string     `gorm:"type:varchar(50)" json:"customerPhone"`
	CustomerEmail   string     `gorm:"type:varchar(255)" json:"customerEmail"`
	GuestCount      int        `gorm:"default:1" json:"guestCount"`
	ReservationDate time.Time  `gorm:"type:date;not null;index" json:"reservationDate"`
	ReservationTime string     `gorm:"type:time;not null" json:"reservationTime"`
	Status          string     `gorm:"type:varchar(50);default:'CONFIRMED'" json:"status"`
	Notes           string     `gorm:"type:text" json:"notes"`
	CreatedAt       time.Time  `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt       time.Time  `gorm:"autoUpdateTime" json:"updatedAt"`
}

func (GormReservationModel) TableName() string { return "reservations" }

func toReservationModel(m *GormReservationModel) *model.Reservation {
	return &model.Reservation{
		ID:              m.ID,
		BranchID:        m.BranchID,
		TableID:         m.TableID,
		CustomerName:    m.CustomerName,
		CustomerPhone:   m.CustomerPhone,
		CustomerEmail:   m.CustomerEmail,
		GuestCount:      m.GuestCount,
		ReservationDate: m.ReservationDate,
		ReservationTime: m.ReservationTime,
		Status:          model.ReservationStatus(m.Status),
		Notes:           m.Notes,
		CreatedAt:       m.CreatedAt,
		UpdatedAt:       m.UpdatedAt,
	}
}

func toReservationGorm(m *model.Reservation) *GormReservationModel {
	return &GormReservationModel{
		ID:              m.ID,
		BranchID:        m.BranchID,
		TableID:         m.TableID,
		CustomerName:    m.CustomerName,
		CustomerPhone:   m.CustomerPhone,
		CustomerEmail:   m.CustomerEmail,
		GuestCount:      m.GuestCount,
		ReservationDate: m.ReservationDate,
		ReservationTime: m.ReservationTime,
		Status:          string(m.Status),
		Notes:           m.Notes,
		CreatedAt:       m.CreatedAt,
		UpdatedAt:       m.UpdatedAt,
	}
}

// NewGormReservationRepository creates a GORM-backed reservation repository
func NewGormReservationRepository(db *gorm.DB) *GormReservationRepository {
	return &GormReservationRepository{db: db}
}

func (r *GormReservationRepository) Create(ctx context.Context, res *model.Reservation) error {
	if res.ID == uuid.Nil {
		res.ID = uuid.New()
	}
	res.CreatedAt = time.Now()
	res.UpdatedAt = time.Now()
	return r.db.WithContext(ctx).Create(toReservationGorm(res)).Error
}

func (r *GormReservationRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Reservation, error) {
	var m GormReservationModel
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&m).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return toReservationModel(&m), nil
}

func (r *GormReservationRepository) GetByBranch(ctx context.Context, branchID uuid.UUID) ([]model.Reservation, error) {
	var models []GormReservationModel
	if err := r.db.WithContext(ctx).
		Where("branch_id = ?", branchID).
		Order("reservation_date DESC, reservation_time DESC").
		Find(&models).Error; err != nil {
		return nil, err
	}
	result := make([]model.Reservation, len(models))
	for i := range models {
		result[i] = *toReservationModel(&models[i])
	}
	return result, nil
}

func (r *GormReservationRepository) ListByBranch(ctx context.Context, branchID uuid.UUID, date *time.Time) ([]model.Reservation, error) {
	var models []GormReservationModel
	query := r.db.WithContext(ctx).Where("branch_id = ?", branchID)
	if date != nil {
		query = query.Where("reservation_date = ?", date.Format("2006-01-02"))
	}
	if err := query.
		Order("reservation_date DESC, reservation_time DESC").
		Find(&models).Error; err != nil {
		return nil, err
	}
	result := make([]model.Reservation, len(models))
	for i := range models {
		result[i] = *toReservationModel(&models[i])
	}
	return result, nil
}

func (r *GormReservationRepository) GetByDate(ctx context.Context, branchID uuid.UUID, date string) ([]model.Reservation, error) {
	var models []GormReservationModel
	if err := r.db.WithContext(ctx).
		Where("branch_id = ? AND reservation_date = ?", branchID, date).
		Order("reservation_time ASC").
		Find(&models).Error; err != nil {
		return nil, err
	}
	result := make([]model.Reservation, len(models))
	for i := range models {
		result[i] = *toReservationModel(&models[i])
	}
	return result, nil
}

func (r *GormReservationRepository) Update(ctx context.Context, res *model.Reservation) error {
	res.UpdatedAt = time.Now()
	return r.db.WithContext(ctx).Save(toReservationGorm(res)).Error
}

func (r *GormReservationRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status model.ReservationStatus) error {
	return r.db.WithContext(ctx).
		Model(&GormReservationModel{}).
		Where("id = ?", id).
		Update("status", string(status)).Error
}

func (r *GormReservationRepository) Seed(ctx context.Context) error {
	return nil // Reservations are typically created by users, not seeded
}

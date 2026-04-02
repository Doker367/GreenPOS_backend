package repository

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/greenpos/backend/internal/model"
)

// ReservationMemoryRepository is an in-memory implementation of ReservationRepository
type ReservationMemoryRepository struct {
	mu           sync.RWMutex
	reservations map[uuid.UUID]*model.Reservation
	byBranch     map[uuid.UUID][]*model.Reservation
}

// NewReservationMemoryRepository creates a new in-memory reservation repository
func NewReservationMemoryRepository() *ReservationMemoryRepository {
	return &ReservationMemoryRepository{
		reservations: make(map[uuid.UUID]*model.Reservation),
		byBranch:    make(map[uuid.UUID][]*model.Reservation),
	}
}

func (r *ReservationMemoryRepository) Create(ctx context.Context, reservation *model.Reservation) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	reservation.CreatedAt = time.Now()
	reservation.UpdatedAt = time.Now()
	if reservation.ID == uuid.Nil {
		reservation.ID = uuid.New()
	}

	r.reservations[reservation.ID] = reservation
	r.byBranch[reservation.BranchID] = append(r.byBranch[reservation.BranchID], reservation)
	return nil
}

func (r *ReservationMemoryRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Reservation, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if res, ok := r.reservations[id]; ok {
		return res, nil
	}
	return nil, ErrNotFound
}

func (r *ReservationMemoryRepository) GetByBranch(ctx context.Context, branchID uuid.UUID) ([]model.Reservation, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	reservations := r.byBranch[branchID]
	result := make([]model.Reservation, len(reservations))
	for i, res := range reservations {
		result[i] = *res
	}
	return result, nil
}

func (r *ReservationMemoryRepository) ListByBranch(ctx context.Context, branchID uuid.UUID, date *time.Time) ([]model.Reservation, error) {
	if date != nil {
		return r.GetByDate(ctx, branchID, date.Format("2006-01-02"))
	}
	return r.GetByBranch(ctx, branchID)
}

func (r *ReservationMemoryRepository) GetByDate(ctx context.Context, branchID uuid.UUID, date string) ([]model.Reservation, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	reservations := r.byBranch[branchID]
	result := make([]model.Reservation, 0)
	for _, res := range reservations {
		if res.ReservationDate.Format("2006-01-02") == date {
			result = append(result, *res)
		}
	}
	return result, nil
}

func (r *ReservationMemoryRepository) Update(ctx context.Context, reservation *model.Reservation) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	reservation.UpdatedAt = time.Now()
	r.reservations[reservation.ID] = reservation
	return nil
}

func (r *ReservationMemoryRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status model.ReservationStatus) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if res, ok := r.reservations[id]; ok {
		res.Status = status
		res.UpdatedAt = time.Now()
		return nil
	}
	return ErrNotFound
}

func (r *ReservationMemoryRepository) Seed(ctx context.Context) error {
	now := time.Now()
	tomorrow := now.AddDate(0, 0, 1)
	reservationDate := time.Date(tomorrow.Year(), tomorrow.Month(), tomorrow.Day(), 0, 0, 0, 0, tomorrow.Location())

	reservations := []*model.Reservation{
		{
			ID:              uuid.MustParse("99999999-9999-9999-9999-999999999991"),
			BranchID:        uuid.MustParse("22222222-2222-2222-2222-222222222222"),
			TableID:         uuidPtr(uuid.MustParse("66666666-6666-6666-6666-666666666664")),
			CustomerName:    "Carol Davis",
			CustomerPhone:   "+1-555-0201",
			CustomerEmail:   "carol@example.com",
			GuestCount:      4,
			ReservationDate: reservationDate,
			ReservationTime: "19:00",
			Status:          model.ReservationConfirmed,
			Notes:           "Birthday celebration",
			CreatedAt:       now,
			UpdatedAt:       now,
		},
		{
			ID:              uuid.MustParse("99999999-9999-9999-9999-999999999992"),
			BranchID:        uuid.MustParse("22222222-2222-2222-2222-222222222222"),
			TableID:         uuidPtr(uuid.MustParse("66666666-6666-6666-6666-666666666662")),
			CustomerName:    "David Wilson",
			CustomerPhone:   "+1-555-0202",
			CustomerEmail:   "",
			GuestCount:      2,
			ReservationDate: reservationDate,
			ReservationTime: "20:00",
			Status:          model.ReservationPending,
			Notes:           "",
			CreatedAt:       now,
			UpdatedAt:       now,
		},
	}

	for _, res := range reservations {
		if err := r.Create(ctx, res); err != nil {
			return err
		}
	}
	return nil
}

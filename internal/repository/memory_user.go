package repository

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/greenpos/backend/internal/model"
	"golang.org/x/crypto/bcrypt"
)

// UserMemoryRepository is an in-memory implementation of UserRepository
type UserMemoryRepository struct {
	mu       sync.RWMutex
	users    map[uuid.UUID]*model.User
	byBranch map[uuid.UUID][]*model.User
	byEmail  map[string]*model.User // key: branchID:email
}

// NewUserMemoryRepository creates a new in-memory user repository
func NewUserMemoryRepository() *UserMemoryRepository {
	return &UserMemoryRepository{
		users:    make(map[uuid.UUID]*model.User),
		byBranch: make(map[uuid.UUID][]*model.User),
		byEmail:  make(map[string]*model.User),
	}
}

func (r *UserMemoryRepository) Create(ctx context.Context, user *model.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()
	if user.ID == uuid.Nil {
		user.ID = uuid.New()
	}

	r.users[user.ID] = user
	r.byBranch[user.BranchID] = append(r.byBranch[user.BranchID], user)
	key := user.BranchID.String() + ":" + user.Email
	r.byEmail[key] = user
	return nil
}

func (r *UserMemoryRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if user, ok := r.users[id]; ok {
		return user, nil
	}
	return nil, ErrNotFound
}

func (r *UserMemoryRepository) GetByEmail(ctx context.Context, branchID uuid.UUID, email string) (*model.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	key := branchID.String() + ":" + email
	if user, ok := r.byEmail[key]; ok {
		return user, nil
	}
	return nil, ErrNotFound
}

func (r *UserMemoryRepository) GetByBranch(ctx context.Context, branchID uuid.UUID) ([]model.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	users := r.byBranch[branchID]
	result := make([]model.User, 0, len(users))
	for _, u := range users {
		if u.IsActive {
			result = append(result, *u)
		}
	}
	return result, nil
}

func (r *UserMemoryRepository) ListByBranch(ctx context.Context, branchID uuid.UUID) ([]model.User, error) {
	return r.GetByBranch(ctx, branchID)
}

func (r *UserMemoryRepository) Update(ctx context.Context, user *model.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	user.UpdatedAt = time.Now()
	r.users[user.ID] = user
	return nil
}

func (r *UserMemoryRepository) UpdatePassword(ctx context.Context, id uuid.UUID, passwordHash string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if user, ok := r.users[id]; ok {
		user.PasswordHash = passwordHash
		user.UpdatedAt = time.Now()
		return nil
	}
	return ErrNotFound
}

func (r *UserMemoryRepository) Delete(ctx context.Context, id uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if user, ok := r.users[id]; ok {
		user.IsActive = false
		user.UpdatedAt = time.Now()
		return nil
	}
	return ErrNotFound
}

func (r *UserMemoryRepository) Seed(ctx context.Context) error {
	now := time.Now()
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
			CreatedAt:    now,
			UpdatedAt:    now,
		},
		{
			ID:           uuid.MustParse("33333333-3333-3333-3333-333333333334"),
			BranchID:     uuid.MustParse("22222222-2222-2222-2222-222222222222"),
			Email:        "admin@greenpos.com",
			PasswordHash: string(password),
			Name:         "Jane Admin",
			Role:         model.RoleAdmin,
			IsActive:     true,
			CreatedAt:    now,
			UpdatedAt:    now,
		},
		{
			ID:           uuid.MustParse("33333333-3333-3333-3333-333333333335"),
			BranchID:     uuid.MustParse("22222222-2222-2222-2222-222222222222"),
			Email:        "waiter@greenpos.com",
			PasswordHash: string(password),
			Name:         "Bob Waiter",
			Role:         model.RoleWaiter,
			IsActive:     true,
			CreatedAt:    now,
			UpdatedAt:    now,
		},
	}

	for _, u := range users {
		if err := r.Create(ctx, u); err != nil {
			return err
		}
	}
	return nil
}

package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/greenpos/backend/internal/model"
)

// PaymentRepositoryInterface defines operations for payments
type PaymentRepositoryInterface interface {
	Create(ctx context.Context, payment *model.Payment) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.Payment, error)
	GetByOrderID(ctx context.Context, orderID uuid.UUID) ([]model.Payment, error)
	Update(ctx context.Context, payment *model.Payment) error
	UpdateStatus(ctx context.Context, id uuid.UUID, status model.PaymentStatus) error
	SetProviderPaymentID(ctx context.Context, id uuid.UUID, providerPaymentID string) error
}

// PaymentRepository implements PaymentRepositoryInterface for PostgreSQL
type PaymentRepository struct {
	db *sql.DB
}

// NewPaymentRepository creates a new PaymentRepository
func NewPaymentRepository(db *sql.DB) *PaymentRepository {
	return &PaymentRepository{db: db}
}

// Create creates a new payment record
func (r *PaymentRepository) Create(ctx context.Context, payment *model.Payment) error {
	query := `
		INSERT INTO payments (id, order_id, amount, method, provider, provider_payment_id, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`
	now := time.Now()
	if payment.ID == uuid.Nil {
		payment.ID = uuid.New()
	}
	payment.CreatedAt = now
	payment.UpdatedAt = now

	_, err := r.db.ExecContext(ctx, query,
		payment.ID,
		payment.OrderID,
		payment.Amount,
		payment.Method,
		payment.Provider,
		payment.ProviderPaymentID,
		payment.Status,
		payment.CreatedAt,
		payment.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create payment: %w", err)
	}
	return nil
}

// GetByID retrieves a payment by ID
func (r *PaymentRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Payment, error) {
	query := `
		SELECT id, order_id, amount, method, provider, provider_payment_id, status, created_at, updated_at
		FROM payments WHERE id = $1
	`
	var payment model.Payment
	var provider, providerPaymentID sql.NullString

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&payment.ID,
		&payment.OrderID,
		&payment.Amount,
		&payment.Method,
		&provider,
		&providerPaymentID,
		&payment.Status,
		&payment.CreatedAt,
		&payment.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to get payment: %w", err)
	}

	if provider.Valid {
		payment.Provider = model.PaymentProvider(provider.String)
	}
	if providerPaymentID.Valid {
		payment.ProviderPaymentID = providerPaymentID.String
	}

	return &payment, nil
}

// GetByOrderID retrieves all payments for an order
func (r *PaymentRepository) GetByOrderID(ctx context.Context, orderID uuid.UUID) ([]model.Payment, error) {
	query := `
		SELECT id, order_id, amount, method, provider, provider_payment_id, status, created_at, updated_at
		FROM payments WHERE order_id = $1 ORDER BY created_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query, orderID)
	if err != nil {
		return nil, fmt.Errorf("failed to query payments: %w", err)
	}
	defer rows.Close()

	var payments []model.Payment
	for rows.Next() {
		var payment model.Payment
		var provider, providerPaymentID sql.NullString

		err := rows.Scan(
			&payment.ID,
			&payment.OrderID,
			&payment.Amount,
			&payment.Method,
			&provider,
			&providerPaymentID,
			&payment.Status,
			&payment.CreatedAt,
			&payment.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan payment: %w", err)
		}

		if provider.Valid {
			payment.Provider = model.PaymentProvider(provider.String)
		}
		if providerPaymentID.Valid {
			payment.ProviderPaymentID = providerPaymentID.String
		}

		payments = append(payments, payment)
	}

	return payments, nil
}

// Update updates an existing payment
func (r *PaymentRepository) Update(ctx context.Context, payment *model.Payment) error {
	query := `
		UPDATE payments 
		SET amount = $2, method = $3, provider = $4, provider_payment_id = $5, status = $6, updated_at = $7
		WHERE id = $1
	`
	payment.UpdatedAt = time.Now()

	result, err := r.db.ExecContext(ctx, query,
		payment.ID,
		payment.Amount,
		payment.Method,
		payment.Provider,
		payment.ProviderPaymentID,
		payment.Status,
		payment.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to update payment: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}

// UpdateStatus updates the payment status
func (r *PaymentRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status model.PaymentStatus) error {
	query := `UPDATE payments SET status = $2, updated_at = $3 WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, id, status, time.Now())
	if err != nil {
		return fmt.Errorf("failed to update payment status: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}

// SetProviderPaymentID sets the provider payment ID
func (r *PaymentRepository) SetProviderPaymentID(ctx context.Context, id uuid.UUID, providerPaymentID string) error {
	query := `UPDATE payments SET provider_payment_id = $2, updated_at = $3 WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, id, providerPaymentID, time.Now())
	if err != nil {
		return fmt.Errorf("failed to set provider payment ID: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}

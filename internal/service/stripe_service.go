package service

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/stripe/stripe-go/v76"
	"github.com/stripe/stripe-go/v76/paymentintent"
	"github.com/stripe/stripe-go/v76/webhook"

	"github.com/greenpos/backend/internal/model"
	"github.com/greenpos/backend/internal/repository"
)

// StripeService handles Stripe payment operations
type StripeService struct {
	payments repository.PaymentRepositoryInterface
	orders   repository.OrderRepositoryInterface
	log      *slog.Logger
	secret   string
}

// NewStripeService creates a new StripeService
func NewStripeService(payments repository.PaymentRepositoryInterface, orders repository.OrderRepositoryInterface, log *slog.Logger) *StripeService {
	stripe.Key = os.Getenv("STRIPE_SECRET_KEY")
	if stripe.Key == "" {
		log.Warn("STRIPE_SECRET_KEY not set, Stripe payments will not work")
	}

	return &StripeService{
		payments: payments,
		orders:   orders,
		log:      log,
		secret:   os.Getenv("STRIPE_WEBHOOK_SECRET"),
	}
}

// CreatePaymentIntent creates a Stripe payment intent for an order
func (s *StripeService) CreatePaymentIntent(ctx context.Context, orderID uuid.UUID) (*model.PaymentIntent, error) {
	// Get order to verify it exists and get amount
	order, err := s.orders.GetByID(ctx, orderID)
	if err != nil {
		return nil, fmt.Errorf("order not found: %w", err)
	}

	// Create payment record with pending status
	payment := &model.Payment{
		ID:       uuid.New(),
		OrderID:  orderID,
		Amount:   order.Total,
		Method:   model.PaymentMethodCard,
		Provider: model.PaymentProviderStripe,
		Status:   model.PaymentStatusPending,
	}

	if err := s.payments.Create(ctx, payment); err != nil {
		return nil, fmt.Errorf("failed to create payment record: %w", err)
	}

	// Convert amount to cents (Stripe uses smallest currency unit)
	amountCents := int64(order.Total * 100)

	// Create Stripe payment intent
	params := &stripe.PaymentIntentParams{
		Amount:   stripe.Int64(amountCents),
		Currency: stripe.String("mxn"), // Mexican pesos
		Metadata: map[string]string{
			"payment_id": payment.ID.String(),
			"order_id":   orderID.String(),
		},
		AutomaticPaymentMethods: &stripe.PaymentIntentAutomaticPaymentMethodsParams{
			Enabled: stripe.Bool(true),
		},
	}

	pi, err := paymentintent.New(params)
	if err != nil {
		// Update payment status to failed
		s.payments.UpdateStatus(ctx, payment.ID, model.PaymentStatusFailed)
		return nil, fmt.Errorf("failed to create Stripe payment intent: %w", err)
	}

	// Update payment with Stripe payment intent ID
	if err := s.payments.SetProviderPaymentID(ctx, payment.ID, pi.ID); err != nil {
		s.log.Error("failed to set provider payment ID", slog.String("error", err.Error()))
	}

	s.log.Info("Stripe payment intent created",
		slog.String("payment_id", payment.ID.String()),
		slog.String("stripe_pi", pi.ID),
		slog.Float64("amount", order.Total),
	)

	return &model.PaymentIntent{
		ID:           pi.ID,
		ClientSecret: pi.ClientSecret,
		Amount:       order.Total,
		Currency:     "mxn",
		Status:       string(pi.Status),
	}, nil
}

// ConfirmPaymentIntent confirms a payment intent and updates the payment status
func (s *StripeService) ConfirmPaymentIntent(ctx context.Context, paymentIntentID string) (*model.Payment, error) {
	// Find payment by provider payment ID
	payments, err := s.payments.GetByOrderID(ctx, uuid.Nil) // We need to search differently
	if err != nil && err != repository.ErrNotFound {
		return nil, fmt.Errorf("failed to get payments: %w", err)
	}

	// For simplicity, let's retrieve payment by metadata lookup
	// In production, you'd store the Stripe PI ID directly
	var targetPayment *model.Payment
	for _, p := range payments {
		if p.ProviderPaymentID == paymentIntentID {
			targetPayment = &p
			break
		}
	}

	if targetPayment == nil {
		return nil, fmt.Errorf("payment not found for payment intent: %s", paymentIntentID)
	}

	// Retrieve the payment intent from Stripe to get current status
	pi, err := paymentintent.Get(paymentIntentID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get Stripe payment intent: %w", err)
	}

	// Update payment status based on Stripe status
	switch pi.Status {
	case stripe.PaymentIntentStatusSucceeded:
		targetPayment.Status = model.PaymentStatusCompleted
	case stripe.PaymentIntentStatusProcessing:
		targetPayment.Status = model.PaymentStatusPending
	case stripe.PaymentIntentStatusRequiresPaymentMethod,
		stripe.PaymentIntentStatusRequiresAction,
		stripe.PaymentIntentStatusRequiresConfirmation:
		targetPayment.Status = model.PaymentStatusPending
	case stripe.PaymentIntentStatusCanceled:
		targetPayment.Status = model.PaymentStatusFailed
	case stripe.PaymentIntentStatusRequiresCapture:
		targetPayment.Status = model.PaymentStatusPending
	}

	if err := s.payments.Update(ctx, targetPayment); err != nil {
		return nil, fmt.Errorf("failed to update payment: %w", err)
	}

	s.log.Info("Stripe payment confirmed",
		slog.String("payment_id", targetPayment.ID.String()),
		slog.String("stripe_pi", paymentIntentID),
		slog.String("status", string(targetPayment.Status)),
	)

	return targetPayment, nil
}

// HandleWebhook processes Stripe webhook events
func (s *StripeService) HandleWebhook(ctx context.Context, payload []byte, signature string) error {
	if s.secret == "" {
		return fmt.Errorf("Stripe webhook secret not configured")
	}

	event, err := webhook.ConstructEvent(payload, signature, s.secret)
	if err != nil {
		s.log.Error("Stripe webhook signature verification failed", slog.String("error", err.Error()))
		return fmt.Errorf("invalid webhook signature: %w", err)
	}

	switch event.Type {
	case "payment_intent.succeeded":
		var pi stripe.PaymentIntent
		if err := stripe.Unmarshal(event.Data.Raw, &pi); err != nil {
			return fmt.Errorf("failed to parse payment intent: %w", err)
		}
		return s.handlePaymentIntentSucceeded(ctx, &pi)

	case "payment_intent.payment_failed":
		var pi stripe.PaymentIntent
		if err := stripe.Unmarshal(event.Data.Raw, &pi); err != nil {
			return fmt.Errorf("failed to parse payment intent: %w", err)
		}
		return s.handlePaymentIntentFailed(ctx, &pi)

	case "payment_intent.canceled":
		var pi stripe.PaymentIntent
		if err := stripe.Unmarshal(event.Data.Raw, &pi); err != nil {
			return fmt.Errorf("failed to parse payment intent: %w", err)
		}
		return s.handlePaymentIntentCanceled(ctx, &pi)

	default:
		s.log.Info("Unhandled Stripe event type", slog.String("type", event.Type))
	}

	return nil
}

func (s *StripeService) handlePaymentIntentSucceeded(ctx context.Context, pi *stripe.PaymentIntent) error {
	paymentIDStr := pi.Metadata["payment_id"]
	if paymentIDStr == "" {
		s.log.Warn("payment_id not found in Stripe metadata")
		return nil
	}

	paymentID, err := uuid.Parse(paymentIDStr)
	if err != nil {
		return fmt.Errorf("invalid payment_id in metadata: %w", err)
	}

	if err := s.payments.UpdateStatus(ctx, paymentID, model.PaymentStatusCompleted); err != nil {
		return fmt.Errorf("failed to update payment status: %w", err)
	}

	// Update order status to PAID
	orderIDStr := pi.Metadata["order_id"]
	if orderIDStr != "" {
		orderID, err := uuid.Parse(orderIDStr)
		if err == nil {
			s.orders.UpdateStatus(ctx, orderID, model.OrderPaid)
		}
	}

	s.log.Info("Payment succeeded via webhook",
		slog.String("payment_id", paymentIDStr),
		slog.String("stripe_pi", pi.ID),
	)

	return nil
}

func (s *StripeService) handlePaymentIntentFailed(ctx context.Context, pi *stripe.PaymentIntent) error {
	paymentIDStr := pi.Metadata["payment_id"]
	if paymentIDStr == "" {
		return nil
	}

	paymentID, err := uuid.Parse(paymentIDStr)
	if err != nil {
		return fmt.Errorf("invalid payment_id: %w", err)
	}

	if err := s.payments.UpdateStatus(ctx, paymentID, model.PaymentStatusFailed); err != nil {
		return fmt.Errorf("failed to update payment status: %w", err)
	}

	s.log.Info("Payment failed via webhook",
		slog.String("payment_id", paymentIDStr),
		slog.String("stripe_pi", pi.ID),
	)

	return nil
}

func (s *StripeService) handlePaymentIntentCanceled(ctx context.Context, pi *stripe.PaymentIntent) error {
	paymentIDStr := pi.Metadata["payment_id"]
	if paymentIDStr == "" {
		return nil
	}

	paymentID, err := uuid.Parse(paymentIDStr)
	if err != nil {
		return fmt.Errorf("invalid payment_id: %w", err)
	}

	if err := s.payments.UpdateStatus(ctx, paymentID, model.PaymentStatusFailed); err != nil {
		return fmt.Errorf("failed to update payment status: %w", err)
	}

	s.log.Info("Payment canceled via webhook",
		slog.String("payment_id", paymentIDStr),
		slog.String("stripe_pi", pi.ID),
	)

	return nil
}

// GetPaymentByOrderID retrieves payments for an order
func (s *StripeService) GetPaymentsByOrderID(ctx context.Context, orderID uuid.UUID) ([]model.Payment, error) {
	return s.payments.GetByOrderID(ctx, orderID)
}

// RefundPayment refunds a payment
func (s *StripeService) RefundPayment(ctx context.Context, paymentID uuid.UUID) error {
	payment, err := s.payments.GetByID(ctx, paymentID)
	if err != nil {
		return fmt.Errorf("payment not found: %w", err)
	}

	if payment.Status != model.PaymentStatusCompleted {
		return fmt.Errorf("can only refund completed payments")
	}

	// In production, you'd call Stripe Refund API here
	// For now, just update the status
	if err := s.payments.UpdateStatus(ctx, paymentID, model.PaymentStatusRefunded); err != nil {
		return fmt.Errorf("failed to update payment status: %w", err)
	}

	s.log.Info("Payment refunded",
		slog.String("payment_id", paymentID.String()),
	)

	return nil
}

// AmountToCents converts a float amount to cents
func AmountToCents(amount float64) int64 {
	return int64(amount * 100)
}

// ParseAmountFromCents converts cents back to float amount
func ParseAmountFromCents(cents int64) float64 {
	return float64(cents) / 100
}

// WaitForPaymentIntent polls Stripe for payment intent status
func (s *StripeService) WaitForPaymentIntent(ctx context.Context, paymentIntentID string, timeout time.Duration) (*stripe.PaymentIntent, error) {
	deadline := time.Now().Add(timeout)
	
	for time.Now().Before(deadline) {
		pi, err := paymentintent.Get(paymentIntentID, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to get payment intent: %w", err)
		}

		switch pi.Status {
		case stripe.PaymentIntentStatusSucceeded:
			return pi, nil
		case stripe.PaymentIntentStatusCanceled, stripe.PaymentIntentStatusFailed:
			return nil, fmt.Errorf("payment failed or was canceled")
		}

		time.Sleep(500 * time.Millisecond)
	}

	return nil, fmt.Errorf("timeout waiting for payment confirmation")
}

// GetStripeKey returns the Stripe public key for client-side initialization
func GetStripeKey() string {
	return os.Getenv("STRIPE_PUBLISHABLE_KEY")
}

// ParseAmount parses an amount string to float
func ParseAmount(amountStr string) (float64, error) {
	return strconv.ParseFloat(amountStr, 64)
}

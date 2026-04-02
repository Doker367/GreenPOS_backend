package graph

import (
	"context"

	"github.com/google/uuid"
	"github.com/greenpos/backend/internal/model"
	"github.com/greenpos/backend/internal/service"
)

// PaymentResult represents the result of a payment operation
type PaymentResult struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Payment *Payment `json:"payment,omitempty"`
}

// Payment represents a payment in GraphQL
type Payment struct {
	ID                string          `json:"id"`
	OrderID           string          `json:"orderId"`
	Amount            float64         `json:"amount"`
	Method            string          `json:"method"`
	Provider          string          `json:"provider"`
	ProviderPaymentID string          `json:"providerPaymentId"`
	Status            string          `json:"status"`
	CreatedAt         string          `json:"createdAt"`
}

// PaymentIntentResult represents a Stripe payment intent
type PaymentIntentResult struct {
	ClientSecret string  `json:"clientSecret"`
	Amount       float64 `json:"amount"`
	Currency     string  `json:"currency"`
}

// PayPalOrderResult represents a PayPal order
type PayPalOrderResult struct {
	ID          string  `json:"id"`
	Status      string  `json:"status"`
	ApprovalURL string  `json:"approvalUrl"`
	Amount      float64 `json:"amount"`
	Currency    string  `json:"currency"`
}

// createStripePaymentIntent is the resolver for creating a Stripe payment intent
func (r *mutationResolver) CreateStripePaymentIntent(ctx context.Context, orderID uuid.UUID) (*PaymentIntentResult, error) {
	stripeSvc := r.Services.Stripe
	if stripeSvc == nil {
		return nil, ErrStripeNotConfigured
	}

	intent, err := stripeSvc.CreatePaymentIntent(ctx, orderID)
	if err != nil {
		return nil, err
	}

	return &PaymentIntentResult{
		ClientSecret: intent.ClientSecret,
		Amount:       intent.Amount,
		Currency:     intent.Currency,
	}, nil
}

// confirmStripePayment is the resolver for confirming a Stripe payment
func (r *mutationResolver) ConfirmStripePayment(ctx context.Context, paymentIntentID string) (*Payment, error) {
	stripeSvc := r.Services.Stripe
	if stripeSvc == nil {
		return nil, ErrStripeNotConfigured
	}

	payment, err := stripeSvc.ConfirmPaymentIntent(ctx, paymentIntentID)
	if err != nil {
		return nil, err
	}

	return convertPaymentToGraphQL(payment), nil
}

// createPayPalOrder is the resolver for creating a PayPal order
func (r *mutationResolver) CreatePayPalOrder(ctx context.Context, orderID uuid.UUID) (*PayPalOrderResult, error) {
	paypalSvc := r.Services.PayPal
	if paypalSvc == nil {
		return nil, ErrPayPalNotConfigured
	}

	ppOrder, err := paypalSvc.CreateOrder(ctx, orderID)
	if err != nil {
		return nil, err
	}

	return &PayPalOrderResult{
		ID:          ppOrder.ID,
		Status:      ppOrder.Status,
		ApprovalURL: ppOrder.ApprovalURL,
		Amount:      ppOrder.Amount,
		Currency:    ppOrder.Currency,
	}, nil
}

// capturePayPalOrder is the resolver for capturing a PayPal order
func (r *mutationResolver) CapturePayPalOrder(ctx context.Context, paypalOrderID string) (*Payment, error) {
	paypalSvc := r.Services.PayPal
	if paypalSvc == nil {
		return nil, ErrPayPalNotConfigured
	}

	payment, err := paypalSvc.CaptureOrder(ctx, paypalOrderID)
	if err != nil {
		return nil, err
	}

	if payment == nil {
		return nil, nil
	}

	return convertPaymentToGraphQL(payment), nil
}

// createCashPayment is the resolver for creating a cash payment
func (r *mutationResolver) CreateCashPayment(ctx context.Context, orderID uuid.UUID, amount float64) (*Payment, error) {
	// Get order
	order, err := r.Services.Order.GetByID(ctx, orderID)
	if err != nil {
		return nil, err
	}

	// Create cash payment
	payment := &model.Payment{
		ID:       uuid.New(),
		OrderID:  orderID,
		Amount:   amount,
		Method:   model.PaymentMethodCash,
		Provider: "", // No provider for cash
		Status:   model.PaymentStatusCompleted, // Cash payments are immediately completed
	}

	// Note: In production, you'd inject the payment repository here
	// For now, this is a placeholder that would need the repository

	// Update order status
	if err := r.Services.Order.UpdateStatus(ctx, orderID, model.OrderPaid); err != nil {
		return nil, err
	}

	s.log.Info("Cash payment created",
		slog.String("order_id", orderID.String()),
		slog.Float64("amount", amount),
	)

	return &Payment{
		ID:                payment.ID.String(),
		OrderID:           order.ID.String(),
		Amount:            payment.Amount,
		Method:            string(payment.Method),
		Provider:          "",
		ProviderPaymentID: "",
		Status:            string(payment.Status),
		CreatedAt:         payment.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}, nil
}

// refundPayment is the resolver for refunding a payment
func (r *mutationResolver) RefundPayment(ctx context.Context, paymentID uuid.UUID) (*PaymentResult, error) {
	stripeSvc := r.Services.Stripe
	if stripeSvc == nil {
		return nil, ErrStripeNotConfigured
	}

	if err := stripeSvc.RefundPayment(ctx, paymentID); err != nil {
		return &PaymentResult{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	return &PaymentResult{
		Success: true,
		Message: "Payment refunded successfully",
	}, nil
}

// getPaymentsByOrder is the resolver for getting payments by order
func (r *queryResolver) GetPaymentsByOrder(ctx context.Context, orderID uuid.UUID) ([]Payment, error) {
	stripeSvc := r.Services.Stripe
	if stripeSvc == nil {
		return nil, ErrStripeNotConfigured
	}

	payments, err := stripeSvc.GetPaymentsByOrderID(ctx, orderID)
	if err != nil {
		return nil, err
	}

	result := make([]Payment, len(payments))
	for i, p := range payments {
		result[i] = *convertPaymentToGraphQL(&p)
	}

	return result, nil
}

// ============ Helper Functions ============

// convertPaymentToGraphQL converts a model.Payment to graph.Payment
func convertPaymentToGraphQL(p *model.Payment) *Payment {
	return &Payment{
		ID:                p.ID.String(),
		OrderID:           p.OrderID.String(),
		Amount:            p.Amount,
		Method:            string(p.Method),
		Provider:          string(p.Provider),
		ProviderPaymentID: p.ProviderPaymentID,
		Status:            string(p.Status),
		CreatedAt:         p.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

// Errors specific to payment operations
var (
	ErrStripeNotConfigured = &PaymentError{Message: "Stripe service not configured"}
	ErrPayPalNotConfigured = &PaymentError{Message: "PayPal service not configured"}
)

// PaymentError represents a payment-related error
type PaymentError struct {
	Message string
}

func (e *PaymentError) Error() string {
	return e.Message
}

// PaymentResolver handles payment-related resolvers
type paymentResolver struct{ *Resolver }

// Order returns OrderResolver implementation
func (r *Resolver) Payment() PaymentResolver { return &paymentResolver{r} }

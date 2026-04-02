package model

import (
	"time"

	"github.com/google/uuid"
)

// PaymentMethod represents the payment method type
type PaymentMethod string

const (
	PaymentMethodCash     PaymentMethod = "cash"
	PaymentMethodCard     PaymentMethod = "card"
	PaymentMethodTransfer PaymentMethod = "transfer"
)

// PaymentProvider represents the payment provider
type PaymentProvider string

const (
	PaymentProviderStripe PaymentProvider = "stripe"
	PaymentProviderPayPal  PaymentProvider = "paypal"
)

// PaymentStatus represents the payment status
type PaymentStatus string

const (
	PaymentStatusPending   PaymentStatus = "pending"
	PaymentStatusCompleted PaymentStatus = "completed"
	PaymentStatusFailed    PaymentStatus = "failed"
	PaymentStatusRefunded  PaymentStatus = "refunded"
)

// Payment represents a payment transaction
type Payment struct {
	ID                uuid.UUID       `json:"id"`
	OrderID           uuid.UUID       `json:"orderId"`
	Amount            float64         `json:"amount"`
	Method            PaymentMethod   `json:"method"`
	Provider          PaymentProvider `json:"provider"`
	ProviderPaymentID string          `json:"providerPaymentId"`
	Status            PaymentStatus   `json:"status"`
	CreatedAt         time.Time       `json:"createdAt"`
	UpdatedAt         time.Time       `json:"updatedAt"`
}

// PaymentIntent represents a Stripe payment intent
type PaymentIntent struct {
	ID           string  `json:"id"`
	ClientSecret string  `json:"clientSecret"`
	Amount       float64 `json:"amount"`
	Currency     string  `json:"currency"`
	Status       string  `json:"status"`
}

// PayPalOrder represents a PayPal order
type PayPalOrder struct {
	ID         string  `json:"id"`
	Status     string  `json:"status"`
	Amount     float64 `json:"amount"`
	Currency   string  `json:"currency"`
	ApprovalURL string `json:"approvalUrl"`
}

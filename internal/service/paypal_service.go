package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/greenpos/backend/internal/model"
	"github.com/greenpos/backend/internal/repository"
)

// PayPalService handles PayPal payment operations
type PayPalService struct {
	payments repository.PaymentRepositoryInterface
	orders   repository.OrderRepositoryInterface
	log      *slog.Logger
	client   *http.Client
	baseURL  string
	clientID string
	secret   string
}

// PayPalAccessToken represents the OAuth access token
type PayPalAccessToken struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
	ExpiresAt   time.Time
}

// NewPayPalService creates a new PayPalService
func NewPayPalService(payments repository.PaymentRepositoryInterface, orders repository.OrderRepositoryInterface, log *slog.Logger) *PayPalService {
	baseURL := "https://api-m.sandbox.paypal.com"
	if os.Getenv("PAYPAL_ENVIRONMENT") == "live" {
		baseURL = "https://api-m.paypal.com"
	}

	return &PayPalService{
		payments: payments,
		orders:   orders,
		log:      log,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL:  baseURL,
		clientID: os.Getenv("PAYPAL_CLIENT_ID"),
		secret:   os.Getenv("PAYPAL_SECRET"),
	}
}

// CreateOrder creates a PayPal order for an order
func (s *PayPalService) CreateOrder(ctx context.Context, orderID uuid.UUID) (*model.PayPalOrder, error) {
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
		Provider: model.PaymentProviderPayPal,
		Status:   model.PaymentStatusPending,
	}

	if err := s.payments.Create(ctx, payment); err != nil {
		return nil, fmt.Errorf("failed to create payment record: %w", err)
	}

	// Get access token
	token, err := s.getAccessToken(ctx)
	if err != nil {
		s.payments.UpdateStatus(ctx, payment.ID, model.PaymentStatusFailed)
		return nil, fmt.Errorf("failed to get PayPal access token: %w", err)
	}

	// Create PayPal order
	ppOrder := s.buildPayPalOrderRequest(payment.ID, orderID, order.Total)
	
	reqBody, err := json.Marshal(ppOrder)
	if err != nil {
		s.payments.UpdateStatus(ctx, payment.ID, model.PaymentStatusFailed)
		return nil, fmt.Errorf("failed to marshal PayPal order: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", s.baseURL+"/v2/checkout/orders", bytes.NewBuffer(reqBody))
	if err != nil {
		s.payments.UpdateStatus(ctx, payment.ID, model.PaymentStatusFailed)
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token.AccessToken)

	resp, err := s.client.Do(req)
	if err != nil {
		s.payments.UpdateStatus(ctx, payment.ID, model.PaymentStatusFailed)
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		s.payments.UpdateStatus(ctx, payment.ID, model.PaymentStatusFailed)
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusCreated {
		s.payments.UpdateStatus(ctx, payment.ID, model.PaymentStatusFailed)
		return nil, fmt.Errorf("PayPal API error: %s - %s", resp.Status, string(body))
	}

	var paypalResp PayPalCreateOrderResponse
	if err := json.Unmarshal(body, &paypalResp); err != nil {
		s.payments.UpdateStatus(ctx, payment.ID, model.PaymentStatusFailed)
		return nil, fmt.Errorf("failed to parse PayPal response: %w", err)
	}

	// Update payment with PayPal order ID
	if err := s.payments.SetProviderPaymentID(ctx, payment.ID, paypalResp.ID); err != nil {
		s.log.Error("failed to set provider payment ID", slog.String("error", err.Error()))
	}

	// Find approval URL
	var approvalURL string
	for _, link := range paypalResp.Links {
		if link.Rel == "approve" {
			approvalURL = link.Href
			break
		}
	}

	s.log.Info("PayPal order created",
		slog.String("payment_id", payment.ID.String()),
		slog.String("paypal_order", paypalResp.ID),
		slog.Float64("amount", order.Total),
	)

	return &model.PayPalOrder{
		ID:          paypalResp.ID,
		Status:      paypalResp.Status,
		Amount:      order.Total,
		Currency:    "MXN",
		ApprovalURL: approvalURL,
	}, nil
}

// CaptureOrder captures a PayPal order after user approval
func (s *PayPalService) CaptureOrder(ctx context.Context, paypalOrderID string) (*model.Payment, error) {
	// Get access token
	token, err := s.getAccessToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get PayPal access token: %w", err)
	}

	// Capture the order
	captureReq := PayPalCaptureRequest{}
	
	reqBody, err := json.Marshal(captureReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal capture request: %w", err)
	}

	url := fmt.Sprintf("%s/v2/checkout/orders/%s/capture", s.baseURL, paypalOrderID)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token.AccessToken)
	req.Header.Set("PayPal-Request-Id", uuid.New().String())

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send capture request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("PayPal capture error: %s - %s", resp.Status, string(body))
	}

	var captureResp PayPalCaptureResponse
	if err := json.Unmarshal(body, &captureResp); err != nil {
		return nil, fmt.Errorf("failed to parse capture response: %w", err)
	}

	// Find payment by provider payment ID and update status
	// Since we don't have direct DB access here, we'll return the payment info
	// The webhook handler will actually update the DB
	if captureResp.Status != "COMPLETED" {
		s.log.Warn("PayPal order not completed", slog.String("status", captureResp.Status))
	}

	// Update payment status
	payments, err := s.payments.GetByOrderID(ctx, uuid.Nil)
	if err == nil {
		for _, p := range payments {
			if p.ProviderPaymentID == paypalOrderID {
				status := model.PaymentStatusCompleted
				if captureResp.Status == "VOIDED" || captureResp.Status == "DECLINED" {
					status = model.PaymentStatusFailed
				}
				s.payments.UpdateStatus(ctx, p.ID, status)
				
				// Update order status
				if status == model.PaymentStatusCompleted {
					s.orders.UpdateStatus(ctx, p.OrderID, model.OrderPaid)
				}
				
				s.log.Info("PayPal order captured",
					slog.String("payment_id", p.ID.String()),
					slog.String("paypal_order", paypalOrderID),
					slog.String("status", string(status)),
				)
				return &p, nil
			}
		}
	}

	s.log.Info("PayPal order captured (payment record update pending via webhook)",
		slog.String("paypal_order", paypalOrderID),
		slog.String("status", captureResp.Status),
	)

	return nil, nil
}

// HandleWebhook processes PayPal webhook events
func (s *PayPalService) HandleWebhook(ctx context.Context, event map[string]interface{}) error {
	eventType, ok := event["event_type"].(string)
	if !ok {
		return fmt.Errorf("missing event_type in webhook payload")
	}

	s.log.Info("PayPal webhook received", slog.String("event_type", eventType))

	switch eventType {
	case "CHECKOUT.ORDER.APPROVED":
		s.log.Info("PayPal order approved (awaiting capture)")
		
	case "PAYMENT.CAPTURE.COMPLETED":
		resource := event["resource"].(map[string]interface{})
		if orderID, ok := resource["custom_id"].(string); ok {
			s.handlePaymentCaptured(ctx, orderID, resource)
		}
		
	case "PAYMENT.CAPTURE.DENIED", "PAYMENT.CAPTURE.DECLINED":
		resource := event["resource"].(map[string]interface{})
		if customID, ok := resource["custom_id"].(string); ok {
			s.handlePaymentDenied(ctx, customID)
		}
		
	default:
		s.log.Info("Unhandled PayPal event type", slog.String("type", eventType))
	}

	return nil
}

func (s *PayPalService) handlePaymentCaptured(ctx context.Context, paymentIDStr string, resource map[string]interface{}) error {
	paymentID, err := uuid.Parse(paymentIDStr)
	if err != nil {
		return fmt.Errorf("invalid payment_id: %w", err)
	}

	if err := s.payments.UpdateStatus(ctx, paymentID, model.PaymentStatusCompleted); err != nil {
		return fmt.Errorf("failed to update payment status: %w", err)
	}

	s.log.Info("PayPal payment captured via webhook",
		slog.String("payment_id", paymentIDStr),
	)

	return nil
}

func (s *PayPalService) handlePaymentDenied(ctx context.Context, paymentIDStr string) error {
	paymentID, err := uuid.Parse(paymentIDStr)
	if err != nil {
		return fmt.Errorf("invalid payment_id: %w", err)
	}

	if err := s.payments.UpdateStatus(ctx, paymentID, model.PaymentStatusFailed); err != nil {
		return fmt.Errorf("failed to update payment status: %w", err)
	}

	s.log.Info("PayPal payment denied via webhook",
		slog.String("payment_id", paymentIDStr),
	)

	return nil
}

// getAccessToken retrieves a PayPal OAuth access token
func (s *PayPalService) getAccessToken(ctx context.Context) (*PayPalAccessToken, error) {
	req, err := http.NewRequestWithContext(ctx, "POST", s.baseURL+"/v1/oauth2/token", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create token request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(s.clientID, s.secret)

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get access token: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("PayPal token error: %s - %s", resp.Status, string(body))
	}

	var token PayPalAccessToken
	if err := json.NewDecoder(resp.Body).Decode(&token); err != nil {
		return nil, fmt.Errorf("failed to parse token response: %w", err)
	}

	token.ExpiresAt = time.Now().Add(time.Duration(token.ExpiresIn) * time.Second)
	return &token, nil
}

// buildPayPalOrderRequest creates the PayPal order request body
func (s *PayPalService) buildPayPalOrderRequest(paymentID, orderID uuid.UUID, amount float64) PayPalCreateOrderRequest {
	return PayPalCreateOrderRequest{
		Intent: "CAPTURE",
		PurchaseUnits: []PayPalPurchaseUnit{
			{
				ReferenceID: paymentID.String(),
				CustomID:    paymentID.String(), // Maps to our payment ID
				InvoiceID:   orderID.String(),
				Amount: PayPalAmount{
					CurrencyCode: "MXN",
					Value:         fmt.Sprintf("%.2f", amount),
				},
			},
		},
		PaymentSource: &PayPalPaymentSource{
			PayPal: &PayPalPayPalPaymentSource{
				ExperienceContext: PayPalExperienceContext{
					PaymentMethodPreference: "IMMEDIATE_PAYMENT_REQUIRED",
					UserAction:              "PAY_NOW",
				},
			},
		},
	}
}

// GetPaymentsByOrderID retrieves payments for an order
func (s *PayPalService) GetPaymentsByOrderID(ctx context.Context, orderID uuid.UUID) ([]model.Payment, error) {
	return s.payments.GetByOrderID(ctx, orderID)
}

// ============ PayPal API Types ============

// PayPalCreateOrderRequest represents a PayPal order creation request
type PayPalCreateOrderRequest struct {
	Intent        string                  `json:"intent"`
	PurchaseUnits []PayPalPurchaseUnit     `json:"purchase_units"`
	PaymentSource *PayPalPaymentSource     `json:"payment_source,omitempty"`
}

// PayPalPurchaseUnit represents a purchase unit in a PayPal order
type PayPalPurchaseUnit struct {
	ReferenceID string       `json:"reference_id"`
	CustomID    string       `json:"custom_id"`
	InvoiceID   string       `json:"invoice_id"`
	Amount      PayPalAmount `json:"amount"`
}

// PayPalAmount represents an amount in PayPal
type PayPalAmount struct {
	CurrencyCode string `json:"currency_code"`
	Value         string `json:"value"`
}

// PayPalPaymentSource represents the payment source
type PayPalPaymentSource struct {
	PayPal *PayPalPayPalPaymentSource `json:"paypal,omitempty"`
}

// PayPalPayPalPaymentSource represents PayPal-specific payment source
type PayPalPayPalPaymentSource struct {
	ExperienceContext PayPalExperienceContext `json:"experience_context"`
}

// PayPalExperienceContext represents the experience context
type PayPalExperienceContext struct {
	PaymentMethodPreference string `json:"payment_method_preference"`
	UserAction              string `json:"user_action"`
}

// PayPalCreateOrderResponse represents a PayPal order creation response
type PayPalCreateOrderResponse struct {
	ID     string        `json:"id"`
	Status string        `json:"status"`
	Links  []PayPalLinks `json:"links"`
}

// PayPalLinks represents a link in PayPal response
type PayPalLinks struct {
	Href   string `json:"href"`
	Rel    string `json:"rel"`
	Method string `json:"method"`
}

// PayPalCaptureRequest represents a PayPal capture request
type PayPalCaptureRequest struct {
}

// PayPalCaptureResponse represents a PayPal capture response
type PayPalCaptureResponse struct {
	ID            string                 `json:"id"`
	Status        string                 `json:"status"`
	PurchaseUnits []PayPalPurchaseUnit   `json:"purchase_units"`
}

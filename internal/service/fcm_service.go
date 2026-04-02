package service

import (
	"context"
	"errors"
	"log/slog"

	"github.com/google/uuid"
	"github.com/greenpos/backend/internal/model"
	"github.com/greenpos/backend/internal/repository"
)

// FCM Service Errors
var (
	ErrFCMTokenNotFound   = errors.New("fcm token not found")
	ErrFCMServiceNotReady = errors.New("fcm service not initialized")
	ErrInvalidToken       = errors.New("invalid fcm token")
)

// FCMService handles Firebase Cloud Messaging operations
type FCMService struct {
	tokens repository.FCMTokenRepositoryInterface
	users  repository.UserRepositoryInterface
	log    *slog.Logger
	// Firebase Admin SDK client (optional - nil for memory implementation)
	firebaseApp interface{}
}

// NewFCMService creates a new FCM service
func NewFCMService(
	tokens repository.FCMTokenRepositoryInterface,
	users repository.UserRepositoryInterface,
	log *slog.Logger,
) *FCMService {
	return &FCMService{
		tokens: tokens,
		users:  users,
		log:    log,
	}
}

// SaveToken saves or updates an FCM token for a user
func (s *FCMService) SaveToken(ctx context.Context, userID uuid.UUID, fcmToken, device string) error {
	token := &model.FCMToken{
		UserID:   userID,
		FCMToken: fcmToken,
		Device:   device,
		IsActive: true,
	}

	if err := s.tokens.Upsert(ctx, token); err != nil {
		return err
	}

	s.log.Info("fcm token saved",
		slog.String("userId", userID.String()),
		slog.String("device", device))
	return nil
}

// RemoveToken removes (deactivates) an FCM token for a user
func (s *FCMService) RemoveToken(ctx context.Context, userID uuid.UUID, fcmToken string) error {
	tokens, err := s.tokens.GetByUserID(ctx, userID)
	if err != nil {
		return err
	}

	for _, t := range tokens {
		if t.FCMToken == fcmToken {
			return s.tokens.Deactivate(ctx, t.ID)
		}
	}

	return ErrFCMTokenNotFound
}

// GetUserTokens retrieves all active FCM tokens for a user
func (s *FCMService) GetUserTokens(ctx context.Context, userID uuid.UUID) ([]model.FCMToken, error) {
	return s.tokens.GetActiveByUserID(ctx, userID)
}

// NotifyOrderReady sends a push notification to the waiter when order is ready
func (s *FCMService) NotifyOrderReady(ctx context.Context, orderID uuid.UUID, waiterID uuid.UUID, tableName string) error {
	tokens, err := s.tokens.GetActiveByUserID(ctx, waiterID)
	if err != nil {
		return err
	}

	if len(tokens) == 0 {
		s.log.Warn("no fcm tokens for waiter", slog.String("waiterId", waiterID.String()))
		return ErrFCMTokenNotFound
	}

	notification := model.FCMNotification{
		Title: "🍽️ Orden Lista",
		Body:  "La orden de la mesa " + tableName + " está lista para servir",
		Data: map[string]string{
			"type":      "ORDER_READY",
			"orderId":   orderID.String(),
			"waiterId":  waiterID.String(),
			"tableName": tableName,
		},
	}

	return s.sendToTokens(ctx, tokens, notification)
}

// NotifyNewOrder sends push notification to kitchen staff when a new order is created
func (s *FCMService) NotifyNewOrder(ctx context.Context, branchID uuid.UUID, orderID uuid.UUID, tableName string, itemCount int) error {
	tokens, err := s.tokens.GetByBranchAndRole(ctx, branchID, model.RoleKitchen)
	if err != nil {
		return err
	}

	if len(tokens) == 0 {
		s.log.Warn("no fcm tokens for kitchen", slog.String("branchId", branchID.String()))
		return ErrFCMTokenNotFound
	}

	notification := model.FCMNotification{
		Title: "📋 Nueva Orden",
		Body:  "Nueva orden para la mesa " + tableName + " (" + string(rune(itemCount+'0')) + " items)",
		Data: map[string]string{
			"type":      "NEW_ORDER",
			"orderId":   orderID.String(),
			"branchId":  branchID.String(),
			"tableName": tableName,
		},
	}

	return s.sendToTokens(ctx, tokens, notification)
}

// NotifyNewReservation sends push notification to managers when a new reservation is created
func (s *FCMService) NotifyNewReservation(ctx context.Context, branchID uuid.UUID, reservationID uuid.UUID, customerName string, guestCount int, time string) error {
	// Get managers and owners
	managerTokens, err := s.tokens.GetByBranchAndRole(ctx, branchID, model.RoleManager)
	if err != nil {
		return err
	}
	ownerTokens, err := s.tokens.GetByBranchAndRole(ctx, branchID, model.RoleOwner)
	if err != nil {
		return err
	}
	adminTokens, err := s.tokens.GetByBranchAndRole(ctx, branchID, model.RoleAdmin)
	if err != nil {
		return err
	}

	allTokens := append(managerTokens, ownerTokens...)
	allTokens = append(allTokens, adminTokens...)

	if len(allTokens) == 0 {
		s.log.Warn("no fcm tokens for managers", slog.String("branchId", branchID.String()))
		return ErrFCMTokenNotFound
	}

	notification := model.FCMNotification{
		Title: "📅 Nueva Reservación",
		Body:  "Reservación de " + customerName + " para " + string(rune(guestCount+'0')) + " personas a las " + time,
		Data: map[string]string{
			"type":          "NEW_RESERVATION",
			"reservationId": reservationID.String(),
			"branchId":      branchID.String(),
		},
	}

	return s.sendToTokens(ctx, allTokens, notification)
}

// SendPushNotification sends a custom push notification to a user
func (s *FCMService) SendPushNotification(ctx context.Context, userID uuid.UUID, title, body string, data map[string]string) error {
	tokens, err := s.tokens.GetActiveByUserID(ctx, userID)
	if err != nil {
		return err
	}

	if len(tokens) == 0 {
		return ErrFCMTokenNotFound
	}

	notification := model.FCMNotification{
		Title: title,
		Body:  body,
		Data:  data,
	}

	return s.sendToTokens(ctx, tokens, notification)
}

// sendToTokens sends a notification to all provided FCM tokens
// This is where the actual Firebase Admin SDK would be called
func (s *FCMService) sendToTokens(ctx context.Context, tokens []model.FCMToken, notification model.FCMNotification) error {
	if s.firebaseApp == nil {
		// Development mode - just log the notification
		s.log.Info("fcm notification (dev mode)",
			slog.String("title", notification.Title),
			slog.String("body", notification.Body),
			slog.Any("data", notification.Data),
			slog.Int("tokenCount", len(tokens)))
		return nil
	}

	// Production mode - send via Firebase Admin SDK
	for _, token := range tokens {
		if err := s.sendViaFirebase(ctx, token.FCMToken, notification); err != nil {
			s.log.Error("failed to send fcm notification",
				slog.String("token", token.FCMToken[:20]+"..."),
				slog.String("error", err.Error()))
			// Continue with other tokens even if one fails
		}
	}

	return nil
}

// sendViaFirebase sends a notification via Firebase Admin SDK
// This is a placeholder - in production, this would use firebase.google.com/go/messaging
func (s *FCMService) sendViaFirebase(ctx context.Context, fcmToken string, notification model.FCMNotification) error {
	// TODO: Implement actual Firebase Admin SDK integration
	// import firebase "firebase.google.com/go/v4"
	// import "firebase.google.com/go/v4/messaging"
	//
	// client, err := messaging.NewClient(s.firebaseApp)
	// if err != nil {
	//     return err
	// }
	//
	// message := &messaging.Message{
	//     Token: fcmToken,
	//     Notification: &messaging.Notification{
	//         Title: notification.Title,
	//         Body:  notification.Body,
	//     },
	//     Data: notification.Data,
	// }
	//
	// return client.Send(ctx, message)

	s.log.Info("fcm send via firebase (placeholder)",
		slog.String("token", fcmToken[:20]+"..."),
		slog.String("title", notification.Title))

	return nil
}

// SetFirebaseApp sets the Firebase Admin SDK app instance
// Call this during server initialization to enable actual FCM sending
func (s *FCMService) SetFirebaseApp(app interface{}) {
	s.firebaseApp = app
	s.log.Info("fcm service configured with firebase app")
}

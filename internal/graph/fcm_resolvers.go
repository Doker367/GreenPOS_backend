package graph

import (
	"context"

	"github.com/google/uuid"
	"github.com/greenpos/backend/internal/model"
)

// SaveFCMToken saves an FCM token for the authenticated user
func (r *mutationResolver) SaveFCMToken(ctx context.Context, token string, device *string) (*SaveFCMTokenResult, error) {
	userID, err := GetUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	deviceName := "unknown"
	if device != nil {
		deviceName = *device
	}

	if err := r.Services.FCM.SaveToken(ctx, userID, token, deviceName); err != nil {
		return nil, err
	}

	return &SaveFCMTokenResult{
		Success: true,
		Message: "Token saved successfully",
	}, nil
}

// RemoveFCMToken removes (deactivates) an FCM token for the authenticated user
func (r *mutationResolver) RemoveFCMToken(ctx context.Context, token string) (*SaveFCMTokenResult, error) {
	userID, err := GetUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if err := r.Services.FCM.RemoveToken(ctx, userID, token); err != nil {
		return &SaveFCMTokenResult{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	return &SaveFCMTokenResult{
		Success: true,
		Message: "Token removed successfully",
	}, nil
}

// SendPushNotification sends a push notification to a specific user
func (r *mutationResolver) SendPushNotification(ctx context.Context, userID uuid.UUID, title string, body string) (*SaveFCMTokenResult, error) {
	data := map[string]string{
		"title": title,
		"body":  body,
	}

	if err := r.Services.FCM.SendPushNotification(ctx, userID, title, body, data); err != nil {
		return &SaveFCMTokenResult{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	return &SaveFCMTokenResult{
		Success: true,
		Message: "Notification sent",
	}, nil
}

// SaveFCMTokenResult is the return type for FCM token mutations
type SaveFCMTokenResult struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

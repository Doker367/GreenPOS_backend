package model

import (
	"time"

	"github.com/google/uuid"
)

// FCMToken represents a user's FCM registration token for push notifications
type FCMToken struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"userId"`
	FCMToken  string    `json:"fcmToken"`
	Device    string    `json:"device"`
	IsActive  bool      `json:"isActive"`
	UpdatedAt time.Time `json:"updatedAt"`
	CreatedAt time.Time `json:"createdAt"`
}

// FCMNotification represents a push notification to be sent
type FCMNotification struct {
	Title string            `json:"title"`
	Body  string            `json:"body"`
	Data  map[string]string `json:"data"`
}

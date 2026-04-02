package graph

import (
	"context"

	"github.com/google/uuid"
	"github.com/greenpos/backend/internal/service"
)

// This file serves as dependency injection for your app, add any dependencies you require here.

// Resolver is the root resolver with all services
type Resolver struct {
	Services  *service.Services
	JWTSecret string
}

// NewResolvers creates a new resolver instance
func NewResolvers(services *service.Services, jwtSecret string) *Resolver {
	return &Resolver{
		Services:  services,
		JWTSecret: jwtSecret,
	}
}

// Context keys for JWT data
type contextKey string

const (
	ContextKeyUserID   contextKey = "userID"
	ContextKeyTenantID contextKey = "tenantID"
	ContextKeyBranchID contextKey = "branchID"
)

// GetUserIDFromContext extracts user ID from JWT context
func GetUserIDFromContext(ctx context.Context) (uuid.UUID, error) {
	if userID, ok := ctx.Value(ContextKeyUserID).(uuid.UUID); ok {
		return userID, nil
	}
	return uuid.Nil, ErrNotAuthenticated
}

// GetTenantIDFromContext extracts tenant ID from JWT context
func GetTenantIDFromContext(ctx context.Context) (uuid.UUID, error) {
	if tenantID, ok := ctx.Value(ContextKeyTenantID).(uuid.UUID); ok {
		return tenantID, nil
	}
	return uuid.Nil, ErrNotAuthenticated
}

// GetBranchIDFromContext extracts branch ID from context
func GetBranchIDFromContext(ctx context.Context) (uuid.UUID, error) {
	if branchID, ok := ctx.Value(ContextKeyBranchID).(uuid.UUID); ok {
		return branchID, nil
	}
	return uuid.Nil, ErrNoBranchContext
}

// Errors
var (
	ErrNotAuthenticated = &AuthenticationError{Message: "not authenticated"}
	ErrNoBranchContext  = &AuthenticationError{Message: "no branch context"}
	ErrForbidden        = &ForbiddenError{Message: "forbidden"}
	ErrNotFound         = &NotFoundError{Message: "not found"}
)

// AuthenticationError represents an authentication error
type AuthenticationError struct {
	Message string
}

func (e *AuthenticationError) Error() string {
	return e.Message
}

// ForbiddenError represents a forbidden error
type ForbiddenError struct {
	Message string
}

func (e *ForbiddenError) Error() string {
	return e.Message
}

// NotFoundError represents a not found error
type NotFoundError struct {
	Message string
}

func (e *NotFoundError) Error() string {
	return e.Message
}

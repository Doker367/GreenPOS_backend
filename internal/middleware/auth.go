package middleware

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/valyala/fasthttp"
	"golang.org/x/crypto/bcrypt"
)

// JWTConfig holds JWT configuration
type JWTConfig struct {
	Secret               []byte
	ExpiryHours          int
	RefreshTokenExpiryDays int
}

// JWTClaims represents the claims in a JWT token
type JWTClaims struct {
	UserID    uuid.UUID `json:"userId"`
	TenantID  uuid.UUID `json:"tenantId"`
	BranchID  uuid.UUID `json:"branchId"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	TokenType string    `json:"tokenType"` // "access" or "refresh"
	jwt.RegisteredClaims
}

// AuthService handles JWT token operations
type AuthService struct {
	config *JWTConfig
	log    *slog.Logger
}

// NewAuthService creates a new auth service
func NewAuthService(config *JWTConfig, log *slog.Logger) *AuthService {
	return &AuthService{
		config: config,
		log:    log,
	}
}

// GenerateTokenPair generates both access and refresh tokens
func (s *AuthService) GenerateTokenPair(user *TokenUser) (accessToken string, refreshToken string, err error) {
	// Generate access token
	accessToken, err = s.generateToken(user, "access", time.Duration(s.config.ExpiryHours)*time.Hour)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate access token: %w", err)
	}

	// Generate refresh token
	refreshToken, err = s.generateToken(user, "refresh", time.Duration(s.config.RefreshTokenExpiryDays)*24*time.Hour)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return accessToken, refreshToken, nil
}

// generateToken creates a JWT token
func (s *AuthService) generateToken(user *TokenUser, tokenType string, expiry time.Duration) (string, error) {
	now := time.Now()

	claims := JWTClaims{
		UserID:    user.ID,
		TenantID:  user.TenantID,
		BranchID:  user.BranchID,
		Email:     user.Email,
		Role:      user.Role,
		TokenType: tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(expiry)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "greenpos",
			Subject:   user.ID.String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.config.Secret)
}

// ValidateToken validates a JWT token and returns the claims
func (s *AuthService) ValidateToken(ctx context.Context, tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.config.Secret, nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}

	return claims, nil
}

// ValidateAccessToken validates an access token
func (s *AuthService) ValidateAccessToken(ctx context.Context, tokenString string) (*JWTClaims, error) {
	claims, err := s.ValidateToken(ctx, tokenString)
	if err != nil {
		return nil, err
	}

	if claims.TokenType != "access" {
		return nil, fmt.Errorf("token is not an access token")
	}

	return claims, nil
}

// ValidateRefreshToken validates a refresh token and returns the claims
func (s *AuthService) ValidateRefreshToken(ctx context.Context, tokenString string) (*JWTClaims, error) {
	claims, err := s.ValidateToken(ctx, tokenString)
	if err != nil {
		return nil, err
	}

	if claims.TokenType != "refresh" {
		return nil, fmt.Errorf("token is not a refresh token")
	}

	return claims, nil
}

// RefreshAccessToken generates a new access token from a valid refresh token
func (s *AuthService) RefreshAccessToken(ctx context.Context, refreshToken string, user *TokenUser) (string, error) {
	// Validate the refresh token first
	claims, err := s.ValidateRefreshToken(ctx, refreshToken)
	if err != nil {
		return "", fmt.Errorf("invalid refresh token: %w", err)
	}

	// Verify the token belongs to the user
	if claims.UserID != user.ID {
		return "", fmt.Errorf("refresh token does not belong to user")
	}

	// Generate new access token
	return s.generateToken(user, "access", time.Duration(s.config.ExpiryHours)*time.Hour)
}

// HashPassword hashes a password using bcrypt with configurable cost
func HashPassword(password string, cost int) (string, error) {
	if cost < 4 {
		cost = bcrypt.DefaultCost
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), cost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(hash), nil
}

// VerifyPassword verifies a password against a hash
func VerifyPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// HashRefreshToken hashes a refresh token for storage
func HashRefreshToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

// TokenUser represents user information needed for token generation
type TokenUser struct {
	ID       uuid.UUID
	TenantID uuid.UUID
	BranchID uuid.UUID
	Email    string
	Role     string
}

// ExtractTenantID extracts tenant ID from request context
func ExtractTenantID(ctx *fasthttp.RequestCtx) uuid.UUID {
	if tenantStr, ok := ctx.UserValue("tenant_id").(string); ok {
		id, _ := uuid.Parse(tenantStr)
		return id
	}
	return uuid.Nil
}

// ExtractUserID extracts user ID from request context
func ExtractUserIDFromCtx(ctx *fasthttp.RequestCtx) uuid.UUID {
	if userStr, ok := ctx.UserValue("user_id").(string); ok {
		id, _ := uuid.Parse(userStr)
		return id
	}
	return uuid.Nil
}

// ExtractBranchID extracts branch ID from request context
func ExtractBranchID(ctx *fasthttp.RequestCtx) uuid.UUID {
	if branchStr, ok := ctx.UserValue("branch_id").(string); ok {
		id, _ := uuid.Parse(branchStr)
		return id
	}
	return uuid.Nil
}

// ExtractRole extracts role from request context
func ExtractRole(ctx *fasthttp.RequestCtx) string {
	if role, ok := ctx.UserValue("role").(string); ok {
		return role
	}
	return ""
}

// RequireAuth creates a middleware that requires valid authentication
func RequireAuth(authService *AuthService) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		auth := string(ctx.Request.Header.Peek("Authorization"))
		if auth == "" {
			ctx.SetStatusCode(fasthttp.StatusUnauthorized)
			ctx.SetContentType("application/json")
			ctx.WriteString(`{"error":"authorization header required"}`)
			return
		}

		tokenString := strings.TrimPrefix(auth, "Bearer ")
		if tokenString == auth {
			ctx.SetStatusCode(fasthttp.StatusUnauthorized)
			ctx.SetContentType("application/json")
			ctx.WriteString(`{"error":"bearer token required"}`)
			return
		}

		claims, err := authService.ValidateAccessToken(context.Background(), tokenString)
		if err != nil {
			ctx.SetStatusCode(fasthttp.StatusUnauthorized)
			ctx.SetContentType("application/json")
			ctx.WriteString(`{"error":"invalid or expired token"}`)
			return
		}

		// Set user values in context
		ctx.SetUserValue("user_id", claims.UserID.String())
		ctx.SetUserValue("tenant_id", claims.TenantID.String())
		ctx.SetUserValue("branch_id", claims.BranchID.String())
		ctx.SetUserValue("role", claims.Role)
		ctx.SetUserValue("email", claims.Email)
	}
}

// RequireRole creates a middleware that requires specific roles
func RequireRole(roles ...string) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		role := ExtractRole(ctx)
		if role == "" {
			ctx.SetStatusCode(fasthttp.StatusForbidden)
			ctx.SetContentType("application/json")
			ctx.WriteString(`{"error":"role not found in context"}`)
			return
		}

		for _, r := range roles {
			if role == r {
				return // Role is allowed
			}
		}

		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetContentType("application/json")
		ctx.WriteString(`{"error":"insufficient permissions"}`)
	}
}

// RequireTenant ensures a tenant ID is present in the context
func RequireTenant() fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		tenantID := ExtractTenantID(ctx)
		if tenantID == uuid.Nil {
			ctx.SetStatusCode(fasthttp.StatusForbidden)
			ctx.SetContentType("application/json")
			ctx.WriteString(`{"error":"tenant context required"}`)
			return
		}
	}
}

// GetUserFromToken extracts user information from a valid JWT token
func (s *AuthService) GetUserFromToken(ctx context.Context, tokenString string) (*TokenUser, error) {
	claims, err := s.ValidateAccessToken(ctx, tokenString)
	if err != nil {
		return nil, err
	}

	return &TokenUser{
		ID:       claims.UserID,
		TenantID:  claims.TenantID,
		BranchID:  claims.BranchID,
		Email:    claims.Email,
		Role:     claims.Role,
	}, nil
}

// IsTokenExpired checks if a token is expired without full validation
func IsTokenExpired(tokenString string, secret []byte) bool {
	token, _, err := jwt.NewParser().ParseUnverified(tokenString, &JWTClaims{})
	if err != nil {
		return true
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok {
		return true
	}

	return claims.ExpiresAt != nil && claims.ExpiresAt.Time.Before(time.Now())
}

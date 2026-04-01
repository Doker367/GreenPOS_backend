package middleware

import (
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/valyala/fasthttp"
)

func ExtractTenantID(ctx *fasthttp.RequestCtx, secret string) uuid.UUID {
	auth := string(ctx.Request.Header.Peek("Authorization"))
	if auth == "" {
		return uuid.Nil
	}

	tokenString := strings.TrimPrefix(auth, "Bearer ")
	if tokenString == auth {
		return uuid.Nil // No Bearer prefix
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	if err != nil || !token.Valid {
		return uuid.Nil
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return uuid.Nil
	}

	tenantStr, ok := claims["tenant"].(string)
	if !ok {
		return uuid.Nil
	}

	tenantID, err := uuid.Parse(tenantStr)
	if err != nil {
		return uuid.Nil
	}

	return tenantID
}

func ExtractUserID(ctx *fasthttp.RequestCtx, secret string) (uuid.UUID, uuid.UUID, uuid.UUID, string) {
	auth := string(ctx.Request.Header.Peek("Authorization"))
	if auth == "" {
		return uuid.Nil, uuid.Nil, uuid.Nil, ""
	}

	tokenString := strings.TrimPrefix(auth, "Bearer ")
	if tokenString == auth {
		return uuid.Nil, uuid.Nil, uuid.Nil, ""
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	if err != nil || !token.Valid {
		return uuid.Nil, uuid.Nil, uuid.Nil, ""
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return uuid.Nil, uuid.Nil, uuid.Nil, ""
	}

	parseUUID := func(key string) uuid.UUID {
		if str, ok := claims[key].(string); ok {
			id, _ := uuid.Parse(str)
			return id
		}
		return uuid.Nil
	}

	userID := parseUUID("sub")
	tenantID := parseUUID("tenant")
	branchID := parseUUID("branch")
	role, _ := claims["role"].(string)

	return userID, tenantID, branchID, role
}

func RequireAuth(secret string) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		userID, _, _, role := ExtractUserID(ctx, secret)
		if userID == uuid.Nil {
			ctx.SetStatusCode(fasthttp.StatusUnauthorized)
			ctx.WriteString(`{"error":"unauthorized"}`)
			return
		}
		ctx.SetUserValue("user_id", userID)
		ctx.SetUserValue("role", role)
	}
}

func RequireRole(roles ...string) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		role, _ := ctx.UserValue("role").(string)
		for _, r := range roles {
			if role == r {
				return
			}
		}
		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.WriteString(`{"error":"forbidden"}`)
	}
}

package middleware

import (
	"context"
	"net/http"
	"strings"

	"marketplace-platform/utils"
)

type contextKey string

const (
	ContextUserID   contextKey = "userID"
	ContextRole     contextKey = "role"
	ContextEmail    contextKey = "email"
)

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			next.ServeHTTP(w, r)
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
			next.ServeHTTP(w, r)
			return
		}

		claims, err := utils.ValidateToken(parts[1])
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}

		ctx := context.WithValue(r.Context(), ContextUserID, claims.ID)
		ctx = context.WithValue(ctx, ContextRole, claims.Role)
		ctx = context.WithValue(ctx, ContextEmail, claims.Email)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func RequireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if GetUserID(r) == 0 {
			http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
			return
		}
		next(w, r)
	}
}

func RequireRole(role string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if GetRole(r) != role {
			http.Error(w, `{"error":"forbidden"}`, http.StatusForbidden)
			return
		}
		next(w, r)
	}
}

func GetUserID(r *http.Request) uint {
	v, _ := r.Context().Value(ContextUserID).(uint)
	return v
}

func GetRole(r *http.Request) string {
	v, _ := r.Context().Value(ContextRole).(string)
	return v
}

func GetEmail(r *http.Request) string {
	v, _ := r.Context().Value(ContextEmail).(string)
	return v
}

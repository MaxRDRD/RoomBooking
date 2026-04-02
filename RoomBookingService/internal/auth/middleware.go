package auth

import (
	"RoomBookingService/internal/httpresp"
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

type contextKey string

const (
	userIDContextKey contextKey = "auth_user_id"
	roleContextKey   contextKey = "auth_role"
)

func AuthMiddleware(tokenService TokenService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				httpresp.WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing authorization header")
				return
			}

			if !strings.HasPrefix(authHeader, "Bearer ") {
				httpresp.WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing Bearer")
				return
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
				httpresp.WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "invalid authorization header")
				return
			}

			userID, role, err := tokenService.ParseToken(parts[1])
			if err != nil {
				httpresp.WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "invalid token")
				return
			}

			ctx := context.WithValue(r.Context(), userIDContextKey, userID)
			ctx = context.WithValue(ctx, roleContextKey, role)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func RequireRole(allowedRole string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			role, err := RoleFromContext(r.Context())
			if err != nil {
				httpresp.WriteError(w, http.StatusForbidden, "FORBIDDEN", "forbidden")
				return
			}
			if role != allowedRole {
				httpresp.WriteError(w, http.StatusForbidden, "FORBIDDEN", "forbidden")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func UserIDFromContext(ctx context.Context) (uuid.UUID, error) {
	userID, ok := ctx.Value(userIDContextKey).(uuid.UUID)
	if !ok {
		return uuid.Nil, errors.New("user id not found in context")
	}
	return userID, nil
}

func RoleFromContext(ctx context.Context) (string, error) {
	role, ok := ctx.Value(roleContextKey).(string)
	if !ok || strings.TrimSpace(role) == "" {
		return "", errors.New("role not found in context")
	}
	return role, nil
}

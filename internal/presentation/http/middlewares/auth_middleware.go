package middlewares

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/farmanexo/order-service/internal/domain/services"
	"github.com/farmanexo/order-service/internal/shared/common"
	"github.com/farmanexo/order-service/internal/shared/constants"
	"github.com/farmanexo/order-service/pkg/mediator"
	"go.uber.org/zap"
)

type contextKey string

const (
	UserIDCtxKey      contextKey = "user_id"
	UserRoleCtxKey    contextKey = "user_role"
	AccessTokenCtxKey contextKey = "access_token"
)

type AuthMiddleware struct {
	jwtService services.JWTService
	logger     *zap.Logger
}

func NewAuthMiddleware(jwtService services.JWTService, logger *zap.Logger) *AuthMiddleware {
	return &AuthMiddleware{jwtService: jwtService, logger: logger}
}

func (m *AuthMiddleware) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			m.respondUnauthorized(w, "Header Authorization es requerido")
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			m.respondUnauthorized(w, "Formato de token inválido. Use: Bearer {token}")
			return
		}

		tokenString := parts[1]
		if tokenString == "" {
			m.respondUnauthorized(w, "Token vacío")
			return
		}

		userID, role, _, err := m.jwtService.ValidateAccessToken(tokenString)
		if err != nil {
			m.respondUnauthorized(w, "Token inválido o expirado")
			return
		}

		ctx := context.WithValue(r.Context(), UserIDCtxKey, userID)
		ctx = context.WithValue(ctx, UserRoleCtxKey, role)
		ctx = context.WithValue(ctx, AccessTokenCtxKey, tokenString)
		ctx = mediator.WithValue(ctx, mediator.UserIDKey, userID)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (m *AuthMiddleware) RequireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		role, ok := GetUserRoleFromContext(r.Context())
		if !ok || role != "admin" {
			m.respondForbidden(w, "No tiene permisos para esta acción. Se requiere rol de administrador")
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (m *AuthMiddleware) RequirePharmacyOwner(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		role, ok := GetUserRoleFromContext(r.Context())
		if !ok || (role != "pharmacy_owner" && role != "admin") {
			m.respondForbidden(w, "No tiene permisos para esta acción. Se requiere rol de encargado de farmacia")
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (m *AuthMiddleware) respondUnauthorized(w http.ResponseWriter, message string) {
	resp := common.UnauthorizedResponse[any](message)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(constants.StatusUnauthorized.Int())
	json.NewEncoder(w).Encode(resp)
}

func (m *AuthMiddleware) respondForbidden(w http.ResponseWriter, message string) {
	resp := common.ForbiddenResponse[any](message)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(constants.StatusForbidden.Int())
	json.NewEncoder(w).Encode(resp)
}

func GetUserIDFromContext(ctx context.Context) (string, bool) {
	val := ctx.Value(UserIDCtxKey)
	if val == nil {
		return "", false
	}
	userID, ok := val.(string)
	return userID, ok
}

func GetUserRoleFromContext(ctx context.Context) (string, bool) {
	val := ctx.Value(UserRoleCtxKey)
	if val == nil {
		return "", false
	}
	role, ok := val.(string)
	return role, ok
}

func GetAccessTokenFromContext(ctx context.Context) (string, bool) {
	val := ctx.Value(AccessTokenCtxKey)
	if val == nil {
		return "", false
	}
	token, ok := val.(string)
	return token, ok
}

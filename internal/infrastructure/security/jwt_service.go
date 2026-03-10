package security

import (
	"fmt"

	"github.com/farmanexo/order-service/internal/domain/services"
	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
)

type JWTServiceImpl struct {
	secret string
	logger *zap.Logger
}

func NewJWTService(secret string, logger *zap.Logger) *JWTServiceImpl {
	return &JWTServiceImpl{secret: secret, logger: logger}
}

type AccessTokenClaims struct {
	UserID string `json:"sub"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

func (s *JWTServiceImpl) ValidateAccessToken(tokenString string) (string, string, string, error) {
	claims := &AccessTokenClaims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.secret), nil
	})

	if err != nil || !token.Valid {
		return "", "", "", fmt.Errorf("token inválido: %w", err)
	}

	return claims.UserID, claims.Role, claims.ID, nil
}

var _ services.JWTService = (*JWTServiceImpl)(nil)

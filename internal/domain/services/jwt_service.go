package services

type JWTService interface {
	ValidateAccessToken(tokenString string) (userID string, role string, jti string, err error)
}

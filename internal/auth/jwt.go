package auth

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWTManager struct {
	secretKey     string
	tokenDuration time.Duration
}

func NewJWTManager(secret string, duration time.Duration) *JWTManager {
	return &JWTManager{
		secretKey:     secret,
		tokenDuration: duration,
	}
}

type AdminClaims struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}

func (m *JWTManager) GenerateAdminToken(username string) (string, error) {
	claims := &AdminClaims{
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(m.tokenDuration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(m.secretKey))
}

func (m *JWTManager) VerifyAdminToken(tokenStr string) (*AdminClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &AdminClaims{}, func(t *jwt.Token) (interface{}, error) {
		return []byte(m.secretKey), nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*AdminClaims)
	if !ok || !token.Valid {
		return nil, jwt.ErrTokenInvalidClaims
	}
	return claims, nil
}

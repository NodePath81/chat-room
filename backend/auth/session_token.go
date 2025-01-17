package auth

import (
	"chat-room/config"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type SessionClaims struct {
	SessionID uint `json:"session_id"`
	jwt.RegisteredClaims
}

func GenerateSessionShareToken(sessionID uint, duration time.Duration) (string, error) {
	claims := SessionClaims{
		SessionID: sessionID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(duration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(config.GetConfig().JWTSecret))
}

func ValidateSessionShareToken(tokenString string) (*SessionClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &SessionClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(config.GetConfig().JWTSecret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*SessionClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token claims")
}

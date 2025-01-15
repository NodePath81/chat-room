package auth

import (
	"net/http"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID uint `json:"user_id"`
	jwt.RegisteredClaims
}

type contextKey string

const UserIDKey contextKey = "userID"

func GetUserIDFromContext(r *http.Request) uint {
	userID, _ := r.Context().Value(UserIDKey).(uint)
	return userID
}

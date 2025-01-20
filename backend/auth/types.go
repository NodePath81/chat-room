package auth

import (
	"net/http"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type Claims struct {
	UserID uuid.UUID `json:"userId"`
	jwt.RegisteredClaims
}

type contextKey string

const UserIDKey contextKey = "userID"

func GetUserIDFromContext(r *http.Request) uuid.UUID {
	userID, _ := r.Context().Value(UserIDKey).(uuid.UUID)
	return userID
}

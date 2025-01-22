package middleware

import (
	"context"
	"net/http"
	"strings"

	"chat-room/token"

	"github.com/google/uuid"
)

type sessionAuthKey struct{}
type sessionIDKey struct{}

// SessionClaims represents the session claims stored in the request context
type SessionClaims struct {
	GroupID uuid.UUID
	Role    string
}

// NewSessionAuth creates a new session authentication middleware
func NewSessionAuth(tokenManager *token.TokenManager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get all cookies
			cookies := r.Cookies()
			var sessionToken string
			var sessionID uuid.UUID

			// Look for session token cookie
			for _, cookie := range cookies {
				if strings.HasPrefix(cookie.Name, "session_token_") {
					sessionToken = cookie.Value
					// Extract session ID from cookie name
					idStr := strings.TrimPrefix(cookie.Name, "session_token_")
					id, err := uuid.Parse(idStr)
					if err == nil {
						sessionID = id
						break
					}
				}
			}

			if sessionToken == "" {
				http.Error(w, "Session token is required", http.StatusUnauthorized)
				return
			}

			// Verify token
			claims, err := tokenManager.VerifyToken(sessionToken)
			if err != nil {
				switch err {
				case token.ErrExpiredToken:
					http.Error(w, "Session token has expired", http.StatusUnauthorized)
				case token.ErrInvalidToken:
					http.Error(w, "Invalid session token", http.StatusUnauthorized)
				default:
					http.Error(w, "Session authentication failed", http.StatusUnauthorized)
				}
				return
			}

			// Verify that the session ID from the cookie matches the one in the token
			if sessionID != claims.GroupID {
				http.Error(w, "Invalid session token", http.StatusUnauthorized)
				return
			}

			// Store claims and session ID in context
			sessionClaims := &SessionClaims{
				GroupID: claims.GroupID,
				Role:    claims.Role,
			}
			ctx := context.WithValue(r.Context(), sessionAuthKey{}, sessionClaims)
			ctx = context.WithValue(ctx, sessionIDKey{}, sessionID)

			// Call next handler with updated context
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetSessionClaims retrieves session claims from the request context
func GetSessionClaims(r *http.Request) *SessionClaims {
	claims, ok := r.Context().Value(sessionAuthKey{}).(*SessionClaims)
	if !ok {
		return nil
	}
	return claims
}

// GetSessionID retrieves the session ID from the request context
func GetSessionID(r *http.Request) uuid.UUID {
	id, ok := r.Context().Value(sessionIDKey{}).(uuid.UUID)
	if !ok {
		return uuid.Nil
	}
	return id
}

// RequireRole creates middleware that checks if the user has the required role
func RequireRole(roles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims := GetSessionClaims(r)
			if claims == nil {
				http.Error(w, "Session authentication required", http.StatusUnauthorized)
				return
			}

			// Check if user's role is in the allowed roles
			hasRole := false
			for _, role := range roles {
				if claims.Role == role {
					hasRole = true
					break
				}
			}

			if !hasRole {
				http.Error(w, "Insufficient permissions", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

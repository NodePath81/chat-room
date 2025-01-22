package token

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"time"

	"github.com/google/uuid"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("token has expired")
)

// Claims represents the payload of a session token
type Claims struct {
	Version   int       `json:"ver"`
	GroupID   uuid.UUID `json:"gid"`
	Role      string    `json:"role"`
	ExpiresAt int64     `json:"exp"`
}

// TokenManager handles token generation and verification
type TokenManager struct {
	gcm cipher.AEAD
}

// NewManager creates a new TokenManager with the given server key
func NewManager(serverKey []byte) (*TokenManager, error) {
	// Create AES cipher
	block, err := aes.NewCipher(serverKey)
	if err != nil {
		return nil, err
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	return &TokenManager{gcm: gcm}, nil
}

// GenerateToken creates a new session token
func (tm *TokenManager) GenerateToken(groupID uuid.UUID, role string, duration time.Duration) (string, error) {
	claims := Claims{
		Version:   1,
		GroupID:   groupID,
		Role:      role,
		ExpiresAt: time.Now().Add(duration).Unix(),
	}

	// Marshal claims to JSON
	payload, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}

	// Create nonce
	nonce := make([]byte, tm.gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	// Encrypt payload
	ciphertext := tm.gcm.Seal(nonce, nonce, payload, nil)

	// Encode to base64
	token := base64.RawURLEncoding.EncodeToString(ciphertext)
	return token, nil
}

// VerifyToken verifies and decodes a session token
func (tm *TokenManager) VerifyToken(token string) (*Claims, error) {
	// Decode from base64
	ciphertext, err := base64.RawURLEncoding.DecodeString(token)
	if err != nil {
		return nil, ErrInvalidToken
	}

	// Extract nonce
	if len(ciphertext) < tm.gcm.NonceSize() {
		return nil, ErrInvalidToken
	}
	nonce := ciphertext[:tm.gcm.NonceSize()]
	ciphertext = ciphertext[tm.gcm.NonceSize():]

	// Decrypt payload
	plaintext, err := tm.gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, ErrInvalidToken
	}

	// Unmarshal claims
	var claims Claims
	if err := json.Unmarshal(plaintext, &claims); err != nil {
		return nil, ErrInvalidToken
	}

	// Check expiration
	if time.Now().Unix() > claims.ExpiresAt {
		return nil, ErrExpiredToken
	}

	return &claims, nil
}

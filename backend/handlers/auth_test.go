package handlers

import (
	"bytes"
	"chat-room/tests"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRegister(t *testing.T) {
	db := tests.SetupTestDB(t)
	defer tests.CleanupTestDB(db)

	handler := NewAuthHandler(db)

	tests := []struct {
		name       string
		input      map[string]string
		wantStatus int
	}{
		{
			name: "valid registration",
			input: map[string]string{
				"username": "testuser",
				"password": "password123",
			},
			wantStatus: http.StatusCreated,
		},
		{
			name: "duplicate username",
			input: map[string]string{
				"username": "testuser",
				"password": "password123",
			},
			wantStatus: http.StatusInternalServerError,
		},
		{
			name: "missing password",
			input: map[string]string{
				"username": "testuser2",
			},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.input)
			req := httptest.NewRequest("POST", "/api/auth/register", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler.Register(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("Register() status = %v, want %v", w.Code, tt.wantStatus)
			}
		})
	}
}

package handlers

import (
	"bytes"
	"chat-room/models"
	"chat-room/tests"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
)

func TestCreateSession(t *testing.T) {
	db := tests.SetupTestDB(t)
	defer tests.CleanupTestDB(db)

	handler := NewSessionHandler(db)

	tests := []struct {
		name       string
		input      map[string]string
		wantStatus int
	}{
		{
			name: "valid session",
			input: map[string]string{
				"name": "Test Room",
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "missing name",
			input: map[string]string{
				"name": "",
			},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.input)
			req := httptest.NewRequest("POST", "/api/sessions", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler.CreateSession(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("CreateSession() status = %v, want %v", w.Code, tt.wantStatus)
			}
		})
	}
}

func TestGetSession(t *testing.T) {
	db := tests.SetupTestDB(t)
	defer tests.CleanupTestDB(db)

	handler := NewSessionHandler(db)

	// Create a test session
	session := models.Session{Name: "Test Room"}
	db.Create(&session)

	tests := []struct {
		name       string
		sessionID  string
		wantStatus int
	}{
		{
			name:       "existing session",
			sessionID:  "1",
			wantStatus: http.StatusOK,
		},
		{
			name:       "non-existing session",
			sessionID:  "999",
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/api/sessions/"+tt.sessionID, nil)
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tt.sessionID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
			w := httptest.NewRecorder()

			handler.GetSession(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("GetSession() status = %v, want %v", w.Code, tt.wantStatus)
			}
		})
	}
}

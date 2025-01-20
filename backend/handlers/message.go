package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"path"
	"time"

	"chat-room/auth"
	"chat-room/config"
	"chat-room/models"
	"chat-room/s3"
	"chat-room/store"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
)

type MessageHandler struct {
	store store.Store
	hub   *WebSocketHandler
}

func NewMessageHandler(store store.Store, hub *WebSocketHandler) *MessageHandler {
	return &MessageHandler{store: store, hub: hub}
}

// UploadMessageImage handles image upload for messages
func (h *MessageHandler) UploadMessageImage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userID := auth.GetUserIDFromContext(r)
	if userID == uuid.Nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get session ID from URL
	sessionID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "Invalid session ID", http.StatusBadRequest)
		return
	}

	// Verify user is a member of the session
	role, err := h.store.GetUserSessionRole(r.Context(), userID, sessionID)
	if err != nil || role == "" {
		http.Error(w, "Not a member of this session", http.StatusForbidden)
		return
	}

	// Parse multipart form
	err = r.ParseMultipartForm(10 << 20) // 10 MB max
	if err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("image")
	if err != nil {
		http.Error(w, "Failed to get file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Generate a unique filename
	filename := fmt.Sprintf("%s%s", uuid.New().String(), path.Ext(header.Filename))
	objectName := fmt.Sprintf("messages/%s", filename)

	cfg := config.GetConfig()
	minioClient := s3.GetClient()

	// Upload the file to MinIO
	opts := minio.PutObjectOptions{
		ContentType: header.Header.Get("Content-Type"),
	}
	_, err = minioClient.PutObject(context.Background(), cfg.MinioBucketName, objectName, file, header.Size, opts)
	if err != nil {
		http.Error(w, "Failed to upload file", http.StatusInternalServerError)
		return
	}

	// Generate the public URL
	publicURL := fmt.Sprintf("http://localhost:9000/%s/%s",
		cfg.MinioBucketName,
		objectName,
	)

	// Create and save the message in the database
	message := &models.Message{
		ID:        uuid.New(),
		Type:      models.MessageTypeImage,
		Content:   publicURL,
		UserID:    userID,
		SessionID: sessionID,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	if err := h.store.CreateMessage(r.Context(), message); err != nil {
		http.Error(w, "Failed to save message", http.StatusInternalServerError)
		return
	}

	// Broadcast the message through WebSocket
	h.hub.broadcast(sessionID, message)

	json.NewEncoder(w).Encode(map[string]string{
		"url": publicURL,
	})
}

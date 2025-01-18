package handlers

import (
	"chat-room/auth"
	"chat-room/config"
	"chat-room/models"
	"chat-room/s3"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"path"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"gorm.io/gorm"
)

type MessageHandler struct {
	db  *gorm.DB
	hub *WebSocketHandler
}

func NewMessageHandler(db *gorm.DB, hub *WebSocketHandler) *MessageHandler {
	return &MessageHandler{db: db, hub: hub}
}

// UploadMessageImage handles image upload for messages
func (h *MessageHandler) UploadMessageImage(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value("claims").(*auth.Claims)
	if !ok || claims == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	// Get session ID from URL
	sessionID := chi.URLParam(r, "id")
	sessionIDUint, err := strconv.ParseUint(sessionID, 10, 32)
	if err != nil {
		http.Error(w, "invalid session ID", http.StatusBadRequest)
		return
	}

	// Parse multipart form
	err = r.ParseMultipartForm(10 << 20) // 10 MB max
	if err != nil {
		http.Error(w, "failed to parse form", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("image")
	if err != nil {
		http.Error(w, "failed to get file", http.StatusBadRequest)
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
		http.Error(w, "failed to upload file", http.StatusInternalServerError)
		return
	}

	// Generate the public URL
	publicURL := fmt.Sprintf("http://localhost:9000/%s/%s",
		cfg.MinioBucketName,
		objectName,
	)

	// Create and save the message in the database
	message := models.Message{
		Type:      models.MessageTypeImage,
		Content:   publicURL,
		UserID:    claims.UserID,
		SessionID: uint(sessionIDUint),
		CreatedAt: time.Now(),
	}

	if err := h.db.Create(&message).Error; err != nil {
		http.Error(w, "failed to save message", http.StatusInternalServerError)
		return
	}

	// Broadcast the message through WebSocket
	h.hub.broadcast(uint(sessionIDUint), Message{
		Type:      "image",
		Content:   message.Content,
		UserID:    message.UserID,
		SessionID: message.SessionID,
		MsgType:   string(message.Type),
		CreatedAt: message.CreatedAt,
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"url": publicURL,
	})
}

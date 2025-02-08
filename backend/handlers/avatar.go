package handlers

import (
	"chat-room/auth"
	"chat-room/config"
	"chat-room/s3"
	"chat-room/store"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"path"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
)

type AvatarHandler struct {
	store store.Store
}

func NewAvatarHandler(store store.Store) *AvatarHandler {
	return &AvatarHandler{store: store}
}

// UploadAvatar handles direct file upload
func (h *AvatarHandler) UploadAvatar(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserIDFromContext(r)
	if userID == uuid.Nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse multipart form
	err := r.ParseMultipartForm(10 << 20) // 10 MB max
	if err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("avatar")
	if err != nil {
		http.Error(w, "Failed to get file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Generate a unique filename
	filename := fmt.Sprintf("%s%s", uuid.New().String(), path.Ext(header.Filename))
	objectName := fmt.Sprintf("user-%s/%s", userID.String(), filename)

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

	// Get current user
	users, err := h.store.GetUsersByIDs(r.Context(), []uuid.UUID{userID})
	if err != nil {
		http.Error(w, "Failed to get user", http.StatusInternalServerError)
		return
	}
	if len(users) == 0 {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}
	user := users[0]

	// Update user's avatar URL
	user.AvatarURL = publicURL
	if err := h.store.UpdateUser(r.Context(), user); err != nil {
		http.Error(w, "Failed to update avatar URL", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"avatarUrl": publicURL,
		"message":   "Avatar uploaded successfully",
	})
}

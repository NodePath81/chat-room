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

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"gorm.io/gorm"
)

type AvatarHandler struct {
	db *gorm.DB
}

func NewAvatarHandler(db *gorm.DB) *AvatarHandler {
	return &AvatarHandler{db: db}
}

// UploadAvatar handles direct file upload
func (h *AvatarHandler) UploadAvatar(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value("claims").(*auth.Claims)
	if !ok || claims == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse multipart form
	err := r.ParseMultipartForm(10 << 20) // 10 MB max
	if err != nil {
		http.Error(w, "failed to parse form", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("avatar")
	if err != nil {
		http.Error(w, "failed to get file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Generate a unique filename
	filename := fmt.Sprintf("%s%s", uuid.New().String(), path.Ext(header.Filename))
	objectName := fmt.Sprintf("user-%d/%s", claims.UserID, filename)

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

	// Update user's avatar URL in the database
	result := h.db.Model(&models.User{}).Where("id = ?", claims.UserID).Update("avatar_url", publicURL)
	if result.Error != nil {
		http.Error(w, "failed to update avatar URL", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"avatarUrl": publicURL,
		"message":   "avatar uploaded successfully",
	})
}

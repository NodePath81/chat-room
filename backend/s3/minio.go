package s3

import (
	"chat-room/config"
	"context"
	"fmt"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

var minioClient *minio.Client

// GetClient returns the initialized MinIO client
func GetClient() *minio.Client {
	return minioClient
}

// Initialize sets up the MinIO client and creates the bucket if it doesn't exist
func Initialize(cfg *config.Config) error {
	var err error
	minioClient, err = minio.New(cfg.MinioEndpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.MinioAccessKey, cfg.MinioSecretKey, ""),
		Secure: cfg.MinioUseSSL,
	})
	if err != nil {
		return fmt.Errorf("failed to create MinIO client: %w", err)
	}

	// Create bucket if it doesn't exist
	err = createBucketIfNotExists(context.Background(), cfg.MinioBucketName)
	if err != nil {
		return fmt.Errorf("error creating bucket: %v", err)
	}

	return nil
}

func createBucketIfNotExists(ctx context.Context, bucketName string) error {
	exists, err := minioClient.BucketExists(ctx, bucketName)
	if err != nil {
		return fmt.Errorf("error checking bucket existence: %v", err)
	}

	if !exists {
		err = minioClient.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{})
		if err != nil {
			return fmt.Errorf("error creating bucket: %v", err)
		}

		// Set bucket policy to allow public read access
		policy := `{
			"Version": "2012-10-17",
			"Statement": [
				{
					"Effect": "Allow",
					"Principal": {"AWS": ["*"]},
					"Action": ["s3:GetObject"],
					"Resource": ["arn:aws:s3:::` + bucketName + `/*"]
				}
			]
		}`

		err = minioClient.SetBucketPolicy(ctx, bucketName, policy)
		if err != nil {
			return fmt.Errorf("error setting bucket policy: %v", err)
		}
	}

	return nil
}

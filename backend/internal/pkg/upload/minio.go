package upload

import (
	"context"
	"fmt"
	"io"
	"path"
	"time"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
)

type MinIOUploader struct {
	client *minio.Client
	bucket string
}

func NewMinIOUploader(client *minio.Client, bucket string) *MinIOUploader {
	return &MinIOUploader{
		client: client,
		bucket: bucket,
	}
}

func (u *MinIOUploader) Upload(ctx context.Context, reader io.Reader, size int64, contentType string, dir string, ext string) (string, error) {
	if u.client == nil {
		return "", fmt.Errorf("MinIO client not initialized")
	}

	objectName := fmt.Sprintf("%s/%s-%s%s", dir, time.Now().Format("20060102"), uuid.New().String(), ext)

	_, err := u.client.PutObject(ctx, u.bucket, objectName, reader, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload to MinIO: %w", err)
	}

	return objectName, nil
}

func (u *MinIOUploader) Delete(ctx context.Context, objectName string) error {
	if u.client == nil {
		return fmt.Errorf("MinIO client not initialized")
	}
	return u.client.RemoveObject(ctx, u.bucket, objectName, minio.RemoveObjectOptions{})
}

func (u *MinIOUploader) GetURL(objectName string) string {
	return fmt.Sprintf("/%s/%s", u.bucket, objectName)
}

func ExtFromContentType(contentType string) string {
	switch contentType {
	case "video/mp4":
		return ".mp4"
	case "video/webm":
		return ".webm"
	case "video/quicktime":
		return ".mov"
	case "image/jpeg":
		return ".jpg"
	case "image/png":
		return ".png"
	case "image/webp":
		return ".webp"
	default:
		return path.Ext(contentType)
	}
}

package service

import (
	"bytes"
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

const (
	PICTURE_BUCKET_NAME = "images"
)

type MediaService struct {
	client *minio.Client
}

// New Funktion mit Bucket-Erstellung
func New(endpoint, accessKeyID, secretAccessKey string) (*MediaService, error) {
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: false,
	})
	if err != nil {
		return nil, err
	}

	// Bucket-Existenz prüfen
	ctx := context.Background()
	exists, err := client.BucketExists(ctx, PICTURE_BUCKET_NAME)
	if err != nil {
		return nil, fmt.Errorf("bucket check failed: %v", err)
	}
	if !exists {
		err = client.MakeBucket(ctx, PICTURE_BUCKET_NAME, minio.MakeBucketOptions{})
		if err != nil {
			return nil, fmt.Errorf("bucket creation failed: %v", err)
		}
	}

	return &MediaService{client: client}, nil
}

// UploadPicture mit Content-Type-Handling
func (m *MediaService) UploadPicture(ctx context.Context, uploader string, contentType string, picture []byte) (string, error) {
	name := uuid.New().String()

	// file size limit to 6MB
	if len(picture) > 6e6 {
		return "", fmt.Errorf("file size to big. max 6mb allowed")
	}

	// Vereinfachte Content-Type-Erkennung anhand der Dateiendung
	switch contentType {
	case "image/jpeg":
		break
	case "image/png":
		break
	case "application/octet-stream":
		break
	default:
		return "", fmt.Errorf("unsupported file type")
	}

	_, err := m.client.PutObject(
		ctx,
		PICTURE_BUCKET_NAME,
		name,
		bytes.NewReader(picture),
		int64(len(picture)),
		minio.PutObjectOptions{
			ContentType: contentType,
			UserMetadata: map[string]string{
				"id": uploader,
			},
		},
	)
	return name, err
}

// GetPicture mit Context-Parameter
func (m *MediaService) GetPicture(ctx context.Context, pictureName string) ([]byte, error) {
	object, err := m.client.GetObject(ctx, PICTURE_BUCKET_NAME, pictureName, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("image not found")
	}
	defer object.Close()

	var buf bytes.Buffer
	if _, err := buf.ReadFrom(object); err != nil {
		return nil, fmt.Errorf("failed to read object: %v", err)
	}
	return buf.Bytes(), nil
}

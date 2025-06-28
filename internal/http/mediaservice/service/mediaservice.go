package service

import (
	"bytes"
	"context"
	"embed"
	"fmt"
	"log"

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

//go:embed impala.jpg
var defaultImage embed.FS

// GetPicture mit Context-Parameter
func (m *MediaService) GetPicture(ctx context.Context, pictureName string) ([]byte, error) {
	object, err := m.client.GetObject(ctx, PICTURE_BUCKET_NAME, pictureName, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("image not found")
	}
	defer func() {
		if err := object.Close(); err != nil {
			log.Println(err)
		}
	}()

	var buf bytes.Buffer
	if _, err := buf.ReadFrom(object); err != nil {
		if err.Error() == "The specified key does not exist." {
			data, err := defaultImage.ReadFile("impala.jpg")
			if err != nil {
				return nil, fmt.Errorf("failed to read default image: %v", err)
			}
			return data, nil
		}
		return nil, fmt.Errorf("failed to read object: %v", err)
	}
	return buf.Bytes(), nil
}

func (m *MediaService) UploadPictureToMulti(ctx context.Context, uploader, compoundID string, contentType string, picture []byte) error {
	name := uuid.New().String()

	// file size limit to 6MB
	if len(picture) > 6e6 {
		return fmt.Errorf("file size to big. max 6mb allowed")
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
		return fmt.Errorf("unsupported file type")
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
				"id":          uploader,
				"Compound-Id": compoundID,
			},
		},
	)
	return err
}

func (m *MediaService) GetMultiPicture(ctx context.Context, id uuid.UUID) ([]string, error) {
	objectCh := m.client.ListObjects(ctx, PICTURE_BUCKET_NAME, minio.ListObjectsOptions{
		Recursive: true,
	})
	var pictureNames []string
	for obj := range objectCh {
		if obj.Err != nil {
			continue
		}

		// Get object metadata
		info, err := m.client.StatObject(ctx, PICTURE_BUCKET_NAME, obj.Key, minio.StatObjectOptions{})
		if err != nil {
			continue
		}

		meta := info.UserMetadata
		if compoundID, ok := meta["Compound-Id"]; ok && compoundID == id.String() {
			pictureNames = append(pictureNames, obj.Key)
		}
	}
	return pictureNames, nil
}

package miniostorage

import (
	"bytes"
	"context"
	"fmt"

	"github.com/minio/minio-go/v7"
	// "github.com/minio/minio-go/v7/pkg/credentials"
)

type MinIoStorage struct {
	minioClient *minio.Client
	bucketName  string
}

func New(minioClient *minio.Client, bucketName string) *MinIoStorage {
	return &MinIoStorage{
		minioClient: minioClient,
		bucketName:  bucketName,
	}
}

func (m *MinIoStorage) UploadPhoto(ctx context.Context, id string, photo []byte, contentType string, fileName string) (string, error) {
	objectName := fmt.Sprintf("%s_%s", id, fileName)

	// Загружаем фото в MinIO
	// Используем bytes.NewReader для создания io.Reader из byte slice
	_, err := m.minioClient.PutObject(ctx, m.bucketName, objectName, bytes.NewReader(photo), int64(len(photo)),
		minio.PutObjectOptions{ContentType: contentType},
	)

	if err != nil {
		return "", fmt.Errorf("failed to upload photo: %w", err)
	}

	url := fmt.Sprintf("%s/%s/%s", m.minioClient.EndpointURL(), m.bucketName, objectName)

	return url, nil
}

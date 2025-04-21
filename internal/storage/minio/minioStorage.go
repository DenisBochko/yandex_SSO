package miniostorage

import (
	"bytes"
	"context"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"

	"github.com/disintegration/imaging"
	"github.com/minio/minio-go/v7"
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

	// Декодируем изображение из []byte
	img, _, err := image.Decode(bytes.NewReader(photo))
	if err != nil {
		return "", fmt.Errorf("failed to decode image: %w", err)
	}

	// Сжимаем изображение до 512x512
	resizedImg := imaging.Fill(img, 512, 512, imaging.Center, imaging.Lanczos)

	// Кодируем обратно в JPEG (можно PNG — зависит от contentType)
	var buf bytes.Buffer
	switch contentType {
	case "image/png":
		err = imaging.Encode(&buf, resizedImg, imaging.PNG)
	default:
		// по умолчанию JPEG
		contentType = "image/jpeg"
		err = imaging.Encode(&buf, resizedImg, imaging.JPEG)
	}

	if err != nil {
		return "", fmt.Errorf("failed to encode resized image: %w", err)
	}

	// Загружаем сжатое изображение в MinIO
	_, err = m.minioClient.PutObject(ctx, m.bucketName, objectName, &buf, int64(buf.Len()),
		minio.PutObjectOptions{ContentType: contentType},
	)

	if err != nil {
		return "", fmt.Errorf("failed to upload photo: %w", err)
	}

	url := fmt.Sprintf("%s/%s/%s", m.minioClient.EndpointURL(), m.bucketName, objectName)
	
	return url, nil
}

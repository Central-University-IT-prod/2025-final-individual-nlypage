package minio

import (
	"bytes"
	"context"
	"fmt"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"io"
	"net/http"
	"strings"
)

type AdImagesRepository interface {
	UploadImage(ctx context.Context, campaignID string, imageData io.Reader) (string, error)
	GetImage(ctx context.Context, campaignID string) (string, error)
	DeleteImage(ctx context.Context, campaignID string) error
}

type repository struct {
	client       *minio.Client
	bucketName   string
	endpoint     string
	httpEndpoint string
}

type Config struct {
	Endpoint     string
	HTTPEndpoint string
	AccessKey    string
	SecretKey    string
	BucketName   string
	UseSSL       bool
}

// NewAdImagesRepository creates a new MinIO repository instance
func NewAdImagesRepository(cfg Config) (AdImagesRepository, error) {
	// Создаем обычный клиент
	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create minio client: %w", err)
	}

	repo := &repository{
		client:       client,
		bucketName:   cfg.BucketName,
		endpoint:     cfg.Endpoint,
		httpEndpoint: cfg.HTTPEndpoint,
	}

	// Проверяем/создаем бакет
	exists, err := client.BucketExists(context.Background(), cfg.BucketName)
	if err != nil {
		return nil, fmt.Errorf("failed to check if bucket exists: %w", err)
	}

	if !exists {
		err = client.MakeBucket(context.Background(), cfg.BucketName, minio.MakeBucketOptions{})
		if err != nil {
			return nil, fmt.Errorf("failed to create bucket: %w", err)
		}
	}

	err = client.SetBucketPolicy(
		context.Background(),
		cfg.BucketName,
		fmt.Sprintf(publicReadPolicy, cfg.BucketName),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to set bucket policy: %w", err)
	}

	return repo, nil
}

// getPublicURL возвращает публичный URL для объекта
func (r *repository) getPublicURL(objectName string) string {
	protocol := "http"
	if strings.HasSuffix(r.httpEndpoint, ":443") {
		protocol = "https"
	}
	return fmt.Sprintf("%s://%s/%s/%s", protocol, r.httpEndpoint, r.bucketName, objectName)
}

// isImageContentType проверяет, является ли тип контента изображением
func isImageContentType(contentType string) bool {
	return strings.HasPrefix(contentType, "image/")
}

// UploadImage uploads an image for an advertising campaign
func (r *repository) UploadImage(ctx context.Context, campaignID string, imageData io.Reader) (string, error) {
	objectName := fmt.Sprintf("campaigns/%s/image", campaignID)

	// Сохраняем в буфер
	var buf bytes.Buffer
	size, err := io.Copy(&buf, imageData)
	if err != nil {
		return "", fmt.Errorf("failed to read image data: %w", err)
	}

	// Определяем тип контента по содержимому
	contentType := http.DetectContentType(buf.Bytes())

	// Проверяем, что файл является изображением
	if !isImageContentType(contentType) {
		return "", ErrFileNotImage
	}

	_, err = r.client.PutObject(ctx, r.bucketName, objectName, &buf, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload image: %w", err)
	}

	return r.getPublicURL(objectName), nil
}

// GetImage returns a public URL for an advertising campaign image
func (r *repository) GetImage(ctx context.Context, campaignID string) (string, error) {
	objectName := fmt.Sprintf("campaigns/%s/image", campaignID)

	// Проверяем существование объекта
	_, err := r.client.StatObject(ctx, r.bucketName, objectName, minio.StatObjectOptions{})
	if err != nil {
		return "", fmt.Errorf("image not found: %w", err)
	}

	return r.getPublicURL(objectName), nil
}

// DeleteImage deletes an image for an advertising campaign
func (r *repository) DeleteImage(ctx context.Context, campaignID string) error {
	objectName := fmt.Sprintf("campaigns/%s/image", campaignID)

	err := r.client.RemoveObject(ctx, r.bucketName, objectName, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete image: %w", err)
	}

	return nil
}

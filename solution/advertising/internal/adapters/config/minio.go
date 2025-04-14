package config

import "github.com/spf13/viper"

type MinioConfig interface {
	Endpoint() string
	HTTPEndpoint() string
	AccessKey() string
	SecretKey() string
	BucketName() string
	UseSSL() bool
}

type minioConfig struct {
	endpoint     string
	httpEndpoint string
	accessKey    string
	secretKey    string
	bucketName   string
	useSSL       bool
}

func NewMinioConfig(v *viper.Viper) MinioConfig {
	return &minioConfig{
		endpoint:     v.GetString("service.minio.endpoint"),
		httpEndpoint: v.GetString("service.minio.http-endpoint"),
		accessKey:    v.GetString("service.minio.access-key"),
		secretKey:    v.GetString("service.minio.secret-key"),
		bucketName:   v.GetString("service.minio.bucket"),
		useSSL:       v.GetBool("service.minio.use-ssl"),
	}
}

func (c *minioConfig) Endpoint() string {
	return c.endpoint
}

func (c *minioConfig) HTTPEndpoint() string {
	return c.httpEndpoint
}

func (c *minioConfig) AccessKey() string {
	return c.accessKey
}

func (c *minioConfig) SecretKey() string {
	return c.secretKey
}

func (c *minioConfig) BucketName() string {
	return c.bucketName
}

func (c *minioConfig) UseSSL() bool {
	return c.useSSL
}

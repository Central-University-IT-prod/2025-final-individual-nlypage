package minio

import "errors"

var (
	ErrImageNotFound = errors.New("image not found")
	ErrFileNotImage  = errors.New("uploaded file is not an image")
)

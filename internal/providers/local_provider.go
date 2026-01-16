package providers

import (
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
)

type LocalUploadProvider struct {
	basePath string
}

func NewLocalUploadProvider(basePath string) *LocalUploadProvider {
	return &LocalUploadProvider{
		basePath: basePath,
	}
}

func (l *LocalUploadProvider) UploadFile(file *multipart.FileHeader, path string) (string, error) {
	fullPath := filepath.Join(l.basePath, path)

	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		return "", err
	}

	// Open source
	source, err := file.Open()
	if err != nil {
		return "", err
	}
	defer source.Close()

	// create destination
	destination, err := os.Create(fullPath)
	if err != nil {
		return "", err
	}
	defer destination.Close()

	// Copy file
	if _, err := io.Copy(destination, source); err != nil {
		return "", err
	}

	return fmt.Sprintf("/uploads/%s", path), nil
}

func (l *LocalUploadProvider) DeleteFile(path string) error {
	fullPath := filepath.Join(l.basePath, path)
	return os.Remove(fullPath)
}

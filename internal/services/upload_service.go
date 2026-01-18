package services

import (
	"fmt"
	"mime/multipart"
	"path/filepath"
	"slices"
	"strings"

	"github.com/google/uuid"
	"github.com/trenchesdeveloper/go-ai-store/internal/interfaces"
)

type UploadService struct {
	provider interfaces.Upload
}

func NewUploadService(provider interfaces.Upload) *UploadService {
	return &UploadService{
		provider: provider,
	}
}

func (s *UploadService) UploadProductImage(productID uint, file *multipart.FileHeader) (string, error) {
	ext := strings.ToLower(filepath.Ext(file.Filename))

	if !isValidImageFormat(ext) {
		return "", fmt.Errorf("invalid image format: %s", ext)
	}

	newFileName := uuid.New().String()

	path := fmt.Sprintf("products/%d/%s%s", productID, newFileName, ext)

	return s.provider.UploadFile(file, path)
}

func isValidImageFormat(ext string) bool {
	validFormats := []string{".png", ".jpg", ".jpeg", ".gif", ".webp"}
	return slices.Contains(validFormats, ext)
}

package services

import (
	"mime/multipart"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockUploadProvider mocks the Upload interface
type MockUploadProvider struct {
	mock.Mock
}

func (m *MockUploadProvider) UploadFile(file *multipart.FileHeader, path string) (string, error) {
	args := m.Called(file, path)
	return args.String(0), args.Error(1)
}

func (m *MockUploadProvider) DeleteFile(path string) error {
	args := m.Called(path)
	return args.Error(0)
}

func TestUploadService_UploadProductImage(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		productID uint
		filename  string
		setupMock func(m *MockUploadProvider)
		wantErr   bool
		errMsg    string
	}{
		{
			name:      "success - jpg image",
			productID: 1,
			filename:  "test.jpg",
			setupMock: func(m *MockUploadProvider) {
				m.On("UploadFile", mock.Anything, mock.Anything).Return("https://cdn.example.com/products/1/test.jpg", nil)
			},
			wantErr: false,
		},
		{
			name:      "success - png image",
			productID: 1,
			filename:  "image.png",
			setupMock: func(m *MockUploadProvider) {
				m.On("UploadFile", mock.Anything, mock.Anything).Return("https://cdn.example.com/products/1/image.png", nil)
			},
			wantErr: false,
		},
		{
			name:      "success - webp image",
			productID: 1,
			filename:  "photo.webp",
			setupMock: func(m *MockUploadProvider) {
				m.On("UploadFile", mock.Anything, mock.Anything).Return("https://cdn.example.com/products/1/photo.webp", nil)
			},
			wantErr: false,
		},
		{
			name:      "error - invalid format txt",
			productID: 1,
			filename:  "document.txt",
			setupMock: func(m *MockUploadProvider) {},
			wantErr:   true,
			errMsg:    "invalid image format",
		},
		{
			name:      "error - invalid format pdf",
			productID: 1,
			filename:  "document.pdf",
			setupMock: func(m *MockUploadProvider) {},
			wantErr:   true,
			errMsg:    "invalid image format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockProvider := new(MockUploadProvider)
			tt.setupMock(mockProvider)

			service := NewUploadService(mockProvider)

			file := &multipart.FileHeader{
				Filename: tt.filename,
			}

			url, err := service.UploadProductImage(tt.productID, file)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				return
			}

			assert.NoError(t, err)
			assert.NotEmpty(t, url)
		})
	}
}

func TestIsValidImageFormat(t *testing.T) {
	t.Parallel()

	tests := []struct {
		ext   string
		valid bool
	}{
		{".jpg", true},
		{".jpeg", true},
		{".png", true},
		{".gif", true},
		{".webp", true},
		{".txt", false},
		{".pdf", false},
		{".bmp", false},
		{".doc", false},
	}

	for _, tt := range tests {
		t.Run(tt.ext, func(t *testing.T) {
			t.Parallel()
			result := isValidImageFormat(tt.ext)
			assert.Equal(t, tt.valid, result)
		})
	}
}

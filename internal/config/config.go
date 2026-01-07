package config

import (
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	JWT      JWTConfig
	AWS      AWSConfig
	Upload   UploadConfig
}

type ServerConfig struct {
	Port    string
	GinMode string
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
}

type JWTConfig struct {
	Secret                string
	ExpiresIn             time.Duration
	RefreshTokenExpiresIn time.Duration
}

type AWSConfig struct {
	S3Endpoint      string
	Region          string
	AccessKeyID     string
	SecretAccessKey string
	S3Bucket        string
}

type UploadConfig struct {
	UploadPath    string
	MaxUploadSize int64
}

func LoadConfig() (*Config, error) {
	_ = godotenv.Load()

	jwtExpiresIn, _ := time.ParseDuration(getEnv("JWT_EXPIRES_IN", "24h"))
	refreshTokenExpiresIn, _ := time.ParseDuration(getEnv("REFRESH_TOKEN_EXPIRES_IN", "72h"))
	maxUploadSize, _ := strconv.ParseInt(getEnv("MAX_UPLOAD_SIZE", "10485760"), 10, 64)

	return &Config{
		Server: ServerConfig{
			Port:    getEnv("PORT", "8000"),
			GinMode: getEnv("GIN_MODE", "debug"),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", "postgres"),
			Name:     getEnv("DB_NAME", "ecommerce"),
			SSLMode:  getEnv("DB_SSL_MODE", "disable"),
		},
		JWT: JWTConfig{
			Secret:                getEnv("JWT_SECRET", "secret"),
			ExpiresIn:             jwtExpiresIn,
			RefreshTokenExpiresIn: refreshTokenExpiresIn,
		},
		AWS: AWSConfig{
			S3Endpoint:      getEnv("AWS_S3_ENDPOINT", "http://localhost:4566"),
			Region:          getEnv("AWS_S3_REGION", "us-east-1"),
			AccessKeyID:     getEnv("AWS_S3_ACCESS_KEY", "localstack"),
			SecretAccessKey: getEnv("AWS_S3_SECRET_KEY", "localstack"),
			S3Bucket:        getEnv("AWS_S3_BUCKET", "ecommerce-uploads"),
		},
		Upload: UploadConfig{
			UploadPath:    getEnv("UPLOAD_PATH", "uploads"),
			MaxUploadSize: maxUploadSize,
		},
	}, nil
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

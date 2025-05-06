package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

var (
	Port             string
	UploadDir        string
	MaxUploadSize    int64
	CacheMaxAge      int
	ServerURL        string
	AllowedMimeTypes string
)

func LoadConfig() error {
	err := godotenv.Load()
	if err != nil {
		return fmt.Errorf("error loading .env file: %w", err)
	}

	Port = os.Getenv("PORT")
	UploadDir = os.Getenv("UPLOAD_DIR")

	if sizeStr := os.Getenv("MAX_UPLOAD_SIZE"); sizeStr != "" {
		MaxUploadSize, _ = strconv.ParseInt(sizeStr, 10, 64)
	}

	if age := os.Getenv("CACHE_MAX_AGE"); age != "" {
		CacheMaxAge, _ = strconv.Atoi(age)
	}

	ServerURL = os.Getenv("SERVER_URL")
	AllowedMimeTypes = os.Getenv("ALLOWED_MIME_TYPES")

	return nil
}

package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

var (
	Port               string
	UploadDir          string
	MaxUploadSize      int64
	CacheMaxAge        int
	ServerURL          string
	AllowedMimeTypes   string
	TurnstileSecretKey string
)

func LoadConfig() error {
	err := godotenv.Load()
	if err != nil {
		// It's okay if .env doesn't exist, we'll use defaults
		fmt.Printf("Notice: .env file not found, using environment variables\n")
	}

	// Set defaults
	if Port == "" {
		Port = "4200"
	}
	if UploadDir == "" {
		UploadDir = "uploads"
	}
	if MaxUploadSize == 0 {
		MaxUploadSize = 500 * 1024 * 1024 // 500MB
	}
	if CacheMaxAge == 0 {
		CacheMaxAge = 86400 // 1 day in seconds
	}
	if ServerURL == "" {
		ServerURL = "http://localhost:4200"
	}
	if AllowedMimeTypes == "" {
		AllowedMimeTypes = "image/"
	}

	// Override with environment variables if they exist
	if port := os.Getenv("PORT"); port != "" {
		Port = port
		if ServerURL == "http://localhost:4200" {
			ServerURL = fmt.Sprintf("http://localhost:%s", port)
		}
	}

	if dir := os.Getenv("UPLOAD_DIR"); dir != "" {
		UploadDir = dir
	}

	if sizeStr := os.Getenv("MAX_UPLOAD_SIZE"); sizeStr != "" {
		MaxUploadSize, _ = strconv.ParseInt(sizeStr, 10, 64)
	}

	if age := os.Getenv("CACHE_MAX_AGE"); age != "" {
		CacheMaxAge, _ = strconv.Atoi(age)
	}

	if url := os.Getenv("SERVER_URL"); url != "" {
		ServerURL = url
	}

	if mime := os.Getenv("ALLOWED_MIME_TYPES"); mime != "" {
		AllowedMimeTypes = mime
	}

	// Optional Turnstile configuration
	TurnstileSecretKey = os.Getenv("TURNSTILE_SECRET_KEY")

	return nil
}

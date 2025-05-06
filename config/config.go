package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	Port             int
	UploadDir        string
	MaxUploadSize    int64
	CacheMaxAge      int
	ServerURL        string
	AllowedMimeTypes string
	TurnstileSecretKey string
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	return &Config{
		Port:             4200,
		UploadDir:        "uploads",
		MaxUploadSize:    500 * 1024 * 1024, // 500MB
		CacheMaxAge:      86400,            // 1 day in seconds
		ServerURL:        "http://localhost:4200",
		AllowedMimeTypes: "image/",
		TurnstileSecretKey: "0x4AAAAAABahtWbEI-SkFM3JEPYprcYay4s",
	}
}

func LoadFromEnv() *Config {
	config := DefaultConfig()

	if port := os.Getenv("PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil && p > 0 {
			config.Port = p
			config.ServerURL = fmt.Sprintf("http://localhost:%d", p)
		}
	} else if port := os.Getenv("SERVER_PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil && p > 0 {
			config.Port = p
			config.ServerURL = fmt.Sprintf("http://localhost:%d", p)
		}
	}

	if dir := os.Getenv("UPLOAD_DIR"); dir != "" {
		config.UploadDir = dir
	}

	if size := os.Getenv("MAX_UPLOAD_SIZE"); size != "" {
		if s, err := strconv.ParseInt(size, 10, 64); err == nil && s > 0 {
			config.MaxUploadSize = s
		}
	}

	if age := os.Getenv("CACHE_MAX_AGE"); age != "" {
		if a, err := strconv.Atoi(age); err == nil && a > 0 {
			config.CacheMaxAge = a
		}
	}

	if url := os.Getenv("SERVER_URL"); url != "" {
		config.ServerURL = url
	}

	if mime := os.Getenv("ALLOWED_MIME_TYPES"); mime != "" {
		config.AllowedMimeTypes = mime
	}

	if turnstile := os.Getenv("TURNSTILE_SECRET_KEY"); turnstile != "" {
		config.TurnstileSecretKey = turnstile
	}

	return config
}

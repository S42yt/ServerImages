package models

import "time"

type ImageResponse struct {
	URL        string    `json:"url"`
	ID         string    `json:"id"`
	Size       int       `json:"size"`
	UploadedAt time.Time `json:"uploaded_at"`
}

type Base64Upload struct {
	Base64 string `json:"base64"`
}

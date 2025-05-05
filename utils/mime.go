package utils

import (
	"io"
	"strings"

	"github.com/gabriel-vasile/mimetype"
)

func GetMimeType(data []byte) (string, error) {
	mime, err := mimetype.DetectReader(strings.NewReader(string(data)))
	if err != nil {
		return "", err
	}
	return mime.String(), nil
}

func GetMimeTypeFromReader(r io.Reader) (string, string, error) {
	mime, err := mimetype.DetectReader(r)
	if err != nil {
		return "", "", err
	}
	return mime.String(), mime.Extension(), nil
}

func GetMimeTypeFromFile(path string) (string, error) {
	mime, err := mimetype.DetectFile(path)
	if err != nil {
		return "", err
	}
	return mime.String(), nil
}

func MimeToExtension(mimeType string) string {
	switch mimeType {
	case "image/jpeg":
		return ".jpg"
	case "image/png":
		return ".png"
	case "image/gif":
		return ".gif"
	case "image/webp":
		return ".webp"
	case "image/svg+xml":
		return ".svg"
	default:
		return ""
	}
}

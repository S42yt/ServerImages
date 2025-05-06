package handlers

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gofiber/fiber/v2"

	"github.com/S42yt/serverimages/config"
	"github.com/S42yt/serverimages/utils"
)

func ServeImage() fiber.Handler {
	return func(c *fiber.Ctx) error {
		filename := c.Params("filename")
		if strings.Contains(filename, "..") {
			return c.Status(fiber.StatusForbidden).SendString("Invalid file path")
		}

		filePath := filepath.Join(config.UploadDir, filename)
		info, err := os.Stat(filePath)
		if os.IsNotExist(err) {
			return c.Status(fiber.StatusNotFound).SendString("Image not found")
		}

		etag := fmt.Sprintf(`"%x-%x"`, info.ModTime().Unix(), info.Size())
		lastModified := info.ModTime().UTC().Format(http.TimeFormat)

		c.Set("Cache-Control", fmt.Sprintf("public, max-age=%d", config.CacheMaxAge))
		c.Set("Last-Modified", lastModified)
		c.Set("ETag", etag)

		if c.Get("If-Modified-Since") == lastModified || c.Get("If-None-Match") == etag {
			return c.SendStatus(fiber.StatusNotModified)
		}

		mimeType, err := utils.GetMimeTypeFromFile(filePath)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("Failed to detect file type")
		}

		c.Set("Content-Type", mimeType)
		return c.SendFile(filePath)
	}
}

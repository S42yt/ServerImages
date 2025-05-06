package handlers

import (
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/S42yt/serverimages/config"
	"github.com/S42yt/serverimages/models"
	"github.com/S42yt/serverimages/utils"
)

func Upload() fiber.Handler {
	return func(c *fiber.Ctx) error {
		var (
			fileData   []byte
			fileExt    string
			uploadedAt = time.Now()
			err        error
		)

		contentType := c.Get("Content-Type")
		switch contentType {
		case "multipart/form-data":
			file, err := c.FormFile("file")
			if err != nil {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error": "No file provided",
				})
			}

			if file.Size > config.MaxUploadSize {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error": fmt.Sprintf("File too large (max %dMB)", config.MaxUploadSize/(1024*1024)),
				})
			}

			src, err := file.Open()
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"error": "Failed to open uploaded file",
				})
			}
			defer src.Close()

			fileData, err = io.ReadAll(src)
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"error": "Failed to read file data",
				})
			}

			fileExt = filepath.Ext(file.Filename)

		case "application/json":
			var body models.Base64Upload
			if err := c.BodyParser(&body); err != nil {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error": "Invalid request body",
				})
			}

			if body.Base64 == "" {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error": "No base64 data provided",
				})
			}

			base64Data := body.Base64
			if idx := strings.Index(base64Data, ";base64,"); idx > 0 {
				mimeType := base64Data[5:idx]
				fileExt = utils.MimeToExtension(mimeType)
				base64Data = base64Data[idx+8:]
			}

			fileData, err = base64.StdEncoding.DecodeString(base64Data)
			if err != nil {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error": "Invalid base64 data",
				})
			}

			if int64(len(fileData)) > config.MaxUploadSize {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error": fmt.Sprintf("File too large (max %dMB)", config.MaxUploadSize/(1024*1024)),
				})
			}

		default:
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Unsupported content type",
			})
		}

		mimeType, err := utils.GetMimeType(fileData)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to detect file type",
			})
		}

		if !strings.HasPrefix(mimeType, config.AllowedMimeTypes) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Only image files are allowed",
			})
		}

		if fileExt == "" {
			fileExt = utils.MimeToExtension(mimeType)
			if fileExt == "" {
				fileExt = ".bin"
			}
		}

		filename := fmt.Sprintf("%s%s", uuid.New().String(), fileExt)
		filePath := filepath.Join(config.UploadDir, filename)

		err = os.WriteFile(filePath, fileData, 0644)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to save file",
			})
		}

		response := models.ImageResponse{
			URL:        fmt.Sprintf("%s/cdn/%s", config.ServerURL, filename),
			ID:         filename,
			Size:       len(fileData),
			UploadedAt: uploadedAt,
		}

		return c.JSON(response)
	}
}

func DeleteImage() fiber.Handler {
	return func(c *fiber.Ctx) error {
		filename := c.Params("filename")
		if strings.Contains(filename, "..") {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "Invalid file path",
			})
		}

		filePath := filepath.Join(config.UploadDir, filename)
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Image not found",
			})
		}

		if err := os.Remove(filePath); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to delete image",
			})
		}

		return c.JSON(fiber.Map{
			"message": "Image deleted successfully",
		})
	}
}

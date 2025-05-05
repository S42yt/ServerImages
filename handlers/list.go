package handlers

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/gofiber/fiber/v2"

	"github.com/S42yt/serverimages/config"
	"github.com/S42yt/serverimages/models"
)

func ListImages(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {

		files, err := os.ReadDir(cfg.UploadDir)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to read upload directory",
			})
		}

		var images []models.ImageResponse

		for _, file := range files {

			if file.IsDir() || strings.HasPrefix(file.Name(), ".") {
				continue
			}

			filePath := filepath.Join(cfg.UploadDir, file.Name())

			info, err := os.Stat(filePath)
			if err != nil {

				continue
			}

			image := models.ImageResponse{
				URL:        fmt.Sprintf("%s/cdn/%s", cfg.ServerURL, file.Name()),
				ID:         file.Name(),
				Size:       int(info.Size()),
				UploadedAt: info.ModTime(),
			}

			images = append(images, image)
		}

		sort.Slice(images, func(i, j int) bool {
			return images[i].UploadedAt.After(images[j].UploadedAt)
		})

		return c.JSON(images)
	}
}

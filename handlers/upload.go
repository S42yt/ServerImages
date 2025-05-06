package handlers

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
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

const turnstileVerifyURL = "https://challenges.cloudflare.com/turnstile/v0/siteverify"

type turnstileRequest struct {
	Secret   string `json:"secret"`
	Response string `json:"response"`
	RemoteIP string `json:"remoteip,omitempty"`
}

type turnstileResponse struct {
	Success     bool      `json:"success"`
	ChallengeTS time.Time `json:"challenge_ts"`
	Hostname    string    `json:"hostname"`
	ErrorCodes  []string  `json:"error-codes"`
	Action      string    `json:"action"`
	CData       string    `json:"cdata"`
}

func verifyTurnstile(secretKey, token, remoteIP string) (bool, error) {
	if secretKey == "" {
		log.Println("Turnstile secret key is not configured.")
		return false, errors.New("turnstile secret key is missing")
	}
	if token == "" {
		return false, errors.New("turnstile token is missing")
	}

	reqBody := turnstileRequest{
		Secret:   secretKey,
		Response: token,
		RemoteIP: remoteIP,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		log.Printf("Error marshalling turnstile request: %v\n", err)
		return false, fmt.Errorf("failed to marshal turnstile request: %w", err)
	}

	resp, err := http.Post(turnstileVerifyURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("Error sending request to turnstile verify endpoint: %v\n", err)
		return false, fmt.Errorf("failed to send request to turnstile verify endpoint: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Turnstile verification failed with status code: %d\n", resp.StatusCode)
		return false, fmt.Errorf("turnstile verification request failed with status: %s", resp.Status)
	}

	var verifyResp turnstileResponse
	if err := json.NewDecoder(resp.Body).Decode(&verifyResp); err != nil {
		log.Printf("Error decoding turnstile response: %v\n", err)
		return false, fmt.Errorf("failed to decode turnstile response: %w", err)
	}

	if !verifyResp.Success {
		log.Printf("Turnstile verification unsuccessful. Error codes: %v\n", verifyResp.ErrorCodes)
		if len(verifyResp.ErrorCodes) > 0 {
			return false, fmt.Errorf("turnstile verification failed: %s", strings.Join(verifyResp.ErrorCodes, ", "))
		}
		return false, errors.New("turnstile verification failed")
	}

	return true, nil
}

func Upload(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		turnstileToken := ""
		contentType := string(c.Request().Header.ContentType())

		if strings.Contains(contentType, "multipart/form-data") {
			turnstileToken = c.FormValue("cf-turnstile-response")
		} else if strings.Contains(contentType, "application/json") {

			turnstileToken = c.FormValue("cf-turnstile-response")
			if turnstileToken == "" {
				log.Println("Turnstile token not found in FormValue for JSON request.")
			}
		}

		clientIP := c.IP()

		verified, err := verifyTurnstile(cfg.TurnstileSecretKey, turnstileToken, clientIP)
		if err != nil {
			log.Printf("Turnstile verification error for IP %s: %v\n", clientIP, err)
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "CAPTCHA verification failed",
			})
		}
		if !verified {
			log.Printf("Turnstile verification failed for IP %s. Token: %s\n", clientIP, turnstileToken)
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "Invalid CAPTCHA token",
			})
		}
		log.Printf("Turnstile verification successful for IP %s\n", clientIP)

		var (
			fileData   []byte
			fileExt    string
			uploadedAt = time.Now()
		)

		if strings.Contains(contentType, "multipart/form-data") {
			file, err := c.FormFile("file")
			if err != nil {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error": "No file provided",
				})
			}

			if file.Size > cfg.MaxUploadSize {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error": fmt.Sprintf("File too large (max %dMB)", cfg.MaxUploadSize/(1024*1024)),
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

		} else if strings.Contains(contentType, "application/json") {
			var body models.Base64Upload
			// Need to read body again if token wasn't in FormValue/Header/Query
			// This might fail if body was already read for token extraction
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

			if int64(len(fileData)) > cfg.MaxUploadSize {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error": fmt.Sprintf("File too large (max %dMB)", cfg.MaxUploadSize/(1024*1024)),
				})
			}

		} else {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Unsupported content type",
			})
		}

		if len(fileData) == 0 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "No file data received",
			})
		}

		mimeType, err := utils.GetMimeType(fileData)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to detect file type",
			})
		}

		if !strings.HasPrefix(mimeType, cfg.AllowedMimeTypes) {
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
		filePath := filepath.Join(cfg.UploadDir, filename)

		err = os.WriteFile(filePath, fileData, 0644)
		if err != nil {
			log.Printf("Failed to save file %s: %v\n", filePath, err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to save file",
			})
		}

		response := models.ImageResponse{
			URL:        fmt.Sprintf("%s/cdn/%s", cfg.ServerURL, filename),
			ID:         filename,
			Size:       len(fileData),
			UploadedAt: uploadedAt,
		}

		return c.JSON(response)
	}
}

func DeleteImage(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		filename := c.Params("filename")

		if strings.Contains(filename, "..") {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "Invalid file path",
			})
		}

		filePath := filepath.Join(cfg.UploadDir, filename)

		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Image not found",
			})
		}

		if err := os.Remove(filePath); err != nil {
			log.Printf("Failed to delete image %s: %v\n", filePath, err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to delete image",
			})
		}

		return c.JSON(fiber.Map{
			"message": "Image deleted successfully",
		})
	}
}


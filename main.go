package main

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"

	"github.com/S42yt/serverimages/config"
	"github.com/S42yt/serverimages/handlers"
)

func main() {

	cfg := config.LoadFromEnv()

	if port := os.Getenv("PORT"); port != "" {
		portNum, err := strconv.Atoi(port)
		if err == nil {
			cfg.Port = portNum
			cfg.ServerURL = fmt.Sprintf("http://localhost:%d", portNum)
		}
	}

	if err := os.MkdirAll(cfg.UploadDir, 0755); err != nil {
		log.Fatalf("Failed to create upload directory: %v", err)
	}

	app := fiber.New(fiber.Config{
		BodyLimit: int(cfg.MaxUploadSize),
	})

	app.Use(logger.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin, Content-Type, Accept",
		AllowMethods: "GET, POST, DELETE",
	}))

	app.Post("/upload", handlers.Upload(cfg))
	app.Get("/cdn/:filename", handlers.ServeImage(cfg))
	app.Get("/images", handlers.ListImages(cfg))
	app.Delete("/cdn/:filename", handlers.DeleteImage(cfg))

	app.Static("/uploads", "./"+cfg.UploadDir)

	log.Printf("ServerImages started on port %d", cfg.Port)
	log.Printf("Upload endpoint: %s/upload", cfg.ServerURL)
	log.Printf("CDN endpoint: %s/cdn/{filename}", cfg.ServerURL)
	log.Printf("List endpoint: %s/images", cfg.ServerURL)
	log.Printf("Upload directory: %s", cfg.UploadDir)
	log.Printf("Max upload size: %d bytes (%d MB)", cfg.MaxUploadSize, cfg.MaxUploadSize/(1024*1024))

	log.Fatal(app.Listen(fmt.Sprintf(":%d", cfg.Port)))
}

package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"

	"github.com/S42yt/serverimages/config"
	"github.com/S42yt/serverimages/handlers"
)

func main() {
	if err := config.LoadConfig(); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	if err := os.MkdirAll(config.UploadDir, 0755); err != nil {
		log.Fatalf("Failed to create upload directory: %v", err)
	}

	app := fiber.New(fiber.Config{
		BodyLimit: int(config.MaxUploadSize),
	})

	app.Use(logger.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin, Content-Type, Accept",
		AllowMethods: "GET, POST, DELETE",
	}))

	app.Post("/upload", handlers.Upload())
	app.Get("/cdn/:filename", handlers.ServeImage())
	app.Get("/images", handlers.ListImages())
	app.Delete("/cdn/:filename", handlers.DeleteImage())

	app.Static("/uploads", "./"+config.UploadDir)

	log.Printf("ServerImages started on port %s", config.Port)
	log.Printf("Upload endpoint: %s/upload", config.ServerURL)
	log.Printf("CDN endpoint: %s/cdn/{filename}", config.ServerURL)
	log.Printf("List endpoint: %s/images", config.ServerURL)
	log.Printf("Upload directory: %s", config.UploadDir)
	log.Printf("Max upload size: %d bytes (%d MB)", config.MaxUploadSize, config.MaxUploadSize/(1024*1024))

	log.Fatal(app.Listen(fmt.Sprintf(":%s", config.Port)))
}

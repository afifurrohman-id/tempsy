package main

import (
	"github.com/afifurrohman-id/files-sync-client/cmd/client"
	"github.com/afifurrohman-id/files-sync-client/cmd/client/middleware"
	"github.com/afifurrohman-id/files-sync-client/internal"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/favicon"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/template/html/v2"
	"github.com/joho/godotenv"
	"os"
	"path"
	"time"
)

func init() {
	if os.Getenv("APP_ENV") != "production" {
		internal.Check(godotenv.Load(path.Join("configs", ".env")))
	}
}

func main() {
	engine := html.New(path.Join("web", "template"), ".go.html")
	app := fiber.New(fiber.Config{
		Views:              engine,
		CaseSensitive:      true,
		ErrorHandler:       middleware.CatchServerError,
		EnableIPValidation: true,
	})

	app.Use(compress.New(compress.Config{
		Level: compress.LevelBestSpeed,
	}), recover.New(), logger.New(), favicon.New(), middleware.Limiter)

	cacheDuration := 10 * time.Second
	if os.Getenv("APP_ENV") != "production" {
		cacheDuration = -1 * time.Second
	}

	app.Static("/public", path.Join("web", "static"), fiber.Static{
		CacheDuration: cacheDuration,
	})

	app.Get("/", middleware.CheckAuth, client.HandleWelcomeClient)

	routeAuth := app.Group("/auth")
	routeAuth.Get("/", client.OAuth2Callback)
	routeAuth.Get("/login", client.AuthLogin)
	routeAuth.Get("/logout", client.AuthLogout)

	routeDashboardUser := app.Group("/dashboard/:username", middleware.CheckAuth)
	routeDashboardUser.Get("/", client.HandleDashboardClient)
	routeDashboardUser.Get("/profile", client.HandleProfileClient)
	routeDashboardUser.Get("/:name", client.HandleDetailDataClient)

	routeDashboardUser.Use(middleware.SetRealIpClient)
	routeDashboardUser.Post("/", client.HandleUploadDashboardClient)
	routeDashboardUser.Put("/:name", client.HandleUpdateDataClient)
	routeDashboardUser.Delete("/profile", client.HandleDeleteAllDataClient)
	routeDashboardUser.Delete("/:name", client.HandleDeleteDataClient)

	if err := app.Listen(":" + os.Getenv("PORT")); err != nil {
		log.Panic(err)
	}
}

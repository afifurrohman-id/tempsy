package main

import (
	"io"
	"os"
	"path"

	"github.com/afifurrohman-id/tempsy/internal/files/utils"
	"github.com/afifurrohman-id/tempsy/pkg/middleware"
	"github.com/afifurrohman-id/tempsy/pkg/router"
	"github.com/gofiber/contrib/swagger"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/etag"
	"github.com/gofiber/fiber/v2/middleware/favicon"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/monitor"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/joho/godotenv"
)

func init() {
	if os.Getenv("APP_ENV") != "production" {
		utils.LogErr(godotenv.Load(path.Join("configs", ".env")))
	}
}

func main() {
	loggerFile, err := os.OpenFile("app.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0200)
	utils.Check(err)
	defer func() {
		utils.LogErr(loggerFile.Close())
	}()

	fileInfo, err := loggerFile.Stat()
	utils.Check(err)

	// truncate file, if size more than 15KB
	if fileInfo.Size() > int64(15<<10) {
		utils.LogErr(os.Truncate(loggerFile.Name(), 0))
	}

	multiWriter := io.MultiWriter(loggerFile, os.Stdout)

	app := fiber.New(fiber.Config{
		CaseSensitive:      true,
		BodyLimit:          middleware.MaxBodyLimit,
		ErrorHandler:       middleware.CatchServerError,
		AppName:            "Tempsy",
		EnableIPValidation: true,
	})

	app.Use(compress.New(compress.Config{
		Level: compress.LevelBestCompression,
	}), recover.New(), favicon.New(), logger.New(logger.Config{
		Output:        multiWriter,
		DisableColors: true,
	}), middleware.Cors, middleware.CheckHttpMethod, swagger.New(swagger.Config{
		FilePath: path.Join("api", "openapi-spec.yaml"),
		Path:     "/",
		Title:    "Tempsy API Documentation",
	}))

	app.Get("/monitor", middleware.Cache, monitor.New(monitor.Config{
		Title: "Tempsy API",
	}))
	app.Options("/*", func(ctx *fiber.Ctx) error {
		defer log.SetOutput(os.Stderr)

		log.SetOutput(os.Stdout)
		log.Info("Options Method from browser")

		return ctx.SendStatus(fiber.StatusNoContent)
	})

	routeAuthApi := app.Group("/auth")
	routeAuthApi.Get("/userinfo/me", middleware.RateLimiterProcessing, etag.New(), router.HandleGetUserInfo)
	routeAuthApi.Get("/guest/token", middleware.RateLimiterGuestToken, router.HandleGetGuestToken)

	routeFilesByUsername := app.Group("/files/:username", middleware.PurgeAnonymousAccount, middleware.AutoDeleteScheduler)
	routeFilesByUsername.Get("/public/:filename", middleware.Cache, router.HandleGetPublicFile)
	routeFilesByUsername.Get("/", middleware.CheckAuth, middleware.RateLimiterProcessing, etag.New(), router.HandleListFilesData)
	routeFilesByUsername.Get("/:filename", middleware.CheckAuth, middleware.RateLimiterProcessing, etag.New(), router.HandleGetFileData)
	routeFilesByUsername.Post("/", middleware.CheckAuth, middleware.RateLimiterProcessing, router.HandleUploadFile)
	routeFilesByUsername.Put("/:filename", middleware.CheckAuth, middleware.RateLimiterProcessing, router.HandleUpdateFile)
	routeFilesByUsername.Delete("/", middleware.CheckAuth, middleware.RateLimiterProcessing, router.HandleDeleteAllFile)
	routeFilesByUsername.Delete("/:filename", middleware.CheckAuth, middleware.RateLimiterProcessing, router.HandleDeleteFile)

	if err := app.Listen(":" + os.Getenv("PORT")); err != nil {
		log.Panic(err)
	}
}

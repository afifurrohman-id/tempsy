package files

import (
	"cloud.google.com/go/storage"
	"context"
	"errors"
	"fmt"
	"github.com/afifurrohman-id/tempsy/internal"
	"github.com/afifurrohman-id/tempsy/internal/models"
	store "github.com/afifurrohman-id/tempsy/internal/storage"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"time"
)

func HandleGetPublicFile(ctx *fiber.Ctx) error {
	storeCtx, cancel := context.WithTimeout(context.Background(), store.DefaultTimeoutCtx)
	defer cancel()

	var (
		username = ctx.Params("username")
		fileName = ctx.Params("filename")
		filePath = fmt.Sprintf("%s/%s", username, fileName)
	)

	fileData, err := store.GetObject(storeCtx, filePath)
	if err != nil {
		if errors.Is(err, storage.ErrObjectNotExist) {
			return ctx.Status(fiber.StatusNotFound).JSON(&models.ApiError{
				Type:        internal.ErrorTypeFileNotPublic,
				Description: fmt.Sprintf("File: %s, Is Not Found Or Not Public", fileName),
			})
		}
		log.Panic(err)
	}
	if !fileData.IsPublic {
		return ctx.Status(fiber.StatusNotFound).JSON(&models.ApiError{
			Type:        internal.ErrorTypeFileNotPublic,
			Description: fmt.Sprintf("File: %s, Is Not Found Or Not Public", fileName),
		})
	}

	agent := fiber.Get(fileData.Url)
	agent.Timeout(10 * time.Second)

	statusCode, fileByte, errs := agent.Bytes()
	if len(errs) > 0 {
		internal.Check(errs[0])
	}

	if statusCode != fiber.StatusOK {
		log.Panic("Unknown Error in Service File")
	}

	ctx.Set(fiber.HeaderContentType, fileData.ContentType)
	ctx.Set(fiber.HeaderContentLength, fmt.Sprintf("%d", len(fileByte))) // maybe unnecessary

	return ctx.Send(fileByte)
}

func HandleGetFileData(ctx *fiber.Ctx) error {
	var (
		username = ctx.Params("username")
		fileName = ctx.Params("filename")
		filePath = fmt.Sprintf("%s/%s", username, fileName)
	)

	storeCtx := context.Background()
	storeCtx, cancel := context.WithTimeout(storeCtx, store.DefaultTimeoutCtx)
	defer cancel()

	fileData, err := store.GetObject(storeCtx, filePath)
	if err != nil {
		if errors.Is(err, storage.ErrObjectNotExist) {
			return ctx.Status(fiber.StatusNotFound).JSON(&models.ApiError{
				Type:        internal.ErrorTypeFileNotFound,
				Description: fmt.Sprintf("File: %s, Is Not Found", fileName),
			})
		}
		log.Panic(err)
	}

	store.Format(fileData)
	return ctx.JSON(&fileData)
}

func HandleGetAllFileData(ctx *fiber.Ctx) error {
	var (
		username = ctx.Params("username")
		storeCtx = context.Background()
	)

	storeCtx, cancel := context.WithTimeout(storeCtx, store.DefaultTimeoutCtx)
	defer cancel()

	filesData, err := store.GetAllObject(storeCtx, username)
	internal.Check(err)

	for i, fileData := range filesData {
		store.Format(fileData)
		filesData[i] = fileData
	}

	return ctx.JSON(&filesData)
}

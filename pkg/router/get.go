package router

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"cloud.google.com/go/storage"
	"github.com/afifurrohman-id/tempsy/internal/files/models"
	store "github.com/afifurrohman-id/tempsy/internal/files/storage"
	"github.com/afifurrohman-id/tempsy/internal/files/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
)

func HandleGetPublicFile(ctx *fiber.Ctx) error {
	storeCtx, cancel := context.WithTimeout(context.Background(), store.DefaultTimeoutCtx)
	defer cancel()

	var (
		fileName = ctx.Params("filename")
		filePath = fmt.Sprintf("%s/%s", ctx.Params("username"), fileName)
	)

	fileData, err := store.GetObject(storeCtx, filePath)
	if err != nil {
		if errors.Is(err, storage.ErrObjectNotExist) {
			return ctx.Status(fiber.StatusNotFound).JSON(&models.ApiError{
				Type:        utils.ErrorTypeFileNotPublic,
				Description: fmt.Sprintf("File: %s, Is Not Found Or Not Public", fileName),
			})
		}
		log.Panic(err)
	}
	if !fileData.IsPublic {
		return ctx.Status(fiber.StatusNotFound).JSON(&models.ApiError{
			Type:        utils.ErrorTypeFileNotPublic,
			Description: fmt.Sprintf("File: %s, Is Not Found Or Not Public", fileName),
		})
	}

	agent := fiber.Get(fileData.Url)
	agent.Timeout(10 * time.Second)

	statusCode, fileByte, errs := agent.Bytes()
	if len(errs) > 0 {
		utils.Check(errs[0])
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
		fileName = ctx.Params("filename")
		filePath = fmt.Sprintf("%s/%s", ctx.Params("username"), fileName)
	)

	storeCtx, cancel := context.WithTimeout(context.Background(), store.DefaultTimeoutCtx)
	defer cancel()

	fileData, err := store.GetObject(storeCtx, filePath)
	if err != nil {
		if errors.Is(err, storage.ErrObjectNotExist) {
			return ctx.Status(fiber.StatusNotFound).JSON(&models.ApiError{
				Type:        utils.ErrorTypeFileNotFound,
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
		mu = new(sync.Mutex)
		wg = new(sync.WaitGroup)
	)

	storeCtx, cancel := context.WithTimeout(context.Background(), store.DefaultTimeoutCtx)
	defer cancel()

	filesData, err := store.GetAllObject(storeCtx, ctx.Params("username"))
	utils.Check(err)

	wg.Add(1)
	go func() {
		defer func() {
			mu.Unlock()
			wg.Done()
		}()
		mu.Lock()

		for i, fileData := range filesData {
			store.Format(fileData)
			filesData[i] = fileData
		}
	}()

	wg.Wait()

	return ctx.JSON(&filesData)
}

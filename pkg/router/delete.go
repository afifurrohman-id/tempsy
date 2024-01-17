package router

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"cloud.google.com/go/storage"
	"github.com/afifurrohman-id/tempsy/internal/files/models"
	store "github.com/afifurrohman-id/tempsy/internal/files/storage"
	"github.com/afifurrohman-id/tempsy/internal/files/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"golang.org/x/sync/errgroup"
)

func HandleDeleteFile(ctx *fiber.Ctx) error {
	var (
		fileName = ctx.Params("filename")
		filePath = fmt.Sprintf("%s/%s", ctx.Params("username"), fileName)
	)

	storeCtx, cancel := context.WithTimeout(context.Background(), store.DefaultTimeoutCtx)
	defer cancel()

	if _, err := store.GetObject(storeCtx, filePath); err != nil {

		if errors.Is(err, storage.ErrObjectNotExist) {
			return ctx.Status(fiber.StatusNotFound).JSON(&models.ApiError{
				Type:        utils.ErrorTypeFileNotFound,
				Description: fmt.Sprintf("File: %s, Is Not Found", fileName),
			})
		}
		utils.Check(err)
	}

	utils.Check(store.DeleteObject(storeCtx, filePath))

	return ctx.SendStatus(fiber.StatusNoContent)
}

func HandleDeleteAllFile(ctx *fiber.Ctx) error {
	username := ctx.Params("username")

	storeCtx, cancel := context.WithTimeout(context.Background(), store.DefaultTimeoutCtx)
	defer cancel()

	filesData, err := store.ListObjects(storeCtx, username)
	utils.Check(err)

	if len(filesData) == 0 {
		return ctx.Status(fiber.StatusBadRequest).JSON(&models.ApiError{
			Type:        utils.ErrorTypeEmptyData,
			Description: "Cannot delete empty data files, no data for user: " + username,
		})
	}

	var (
		eg = new(errgroup.Group)
		mu = new(sync.Mutex)
	)

	eg.Go(func() error {
		defer mu.Unlock()

		mu.Lock()
		for _, fileData := range filesData {
			if err = store.DeleteObject(storeCtx, fileData.Name); err != nil {
				return err
			}
		}

		return nil
	})

	if err = eg.Wait(); err != nil {
		log.Panic(err)
	}

	return ctx.SendStatus(fiber.StatusNoContent)
}

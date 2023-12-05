package files

import (
	"cloud.google.com/go/storage"
	"context"
	"errors"
	"fmt"
	"github.com/afifurrohman-id/tempsy/internal"
	"github.com/afifurrohman-id/tempsy/internal/models"
	"github.com/afifurrohman-id/tempsy/internal/storage"
	"github.com/gofiber/fiber/v2"
)

func HandleDeleteFile(ctx *fiber.Ctx) error {
	var (
		username = ctx.Params("username")
		fileName = ctx.Params("filename")
		storeCtx = context.Background()
		filePath = fmt.Sprintf("%s/%s", username, fileName)
	)

	storeCtx, cancel := context.WithTimeout(storeCtx, store.DefaultTimeoutCtx)
	defer cancel()

	if _, err := store.GetObject(storeCtx, filePath); err != nil {

		if errors.Is(err, storage.ErrObjectNotExist) {
			return ctx.Status(fiber.StatusNotFound).JSON(&models.ApiError{
				Type:        internal.ErrorTypeFileNotFound,
				Description: fmt.Sprintf("File: %s, Is Not Found", fileName),
			})
		}
		internal.Check(err)
	}

	internal.Check(store.DeleteObject(storeCtx, filePath))

	return ctx.SendStatus(fiber.StatusNoContent)
}

func HandleDeleteAllFile(ctx *fiber.Ctx) error {
	var (
		username = ctx.Params("username")
		storeCtx = context.Background()
	)

	storeCtx, cancel := context.WithTimeout(storeCtx, store.DefaultTimeoutCtx)
	defer cancel()

	filesData, err := store.GetAllObject(storeCtx, username)
	internal.Check(err)

	if len(filesData) == 0 {
		return ctx.Status(fiber.StatusBadRequest).JSON(&models.ApiError{
			Type:        internal.ErrorTypeEmptyData,
			Description: fmt.Sprintf("Cannot delete empty data files, no data for user: %s", username),
		})
	}

	//TODO: More efficient way to delete all files
	for _, fileData := range filesData {
		internal.Check(store.DeleteObject(storeCtx, fileData.Name))
	}

	return ctx.SendStatus(fiber.StatusNoContent)
}

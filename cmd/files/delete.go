package files

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"cloud.google.com/go/storage"
	"github.com/afifurrohman-id/tempsy/internal"
	"github.com/afifurrohman-id/tempsy/internal/models"
	"github.com/afifurrohman-id/tempsy/internal/storage"
	"github.com/gofiber/fiber/v2"
	"golang.org/x/sync/errgroup"
  "github.com/gofiber/fiber/v2/log"
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
    eg = new(errgroup.Group)
    mu = new(sync.Mutex)
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

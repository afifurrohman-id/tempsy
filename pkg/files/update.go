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
	"github.com/gofiber/fiber/v2/log"
	"strings"
)

// HandleUpdateFile Updates single file by name
func HandleUpdateFile(ctx *fiber.Ctx) error {
	var (
		fileName = ctx.Params("filename")
		filePath = fmt.Sprintf("%s/%s", ctx.Params("username"), fileName)
	)

	storeCtx, cancel := context.WithTimeout(context.Background(), store.DefaultTimeoutCtx)
	defer cancel()

	if len(ctx.Body()) < 1 {
		return ctx.Status(fiber.StatusBadRequest).JSON(&models.ApiError{
			Type:        internal.ErrorTypeEmptyFile,
			Description: "Cannot Update File with Empty File",
		})
	}

	// Check if file exists
	file, err := store.GetObject(storeCtx, filePath)
	if err != nil {
		if errors.Is(err, storage.ErrObjectNotExist) {
			return ctx.Status(fiber.StatusNotFound).JSON(&models.ApiError{
				Type:        internal.ErrorTypeFileNotFound,
				Description: fmt.Sprintf("File %s Is Not Found", fileName),
			})
		}
		log.Panic(err)
	}

	if !strings.Contains(file.ContentType, ctx.Get(fiber.HeaderContentType)) {
		return ctx.Status(fiber.StatusBadRequest).JSON(&models.ApiError{
			Type:        internal.ErrorTypeMismatchType,
			Description: "Please use the same content type as the original file",
		})
	}

	fileHeader := mapFileHeader(ctx.GetReqHeaders())

	fileMetadata := new(models.DataFile)
	fileMetadata.ContentType = fileHeader[fiber.HeaderContentType]

	if err = store.UnmarshalMetadata(fileHeader, fileMetadata); err != nil {
		log.Errorf("Error Unmarshal File Metadata: %s", err.Error())

		return ctx.Status(fiber.StatusUnprocessableEntity).JSON(&models.ApiError{
			Type:        internal.ErrorTypeInvalidHeaderFile,
			Description: strings.Join(strings.Split(err.Error(), "_"), " "),
		})
	}

	if err = validateExpiry(fileMetadata.PrivateUrlExpires, fileMetadata.AutoDeletedAt); err != nil {
		log.Errorf("Error Validate Expiry: %s", err.Error())

		return ctx.Status(fiber.StatusUnprocessableEntity).JSON(&models.ApiError{
			Type:        internal.ErrorTypeInvalidHeaderFile,
			Description: strings.Join(strings.Split(err.Error(), "_"), " "),
		})
	}

	fileMetadata.Name = fileName // Bypass file name, for preventing file name change

	internal.Check(store.DeleteObject(storeCtx, filePath))
	internal.Check(store.UploadObject(storeCtx, filePath, ctx.Body(), fileMetadata))

	fileData, err := store.GetObject(storeCtx, filePath)
	internal.Check(err)

	store.Format(fileData)
	return ctx.JSON(&fileData)
}
package router

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"cloud.google.com/go/storage"
	"github.com/afifurrohman-id/tempsy/internal/files/models"
	store "github.com/afifurrohman-id/tempsy/internal/files/storage"
	"github.com/afifurrohman-id/tempsy/internal/files/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
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
			Error: &models.Error{
				Kind:        utils.ErrorTypeEmptyFile,
				Description: "Cannot Update File with Empty File",
			},
		})
	}

	// Check if file exists
	file, err := store.GetObject(storeCtx, filePath)
	if err != nil {
		if errors.Is(err, storage.ErrObjectNotExist) {
			return ctx.Status(fiber.StatusNotFound).JSON(&models.ApiError{
				Error: &models.Error{
					Kind:        utils.ErrorTypeFileNotFound,
					Description: fmt.Sprintf("File %s Is Not Found", fileName),
				},
			})
		}
		log.Panic(err)
	}

	if !strings.Contains(file.MimeType, ctx.Get(fiber.HeaderContentType)) {
		return ctx.Status(fiber.StatusBadRequest).JSON(&models.ApiError{
			Error: &models.Error{
				Kind:        utils.ErrorTypeMismatchType,
				Description: "Please use the same content type as the original file",
			},
		})
	}

	fileHeader := store.MapFileHeader(ctx.GetReqHeaders())

	fileMetadata := new(models.DataFile)
	fileMetadata.MimeType = fileHeader.Get(fiber.HeaderContentType)

	if err = store.UnmarshalMetadata(fileHeader, fileMetadata); err != nil {
		log.Error("Error Unmarshal File Metadata: " + err.Error())

		return ctx.Status(fiber.StatusUnprocessableEntity).JSON(&models.ApiError{
			Error: &models.Error{
				Kind:        utils.ErrorTypeInvalidHeaderFile,
				Description: strings.Join(strings.Split(err.Error(), "_"), " "),
			},
		})
	}

	if err = validateExpiry(fileMetadata.PrivateUrlExpires, fileMetadata.AutoDeleteAt); err != nil {
		log.Error("Error Validate Expiry: " + err.Error())

		return ctx.Status(fiber.StatusUnprocessableEntity).JSON(&models.ApiError{
			Error: &models.Error{
				Kind:        utils.ErrorTypeInvalidHeaderFile,
				Description: strings.Join(strings.Split(err.Error(), "_"), " "),
			},
		})
	}

	fileMetadata.Name = fileName // Bypass file name, for preventing file name change

	utils.Check(store.DeleteObject(storeCtx, filePath))
	utils.Check(store.UploadObject(storeCtx, filePath, ctx.Body(), fileMetadata))

	fileData, err := store.GetObject(storeCtx, filePath)
	utils.Check(err)

	store.Format(fileData)
	return ctx.JSON(&fileData)
}

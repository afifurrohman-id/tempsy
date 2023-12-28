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
	"golang.org/x/exp/slices"
	"regexp"
	"strings"
	"time"
)

// TODO: Make parameter as pointer ??
func mapFileHeader(header map[string][]string) map[string]string {
	fileHeader := make(map[string]string)
	for key, value := range header {
		fileHeader[key] = value[0]
	}

	return fileHeader
}

func HandleUploadFile(ctx *fiber.Ctx) error {

	var (
		storeCtx = context.Background()
		username = ctx.Params("username")
    fileName = ctx.Get(store.HeaderFileName)
		filePath = fmt.Sprintf("%s/%s", username, fileName)
	)

	storeCtx, cancel := context.WithTimeout(storeCtx, store.DefaultTimeoutCtx)
	defer cancel()

	if len(ctx.Body()) < 1 {
		return ctx.Status(fiber.StatusBadRequest).JSON(&models.ApiError{
			Type:        internal.ErrorTypeEmptyFile,
			Description: "Cannot Upload Empty File",
		})
	}

	match, err := regexp.MatchString(`^[a-zA-Z0-9_-]+\.+[a-zA-Z0-9_-]+$`, fileName)
	internal.Check(err)

	if !match {
		return ctx.Status(fiber.StatusBadRequest).JSON(&models.ApiError{
			Type:        internal.ErrorTypeInvalidFileName,
			Description: "File name must be alphanumeric and contain extension separated by dot, underscore, or dash",
		})
	}

	// Check if file already exists
	dataFile, err := store.GetObject(storeCtx, filePath)
	if err != nil {
		if errors.Is(err, storage.ErrObjectNotExist) {
			fileHeader := mapFileHeader(ctx.GetReqHeaders())

			fileMetadata := new(models.DataFile)
			if !slices.Contains(store.AcceptedContentType, fileHeader[fiber.HeaderContentType]) {

				return ctx.Status(fiber.StatusUnsupportedMediaType).JSON(&models.ApiError{
					Type:        internal.ErrorTypeUnsupportedType,
					Description: fmt.Sprintf("Unsupported Content-Type: %s", fileHeader[fiber.HeaderContentType]),
				})
			}

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

			internal.Check(store.UploadObject(storeCtx, filePath, ctx.Body(), fileMetadata))

			dataFile, err = store.GetObject(storeCtx, filePath)
			internal.Check(err)

			store.Format(dataFile)
			return ctx.Status(fiber.StatusCreated).JSON(&dataFile)

		}
		log.Panic(err)
	}

	return ctx.Status(fiber.StatusConflict).JSON(&models.ApiError{
		Type:        internal.ErrorTypeFileExists,
		Description: fmt.Sprintf("File: %s Already Exists", fileName),
	})
}

func validateExpiry(urlExp int, autoDel int64) error {
	if time.Now().Add(time.Duration(urlExp)*time.Second).UnixMilli() > autoDel {
		return errors.New("private_url_expires_cannot_be_later_than_auto_deleted_at_starting_from_now")
	}

	// cutoff one year from now
	if cutoff := time.Now().Add(8766 * time.Hour); !time.UnixMilli(autoDel).Before(cutoff) {
		return errors.New("auto_deleted_at_cannot_be_later_than_1_year_from_now")
	}

	return nil
}

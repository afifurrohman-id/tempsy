package router

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"cloud.google.com/go/storage"
	"github.com/afifurrohman-id/tempsy/internal/files/models"
	store "github.com/afifurrohman-id/tempsy/internal/files/storage"
	"github.com/afifurrohman-id/tempsy/internal/files/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"golang.org/x/exp/slices"
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
		fileName = ctx.Get(store.HeaderFileName)
		filePath = fmt.Sprintf("%s/%s", ctx.Params("username"), fileName)
	)

	storeCtx, cancel := context.WithTimeout(context.Background(), store.DefaultTimeoutCtx)
	defer cancel()

	if len(ctx.Body()) < 1 {
		return ctx.Status(fiber.StatusBadRequest).JSON(&models.ApiError{
			Error: &models.Error{
				Kind:        utils.ErrorTypeEmptyFile,
				Description: "Cannot Upload Empty File",
			},
		})
	}

	match, err := regexp.MatchString(`^[a-zA-Z0-9_-]+\.+[a-zA-Z0-9_-]+$`, fileName)
	utils.Check(err)

	if !match {
		return ctx.Status(fiber.StatusBadRequest).JSON(&models.ApiError{
			Error: &models.Error{
				Kind:        utils.ErrorTypeInvalidFileName,
				Description: "File name must be alphanumeric lowercase or uppercase split by underscore, or dash and contain extension separated by dot",
			},
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
					Error: &models.Error{
						Kind:        utils.ErrorTypeUnsupportedType,
						Description: "Unsupported Content-Type: " + fileHeader[fiber.HeaderContentType],
					},
				})
			}

			fileMetadata.MimeType = fileHeader[fiber.HeaderContentType]

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

			utils.Check(store.UploadObject(storeCtx, filePath, ctx.Body(), fileMetadata))

			dataFile, err = store.GetObject(storeCtx, filePath)
			utils.Check(err)

			store.Format(dataFile)
			return ctx.Status(fiber.StatusCreated).JSON(&dataFile)

		}
		log.Panic(err)
	}

	return ctx.Status(fiber.StatusConflict).JSON(&models.ApiError{
		Error: &models.Error{
			Kind:        utils.ErrorTypeFileExists,
			Description: fmt.Sprintf("File: %s Already Exists", fileName),
		},
	})
}

func validateExpiry(urlExp uint, autoDel int64) error {
	if time.Now().Add(time.Duration(urlExp)*time.Second).UnixMilli() > autoDel {
		return errors.New("private_url_expires_cannot_be_later_than_auto_deleted_at_starting_from_now")
	}

	// cutoff one year from now
	if cutoff := time.Now().Add(8766 * time.Hour); !time.UnixMilli(autoDel).Before(cutoff) {
		return errors.New("auto_deleted_at_cannot_be_later_than_1_year_from_now")
	}

	if urlExp > 604800 || urlExp < 2 {
		return errors.New("private_url_expires_must_be_less_than_7_days_in_seconds_and_more_than_2_seconds")
	}

	return nil
}

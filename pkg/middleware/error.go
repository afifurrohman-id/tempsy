package middleware

import (
	"errors"
	"fmt"

	"github.com/afifurrohman-id/tempsy/internal/files/auth"
	"github.com/afifurrohman-id/tempsy/internal/files/models"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"golang.org/x/exp/slices"
)

func CatchServerError(ctx *fiber.Ctx, err error) error {
	fiberErr := new(fiber.Error)
	if errors.As(err, &fiberErr) {
		log.Error("Fiber - ", fiberErr)

		if fiberErr.Code == fiber.StatusMethodNotAllowed && slices.Contains[[]string](auth.AllowedHttpMethod, ctx.Method()) || fiberErr.Code == fiber.StatusNotFound {
			return ctx.Status(fiber.StatusNotFound).JSON(&models.ApiError{
				Type:        "resource_not_found",
				Description: fmt.Sprintf("Path %s for Http Method %s Is Not Found", ctx.Path(), ctx.Method()),
			})
		}

		if fiberErr.Code == fiber.StatusRequestEntityTooLarge {
			return ctx.Status(fiberErr.Code).JSON(&models.ApiError{
				Type:        "request_entity_too_large",
				Description: "Maximum Entity Exceeded, Max Request Body is 30MB For File",
			})
		}

		return ctx.Status(fiberErr.Code).JSON(&models.ApiError{
			Type:        fmt.Sprintf("%d", fiberErr.Code),
			Description: fiberErr.Message,
		})
	}

	log.Error("Server - ", err)
	return ctx.Status(fiber.StatusInternalServerError).JSON(&models.ApiError{
		Type:        "unknown_server_error",
		Description: "Unknown Internal Server Error",
	})
}

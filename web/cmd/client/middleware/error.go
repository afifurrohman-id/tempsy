package middleware

import (
	"errors"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
)

func CatchServerError(ctx *fiber.Ctx, err error) error {

	fiberError := new(fiber.Error)
	if errors.As(err, &fiberError) {
		log.Error("Fiber Error - ", fiberError)

		if fiberError.Code == fiber.StatusNotFound {
			return ctx.Status(fiberError.Code).Render("pages/error", map[string]any{
				"title":   fmt.Sprintf("Error - %d", fiberError.Code),
				"message": fmt.Sprintf("Page for Path: %s Is Not Found", ctx.Path()),
			})
		}
		return ctx.Status(fiberError.Code).Render("pages/error", map[string]any{
			"title":   fmt.Sprintf("Error - %d", fiberError.Code),
			"message": fiberError.Code,
		})
	}

	log.Error("Server Error - ", err)
	return ctx.Status(fiber.StatusInternalServerError).Render("pages/error", map[string]any{
		"title":   fmt.Sprintf("Error - %d", fiber.StatusInternalServerError),
		"message": "Internal Server Error",
	})
}

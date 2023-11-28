package middleware

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
)

var Limiter = limiter.New(limiter.Config{
	Max: 100,
	LimitReached: func(ctx *fiber.Ctx) error {
		return ctx.Status(fiber.StatusTooManyRequests).Render("pages/error", map[string]any{
			"title":   fmt.Sprintf("Error - %d", fiber.StatusTooManyRequests),
			"message": "Too Many Requests, Please Try Again Later",
		})
	},
})

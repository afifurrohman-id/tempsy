package middleware

import (
	"fmt"
	"github.com/afifurrohman-id/tempsy/internal/auth"
	"github.com/afifurrohman-id/tempsy/internal/models"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
)

const (
	MaxBodyLimit               = 30 << 20 // 30MB
	MaxReqProcsPerSeconds      = 30
	MaxReqGuestTokenPerSeconds = 3
)

var RateLimiterProcessing = limiter.New(limiter.Config{
	KeyGenerator: func(ctx *fiber.Ctx) string {
		return ctx.Get(fiber.HeaderAuthorization)
	},
	Max: MaxReqProcsPerSeconds,
	LimitReached: func(ctx *fiber.Ctx) error {
		return ctx.Status(fiber.StatusTooManyRequests).JSON(&models.ApiError{
			Type:        "too_many_request",
			Description: fmt.Sprintf("Maximum Request Exceeded, Maximum %d Request per seconds for user", MaxReqProcsPerSeconds),
		})
	},
})

var RateLimiterGuestToken = limiter.New(limiter.Config{
	Max: MaxReqGuestTokenPerSeconds,
	KeyGenerator: func(ctx *fiber.Ctx) string {
		if ctx.Get(auth.HeaderRealIp) != "" {
			return ctx.Get(auth.HeaderRealIp)
		}

		if ctx.Get(auth.HeaderXRealIp) != "" {
			return ctx.Get(auth.HeaderXRealIp)
		}

		return ctx.IP()
	},
	LimitReached: func(ctx *fiber.Ctx) error {
		return ctx.Status(fiber.StatusTooManyRequests).JSON(&models.ApiError{
			Type:        "too_many_request_token",
			Description: fmt.Sprintf("Maximum Request Exceeded, Maximum %d Request per seconds for guest token", MaxReqGuestTokenPerSeconds),
		})
	},
})

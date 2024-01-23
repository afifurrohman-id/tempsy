package middleware

import (
	"fmt"

	"github.com/afifurrohman-id/tempsy/internal/files/auth"
	"github.com/afifurrohman-id/tempsy/internal/files/models"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/utils"
)

const (
	MaxBodyLimit               = 30 << 20 // 30MB
	MaxReqProcsPerSeconds      = 30
	MaxReqGuestTokenPerSeconds = 3
)

var RateLimiterProcessing = limiter.New(limiter.Config{
	KeyGenerator: func(ctx *fiber.Ctx) string {
		/* Go fiber is immutable by default,
		need to copy the string to prevent unexpected behavior
		*/
		return utils.CopyString(ctx.Get(fiber.HeaderAuthorization))
	},
	Max: MaxReqProcsPerSeconds,
	LimitReached: func(ctx *fiber.Ctx) error {
		return ctx.Status(fiber.StatusTooManyRequests).JSON(&models.ApiError{
			Error: &models.Error{
				Kind:        "too_many_request",
				Description: fmt.Sprintf("Maximum Request Exceeded, Maximum %d Request per seconds for user", MaxReqProcsPerSeconds),
			},
		})
	},
})

var RateLimiterGuestToken = limiter.New(limiter.Config{
	Max: MaxReqGuestTokenPerSeconds,
	KeyGenerator: func(ctx *fiber.Ctx) string {
		var (
			realIp  = ctx.Get(auth.HeaderRealIp)
			xRealIp = ctx.Get(auth.HeaderXRealIp)
		)

		if realIp != "" {
			return utils.CopyString(realIp)
		}

		if xRealIp != "" {
			return utils.CopyString(xRealIp)
		}

		// ctx.IP() is copy by default
		return ctx.IP()
	},
	LimitReached: func(ctx *fiber.Ctx) error {
		return ctx.Status(fiber.StatusTooManyRequests).JSON(&models.ApiError{
			Error: &models.Error{
				Kind:        "too_many_request_token",
				Description: fmt.Sprintf("Maximum Request Exceeded, Maximum %d Request per seconds for guest token", MaxReqGuestTokenPerSeconds),
			},
		})
	},
})

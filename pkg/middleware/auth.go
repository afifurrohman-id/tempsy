package middleware

import (
	"fmt"
	"strings"

	"github.com/afifurrohman-id/tempsy/internal/files/auth"
	"github.com/afifurrohman-id/tempsy/internal/files/auth/guest"
	"github.com/afifurrohman-id/tempsy/internal/files/auth/oauth2"
	"github.com/afifurrohman-id/tempsy/internal/files/models"
	"github.com/afifurrohman-id/tempsy/internal/files/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"golang.org/x/exp/slices"
)

func CheckHttpMethod(ctx *fiber.Ctx) error {
	if !slices.Contains(auth.AllowedHttpMethod, ctx.Method()) {
		return ctx.Status(fiber.StatusMethodNotAllowed).JSON(&models.ApiError{
			Type:        "method_not_allowed",
			Description: fmt.Sprintf("Http Method: %s Is Not Allowed", ctx.Method()),
		})
	}

	return ctx.Next()
}

func CheckAuth(ctx *fiber.Ctx) error {
	var (
		username  = ctx.Params("username")
		authToken = ctx.Get(fiber.HeaderAuthorization)
	)

	if strings.HasPrefix(authToken, auth.BearerPrefix) {
		if strings.HasPrefix(username, guest.UsernamePrefix) {
			tokenMap, err := guest.ParseToken(strings.TrimPrefix(authToken, auth.BearerPrefix))

			if err == nil && tokenMap["jti"] == username {
				return ctx.Next()
			}
			utils.LogErr(err)
		} else {
			accountInfo, err := oauth2.GetGoogleAccountInfo(strings.TrimPrefix(authToken, auth.BearerPrefix))

			if err == nil && username == accountInfo.UserName && accountInfo.VerifiedEmail {
				return ctx.Next()
			}
			utils.LogErr(err)
		}
	}

	return ctx.Status(fiber.StatusUnauthorized).JSON(&models.ApiError{Type: "unauthorized", Description: "You don't have right access to this resources"})
}

var Cors = cors.New(cors.Config{
	AllowMethods: strings.Join(auth.AllowedHttpMethod, ","),
	AllowHeaders: strings.Join([]string{fiber.HeaderContentType, fiber.HeaderContentLength, fiber.HeaderAccept, fiber.HeaderUserAgent, fiber.HeaderAcceptEncoding, fiber.HeaderAcceptCharset, fiber.HeaderAuthorization, fiber.HeaderOrigin, fiber.HeaderLocation, fiber.HeaderKeepAlive}, ","),
})

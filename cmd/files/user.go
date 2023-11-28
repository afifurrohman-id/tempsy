package files

import (
	"context"
	"github.com/afifurrohman-id/tempsy/internal"
	"github.com/afifurrohman-id/tempsy/internal/auth"
	"github.com/afifurrohman-id/tempsy/internal/auth/guest"
	"github.com/afifurrohman-id/tempsy/internal/auth/oauth2"
	"github.com/afifurrohman-id/tempsy/internal/models"
	store "github.com/afifurrohman-id/tempsy/internal/storage"
	"github.com/gofiber/fiber/v2"
	"strings"
)

func HandleGetGuestToken(ctx *fiber.Ctx) error {
	if strings.Contains(ctx.Get(fiber.HeaderAuthorization), auth.BearerPrefix) {
		return ctx.Status(fiber.StatusBadRequest).JSON(&models.ApiError{
			Type:        internal.ErrorTypeHaveToken,
			Description: "You already have token",
		})
	}
	username := guest.GenerateUsername()

	token, err := guest.CreateToken(username)
	internal.Check(err)

	return ctx.JSON(&models.Token{
		AccessToken: token,
		TokenType:   strings.TrimSpace(auth.BearerPrefix),
		ExpiresIn:   604800, // 7 days in seconds
	})
}

func HandleGetUserInfo(ctx *fiber.Ctx) error {
	var (
		storeCtx = context.Background()
		userinfo = new(models.User)
	)
	token := strings.TrimPrefix(ctx.Get(fiber.HeaderAuthorization), auth.BearerPrefix)

	if claims, err := guest.ParseToken(token); err == nil {
		userinfo.UserName = claims["jti"].(string)
	} else {
		goUser, err := oauth2.GetGoogleAccountInfo(token)
		if err != nil {
			return ctx.Status(fiber.StatusBadRequest).JSON(&models.ApiError{
				Type:        internal.ErrorTypeInvalidToken,
				Description: "Token is not valid, Cannot get user info",
			})
		}
		userinfo.UserName = goUser.UserName
	}

	storeCtx, cancel := context.WithTimeout(storeCtx, timeoutCtx)
	defer cancel()

	files, err := store.GetAllObject(storeCtx, userinfo.UserName)
	internal.Check(err)

	userinfo.TotalFiles = len(files)

	return ctx.JSON(&userinfo)
}

package router

import (
	"context"
	"strings"

	"github.com/afifurrohman-id/tempsy/internal/files/auth"
	"github.com/afifurrohman-id/tempsy/internal/files/auth/guest"
	"github.com/afifurrohman-id/tempsy/internal/files/auth/oauth2"
	"github.com/afifurrohman-id/tempsy/internal/files/models"
	store "github.com/afifurrohman-id/tempsy/internal/files/storage"
	"github.com/afifurrohman-id/tempsy/internal/files/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
)

func HandleGetGuestToken(ctx *fiber.Ctx) error {
	trimAuth := strings.TrimPrefix(ctx.Get(fiber.HeaderAuthorization), auth.BearerPrefix)

	if _, err := guest.ParseToken(trimAuth); err == nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(&models.ApiError{
			Type:        utils.ErrorTypeHaveToken,
			Description: "You already have valid token",
		})
	}

	if _, err := oauth2.GetGoogleAccountInfo(trimAuth); err == nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(&models.ApiError{
			Type:        utils.ErrorTypeHaveToken,
			Description: "You already have valid token",
		})
	}

	token, err := guest.CreateToken(guest.GenerateUsername())
	utils.Check(err)

	return ctx.JSON(&models.GuestToken{
		AccessToken: token,
		ExpiresIn:   604800, // 7 days in seconds
	})
}

func HandleGetUserInfo(ctx *fiber.Ctx) error {
	userinfo := new(models.User)

	token := strings.TrimPrefix(ctx.Get(fiber.HeaderAuthorization), auth.BearerPrefix)

	if claims, err := guest.ParseToken(token); err == nil {
		userinfo.UserName = claims["jti"].(string)
	} else {
		log.Error(err)

		goUser, err := oauth2.GetGoogleAccountInfo(token)
		if err != nil {
			return ctx.Status(fiber.StatusBadRequest).JSON(&models.ApiError{
				Type:        utils.ErrorTypeInvalidToken,
				Description: "GuestToken is not valid, Cannot get user info",
			})
		}
		userinfo.UserName = goUser.UserName
	}

	storeCtx, cancel := context.WithTimeout(context.Background(), store.DefaultTimeoutCtx)
	defer cancel()

	files, err := store.GetAllObject(storeCtx, userinfo.UserName)
	utils.Check(err)

	userinfo.TotalFiles = len(files)

	return ctx.JSON(&userinfo)
}

package client

import (
	"errors"
	"fmt"
	"github.com/afifurrohman-id/files-sync-client/internal"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"golang.org/x/sync/errgroup"
	"os"
	"time"
)

// Six months in seconds
const maxAgeCookie = 6 * 30 * 24 * 60 * 60

func OAuth2Callback(ctx *fiber.Ctx) error {
	code := ctx.Query("code")

	oAuth2, err := internal.New()
	internal.Check(err)

	tokens, err := oAuth2.ExchangeCode(code)
	if err != nil {
		if errors.Is(err, internal.GOAuth2Error) {
			return ctx.Redirect("/auth/login")
		}
		log.Panic(err)
	}

	user, err := internal.GetGoogleAccountInfo(tokens.AccessToken)
	internal.Check(err)

	ctx.Cookie(&fiber.Cookie{
		Name:     "token",
		Value:    tokens.RefreshToken,
		Path:     "/",
		MaxAge:   maxAgeCookie,
		Secure:   os.Getenv("APP_ENV") == "production",
		HTTPOnly: os.Getenv("APP_ENV") == "production",
	})

	ctx.Cookie(&fiber.Cookie{
		Name:     "last_login",
		Value:    fmt.Sprintf("%d", time.Now().UnixMilli()),
		Path:     "/",
		MaxAge:   maxAgeCookie,
		Secure:   os.Getenv("APP_ENV") == "production",
		HTTPOnly: os.Getenv("APP_ENV") == "production",
	})

	return ctx.Redirect(fmt.Sprintf("/dashboard/%s", user.UserName))
}

// AuthLogin
// TODO: More validation
func AuthLogin(ctx *fiber.Ctx) error {
	if ctx.Cookies("token") != "" {
		return ctx.Redirect("/")
	}

	if ctx.Query("type", "oauth2") == "guest" {
		agent := fiber.Get(os.Getenv("API_SERVER_URI") + "/files/auth/guest/token")

		apiRes := new(internal.Token)
		statusCode, body, errs := agent.Struct(&apiRes)
		if len(errs) > 0 {
			log.Panic()
		}

		if statusCode != fiber.StatusOK {
			return ctx.Render("pages/error", map[string]any{
				"title":   fmt.Sprintf("Error - %d", statusCode),
				"message": string(body),
			})
		}

		ctx.Cookie(&fiber.Cookie{
			Name:     "token",
			Value:    apiRes.AccessToken,
			Path:     "/",
			MaxAge:   apiRes.ExpiresIn, // convert to seconds
			Secure:   os.Getenv("APP_ENV") == "production",
			HTTPOnly: os.Getenv("APP_ENV") == "production",
		})

		ctx.Cookie(&fiber.Cookie{
			Name:     "last_login",
			Value:    fmt.Sprintf("%d", time.Now().UnixMilli()),
			Path:     "/",
			MaxAge:   apiRes.ExpiresIn, // convert to seconds
			Secure:   os.Getenv("APP_ENV") == "production",
			HTTPOnly: os.Getenv("APP_ENV") == "production",
		})

		agent = fiber.Get(fmt.Sprintf("%s/files/auth/userinfo/me", os.Getenv("API_SERVER_URI")))
		agent.Set(fiber.HeaderAuthorization, apiRes.AccessToken)

		apiResUser := new(internal.User)
		statusCode, body, errs = agent.Struct(&apiResUser)
		if len(errs) > 0 {
			log.Panic(errs[0])
		}

		if statusCode != fiber.StatusOK {

			return ctx.Render("pages/error", map[string]any{
				"title":   fmt.Sprintf("Error - %d", statusCode),
				"message": string(body),
			})
		}

		return ctx.Redirect(fmt.Sprintf("/dashboard/%s", apiResUser.UserName))
	}

	oAuth2, err := internal.New()
	internal.Check(err)

	return ctx.Redirect(oAuth2.RedirectUrl())
}

func AuthLogout(ctx *fiber.Ctx) error {
	token := ctx.Cookies("token") // refresh token if oauth2 or access token if guest

	oAuth2, err := internal.New()
	internal.Check(err)

	tokens, err := oAuth2.AccessToken(token)
	if err != nil {
		if !errors.Is(err, internal.GOAuth2Error) {
			log.Panic(err)
		}
	} else {
		errorGroup := new(errgroup.Group)

		errorGroup.Go(func() error {
			return oAuth2.RevokeAccessToken(tokens.AccessToken)
		})

		if err = errorGroup.Wait(); err != nil {
			log.Error(err)
		}
	}

	ctx.Cookie(&fiber.Cookie{
		Name:     "token",
		Path:     "/",
		Expires:  time.Now().Add(-3 * time.Second), // 3 seconds ago, to delete cookie
		Secure:   os.Getenv("APP_ENV") == "production",
		HTTPOnly: os.Getenv("APP_ENV") == "production",
	})

	ctx.Cookie(&fiber.Cookie{
		Name:     "last_login",
		Path:     "/",
		Expires:  time.Now().Add(-3 * time.Second), // 3 seconds ago, to delete cookie
		Secure:   os.Getenv("APP_ENV") == "production",
		HTTPOnly: os.Getenv("APP_ENV") == "production",
	})

	return ctx.Redirect("/")
}

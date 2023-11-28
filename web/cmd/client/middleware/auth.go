package middleware

import (
	"errors"
	"fmt"
	"github.com/afifurrohman-id/files-sync-client/internal"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"os"
	"strings"
	"time"
)

func CheckAuth(ctx *fiber.Ctx) error {
	user := new(internal.User)
	token := ctx.Cookies("token") // refresh token if oauth2 or access token if guest

	o2, err := internal.New()
	internal.Check(err)

	tokens, err := o2.AccessToken(token)

	//TODO:Validate if user is authorized
	if err != nil {
		// try to get guest user
		if errors.Is(err, internal.GOAuth2Error) {
			agent := fiber.Get(fmt.Sprintf("%s/files/auth/userinfo/me", os.Getenv("API_SERVER_URI")))
			agent.Set(fiber.HeaderAuthorization, internal.BearerPrefix+token)
			apiRes := new(internal.User)

			agent.Timeout(10 * time.Second)
			statusCode, _, errs := agent.Struct(&apiRes)
			if len(errs) > 0 {
				log.Panic(errs[0])
			}

			if statusCode != fiber.StatusOK || !strings.HasPrefix(apiRes.UserName, internal.GuestUsernamePrefix) {
				if ctx.Path() == "/" {
					return ctx.Next()
				}
				return ctx.Redirect("/auth/login")
			}
			user = apiRes
			user.Picture = "https://placehold.co/96x96/png?text=a"
		} else {
			log.Panic(err)
		}
	} else {
		userInfo, err := internal.GetGoogleAccountInfo(tokens.AccessToken)
		internal.Check(err)

		if !userInfo.VerifiedEmail {
			if ctx.Path() == "/" {
				return ctx.Next()
			}
			return ctx.Redirect("/auth/login")
		}

		user = userInfo.User

		//change to access token
		token = tokens.AccessToken
	}
	ctx.Locals("user", user)
	ctx.Locals("token", token) //access token
	if ctx.Path() != "/" && user.UserName != ctx.Params("username") {
		return ctx.Redirect("/dashboard/" + user.UserName)
	}
	return ctx.Next()
}

func SetRealIpClient(ctx *fiber.Ctx) error {
	ctx.Request().Header.Set(internal.HeaderRealIp, ctx.IP())
	return ctx.Next()
}

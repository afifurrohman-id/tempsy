package client

import (
	"encoding/json"
	"fmt"
	"github.com/afifurrohman-id/files-sync-client/internal"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"os"
	"strconv"
	"strings"
	"time"
)

func HandleWelcomeClient(ctx *fiber.Ctx) error {
	return ctx.Render("pages/index", map[string]any{
		"user": ctx.Locals("user"),
	})
}

func HandleDashboardClient(ctx *fiber.Ctx) error {
	user, ok := ctx.Locals("user").(*internal.User)
	if !ok {
		log.Panic("invalid_user_struct")
	}

	agent := fiber.Get(fmt.Sprintf("%s/files/%s", os.Getenv("API_SERVER_URI"), ctx.Params("username")))
	agent.Set(fiber.HeaderAuthorization, internal.BearerPrefix+ctx.Locals("token").(string))

	apiRes := new([]*internal.DataFile)

	statusCode, body, errs := agent.Struct(&apiRes)
	if len(errs) > 0 {
		internal.Check(errs[0])
	}

	if statusCode != fiber.StatusOK {
		apiErr := new(internal.ApiError)
		internal.Check(json.Unmarshal(body, &apiErr))

		return ctx.Render("pages/error", map[string]any{
			"title":   fmt.Sprintf("Error - %d", statusCode),
			"message": apiErr.Description,
		})
	}

	return ctx.Render("pages/dashboard", map[string]any{
		"user":  user,
		"files": apiRes,
		"type":  "Upload",
	})
}

func HandleProfileClient(ctx *fiber.Ctx) error {
	user, ok := ctx.Locals("user").(*internal.User)
	if !ok {
		log.Panic("invalid_user_struct")
	}

	lastLoginMs, err := strconv.ParseInt(ctx.Cookies("last_login", fmt.Sprintf("%d", time.Now().UnixMilli())), 10, 64)
	internal.Check(err)

	agent := fiber.Get(fmt.Sprintf("%s/files/auth/userinfo/me", os.Getenv("API_SERVER_URI")))
	agent.Set(fiber.HeaderAuthorization, internal.BearerPrefix+ctx.Locals("token").(string))

	apiRes := new(internal.User)
	statusCode, body, errs := agent.Struct(&apiRes)
	if len(errs) > 0 {
		log.Panic(errs[0])
	}

	if statusCode != fiber.StatusOK {
		apiErr := new(internal.ApiError)
		internal.Check(json.Unmarshal(body, &apiErr))

		return ctx.Render("pages/error", map[string]any{
			"title":   fmt.Sprintf("Error - %d", statusCode),
			"message": apiErr.Description,
		})
	}

	return ctx.Render("pages/profile", map[string]any{
		"user":      user,
		"lastLogin": strings.Split(strings.SplitN(time.UnixMilli(lastLoginMs).Local().String(), "+", 2)[0], ".")[0],
		"total":     apiRes.TotalFiles,
	})
}

func HandleDetailDataClient(ctx *fiber.Ctx) error {
	user, ok := ctx.Locals("user").(*internal.User)
	if !ok {
		log.Panic("invalid_user_struct")
	}

	agent := fiber.Get(fmt.Sprintf("%s/files/%s/%s", os.Getenv("API_SERVER_URI"), user.UserName, ctx.Params("name")))
	agent.Set(fiber.HeaderAuthorization, internal.BearerPrefix+ctx.Locals("token").(string))

	apiRes := new(internal.DataFile)
	statusCode, body, errs := agent.Struct(&apiRes)
	if len(errs) > 0 {
		for _, err := range errs {
			internal.Check(err)
		}
	}

	if statusCode != fiber.StatusOK {
		apiErr := new(internal.ApiError)
		internal.Check(json.Unmarshal(body, &apiErr))

		return ctx.Render("pages/error", map[string]any{
			"title":   fmt.Sprintf("Error - %d", statusCode),
			"message": apiErr.Description,
		})
	}

	return ctx.Render("pages/details", map[string]any{
		"user": user,
		"file": struct {
			*internal.DataFile
			UploadedAt        string
			UpdatedAt         string
			AutoDeletedAt     string
			PrivateUrlExpires string
		}{
			DataFile:          apiRes,
			UploadedAt:        internal.FormatDate(apiRes.UploadedAt),
			UpdatedAt:         internal.FormatDate(apiRes.UpdatedAt),
			AutoDeletedAt:     internal.FormatDate(apiRes.AutoDeletedAt),
			PrivateUrlExpires: internal.FormatDate(time.Now().Add(time.Duration(apiRes.PrivateUrlExpires) * time.Second).UnixMilli()),
		},
		"type": "Update",
	})
}

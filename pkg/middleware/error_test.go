package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http/httptest"
	"testing"

	"github.com/afifurrohman-id/tempsy/internal/files/models"
	"github.com/afifurrohman-id/tempsy/internal/files/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestErrorServer(test *testing.T) {
	app := fiber.New(fiber.Config{
		ErrorHandler: CatchServerError,
		BodyLimit:    MaxBodyLimit,
	})

	app.Use(recover.New())

	test.Run("TestInternalServerError", func(test *testing.T) {
		app.Get("/error", func(ctx *fiber.Ctx) error {
			panic("test")
		})

		req := httptest.NewRequest(fiber.MethodGet, "/error", nil)

		res, err := app.Test(req, 1500*10) // 15 seconds
		require.NoError(test, err)

		test.Cleanup(func() {
			utils.LogErr(res.Body.Close())
		})

		require.Equal(test, fiber.StatusInternalServerError, res.StatusCode)
	})

	test.Run("TestNotFound", func(test *testing.T) {
		req := httptest.NewRequest(fiber.MethodGet, "/not-found", nil)

		res, err := app.Test(req, 1500*10) // 15 seconds
		require.NoError(test, err)

		test.Cleanup(func() {
			utils.LogErr(res.Body.Close())
		})

		body, err := io.ReadAll(res.Body)
		require.NoError(test, err)
		require.NotEmpty(test, body)

		apiErr := new(models.ApiError)
		require.NoError(test, json.Unmarshal(body, &apiErr))

		require.Equal(test, fiber.StatusNotFound, res.StatusCode)
		assert.Equal(test, "resource_not_found", apiErr.Error.Kind)
	})

	test.Run("TestBodyLimit", func(test *testing.T) {
		req := httptest.NewRequest(fiber.MethodPost, "/body-limit", bytes.NewReader(make([]byte, MaxBodyLimit+1)))

		res, err := app.Test(req, 1500*10) // 15 seconds

		require.Error(test, err)
		require.Nil(test, res)
	})
}

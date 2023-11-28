package middleware

import (
	"encoding/json"
	"github.com/afifurrohman-id/tempsy/internal"
	"github.com/afifurrohman-id/tempsy/internal/auth"
	"github.com/afifurrohman-id/tempsy/internal/auth/guest"
	"github.com/afifurrohman-id/tempsy/internal/auth/oauth2"
	"github.com/afifurrohman-id/tempsy/internal/models"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"
	"io"
	"net/http/httptest"
	"os"
	"testing"
)

func TestCheckHttpMethod(test *testing.T) {
	app := fiber.New()

	app.All("/", CheckHttpMethod, func(ctx *fiber.Ctx) error {
		return nil
	})

	test.Run("TestAllowedHttpMethod", func(test *testing.T) {
		for _, method := range auth.AllowedHttpMethod {
			req := httptest.NewRequest(method, "/", nil)
			res, err := app.Test(req, 1500*10) // 15 seconds
			require.NoError(test, err)

			test.Cleanup(func() {
				internal.LogErr(res.Body.Close())
			})

			require.Equal(test, fiber.StatusOK, res.StatusCode)
		}
	})

	test.Run("TestNotAllowedHttpMethod", func(test *testing.T) {
		req := httptest.NewRequest(fiber.MethodTrace, "/", nil)
		res, err := app.Test(req, 1500*10) // 15 seconds
		require.NoError(test, err)

		test.Cleanup(func() {
			internal.LogErr(res.Body.Close())
		})

		body, err := io.ReadAll(res.Body)
		require.NoError(test, err)
		require.NotEmpty(test, body)

		apiErr := new(models.ApiError)
		require.NoError(test, json.Unmarshal(body, &apiErr))

		require.Equal(test, fiber.StatusMethodNotAllowed, res.StatusCode)
		require.Equal(test, "method_not_allowed", apiErr.Type)
	})
}

func TestCheckAuth(test *testing.T) {
	app := fiber.New()

	app.Get("/:username", CheckAuth, func(ctx *fiber.Ctx) error {
		return nil
	})

	test.Run("TestGOAuth2", func(test *testing.T) {
		test.Run("TestOk", func(test *testing.T) {
			tokens, err := oauth2.GetAccessToken(os.Getenv("GOOGLE_OAUTH2_REFRESH_TOKEN_TEST"))
			require.NoError(test, err)

			userInfo, err := oauth2.GetGoogleAccountInfo(tokens.AccessToken)
			require.NoError(test, err)

			req := httptest.NewRequest(fiber.MethodGet, "/"+userInfo.UserName, nil)
			req.Header.Set(fiber.HeaderAuthorization, auth.BearerPrefix+tokens.AccessToken)
			res, err := app.Test(req, 1500*10) // 15 seconds
			require.NoError(test, err)

			test.Cleanup(func() {
				internal.LogErr(res.Body.Close())
			})

			require.Equal(test, fiber.StatusOK, res.StatusCode)
		})
		test.Run("TestUnauthorized", func(test *testing.T) {
			req := httptest.NewRequest(fiber.MethodGet, "/test", nil)
			res, err := app.Test(req, 1500*10) // 15 seconds
			require.NoError(test, err)

			test.Cleanup(func() {
				internal.LogErr(res.Body.Close())
			})

			require.Equal(test, fiber.StatusUnauthorized, res.StatusCode)
		})
	})

	test.Run("TestGuest", func(test *testing.T) {
		test.Run("TestOk", func(test *testing.T) {
			username := guest.GenerateUsername()

			token, err := guest.CreateToken(username)
			require.NoError(test, err)

			req := httptest.NewRequest(fiber.MethodGet, "/"+username, nil)
			req.Header.Set(fiber.HeaderAuthorization, auth.BearerPrefix+token)
			res, err := app.Test(req, 1500*10) // 15 seconds
			require.NoError(test, err)

			test.Cleanup(func() {
				internal.LogErr(res.Body.Close())
			})
		})
		test.Run("TestUnauthorized", func(test *testing.T) {
			req := httptest.NewRequest(fiber.MethodGet, "/test", nil)
			res, err := app.Test(req, 1500*10) // 15 seconds
			require.NoError(test, err)

			test.Cleanup(func() {
				internal.LogErr(res.Body.Close())
			})

			require.Equal(test, fiber.StatusUnauthorized, res.StatusCode)
		})
	})

}

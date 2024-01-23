package router

import (
	"encoding/json"
	"io"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/afifurrohman-id/tempsy/internal/files/auth"
	"github.com/afifurrohman-id/tempsy/internal/files/auth/guest"
	"github.com/afifurrohman-id/tempsy/internal/files/auth/oauth2"
	"github.com/afifurrohman-id/tempsy/internal/files/models"
	"github.com/afifurrohman-id/tempsy/internal/files/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetGuestToken(test *testing.T) {
	app := fiber.New()
	app.Get("/token", HandleGetGuestToken)

	test.Run("TestOk", func(test *testing.T) {
		req := httptest.NewRequest(fiber.MethodGet, "/token", nil)
		res, err := app.Test(req, 1500*10) // 15 seconds
		require.NoError(test, err)

		test.Cleanup(func() {
			utils.LogErr(res.Body.Close())
		})

		body, err := io.ReadAll(res.Body)
		require.NoError(test, err)

		apiRes := new(models.GuestToken)
		require.NoError(test, json.Unmarshal(body, &apiRes))

		assert.NotNil(test, apiRes)
		assert.NotEmpty(test, apiRes.AccessToken)
		assert.Greater(test, time.Now().Add(time.Duration(apiRes.ExpiresIn)*time.Second).UnixMilli(), time.Now().UnixMilli())
		assert.Equal(test, fiber.StatusOK, res.StatusCode)
	})

	test.Run("TestAlreadyValidHaveToken", func(test *testing.T) {
		req := httptest.NewRequest(fiber.MethodGet, "/token", nil)

		token, err := guest.CreateToken(guest.GenerateUsername())
		require.NoError(test, err)

		req.Header.Set(fiber.HeaderAuthorization, auth.BearerPrefix+token)
		res, err := app.Test(req, 1500*10) // 15 seconds
		require.NoError(test, err)

		test.Cleanup(func() {
			utils.LogErr(res.Body.Close())
		})

		body, err := io.ReadAll(res.Body)
		require.NoError(test, err)

		apiErr := new(models.ApiError)
		require.NoError(test, json.Unmarshal(body, &apiErr))

		assert.NotNil(test, apiErr)
		assert.Equal(test, fiber.StatusBadRequest, res.StatusCode)
		assert.Equal(test, utils.ErrorTypeHaveToken, apiErr.Error.Kind)
	})
}

func TestHandleGetUserInfo(test *testing.T) {
	var (
		app      = fiber.New()
		username = guest.GenerateUsername()
	)

	app.Get("/userinfo/me", HandleGetUserInfo)

	token, err := guest.CreateToken(username)
	require.NoError(test, err)

	tokens, err := oauth2.GetAccessToken(os.Getenv("GOOGLE_OAUTH2_REFRESH_TOKEN_TEST"))
	require.NoError(test, err)

	tablesOk := []struct {
		name  string
		token string
	}{
		{
			name:  "TestGuest",
			token: token,
		},
		{
			name:  "TestGOAuth2",
			token: tokens.AccessToken,
		},
	}

	for _, table := range tablesOk {
		test.Run(table.name, func(test *testing.T) {
			req := httptest.NewRequest(fiber.MethodGet, "/userinfo/me", nil)
			req.Header.Set(fiber.HeaderAuthorization, auth.BearerPrefix+table.token)
			res, err := app.Test(req, 1500*10) // 15 seconds
			require.NoError(test, err)

			test.Cleanup(func() {
				utils.LogErr(res.Body.Close())
			})

			body, err := io.ReadAll(res.Body)
			require.NoError(test, err)

			apiRes := new(models.User)
			require.NoError(test, json.Unmarshal(body, &apiRes))

			assert.NotNil(test, apiRes)
			if table.name != "TestGOAuth2" {
				assert.Equal(test, username, apiRes.UserName)
			}
			assert.Empty(test, apiRes.TotalFiles)
			assert.Equal(test, fiber.StatusOK, res.StatusCode)
		})
	}

	test.Run("TestInvalidToken", func(test *testing.T) {
		req := httptest.NewRequest(fiber.MethodGet, "/userinfo/me", nil)
		req.Header.Set(fiber.HeaderAuthorization, auth.BearerPrefix+"test")

		res, err := app.Test(req, 1500*10) // 15 seconds
		require.NoError(test, err)

		test.Cleanup(func() {
			utils.LogErr(res.Body.Close())
		})

		body, err := io.ReadAll(res.Body)
		require.NoError(test, err)

		apiErr := new(models.ApiError)
		require.NoError(test, json.Unmarshal(body, &apiErr))

		assert.NotNil(test, apiErr)
		assert.Equal(test, fiber.StatusBadRequest, res.StatusCode)
		assert.Equal(test, utils.ErrorTypeInvalidToken, apiErr.Error.Kind)
	})
}

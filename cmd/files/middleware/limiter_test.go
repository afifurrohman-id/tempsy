package middleware

import (
	"github.com/afifurrohman-id/tempsy/internal"
	"github.com/afifurrohman-id/tempsy/internal/auth"
	"github.com/afifurrohman-id/tempsy/internal/auth/guest"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http/httptest"
	"testing"
)

func TestLimitAuthTokenProcess(test *testing.T) {
	app := fiber.New()
	app.Get("/auth", RateLimiterProcessing, func(ctx *fiber.Ctx) error {
		return nil
	})

	test.Run("TestLimit", func(test *testing.T) {
		req := httptest.NewRequest(fiber.MethodGet, "/auth", nil)
		req.Header.Set(fiber.HeaderAuthorization, auth.BearerPrefix+"test")

		for i := 0; i <= MaxReqProcsPerSeconds; i++ {
			res, err := app.Test(req, 1500*10) // 15 seconds
			require.NoError(test, err)

			test.Cleanup(func() {
				internal.LogErr(res.Body.Close())
			})

			if i < MaxReqProcsPerSeconds {
				assert.Equal(test, fiber.StatusOK, res.StatusCode)
			} else {
				assert.Equal(test, fiber.StatusTooManyRequests, res.StatusCode)
			}
		}
	})

	test.Run("TestOnDifferent", func(test *testing.T) {
		req := httptest.NewRequest(fiber.MethodGet, "/auth", nil)

		for i := 0; i <= MaxReqProcsPerSeconds; i++ {
			req.Header.Set(fiber.HeaderAuthorization, auth.BearerPrefix+guest.GenerateUsername())
			res, err := app.Test(req, 1500*10) // 15 seconds
			require.NoError(test, err)

			test.Cleanup(func() {
				internal.LogErr(res.Body.Close())
			})

			assert.Equal(test, fiber.StatusOK, res.StatusCode)
		}
	})
}

func TestLimitGuestToken(test *testing.T) {
	app := fiber.New()
	app.Get("/guest", RateLimiterGuestToken, func(ctx *fiber.Ctx) error {
		return nil
	})

	test.Run("TestOnRealIp", func(test *testing.T) {
		req := httptest.NewRequest(fiber.MethodGet, "/guest", nil)

		for i := 0; i <= MaxReqGuestTokenPerSeconds; i++ {
			res, err := app.Test(req, 1500*10) // 15 seconds
			require.NoError(test, err)

			test.Cleanup(func() {
				internal.LogErr(res.Body.Close())
			})

			if i < MaxReqGuestTokenPerSeconds {
				assert.Equal(test, fiber.StatusOK, res.StatusCode)
			} else {
				assert.Equal(test, fiber.StatusTooManyRequests, res.StatusCode)
			}
		}
	})

	test.Run("TestOnProxy", func(test *testing.T) {
		testsTable := []struct {
			Name   string
			Header string
			Value  string
		}{
			{
				Name:   "TestOnXRealIp",
				Header: auth.HeaderXRealIp,
				Value:  "1.1.1.1",
			},
			{
				Name:   "TestOnRealIp",
				Header: auth.HeaderRealIp,
				Value:  "8.8.8.8",
			},
		}

		for _, table := range testsTable {
			test.Run(table.Name, func(test *testing.T) {
				req := httptest.NewRequest(fiber.MethodGet, "/guest", nil)
				req.Header.Set(table.Header, table.Value)

				for i := 0; i <= MaxReqGuestTokenPerSeconds; i++ {
					res, err := app.Test(req, 1500*10) // 15 seconds
					require.NoError(test, err)

					test.Cleanup(func() {
						internal.LogErr(res.Body.Close())
					})

					if i < MaxReqGuestTokenPerSeconds {
						assert.Equal(test, fiber.StatusOK, res.StatusCode)
					} else {
						assert.Equal(test, fiber.StatusTooManyRequests, res.StatusCode)
					}
				}
			})
		}

	})
}

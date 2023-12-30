package files

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/afifurrohman-id/tempsy/internal"
	"github.com/afifurrohman-id/tempsy/internal/models"
	store "github.com/afifurrohman-id/tempsy/internal/storage"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandleDelete(test *testing.T) {
	var (
		app      = fiber.New()
		storeCtx = context.Background()
		username = "test-handle-delete"
		fileByte = []byte(test.Name())
	)

	storeCtx, cancel := context.WithTimeout(storeCtx, store.DefaultTimeoutCtx)

	test.Cleanup(func() {
		defer cancel()

		dataFiles, err := store.GetAllObject(storeCtx, username)
		internal.Check(err)

		for _, dataFile := range dataFiles {
			internal.LogErr(store.DeleteObject(storeCtx, dataFile.Name))
		}
	})

	for i := 1; i <= 3; i++ {
		fileName := fmt.Sprintf("%s/example-%d.txt", username, i)

		require.NoError(test, store.UploadObject(storeCtx, fileName, fileByte, &models.DataFile{
			Name:              fileName,
			AutoDeletedAt:     time.Now().Add(1 * time.Minute).UnixMilli(),
			PrivateUrlExpires: 10, // 10 seconds
			IsPublic:          false,
			ContentType:       fiber.MIMETextPlainCharsetUTF8,
		}))
	}

	routeUsernameBase := app.Group("/api/files/:username")
	routeUsernameBase.Delete("/", HandleDeleteAllFile)
	routeUsernameBase.Delete("/:filename", HandleDeleteFile)

	test.Run("TestHandleDelete", func(test *testing.T) {
		test.Run("TestOk", func(test *testing.T) {
			req := httptest.NewRequest(fiber.MethodDelete, fmt.Sprintf("/api/files/%s/example-1.txt", username), nil)
			res, err := app.Test(req, 1500*10) // 15 seconds
			require.NoError(test, err)

			test.Cleanup(func() {
				internal.LogErr(res.Body.Close())
			})

			assert.Equal(test, fiber.StatusNoContent, res.StatusCode)
		})

		test.Run("TestNotFound", func(test *testing.T) {
			req := httptest.NewRequest(fiber.MethodDelete, fmt.Sprintf("/api/files/%s/example-1.json", username), nil)
			res, err := app.Test(req, 1500*10) // 15 seconds
			require.NoError(test, err)

			test.Cleanup(func() {
				internal.LogErr(res.Body.Close())
			})

			assert.Equal(test, fiber.StatusNotFound, res.StatusCode)
		})
	})

	test.Run("TestHandleDeleteAll", func(test *testing.T) {
		test.Run("TestOk", func(test *testing.T) {
			req := httptest.NewRequest(fiber.MethodDelete, "/api/files/"+username, nil)
			res, err := app.Test(req, 1500*10) // 15 seconds
			require.NoError(test, err)

			test.Cleanup(func() {
				internal.LogErr(res.Body.Close())
			})

			assert.Equal(test, fiber.StatusNoContent, res.StatusCode)
		})

		test.Run("TestOnEmptyData", func(test *testing.T) {
			req := httptest.NewRequest(fiber.MethodDelete, "/api/files/"+username, nil)
			res, err := app.Test(req, 1500*10) // 15 seconds
			require.NoError(test, err)

			test.Cleanup(func() {
				internal.LogErr(res.Body.Close())
			})

			apiErr := new(models.ApiError)

			body, err := io.ReadAll(res.Body)
			require.NoError(test, err)
			require.NotEmpty(test, body)

			require.NoError(test, json.Unmarshal(body, &apiErr))

			assert.Equal(test, fiber.StatusBadRequest, res.StatusCode)
			assert.Equal(test, internal.ErrorTypeEmptyData, apiErr.Type)
		})
	})
}

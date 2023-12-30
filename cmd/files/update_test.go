package files

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/afifurrohman-id/tempsy/internal"
	"github.com/afifurrohman-id/tempsy/internal/models"
	store "github.com/afifurrohman-id/tempsy/internal/storage"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandleUpdateFile(test *testing.T) {
	const username = "update-test"

	var (
		app      = fiber.New()
		storeCtx = context.Background()
		fileName = fmt.Sprintf("%s.txt", strings.ToLower(test.Name()))
		filePath = fmt.Sprintf("%s/%s", username, fileName)
		fileByte = []byte(test.Name())
	)
	storeCtx, cancel := context.WithTimeout(storeCtx, store.DefaultTimeoutCtx)

	test.Cleanup(func() {
		defer cancel()

		internal.Check(store.DeleteObject(storeCtx, filePath))
	})

	app.Put("/api/files/:username/:filename", HandleUpdateFile)

	require.NoError(test, store.UploadObject(storeCtx, filePath, fileByte, &models.DataFile{
		Name:              filePath,
		AutoDeletedAt:     time.Now().Add(1 * time.Minute).UnixMilli(),
		PrivateUrlExpires: 10, // 10 seconds
		IsPublic:          true,
		ContentType:       fiber.MIMETextPlainCharsetUTF8,
	}))

	test.Run("TestOk", func(test *testing.T) {
		req := httptest.NewRequest(fiber.MethodPut, "/api/files/"+filePath, bytes.NewReader(fileByte))
		req.Header.Set(fiber.HeaderContentType, fiber.MIMETextPlainCharsetUTF8)
		req.Header.Set(store.HeaderIsPublic, "-1")
		req.Header.Set(store.HeaderAutoDeletedAt, fmt.Sprintf("%d", time.Now().Add(3*time.Minute).UnixMilli()))
		req.Header.Set(store.HeaderPrivateUrlExpires, "10") // 10 seconds

		res, err := app.Test(req, 1500*10) // 15 seconds
		require.NoError(test, err)

		test.Cleanup(func() {
			internal.LogErr(res.Body.Close())
		})

		body, err := io.ReadAll(res.Body)
		require.NoError(test, err)

		apiRes := new(models.DataFile)
		require.NoError(test, json.Unmarshal(body, &apiRes))

		require.Equal(test, fiber.StatusOK, res.StatusCode)
		require.NotEmpty(test, apiRes)
		require.Equal(test, fiber.MIMETextPlainCharsetUTF8, apiRes.ContentType)
		require.Equal(test, fileName, apiRes.Name)
		require.Contains(test, apiRes.Url, username+"/public/"+fileName)
	})

	test.Run("TestOnDifferentFileNameHeader", func(test *testing.T) {
		req := httptest.NewRequest(fiber.MethodPut, "/api/files/"+filePath, bytes.NewReader(fileByte))
		req.Header.Set(fiber.HeaderContentType, fiber.MIMETextPlainCharsetUTF8)
		req.Header.Set(store.HeaderFileName, "different.txt")
		req.Header.Set(store.HeaderIsPublic, "0")
		req.Header.Set(store.HeaderAutoDeletedAt, fmt.Sprintf("%d", time.Now().Add(3*time.Minute).UnixMilli()))
		req.Header.Set(store.HeaderPrivateUrlExpires, "10") // 10 seconds

		res, err := app.Test(req, 1500*10) // 15 seconds
		require.NoError(test, err)

		test.Cleanup(func() {
			internal.LogErr(res.Body.Close())
		})

		body, err := io.ReadAll(res.Body)
		require.NoError(test, err)

		apiRes := new(models.DataFile)
		require.NoError(test, json.Unmarshal(body, &apiRes))

		require.Equal(test, fiber.StatusOK, res.StatusCode)
		require.NotEmpty(test, apiRes)
		assert.Equal(test, fiber.MIMETextPlainCharsetUTF8, apiRes.ContentType)
		assert.Equal(test, fileName, apiRes.Name)
		assert.NotContains(test, apiRes.Url, username+"/public/")
	})

	test.Run("TestOnFileNotFound", func(test *testing.T) {
		req := httptest.NewRequest(fiber.MethodPut, fmt.Sprintf("/api/files/%s/notfound.json", username), bytes.NewReader(fileByte))
		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSONCharsetUTF8)
		req.Header.Set(store.HeaderAutoDeletedAt, fmt.Sprintf("%d", time.Now().Add(3*time.Minute).UnixMilli()))
		req.Header.Set(store.HeaderPrivateUrlExpires, fmt.Sprintf("%d", 10)) // 10 seconds

		res, err := app.Test(req, 1500*10) // 15 seconds
		require.NoError(test, err)

		test.Cleanup(func() {
			internal.LogErr(res.Body.Close())
		})

		body, err := io.ReadAll(res.Body)
		require.NoError(test, err)

		apiRes := new(models.ApiError)
		require.NoError(test, json.Unmarshal(body, &apiRes))

		require.Equal(test, fiber.StatusNotFound, res.StatusCode)
		require.NotEmpty(test, apiRes)
		require.Equal(test, internal.ErrorTypeFileNotFound, apiRes.Type)
	})

	test.Run("TestOnEmptyFileUpdate", func(test *testing.T) {
		req := httptest.NewRequest(fiber.MethodPut, "/api/files/"+filePath, nil)
		req.Header.Set(fiber.HeaderContentType, fiber.MIMETextPlainCharsetUTF8)
		req.Header.Set(store.HeaderAutoDeletedAt, fmt.Sprintf("%d", time.Now().Add(3*time.Minute).UnixMilli()))
		req.Header.Set(store.HeaderPrivateUrlExpires, fmt.Sprintf("%d", 10)) // 10 seconds

		res, err := app.Test(req, 1500*10) // 15 seconds
		require.NoError(test, err)

		test.Cleanup(func() {
			internal.LogErr(res.Body.Close())
		})

		body, err := io.ReadAll(res.Body)
		require.NoError(test, err)

		apiRes := new(models.ApiError)
		require.NoError(test, json.Unmarshal(body, &apiRes))

		require.Equal(test, fiber.StatusBadRequest, res.StatusCode)
		require.NotEmpty(test, apiRes)
		require.Equal(test, internal.ErrorTypeEmptyFile, apiRes.Type)
	})

	test.Run("TestOnInvalidHeaderFile", func(test *testing.T) {
		req := httptest.NewRequest(fiber.MethodPut, "/api/files/"+filePath, bytes.NewReader([]byte("invalid")))
		req.Header.Set(fiber.HeaderContentType, fiber.MIMETextPlainCharsetUTF8)
		req.Header.Set(store.HeaderAutoDeletedAt, fmt.Sprintf("%d", time.Now().Add(3*time.Minute).UnixMilli()))
		req.Header.Set(store.HeaderPrivateUrlExpires, fmt.Sprintf("%d", 10)) // 10 seconds

		res, err := app.Test(req, 1500*10) // 15 seconds
		require.NoError(test, err)

		test.Cleanup(func() {
			internal.LogErr(res.Body.Close())
		})

		body, err := io.ReadAll(res.Body)
		require.NoError(test, err)

		apiRes := new(models.ApiError)
		require.NoError(test, json.Unmarshal(body, &apiRes))

		require.Equal(test, fiber.StatusUnprocessableEntity, res.StatusCode)
		require.NotEmpty(test, apiRes)
		require.Equal(test, internal.ErrorTypeInvalidHeaderFile, apiRes.Type)
	})

	test.Run("TestOnMismatchContentType", func(test *testing.T) {
		req := httptest.NewRequest(fiber.MethodPut, "/api/files/"+filePath, bytes.NewReader(fileByte))
		req.Header.Set(fiber.HeaderContentType, fiber.MIMETextXML)
		req.Header.Set(store.HeaderAutoDeletedAt, fmt.Sprintf("%d", time.Now().Add(3*time.Minute).UnixMilli()))
		req.Header.Set(store.HeaderPrivateUrlExpires, fmt.Sprintf("%d", 10)) // 10 seconds

		res, err := app.Test(req, 1500*10) // 15 seconds
		require.NoError(test, err)

		test.Cleanup(func() {
			internal.LogErr(res.Body.Close())
		})

		body, err := io.ReadAll(res.Body)
		require.NoError(test, err)

		apiRes := new(models.ApiError)
		require.NoError(test, json.Unmarshal(body, &apiRes))

		require.Equal(test, fiber.StatusBadRequest, res.StatusCode)
		require.NotEmpty(test, apiRes)
		require.Equal(test, internal.ErrorTypeMismatchType, apiRes.Type)
	})
}

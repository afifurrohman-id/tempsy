package files

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/afifurrohman-id/tempsy/internal"
	"github.com/afifurrohman-id/tempsy/internal/models"
	store "github.com/afifurrohman-id/tempsy/internal/storage"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestMapFileHeader(test *testing.T) {
	fileHeader := mapFileHeader(map[string][]string{
		fiber.HeaderContentType: {fiber.MIMEApplicationJSONCharsetUTF8},
	})
	require.NotEmpty(test, fileHeader)

	for key, value := range fileHeader {
		assert.Equal(test, fiber.HeaderContentType, key)
		assert.Equal(test, fiber.MIMEApplicationJSONCharsetUTF8, value)
	}
}

func TestHandleUploadFile(test *testing.T) {
	const username = "upload-test"

	var (
		app      = fiber.New()
		storeCtx = context.Background()
		fileName = fmt.Sprintf("%s.txt", strings.ToLower(test.Name()))
		fileByte = []byte(test.Name())
	)

	storeCtx, cancel := context.WithTimeout(storeCtx, store.DefaultTimeoutCtx)

	test.Cleanup(func() {
		defer cancel()

		internal.Check(store.DeleteObject(storeCtx, fmt.Sprintf("%s/%s", username, fileName)))
	})

	app.Post("/api/files/:username", HandleUploadFile)

	test.Run("TestOk", func(test *testing.T) {
		req := httptest.NewRequest(fiber.MethodPost, "/api/files/"+username, bytes.NewReader(fileByte))
		req.Header.Set(store.HeaderFileName, fileName)
		req.Header.Set(fiber.HeaderContentType, fiber.MIMETextPlainCharsetUTF8)
		req.Header.Set(store.HeaderIsPublic, "1")
		req.Header.Set(store.HeaderAutoDeletedAt, fmt.Sprintf("%d", time.Now().Add(3*time.Minute).UnixMilli()))
		req.Header.Set(store.HeaderPrivateUrlExpires, fmt.Sprintf("%d", 10)) // 10 seconds

		res, err := app.Test(req, 1500*10) // 15 seconds
		require.NoError(test, err)
		require.NotEmpty(test, res)
		test.Cleanup(func() {
			internal.LogErr(res.Body.Close())
		})

		body, err := io.ReadAll(res.Body)
		require.NoError(test, err)
		require.NotEmpty(test, body)

		apiRes := new(models.DataFile)
		require.NoError(test, json.Unmarshal(body, &apiRes))

		assert.Equal(test, fiber.StatusCreated, res.StatusCode)
		assert.NotEmpty(test, apiRes)
		assert.Equal(test, fiber.MIMETextPlainCharsetUTF8, apiRes.ContentType)
		assert.Equal(test, fileName, apiRes.Name)
		assert.Contains(test, apiRes.Url, fmt.Sprintf("%s/public/%s", username, fileName))
	})

	test.Run("TestOnFileAlreadyExists", func(test *testing.T) {
		req := httptest.NewRequest(fiber.MethodPost, "/api/files/"+username, bytes.NewReader(fileByte))
		req.Header.Set(store.HeaderFileName, fileName)
		req.Header.Set(fiber.HeaderContentType, fiber.MIMETextPlainCharsetUTF8)
		req.Header.Set(store.HeaderIsPublic, "1")
		req.Header.Set(store.HeaderAutoDeletedAt, fmt.Sprintf("%d", time.Now().Add(3*time.Minute).UnixMilli()))
		req.Header.Set(store.HeaderPrivateUrlExpires, fmt.Sprintf("%d", 10)) // 10 seconds

		res, err := app.Test(req, 1500*10) // 15 seconds
		require.NoError(test, err)
		require.NotEmpty(test, res)
		test.Cleanup(func() {
			internal.LogErr(res.Body.Close())
		})

		body, err := io.ReadAll(res.Body)
		require.NoError(test, err)
		require.NotEmpty(test, body)

		apiErr := new(models.ApiError)
		require.NoError(test, json.Unmarshal(body, &apiErr))

		assert.Equal(test, fiber.StatusConflict, res.StatusCode)
		assert.NotEmpty(test, apiErr)
		assert.Equal(test, internal.ErrorTypeFileExists, apiErr.Type)

	})

	test.Run("TestOnInvalidEmptyFile", func(test *testing.T) {
		req := httptest.NewRequest(fiber.MethodPost, "/api/files/"+username, nil)
		req.Header.Set(store.HeaderFileName, fileName)
		req.Header.Set(fiber.HeaderContentType, fiber.MIMETextPlainCharsetUTF8)
		req.Header.Set(store.HeaderIsPublic, "1")
		req.Header.Set(store.HeaderAutoDeletedAt, fmt.Sprintf("%d", time.Now().Add(3*time.Minute).UnixMilli()))
		req.Header.Set(store.HeaderPrivateUrlExpires, fmt.Sprintf("%d", 10)) // 10 seconds

		res, err := app.Test(req, 1500*10) // 15 seconds
		require.NoError(test, err)
		require.NotEmpty(test, res)
		test.Cleanup(func() {
			internal.LogErr(res.Body.Close())
		})

		body, err := io.ReadAll(res.Body)
		require.NoError(test, err)
		require.NotEmpty(test, body)

		apiRes := new(models.ApiError)
		require.NoError(test, json.Unmarshal(body, &apiRes))

		assert.Equal(test, fiber.StatusBadRequest, res.StatusCode)
		assert.NotEmpty(test, apiRes)
		assert.Equal(test, internal.ErrorTypeEmptyFile, apiRes.Type)
	})

	test.Run("TestInvalidFileName", func(test *testing.T) {
		req := httptest.NewRequest(fiber.MethodPost, "/api/files/"+username, bytes.NewReader(fileByte))
		req.Header.Set(store.HeaderFileName, "example")
		req.Header.Set(fiber.HeaderContentType, fiber.MIMETextPlainCharsetUTF8)
		req.Header.Set(store.HeaderIsPublic, "1")
		req.Header.Set(store.HeaderAutoDeletedAt, fmt.Sprintf("%d", time.Now().Add(3*time.Minute).UnixMilli()))
		req.Header.Set(store.HeaderPrivateUrlExpires, fmt.Sprintf("%d", 10)) // 10 seconds

		res, err := app.Test(req, 1500*10) // 15 seconds
		require.NoError(test, err)
		require.NotEmpty(test, res)
		test.Cleanup(func() {
			internal.LogErr(res.Body.Close())
		})

		body, err := io.ReadAll(res.Body)
		require.NoError(test, err)
		require.NotEmpty(test, body)

		apiRes := new(models.ApiError)
		require.NoError(test, json.Unmarshal(body, &apiRes))

		assert.Equal(test, fiber.StatusBadRequest, res.StatusCode)
		assert.NotEmpty(test, apiRes)
		assert.Equal(test, internal.ErrorTypeInvalidFileName, apiRes.Type)
	})

	test.Run("TestInvalidContentType", func(test *testing.T) {
		req := httptest.NewRequest(fiber.MethodPost, "/api/files/"+username, bytes.NewReader(fileByte))
		req.Header.Set(store.HeaderFileName, "1.json")
		req.Header.Set(fiber.HeaderContentType, fiber.MIMEOctetStream)
		req.Header.Set(store.HeaderIsPublic, "1")
		req.Header.Set(store.HeaderAutoDeletedAt, fmt.Sprintf("%d", time.Now().Add(3*time.Minute).UnixMilli()))
		req.Header.Set(store.HeaderPrivateUrlExpires, fmt.Sprintf("%d", 10)) // 10 seconds

		res, err := app.Test(req, 1500*10) // 15 seconds
		require.NoError(test, err)
		require.NotEmpty(test, res)
		test.Cleanup(func() {
			internal.LogErr(res.Body.Close())
		})

		body, err := io.ReadAll(res.Body)
		require.NoError(test, err)
		require.NotEmpty(test, body)

		apiRes := new(models.ApiError)
		require.NoError(test, json.Unmarshal(body, &apiRes))

		assert.Equal(test, fiber.StatusUnsupportedMediaType, res.StatusCode)
		assert.NotEmpty(test, apiRes)
		assert.Equal(test, internal.ErrorTypeUnsupportedType, apiRes.Type)
	})

	test.Run("TestInvalidHeaderFile", func(test *testing.T) {
		req := httptest.NewRequest(fiber.MethodPost, "/api/files/"+username, bytes.NewReader(fileByte))
		req.Header.Set(store.HeaderFileName, "test.json")
		req.Header.Set(fiber.HeaderContentType, fiber.MIMETextPlainCharsetUTF8)
		req.Header.Set(store.HeaderIsPublic, "1")
		req.Header.Set(store.HeaderPrivateUrlExpires, fmt.Sprintf("%d", 10)) // 10 seconds
		req.Header.Set(store.HeaderPrivateUrlExpires, fmt.Sprintf("%d", 10)) // 10 seconds

		res, err := app.Test(req, 1500*10) // 15 seconds
		require.NoError(test, err)
		require.NotEmpty(test, res)
		test.Cleanup(func() {
			internal.LogErr(res.Body.Close())
		})

		body, err := io.ReadAll(res.Body)
		require.NoError(test, err)
		require.NotEmpty(test, body)

		apiRes := new(models.ApiError)
		require.NoError(test, json.Unmarshal(body, &apiRes))

		assert.Equal(test, fiber.StatusUnprocessableEntity, res.StatusCode)
		assert.NotEmpty(test, apiRes)
		assert.Equal(test, internal.ErrorTypeInvalidHeaderFile, apiRes.Type)
	})
}

func TestValidateUrlExpires(test *testing.T) {
	test.Run("TestOk", func(test *testing.T) {
		assert.True(test, checkCompareUrlExpires(10, time.Now().Add(11*time.Second).UnixMilli()))
	})

	test.Run("TestInvalid", func(test *testing.T) {
		assert.False(test, checkCompareUrlExpires(10, time.Now().UnixMilli()))
	})

}

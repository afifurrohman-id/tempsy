package router

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

	"github.com/afifurrohman-id/tempsy/internal/files/models"
	store "github.com/afifurrohman-id/tempsy/internal/files/storage"
	"github.com/afifurrohman-id/tempsy/internal/files/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandleUpdateFile(test *testing.T) {
	const username = "update-test"

	var (
		app      = fiber.New()
		fileName = strings.ToLower(test.Name()) + ".txt"
		filePath = fmt.Sprintf("%s/%s", username, fileName)
		fileByte = []byte(test.Name())
	)
	storeCtx, cancel := context.WithTimeout(context.Background(), store.DefaultTimeoutCtx)

	test.Cleanup(func() {
		defer cancel()

		utils.Check(store.DeleteObject(storeCtx, filePath))
	})

	app.Put("/api/files/:username/:filename", HandleUpdateFile)

	require.NoError(test, store.UploadObject(storeCtx, filePath, fileByte, &models.DataFile{
		Name:              filePath,
		AutoDeleteAt:      time.Now().Add(1 * time.Minute).UnixMilli(),
		PrivateUrlExpires: 10, // 10 seconds
		IsPublic:          true,
		ContentType:       fiber.MIMETextPlainCharsetUTF8,
	}))

	test.Run("TestOk", func(test *testing.T) {
		req := httptest.NewRequest(fiber.MethodPut, "/api/files/"+filePath, bytes.NewReader(fileByte))
		req.Header.Set(fiber.HeaderContentType, fiber.MIMETextPlainCharsetUTF8)
		req.Header.Set(store.HeaderIsPublic, "-1")
		req.Header.Set(store.HeaderAutoDeleteAt, fmt.Sprintf("%d", time.Now().Add(3*time.Minute).UnixMilli()))
		req.Header.Set(store.HeaderPrivateUrlExpires, "10") // 10 seconds

		res, err := app.Test(req, 1500*10) // 15 seconds
		require.NoError(test, err)

		test.Cleanup(func() {
			utils.LogErr(res.Body.Close())
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
		req.Header.Set(store.HeaderAutoDeleteAt, fmt.Sprintf("%d", time.Now().Add(3*time.Minute).UnixMilli()))
		req.Header.Set(store.HeaderPrivateUrlExpires, "10") // 10 seconds

		res, err := app.Test(req, 1500*10) // 15 seconds
		require.NoError(test, err)

		test.Cleanup(func() {
			utils.LogErr(res.Body.Close())
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

	tableErrs := []struct {
		headers    map[string]string
		fileName   string
		name       string
		errType    string
		file       []byte
		statusCode int
	}{
		{
			name:     "TestOnFileNotFound",
			file:     fileByte,
			fileName: username + "/not-found.json",
			headers: map[string]string{
				fiber.HeaderContentType:       fiber.MIMEApplicationJSONCharsetUTF8,
				store.HeaderAutoDeleteAt:     fmt.Sprintf("%d", time.Now().Add(3*time.Minute).UnixMilli()),
				store.HeaderPrivateUrlExpires: "10", // 10 seconds
			},
			errType:    utils.ErrorTypeFileNotFound,
			statusCode: fiber.StatusNotFound,
		},
		{
			name:     "TestOnInvalidEmptyFile",
			file:     make([]byte, 0),
			fileName: filePath,
			headers: map[string]string{
				fiber.HeaderContentType:       fiber.MIMETextPlainCharsetUTF8,
				store.HeaderAutoDeleteAt:     fmt.Sprintf("%d", time.Now().Add(3*time.Minute).UnixMilli()),
				store.HeaderPrivateUrlExpires: "10", // 10 seconds
			},
			errType:    utils.ErrorTypeEmptyFile,
			statusCode: fiber.StatusBadRequest,
		},
		{
			name:     "TestOnInvalidHeaderFile",
			file:     fileByte,
			fileName: filePath,
			headers: map[string]string{
				fiber.HeaderContentType:       fiber.MIMETextPlainCharsetUTF8,
				store.HeaderAutoDeleteAt:     "test",
				store.HeaderPrivateUrlExpires: "10", // 10 seconds
			},
			errType:    utils.ErrorTypeInvalidHeaderFile,
			statusCode: fiber.StatusUnprocessableEntity,
		},
		{
			name:     "TestOnMismatchContentType",
			file:     fileByte,
			fileName: filePath,
			headers: map[string]string{
				fiber.HeaderContentType:       fiber.MIMETextXML,
				store.HeaderAutoDeleteAt:     fmt.Sprintf("%d", time.Now().Add(3*time.Minute).UnixMilli()),
				store.HeaderPrivateUrlExpires: "10", // 10 seconds
			},
			errType:    utils.ErrorTypeMismatchType,
			statusCode: fiber.StatusBadRequest,
		},
	}

	for _, table := range tableErrs {
		test.Run(table.name, func(test *testing.T) {
			req := httptest.NewRequest(fiber.MethodPut, "/api/files/"+table.fileName, bytes.NewReader(table.file))
			for key, val := range table.headers {
				req.Header.Set(key, val)
			}
			res, err := app.Test(req, 1500*10) // 15 seconds
			require.NoError(test, err)

			test.Cleanup(func() {
				utils.LogErr(res.Body.Close())
			})

			body, err := io.ReadAll(res.Body)
			require.NoError(test, err)

			apiRes := new(models.ApiError)
			require.NoError(test, json.Unmarshal(body, &apiRes))

			require.Equal(test, table.statusCode, res.StatusCode)
			require.NotEmpty(test, apiRes)
			require.Equal(test, table.errType, apiRes.Error.Kind)
		})
	}
}

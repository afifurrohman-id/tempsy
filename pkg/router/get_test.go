package router

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http/httptest"
	"path"
	"strings"
	"testing"
	"time"

	"github.com/afifurrohman-id/tempsy/internal/files/models"
	store "github.com/afifurrohman-id/tempsy/internal/files/storage"
	"github.com/afifurrohman-id/tempsy/internal/files/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	utils.LogErr(godotenv.Load(path.Join("..", "..", "configs", ".env")))
}

func TestHandleGetAllFileData(test *testing.T) {
	const (
		username   = "test-get-all"
		filesCount = 3
	)

	var (
		app      = fiber.New()
		fileByte = []byte(test.Name())
	)
	storeCtx, cancel := context.WithTimeout(context.Background(), store.DefaultTimeoutCtx)

	app.Get("/:username", HandleListFilesData)

	test.Cleanup(func() {
		defer cancel()

		dataFiles, err := store.ListObjects(storeCtx, username)
		utils.Check(err)

		for _, dataFile := range dataFiles {
			utils.LogErr(store.DeleteObject(storeCtx, dataFile.Name))
		}
	})

	for i := 1; i <= filesCount; i++ {
		filePath := fmt.Sprintf("%s/%s-%d.txt", username, strings.ToLower(test.Name()), i)

		require.NoError(test, store.UploadObject(storeCtx, filePath, fileByte, &models.DataFile{
			Name:              filePath,
			AutoDeleteAt:      time.Now().Add(1 * time.Minute).UnixMilli(),
			PrivateUrlExpires: 10, // 10 seconds
			IsPublic:          true,
			MimeType:       fiber.MIMETextPlainCharsetUTF8,
		}))
	}

	test.Run("TestOk", func(test *testing.T) {
		req := httptest.NewRequest(fiber.MethodGet, "/"+username, nil)

		res, err := app.Test(req, 1500*10) // 15 seconds
		require.NoError(test, err)

		test.Cleanup(func() {
			utils.LogErr(res.Body.Close())
		})

		body, err := io.ReadAll(res.Body)
		require.NoError(test, err)
		require.NotEmpty(test, body)

		apiRes := new([]*models.DataFile)
		require.NoError(test, json.Unmarshal(body, &apiRes))

		assert.Equal(test, fiber.StatusOK, res.StatusCode)
		require.NotNil(test, apiRes)
		assert.Equal(test, filesCount, len(*apiRes))
	})

	test.Run("TestNotFound", func(test *testing.T) {
		req := httptest.NewRequest(fiber.MethodGet, "/hello", nil)

		res, err := app.Test(req, 1500*10) // 15 seconds
		require.NoError(test, err)

		test.Cleanup(func() {
			utils.LogErr(res.Body.Close())
		})

		body, err := io.ReadAll(res.Body)
		require.NoError(test, err)

		result := new(any)
		require.NoError(test, json.Unmarshal(body, result))

		assert.NotNil(test, result)
		assert.Empty(test, result)
		assert.Equal(test, fiber.StatusOK, res.StatusCode)
	})
}

func TestHandleGetFileData(test *testing.T) {
	const username = "get-data"

	var (
		app      = fiber.New()
		fileName = strings.ToLower(test.Name()) + ".txt"
		filePath = fmt.Sprintf("%s/%s", username, fileName)
		fileByte = []byte(test.Name())
	)
	storeCtx, cancel := context.WithTimeout(context.Background(), store.DefaultTimeoutCtx)

	app.Get("/:username/:filename", HandleGetFileData)

	test.Cleanup(func() {
		defer cancel()

		utils.Check(store.DeleteObject(storeCtx, filePath))
	})

	require.NoError(test, store.UploadObject(storeCtx, filePath, fileByte, &models.DataFile{
		Name:              filePath,
		AutoDeleteAt:      time.Now().Add(1 * time.Minute).UnixMilli(),
		PrivateUrlExpires: 10, // 10 seconds
		IsPublic:          false,
		MimeType:       fiber.MIMETextPlainCharsetUTF8,
	}))

	test.Run("TestOk", func(test *testing.T) {
		req := httptest.NewRequest(fiber.MethodGet, "/"+filePath, nil)

		res, err := app.Test(req, 1500*10) // 15 seconds
		require.NoError(test, err)
		test.Cleanup(func() {
			utils.LogErr(res.Body.Close())
		})

		body, err := io.ReadAll(res.Body)
		require.NoError(test, err)
		require.NotEmpty(test, body)

		apiRes := new(models.DataFile)
		require.NoError(test, json.Unmarshal(body, &apiRes))

		assert.NotNil(test, apiRes)
		assert.Equal(test, fiber.StatusOK, res.StatusCode)
		assert.Equal(test, fiber.MIMETextPlainCharsetUTF8, apiRes.MimeType)
	})

	test.Run("TestNotFound", func(test *testing.T) {
		req := httptest.NewRequest(fiber.MethodGet, fmt.Sprintf("/%s/app.json", username), nil)

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

		assert.NotNil(test, apiErr)
		assert.Equal(test, utils.ErrorTypeFileNotFound, apiErr.Error.Kind)
		assert.Equal(test, fiber.StatusNotFound, res.StatusCode)
	})
}

func TestHandleGetPublicFile(test *testing.T) {
	const username = "public-get"

	var (
		app      = fiber.New()
		fileByte = []byte(test.Name())
	)

	app.Get("/:username/public/:filename", HandleGetPublicFile)

	storeCtx, cancel := context.WithTimeout(context.Background(), store.DefaultTimeoutCtx)

	test.Cleanup(func() {
		defer cancel()

		dataFiles, err := store.ListObjects(storeCtx, username)
		utils.Check(err)

		for _, dataFile := range dataFiles {
			utils.LogErr(store.DeleteObject(storeCtx, dataFile.Name))
		}
	})

	filesToUpload := []*models.DataFile{
		{
			Name:              username + "/app.txt",
			AutoDeleteAt:      time.Now().Add(1 * time.Minute).UnixMilli(),
			PrivateUrlExpires: 10, // 10 seconds
			IsPublic:          true,
			MimeType:       fiber.MIMETextPlainCharsetUTF8,
		},
		{
			Name:              fmt.Sprintf("%s/%s.txt", username, strings.ToLower(test.Name())),
			AutoDeleteAt:      time.Now().Add(1 * time.Minute).UnixMilli(),
			PrivateUrlExpires: 10, // 10 seconds
			IsPublic:          false,
			MimeType:       fiber.MIMETextPlainCharsetUTF8,
		},
	}

	for _, file := range filesToUpload {
		require.NoError(test, store.UploadObject(storeCtx, file.Name, fileByte, file))
	}

	test.Run("TestOk", func(test *testing.T) {
		req := httptest.NewRequest(fiber.MethodGet, fmt.Sprintf("/%s/public/%s", username, strings.Split(filesToUpload[0].Name, "/")[1]), nil)

		res, err := app.Test(req, 1500*10) // 15 seconds
		require.NoError(test, err)
		test.Cleanup(func() {
			utils.LogErr(res.Body.Close())
		})

		body, err := io.ReadAll(res.Body)
		require.NoError(test, err)
		require.NotEmpty(test, body)

		assert.NotEmpty(test, body)
		assert.Equal(test, fileByte, body)
		assert.Equal(test, fiber.MIMETextPlainCharsetUTF8, res.Header.Get(fiber.HeaderContentType))
		assert.Equal(test, fmt.Sprintf("%d", len(body)), res.Header.Get(fiber.HeaderContentLength))
		assert.Equal(test, fiber.StatusOK, res.StatusCode)
	})

	tableFails := []struct {
		name     string
		fileName string
	}{
		{
			name:     "TestNotFound",
			fileName: "not-found.json",
		},
		{
			name:     "TestIsNotPublic",
			fileName: strings.Split(filesToUpload[1].Name, "/")[1],
		},
	}

	for _, table := range tableFails {
		test.Run(table.name, func(test *testing.T) {
			req := httptest.NewRequest(fiber.MethodGet, fmt.Sprintf("/%s/public/%s", username, table.fileName), nil)

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

			assert.NotNil(test, apiErr)
			assert.Equal(test, fiber.StatusNotFound, res.StatusCode)
			assert.Equal(test, utils.ErrorTypeFileNotPublic, apiErr.Error.Kind)
		})
	}
}

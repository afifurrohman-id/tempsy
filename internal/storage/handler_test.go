package store

import (
	"cloud.google.com/go/storage"
	"context"
	"errors"
	"github.com/afifurrohman-id/tempsy/internal"
	"github.com/afifurrohman-id/tempsy/internal/models"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"strings"
	"testing"
	"time"
)

const timeoutCtx = 25 * time.Second

func TestCreateStorageClient(test *testing.T) {
	storeCtx := context.Background()
	storeCtx, cancel := context.WithTimeout(storeCtx, 8*time.Second)

	client, err := createClient(storeCtx)
	require.NoError(test, err)

	test.Cleanup(func() {
		defer cancel()

		internal.LogErr(client.Close())
	})

	assert.NotEmpty(test, client)
}

func TestGetAllObject(test *testing.T) {
	const username = "try"

	var (
		fileNames = []string{username + "/app.txt", username + "/test.txt"}
		storeCtx  = context.Background()
	)

	storeCtx, cancel := context.WithTimeout(storeCtx, timeoutCtx)

	test.Cleanup(func() {
		defer cancel()

		for _, fileName := range fileNames {
			internal.LogErr(DeleteObject(storeCtx, fileName))
		}
	})

	for _, fileName := range fileNames {
		require.NoError(test, UploadObject(storeCtx, fileName, []byte(strings.Split(fileName, "/")[0]), &models.DataFile{
			AutoDeletedAt:     time.Now().Add(2 * time.Minute).UnixMilli(),
			IsPublic:          true,
			PrivateUrlExpires: 30, // 30 seconds
			ContentType:       fiber.MIMETextPlainCharsetUTF8,
		}))
	}

	test.Run("TestOk", func(test *testing.T) {
		dataFiles, err := GetAllObject(storeCtx, username)
		require.NoError(test, err)
		assert.NotEmpty(test, dataFiles)
		assert.Len(test, dataFiles, len(fileNames))
	})

	test.Run("TestNotFound", func(test *testing.T) {
		dataFiles, err := GetAllObject(storeCtx, "not_found")
		require.NoError(test, err)
		assert.Empty(test, dataFiles)
	})
}

func TestGetObject(test *testing.T) {
	var (
		filePath = strings.ToLower(test.Name()) + "/ok.txt"
		objByte  = []byte("is ok")
	)

	storeCtx := context.Background()
	storeCtx, cancel := context.WithTimeout(storeCtx, timeoutCtx)

	test.Cleanup(func() {
		defer cancel()

		internal.LogErr(DeleteObject(storeCtx, filePath))
	})

	require.NoError(test, UploadObject(storeCtx, filePath, objByte, &models.DataFile{
		AutoDeletedAt:     time.Now().Add(2 * time.Minute).UnixMilli(),
		PrivateUrlExpires: 30, // 30 seconds
		ContentType:       fiber.MIMETextPlainCharsetUTF8,
	}))

	test.Run("TestOk", func(test *testing.T) {
		fileData, err := GetObject(storeCtx, filePath)
		require.NoError(test, err)
		require.NotEmpty(test, fileData)

		log.Info(fileData.Url)

		agent := fiber.Get(fileData.Url)

		statusCode, body, errs := agent.Bytes()
		require.Empty(test, errs)
		assert.Equal(test, fiber.StatusOK, statusCode)
		assert.NotEmpty(test, body)
		assert.Equal(test, objByte, body)

		Format(fileData)
		//TODO: Local time is different with server time
		//assert.Less(test, fileData.UploadedAt, time.Now().UnixMilli())
		assert.Equal(test, fileData.UpdatedAt, fileData.UploadedAt)
		assert.Greater(test, fileData.AutoDeletedAt, time.Now().UnixMilli())
		assert.Equal(test, fiber.MIMETextPlainCharsetUTF8, fileData.ContentType)
		assert.Less(test, time.Now().Add(time.Duration(fileData.PrivateUrlExpires)*time.Second).UnixMilli(), fileData.AutoDeletedAt)
		assert.Greater(test, time.Now().Add(time.Duration(fileData.PrivateUrlExpires)*time.Second).UnixMilli(), time.Now().UnixMilli())
	})

	test.Run("TestNotFound", func(test *testing.T) {
		dataFile, err := GetObject(storeCtx, "not_found.txt")
		assert.Error(test, err)
		assert.Empty(test, dataFile)
	})
}

func TestUploadObject(test *testing.T) {
	var (
		filePath = strings.ToLower(test.Name()) + "/up.txt"
		storeCtx = context.Background()
	)

	storeCtx, cancel := context.WithTimeout(storeCtx, timeoutCtx)

	test.Cleanup(func() {
		defer cancel()

		internal.LogErr(DeleteObject(storeCtx, filePath))
	})

	test.Run("TestOk", func(test *testing.T) {
		assert.NoError(test, UploadObject(storeCtx, filePath, []byte("is ok"), &models.DataFile{
			AutoDeletedAt:     time.Now().Add(2 * time.Minute).UnixMilli(),
			IsPublic:          true,
			PrivateUrlExpires: 30, // 30 seconds
			ContentType:       fiber.MIMEApplicationJSONCharsetUTF8,
		}))

	})

	test.Run("TestInvalidObjectPath", func(test *testing.T) {
		err := UploadObject(storeCtx, "invalid", []byte("hello"), &models.DataFile{
			AutoDeletedAt:     time.Now().Add(5 * time.Minute).UnixMilli(),
			IsPublic:          true,
			PrivateUrlExpires: 5, // 5 seconds
			ContentType:       fiber.MIMEApplicationJSONCharsetUTF8,
		})

		require.Error(test, err)
		assert.Contains(test, err.Error(), "file_path_must_be_in_format_username_and_slash_filename")
	})
}

func TestDeleteObject(test *testing.T) {

	var (
		filePath = strings.ToLower(test.Name()) + "/app.txt"
		storeCtx = context.Background()
	)

	storeCtx, cancel := context.WithTimeout(storeCtx, timeoutCtx)
	client, err := createClient(storeCtx)
	require.NoError(test, err)

	test.Cleanup(func() {
		defer cancel()

		internal.LogErr(client.Bucket(os.Getenv("GOOGLE_CLOUD_STORAGE_BUCKET")).Object(filePath).Delete(storeCtx))
		internal.LogErr(client.Close())
	})

	w := client.Bucket(os.Getenv("GOOGLE_CLOUD_STORAGE_BUCKET")).Object(filePath).NewWriter(storeCtx)

	_, err = w.Write([]byte(strings.Split(filePath, "/")[0]))
	require.NoError(test, err)

	require.NoError(test, w.Close())

	test.Run("TestOk", func(test *testing.T) {
		err := DeleteObject(storeCtx, filePath)
		assert.NoError(test, err)
	})

	test.Run("TestNotFound", func(test *testing.T) {
		err := DeleteObject(storeCtx, "not_found.txt")
		assert.Error(test, err)
		assert.True(test, errors.Is(err, storage.ErrObjectNotExist))
	})
}

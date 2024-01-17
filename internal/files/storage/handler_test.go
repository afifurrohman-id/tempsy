package store

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	"cloud.google.com/go/storage"
	"github.com/afifurrohman-id/tempsy/internal/files/models"
	"github.com/afifurrohman-id/tempsy/internal/files/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateStorageClient(test *testing.T) {
	storeCtx, cancel := context.WithTimeout(context.Background(), 8*time.Second)

	client, err := createClient(storeCtx)
	require.NoError(test, err)

	test.Cleanup(func() {
		defer cancel()

		utils.LogErr(client.Close())
	})

	assert.NotEmpty(test, client)
}

func TestGetAllObject(test *testing.T) {
	const username = "try"

	fileNames := []string{username + "/app.txt", username + "/test.txt"}

	storeCtx, cancel := context.WithTimeout(context.Background(), DefaultTimeoutCtx)

	test.Cleanup(func() {
		defer cancel()

		for _, fileName := range fileNames {
			utils.LogErr(DeleteObject(storeCtx, fileName))
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
		dataFiles, err := ListObjects(storeCtx, username)
		require.NoError(test, err)
		assert.NotEmpty(test, dataFiles)
		assert.Len(test, dataFiles, len(fileNames))
	})

	test.Run("TestNotFound", func(test *testing.T) {
		dataFiles, err := ListObjects(storeCtx, "not_found")
		require.NoError(test, err)
		assert.Empty(test, dataFiles)
	})
}

func TestGetObject(test *testing.T) {
	var (
		filePath = strings.ToLower(test.Name()) + "/ok.txt"
		objByte  = []byte("is ok")
	)

	storeCtx, cancel := context.WithTimeout(context.Background(), DefaultTimeoutCtx)

	test.Cleanup(func() {
		defer cancel()

		utils.LogErr(DeleteObject(storeCtx, filePath))
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
		assert.Less(test, fileData.UploadedAt, time.Now().UnixMilli())
		assert.Equal(test, fileData.UpdatedAt, fileData.UploadedAt)
		assert.Greater(test, fileData.AutoDeletedAt, time.Now().UnixMilli())
		assert.Equal(test, fiber.MIMETextPlainCharsetUTF8, fileData.ContentType)
		assert.Less(test, time.Now().Add(time.Duration(fileData.PrivateUrlExpires)*time.Second).UnixMilli(), fileData.AutoDeletedAt)
		assert.Greater(test, time.Now().Add(time.Duration(fileData.PrivateUrlExpires)*time.Second).UnixMilli(), time.Now().UnixMilli())
	})

	test.Run("TestNotFound", func(test *testing.T) {
		dataFile, err := GetObject(storeCtx, "not_found.txt")
		require.Error(test, err)
		assert.Empty(test, dataFile)
	})
}

func TestUploadObject(test *testing.T) {
	filePath := strings.ToLower(test.Name()) + "/up.txt"

	storeCtx, cancel := context.WithTimeout(context.Background(), DefaultTimeoutCtx)

	test.Cleanup(func() {
		defer cancel()

		utils.LogErr(DeleteObject(storeCtx, filePath))
	})

	test.Run("TestOk", func(test *testing.T) {
		require.NoError(test, UploadObject(storeCtx, filePath, []byte("is ok"), &models.DataFile{
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
		assert.Contains(test, err.Error(), "invalid_file_path")
	})
}

func TestDeleteObject(test *testing.T) {
	filePath := strings.ToLower(test.Name()) + "/app.txt"

	storeCtx, cancel := context.WithTimeout(context.Background(), DefaultTimeoutCtx)
	client, err := createClient(storeCtx)
	require.NoError(test, err)

	test.Cleanup(func() {
		defer cancel()

		utils.LogErr(client.Bucket(os.Getenv("GOOGLE_CLOUD_STORAGE_BUCKET")).Object(filePath).Delete(storeCtx))
		utils.LogErr(client.Close())
	})

	writer := client.Bucket(os.Getenv("GOOGLE_CLOUD_STORAGE_BUCKET")).Object(filePath).NewWriter(storeCtx)

	_, err = writer.Write([]byte(strings.Split(filePath, "/")[0]))
	require.NoError(test, err)

	require.NoError(test, writer.Close())

	test.Run("TestOk", func(test *testing.T) {
		err := DeleteObject(storeCtx, filePath)
		require.NoError(test, err)
	})

	test.Run("TestNotFound", func(test *testing.T) {
		err := DeleteObject(storeCtx, "not_found.txt")
		require.Error(test, err)
		assert.True(test, errors.Is(err, storage.ErrObjectNotExist))
	})
}

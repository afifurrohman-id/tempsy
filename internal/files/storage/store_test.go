package store

import (
	"fmt"
	"path"
	"strings"
	"testing"
	"time"

	"github.com/afifurrohman-id/tempsy/internal/files/models"
	"github.com/afifurrohman-id/tempsy/internal/files/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	utils.LogErr(godotenv.Load(path.Join("..", "..", "..", "configs", ".env")))
}

func TestUnmarshalMetadata(test *testing.T) {
	metadata := map[string]string{
		HeaderAutoDeleteAt:      fmt.Sprintf("%d", time.Now().Add(1*time.Minute).UnixMilli()),
		HeaderIsPublic:          "1",
		HeaderPrivateUrlExpires: "2", // 2 seconds
	}
	dataFile := new(models.DataFile)

	test.Run("TestOk", func(test *testing.T) {
		require.NoError(test, UnmarshalMetadata(metadata, dataFile))

		assert.NotEmpty(test, dataFile)
		assert.Greater(test, dataFile.AutoDeleteAt, time.Now().UnixMilli())
		assert.Less(test, time.Now().Add(time.Duration(dataFile.PrivateUrlExpires)*time.Second).UnixMilli(), dataFile.AutoDeleteAt)
		assert.Greater(test, time.Now().Add(time.Duration(dataFile.PrivateUrlExpires)*time.Second).UnixMilli(), time.Now().UnixMilli())
		assert.True(test, dataFile.IsPublic)
	})

	test.Run("TestInvalid", func(test *testing.T) {
		test.Run("TestInvalidAutoDeleteAt", func(test *testing.T) {
			metadata[HeaderAutoDeleteAt] = "invalid"
			require.Error(test, UnmarshalMetadata(metadata, dataFile))
		})

		test.Run("TestInvalidIsPublic", func(test *testing.T) {
			metadata[HeaderIsPublic] = "invalid"

			require.Error(test, UnmarshalMetadata(metadata, dataFile))
		})

		test.Run("TestInvalidPrivateUrlExpiredAt", func(test *testing.T) {
			metadata[HeaderPrivateUrlExpires] = "invalid"

			require.Error(test, UnmarshalMetadata(metadata, dataFile))
		})
	})
}

func TestFormat(test *testing.T) {
	dataFile := &models.DataFile{
		AutoDeleteAt:      time.Now().Add(1 * time.Minute).UnixMilli(),
		PrivateUrlExpires: 10, // 10 seconds
		MimeType:          fiber.MIMEApplicationJSONCharsetUTF8,
	}

	test.Run("TestFormatPrivate", func(test *testing.T) {
		dataFile.Name = "testing/example.json"
		dataFile.IsPublic = false
		dataFile.Url = "https://example.com/api/files/example.json"

		before := *dataFile

		Format(dataFile)

		assert.NotEqual(test, &before, dataFile)
		assert.NotContains(test, dataFile.Url, "/public/")
		assert.NotContains(test, dataFile.Name, "/")
	})

	test.Run("TestFormatPublic", func(test *testing.T) {
		dataFile.Name = "test/example.json"
		dataFile.IsPublic = true

		before := *dataFile

		Format(dataFile)

		assert.NotEqual(test, &before, dataFile)
		assert.Contains(test, dataFile.Url, "/public/")
		assert.NotContains(test, dataFile.Name, "/")
	})

	test.Run("TestFormatOnInvalidName", func(test *testing.T) {
		dataFile.Name = "test.txt"
		before := *dataFile
		Format(dataFile)

		assert.Equal(test, &before, dataFile)
	})
}

func TestMapFileHeader(test *testing.T) {
	tName := "unknown"

	fileHeader := MapFileHeader(map[string][]string{
		fiber.HeaderContentType: {fiber.MIMEApplicationJSONCharsetUTF8},
		HeaderFileName:          {tName},
		"UPPER":                 {"test"},
	})

	require.NotEmpty(test, fileHeader)

	for key, val := range fileHeader {
		assert.Equal(test, strings.ToLower(key), key)
		assert.Equal(test, val, fileHeader.Get(key))
		assert.Equal(test, val, fileHeader.Get(strings.ToUpper(key)))
	}
}

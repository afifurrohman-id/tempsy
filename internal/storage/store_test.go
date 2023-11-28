package store

import (
	"fmt"
	"github.com/afifurrohman-id/tempsy/internal"
	"github.com/afifurrohman-id/tempsy/internal/models"
	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"path"
	"testing"
	"time"
)

func init() {
	internal.LogErr(godotenv.Load(path.Join("..", "..", "deployments", ".env")))
}

func TestUnmarshalMetadata(test *testing.T) {
	metadata := map[string]string{
		HeaderAutoDeletedAt:     fmt.Sprintf("%d", time.Now().Add(1*time.Minute).UnixMilli()),
		HeaderIsPublic:          "1",
		HeaderPrivateUrlExpires: "10", // 10 seconds
	}
	dataFile := new(models.DataFile)

	test.Run("TestOk", func(test *testing.T) {
		require.NoError(test, UnmarshalMetadata(metadata, dataFile))

		assert.NotEmpty(test, dataFile)
		assert.Greater(test, dataFile.AutoDeletedAt, time.Now().UnixMilli())
		assert.Less(test, time.Now().Add(time.Duration(dataFile.PrivateUrlExpires)*time.Second).UnixMilli(), dataFile.AutoDeletedAt)
		assert.Greater(test, time.Now().Add(time.Duration(dataFile.PrivateUrlExpires)*time.Second).UnixMilli(), time.Now().UnixMilli())
		assert.True(test, dataFile.IsPublic)
	})

	test.Run("TestInvalid", func(test *testing.T) {
		test.Run("TestInvalidAutoDeletedAt", func(test *testing.T) {
			metadata[HeaderAutoDeletedAt] = "invalid"
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

		test.Run("TestOnPrivateUrlNotWithin7DaysFromNow", func(test *testing.T) {
			metadata[HeaderPrivateUrlExpires] = "604801" // 7 days + 1 seconds
			metadata[HeaderIsPublic] = "0"
			metadata[HeaderAutoDeletedAt] = fmt.Sprintf("%d", time.Now().Add(8*time.Hour).UnixMilli())

			err := UnmarshalMetadata(metadata, dataFile)
			require.Error(test, err)
			assert.Contains(test, err.Error(), "expired_url_should_be_within_7_day_from_now")
		})

	})

}

func TestFormat(test *testing.T) {
	dataFile := &models.DataFile{
		AutoDeletedAt:     time.Now().Add(1 * time.Minute).UnixMilli(),
		PrivateUrlExpires: 10, // 10 seconds
		ContentType:       fiber.MIMEApplicationJSONCharsetUTF8,
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
		dataFile.Name = fmt.Sprintf("test/example.json")
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

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
		fileName = fmt.Sprintf("%s.txt", strings.ToLower(test.Name()))
		fileByte = []byte(test.Name())
	)

	storeCtx, cancel := context.WithTimeout(context.Background(), store.DefaultTimeoutCtx)

	test.Cleanup(func() {
		defer cancel()

		utils.Check(store.DeleteObject(storeCtx, fmt.Sprintf("%s/%s", username, fileName)))
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
			utils.LogErr(res.Body.Close())
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

	tableErrs := []struct {
		headers    map[string]string
		name       string
		errType    string
		file       []byte
		statusCode int
	}{
		{
			name: "TestOnFileAlreadyExists",
			file: fileByte,
			headers: map[string]string{
				store.HeaderFileName:          fileName,
				fiber.HeaderContentType:       fiber.MIMETextPlainCharsetUTF8,
				store.HeaderIsPublic:          "1",
				store.HeaderAutoDeletedAt:     fmt.Sprintf("%d", time.Now().Add(3*time.Minute).UnixMilli()),
				store.HeaderPrivateUrlExpires: fmt.Sprintf("%d", 10), // 10 seconds
			},
			errType:    utils.ErrorTypeFileExists,
			statusCode: fiber.StatusConflict,
		},
		{
			name: "TestOnInvalidEmptyFile",
			file: make([]byte, 0),
			headers: map[string]string{
				store.HeaderFileName:          fileName,
				fiber.HeaderContentType:       fiber.MIMETextPlainCharsetUTF8,
				store.HeaderIsPublic:          "1",
				store.HeaderAutoDeletedAt:     fmt.Sprintf("%d", time.Now().Add(3*time.Minute).UnixMilli()),
				store.HeaderPrivateUrlExpires: fmt.Sprintf("%d", 10), // 10 seconds
			},
			errType:    utils.ErrorTypeEmptyFile,
			statusCode: fiber.StatusBadRequest,
		},
		{
			name: "TestOnInvalidFileName",
			file: fileByte,
			headers: map[string]string{
				store.HeaderFileName:          "example",
				fiber.HeaderContentType:       fiber.MIMETextPlainCharsetUTF8,
				store.HeaderIsPublic:          "1",
				store.HeaderAutoDeletedAt:     fmt.Sprintf("%d", time.Now().Add(3*time.Minute).UnixMilli()),
				store.HeaderPrivateUrlExpires: fmt.Sprintf("%d", 10), // 10 seconds
			},
			errType:    utils.ErrorTypeInvalidFileName,
			statusCode: fiber.StatusBadRequest,
		},
		{
			name: "TestOnInvalidContentType",
			file: fileByte,
			headers: map[string]string{
				store.HeaderFileName:          "1.json",
				fiber.HeaderContentType:       fiber.MIMEOctetStream,
				store.HeaderIsPublic:          "1",
				store.HeaderAutoDeletedAt:     fmt.Sprintf("%d", time.Now().Add(3*time.Minute).UnixMilli()),
				store.HeaderPrivateUrlExpires: fmt.Sprintf("%d", 10), // 10 seconds
			},
			errType:    utils.ErrorTypeUnsupportedType,
			statusCode: fiber.StatusUnsupportedMediaType,
		},
		{
			name: "TestInvalidHeaderFile",
			file: fileByte,
			headers: map[string]string{
				store.HeaderFileName:          "test.json",
				fiber.HeaderContentType:       fiber.MIMETextPlainCharsetUTF8,
				store.HeaderIsPublic:          "1",
				store.HeaderPrivateUrlExpires: fmt.Sprintf("%d", 10), // 10 seconds
			},
			errType:    utils.ErrorTypeInvalidHeaderFile,
			statusCode: fiber.StatusUnprocessableEntity,
		},
	}

	for _, tableE := range tableErrs {
		test.Run(tableE.name, func(test *testing.T) {
			req := httptest.NewRequest(fiber.MethodPost, "/api/files/"+username, bytes.NewReader(tableE.file))
			for key, value := range tableE.headers {
				req.Header.Set(key, value)
			}
			res, err := app.Test(req, 1500*10) // 15 seconds
			require.NoError(test, err)
			require.NotEmpty(test, res)

			test.Cleanup(func() {
				utils.LogErr(res.Body.Close())
			})

			body, err := io.ReadAll(res.Body)
			require.NoError(test, err)
			require.NotEmpty(test, body)

			apiRes := new(models.ApiError)
			require.NoError(test, json.Unmarshal(body, &apiRes))

			assert.Equal(test, tableE.statusCode, res.StatusCode)
			assert.NotEmpty(test, apiRes)
			assert.Equal(test, tableE.errType, apiRes.Type)
		})
	}
}

func TestValidateExpiry(test *testing.T) {
	tableTests := []struct {
		err     string
		name    string
		urlExp  uint
		autoDel int64
	}{
		{
			name:    "TestOkPrivateUrlNotLaterThanAutoDeletedAt",
			urlExp:  10,
			autoDel: time.Now().Add(1 * time.Minute).UnixMilli(),
		},
		{
			name:    "TestOKAutoDeletedAtNotLaterThan1YearFromNow",
			urlExp:  10,
			autoDel: time.Now().Add(8766 * time.Hour).UnixMilli(),
		},
		{
			name:    "TestOnPrivateUrlLaterThanAutoDeletedAt",
			urlExp:  10,
			autoDel: time.Now().Add(9 * time.Second).UnixMilli(),
			err:     "private_url_expires_cannot_be_later_than_auto_deleted_at_starting_from_now",
		},
		{
			name:    "TestOnAutoDeletedAtLaterThan1YearFromNow",
			urlExp:  10,
			autoDel: time.Now().Add(8767 * time.Hour).UnixMilli(),
			err:     "auto_deleted_at_cannot_be_later_than_1_year_from_now",
		},
		{
			name:    "TestOnPrivateUrlNotWithin7DaysFromNow",
			urlExp:  604801, // 7 days + 1 seconds (in seconds)
			autoDel: time.Now().Add(1000 * time.Hour).UnixMilli(),
			err:     "private_url_expires_must_be_less_than_7_days_in_seconds_and_more_than_2_seconds",
		},
		{
			name:    "TestOnPrivateUrlLessThan2Seconds",
			urlExp:  1,
			autoDel: time.Now().Add(1000 * time.Hour).UnixMilli(),
			err:     "private_url_expires_must_be_less_than_7_days_in_seconds_and_more_than_2_seconds",
		},
	}

	for _, tt := range tableTests {
		test.Run(tt.name, func(test *testing.T) {
			err := validateExpiry(tt.urlExp, tt.autoDel)
			if tt.err != "" {
				require.Error(test, err)
				assert.Equal(test, tt.err, err.Error())
			} else {
				assert.NoError(test, err)
			}
		})
	}
}

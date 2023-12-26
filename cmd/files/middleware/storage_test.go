package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/afifurrohman-id/tempsy/internal"
	"github.com/afifurrohman-id/tempsy/internal/auth/guest"
	"github.com/afifurrohman-id/tempsy/internal/models"
	store "github.com/afifurrohman-id/tempsy/internal/storage"
	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/require"
	"io"
	"net/http/httptest"
	"path"
	"strings"
	"testing"
	"time"
)

func init() {
	internal.LogErr(godotenv.Load(path.Join("..", "..", "..", "configs", ".env")))
}

func TestPurgeAnonymousAccount(test *testing.T) {
	var (
		app      = fiber.New()
		storeCtx = context.Background()
		byteFile = []byte(test.Name())
	)

	testsTables := []struct {
		name     string
		username string
		empty    bool
	}{
		{
			"TestOK",
			fmt.Sprintf("%s%d-%s", guest.UsernamePrefix, time.Now().Add(-1*time.Second).UnixMilli(), strings.ToLower(test.Name())),
			true,
		},
		{
			"TestUserNotAnonymous",
			"unknown",
			false,
		},
		{
			"TestNotDeletedNow",
			fmt.Sprintf("%s%d-%s", guest.UsernamePrefix, time.Now().Add(1*time.Minute).UnixMilli(), strings.ToLower(test.Name())),
			false,
		},
	}

	storeCtx, cancel := context.WithTimeout(storeCtx, 25*time.Second)
	test.Cleanup(func() {
		defer cancel()

		for _, table := range testsTables {
			dataFiles, err := store.GetAllObject(storeCtx, table.username)
			internal.Check(err)

			for _, dataFile := range dataFiles {
				internal.LogErr(store.DeleteObject(storeCtx, dataFile.Name))
			}

		}
	})

	app.Get("/purge/:username", PurgeAnonymousAccount, func(ctx *fiber.Ctx) error {
		files, err := store.GetAllObject(storeCtx, ctx.Params("username"))
		internal.Check(err)

		return ctx.JSON(&files)
	})

	for i, table := range testsTables {
		err := store.UploadObject(storeCtx, fmt.Sprintf("%s/%s-%d.txt", table.username, strings.ToLower(test.Name()), i), byteFile, &models.DataFile{
			AutoDeletedAt:     time.Now().Add(1 * time.Minute).UnixMilli(),
			PrivateUrlExpires: 25,
			IsPublic:          false,
			ContentType:       fiber.MIMETextPlainCharsetUTF8,
		})
		require.NoError(test, err)
	}

	for _, table := range testsTables {
		test.Run(table.name, func(test *testing.T) {
			req := httptest.NewRequest(fiber.MethodGet, fmt.Sprintf("/purge/%s", table.username), nil)

			res, err := app.Test(req, 1500*10) // 15 seconds
			require.NoError(test, err)

			test.Cleanup(func() {
				internal.LogErr(res.Body.Close())
			})

			body, err := io.ReadAll(res.Body)
			require.NoError(test, err)

			apiRes := new([]*models.DataFile)

			require.NoError(test, json.Unmarshal(body, &apiRes))

			require.Equal(test, fiber.StatusOK, res.StatusCode)

			if table.empty {
				require.Empty(test, apiRes)
			} else {
				require.NotEmpty(test, apiRes)
			}
		})
	}
}

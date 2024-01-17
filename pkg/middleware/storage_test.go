package middleware

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

	"github.com/afifurrohman-id/tempsy/internal/files/auth/guest"
	"github.com/afifurrohman-id/tempsy/internal/files/models"
	store "github.com/afifurrohman-id/tempsy/internal/files/storage"
	"github.com/afifurrohman-id/tempsy/internal/files/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/require"
)

func init() {
	utils.LogErr(godotenv.Load(path.Join("..", "..", "configs", ".env")))
}

func TestPurgeAnonymousAccount(test *testing.T) {
	var (
		app      = fiber.New()
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

	storeCtx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
	test.Cleanup(func() {
		defer cancel()

		for _, table := range testsTables {
			dataFiles, err := store.ListObjects(storeCtx, table.username)
			utils.Check(err)

			for _, dataFile := range dataFiles {
				utils.LogErr(store.DeleteObject(storeCtx, dataFile.Name))
			}

		}
	})

	app.Get("/purge/:username", PurgeAnonymousAccount, func(ctx *fiber.Ctx) error {
		files, err := store.ListObjects(storeCtx, ctx.Params("username"))
		utils.Check(err)

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
			req := httptest.NewRequest(fiber.MethodGet, "/purge/"+table.username, nil)

			res, err := app.Test(req, 1500*10) // 15 seconds
			require.NoError(test, err)

			test.Cleanup(func() {
				utils.LogErr(res.Body.Close())
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

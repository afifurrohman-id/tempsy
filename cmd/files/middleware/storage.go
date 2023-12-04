package middleware

import (
	"context"
	"github.com/afifurrohman-id/tempsy/internal"
	"github.com/afifurrohman-id/tempsy/internal/auth/guest"
	"github.com/afifurrohman-id/tempsy/internal/storage"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/gofiber/fiber/v2/middleware/cache"
	"strconv"
	"strings"
	"time"
)

// PurgeAnonymousAccount
// TODO: More efficient way to purge anonymous account
func PurgeAnonymousAccount(ctx *fiber.Ctx) error {
	var (
		username = ctx.Params("username")
		storeCtx = context.Background()
	)

	if strings.HasPrefix(username, guest.UsernamePrefix) {
		if lU := strings.SplitN(username, "-", 3); len(lU) > 2 {
			autoDeletedAccount, err := strconv.ParseInt(lU[1], 10, 64)
			if err == nil {
				if autoDeletedAccount < time.Now().UnixMilli() {
					timeout := 15 * time.Second
					storeCtx, cancel := context.WithTimeout(storeCtx, timeout)
					defer cancel()

					filesData, err := store.GetAllObject(storeCtx, username)
					if err != nil {
						log.Error(err)
						return ctx.Next()
					}

					for _, fileData := range filesData {
						internal.LogErr(store.DeleteObject(storeCtx, fileData.Name))
					}
				}
			} else {
				log.Error(err)
			}
		}
	}

	return ctx.Next()
}

func AutoDeleteScheduler(ctx *fiber.Ctx) error {
	var (
		username = ctx.Params("username")
		storeCtx = context.Background()
	)

	storeCtx, cancel := context.WithTimeout(storeCtx, 10*time.Second)
	defer cancel()

	filesData, err := store.GetAllObject(storeCtx, username)
	if err != nil {
		log.Error(err)
		return ctx.Next()
	}

	for _, fileData := range filesData {
		if fileData.AutoDeletedAt < time.Now().UnixMilli() {
			internal.LogErr(store.DeleteObject(storeCtx, fileData.Name))
		}
	}

	return ctx.Next()
}

var Cache = cache.New(cache.Config{
	Expiration:   10 * time.Second,
	CacheControl: true,
})

package middleware

import (
	"context"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/afifurrohman-id/tempsy/internal"
	"github.com/afifurrohman-id/tempsy/internal/auth/guest"
	store "github.com/afifurrohman-id/tempsy/internal/storage"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/gofiber/fiber/v2/middleware/cache"
)

func PurgeAnonymousAccount(ctx *fiber.Ctx) error {
	var (
		username = ctx.Params("username")
		storeCtx = context.Background()
	)

	if strings.HasPrefix(username, guest.UsernamePrefix) {
		if nameSplit := strings.SplitN(username, "-", 3); len(nameSplit) > 2 {
			autoDeletedAccount, err := strconv.ParseInt(nameSplit[1], 10, 64)
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

					var (
						eg = new(errgroup.Group)
						mu = new(sync.Mutex)
					)

					eg.Go(func() error {
						defer mu.Unlock()

						mu.Lock()
						for _, fileData := range filesData {
							if err = store.DeleteObject(storeCtx, fileData.Name); err != nil {
								return err
							}
						}
						return nil
					})

					internal.LogErr(eg.Wait())

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

	var (
		mu = new(sync.Mutex)
		eg = new(errgroup.Group)
	)

	eg.Go(func() error {
		defer mu.Unlock()

		mu.Lock()
		for _, fileData := range filesData {
			if fileData.AutoDeletedAt < time.Now().UnixMilli() {
				if err = store.DeleteObject(storeCtx, fileData.Name); err != nil {
					return err
				}
			}
		}
		return nil
	})

	internal.LogErr(eg.Wait())

	return ctx.Next()
}

var Cache = cache.New(cache.Config{
	Expiration:   10 * time.Second,
	CacheControl: true,
})

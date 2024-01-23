package middleware

import (
	"context"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/afifurrohman-id/tempsy/internal/files/auth/guest"
	store "github.com/afifurrohman-id/tempsy/internal/files/storage"
	"github.com/afifurrohman-id/tempsy/internal/files/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/gofiber/fiber/v2/middleware/cache"
)

func PurgeAnonymousAccount(ctx *fiber.Ctx) error {
	username := ctx.Params("username")

	if strings.HasPrefix(username, guest.UsernamePrefix) {
		if nameSplit := strings.SplitN(username, "-", 3); len(nameSplit) > 2 {
			autoDeletedAccount, err := strconv.ParseInt(nameSplit[1], 10, 64)
			if err == nil {
				if autoDeletedAccount < time.Now().UnixMilli() {
					timeout := 15 * time.Second
					storeCtx, cancel := context.WithTimeout(context.Background(), timeout)
					defer cancel()

					filesData, err := store.ListObjects(storeCtx, username)
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

					utils.LogErr(eg.Wait())

				}
			} else {
				log.Error(err)
			}
		}
	}

	return ctx.Next()
}

func AutoDeleteScheduler(ctx *fiber.Ctx) error {
	storeCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filesData, err := store.ListObjects(storeCtx, ctx.Params("username"))
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
			if fileData.AutoDeleteAt < time.Now().UnixMilli() {
				if err = store.DeleteObject(storeCtx, fileData.Name); err != nil {
					return err
				}
			}
		}
		return nil
	})

	utils.LogErr(eg.Wait())

	return ctx.Next()
}

var Cache = cache.New(cache.Config{
	Expiration:   10 * time.Second,
	CacheControl: true,
})

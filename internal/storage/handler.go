package store

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"cloud.google.com/go/storage"
	"github.com/afifurrohman-id/tempsy/internal"
	"github.com/afifurrohman-id/tempsy/internal/models"
	"github.com/gofiber/fiber/v2"
	"golang.org/x/sync/errgroup"
	"google.golang.org/api/iterator"
)

func GetAllObject(ctx context.Context, path string) ([]*models.DataFile, error) {
	client, err := createClient(ctx)
	if err != nil {
		return nil, err
	}
	defer internal.LogErr(client.Close())

	var (
		bucket  = client.Bucket(os.Getenv("GOOGLE_CLOUD_STORAGE_BUCKET"))
		objects = bucket.Objects(ctx, &storage.Query{Prefix: path})
	)

	var (
		eg          = new(errgroup.Group)
		mu          = new(sync.Mutex)
		dataFiles   = new([]*models.DataFile)
		objectNames = new([]string)
	)

	eg.Go(func() error {
		defer mu.Unlock()
		mu.Lock()
		for {
			obj, err := objects.Next()
			if err != nil {
				if errors.Is(err, iterator.Done) {
					break
				}
				return err
			}

			*objectNames = append(*objectNames, obj.Name)
		}

		for _, objectName := range *objectNames {
			dataFile, err := GetObject(ctx, objectName)
			if err != nil {
				return err
			}

			*dataFiles = append(*dataFiles, dataFile)
		}

		if len(*dataFiles) == 0 {
			*dataFiles = make([]*models.DataFile, 0)
		}

		return nil
	})

	if err = eg.Wait(); err != nil {
		return nil, err
	}

	return *dataFiles, nil
}

// GetObject
// return Name object will be in format `username/filename` as standard format in upload file
func GetObject(ctx context.Context, filePath string) (*models.DataFile, error) {
	client, err := createClient(ctx)
	if err != nil {
		return nil, err
	}

	attrs, err := client.Bucket(os.Getenv("GOOGLE_CLOUD_STORAGE_BUCKET")).Object(filePath).Attrs(ctx)
	if err != nil {
		return nil, err
	}
	fileData := &models.DataFile{
		Name:        attrs.Name,
		UploadedAt:  attrs.Created.UnixMilli(),
		UpdatedAt:   attrs.Updated.UnixMilli(),
		ContentType: attrs.ContentType,
		Size:        attrs.Size,
	}

	if err = UnmarshalMetadata(attrs.Metadata, fileData); err != nil {
		return nil, err
	}

	url, err := client.Bucket(os.Getenv("GOOGLE_CLOUD_STORAGE_BUCKET")).SignedURL(filePath, &storage.SignedURLOptions{
		Method:   fiber.MethodGet,
		Scheme:   storage.SigningSchemeV4,
		Expires:  time.Now().Add(time.Duration(fileData.PrivateUrlExpires) * time.Second),
		Insecure: os.Getenv("APP_ENV") != "production",
	})
	if err != nil {
		return nil, err
	}

	fileData.Url = url

	defer internal.LogErr(client.Close())
	return fileData, nil
}

// UploadObject filePath must be in format `username/filename`
func UploadObject(ctx context.Context, filePath string, fileByte []byte, fileData *models.DataFile) error {
	if !strings.Contains(filePath, "/") {
		return errors.New("file_path_must_be_in_format_username_and_slash_filename")
	}

	client, err := createClient(ctx)
	if err != nil {
		return err
	}
	defer internal.LogErr(client.Close())

	obj := client.Bucket(os.Getenv("GOOGLE_CLOUD_STORAGE_BUCKET")).Object(filePath)

	obj = obj.If(storage.Conditions{DoesNotExist: true})

	writer := obj.NewWriter(ctx)

	writer.Metadata = map[string]string{
		HeaderAutoDeletedAt:     fmt.Sprintf("%d", fileData.AutoDeletedAt),
		HeaderIsPublic:          fmt.Sprintf("%t", fileData.IsPublic),
		HeaderPrivateUrlExpires: fmt.Sprintf("%d", fileData.PrivateUrlExpires),
	}

	writer.ContentType = fileData.ContentType

	if _, err = writer.Write(fileByte); err != nil {
		return err
	}

	return writer.Close()
}

func DeleteObject(ctx context.Context, filePath string) error {
	client, err := createClient(ctx)
	if err != nil {
		return err
	}
	defer internal.LogErr(client.Close())

	obj := client.Bucket(os.Getenv("GOOGLE_CLOUD_STORAGE_BUCKET")).Object(filePath)

	attrs, err := obj.Attrs(ctx)
	if err != nil {
		return err
	}
	obj = obj.If(storage.Conditions{GenerationMatch: attrs.Generation})
	return obj.Delete(ctx)
}

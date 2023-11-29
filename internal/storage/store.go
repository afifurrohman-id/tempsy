package store

import (
	"cloud.google.com/go/storage"
	"context"
	"errors"
	"fmt"
	"github.com/afifurrohman-id/tempsy/internal/models"
	"github.com/gofiber/fiber/v2"
	"google.golang.org/api/option"
	"os"
	"strconv"
	"strings"
	"time"
)

var AcceptedContentType = []string{
	fiber.MIMEApplicationJSONCharsetUTF8, // not sure
	fiber.MIMEApplicationJSON,
	fiber.MIMETextHTMLCharsetUTF8, // not sure
	fiber.MIMETextHTML,
	fiber.MIMETextPlainCharsetUTF8, // not sure
	fiber.MIMETextPlain,
	fiber.MIMETextJavaScriptCharsetUTF8, // not sure
	fiber.MIMETextJavaScript,
	fiber.MIMEApplicationXMLCharsetUTF8, // not sure
	fiber.MIMEApplicationXML,            // Standard
	fiber.MIMETextXML,                   // Common Major browsers
	fiber.MIMETextXMLCharsetUTF8,        // not sure
	"text/csv",
	"text/css",
	"video/mpeg",
	"audio/mpeg",
	"application/epub+zip", // standard
	"application/epub",     // Chrome
	"image/gif",
	"image/jpeg",
	"application/pdf",
	"audio/wav",
	"audio/ogg",
	"image/png",
	"application/font-woff",    // Chrome
	"font/woff",                // standard
	"font/woff2",               // standard
	"application/x-compressed", // Chrome (7z, rar)
	"application/x-7z-compressed",

	// excel (xls, xlsx)
	"application/vnd.ms-excel",
	"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
	// word (doc, docx)
	"application/msword",
	"application/vnd.openxmlformats-officedocument.wordprocessingml.document",
	// powerpoint (ppt, pptx)
	"application/vnd.ms-powerpoint",
	"application/vnd.openxmlformats-officedocument.presentationml.presentation",

	"application/x-sh",
	"image/svg+xml",
	"application/x-tar",
	"application/x-gzip", // common Major browsers
	"application/gzip",   // standard
	"image/webp",
	"image/x-icon",             // Common Major browsers
	"image/vnd.microsoft.icon", // standard
	"image/avif",
	"application/wasm",
	"application/x-zip-compressed", // common Major browsers
	"application/zip",              // standard
}

const (
	HeaderAutoDeletedAt     = "File-Auto-Deleted-At"
	HeaderPrivateUrlExpires = "File-Private-Url-Expires"
	HeaderIsPublic          = "File-Is-Public"
	HeaderFileName          = "File-Name"
	DefaultTimeoutCtx       = 25 * time.Second
)

func createClient(ctx context.Context) (*storage.Client, error) {
	opt := []option.ClientOption{
		option.WithCredentialsJSON([]byte(os.Getenv("GOOGLE_CLOUD_STORAGE_SERVICE_ACCOUNT"))),
	}
	if os.Getenv("APP_ENV") != "production" {
		opt = append(opt, option.WithEndpoint(os.Getenv("GOOGLE_CLOUD_STORAGE_EMULATOR_ENDPOINT")))
	}

	return storage.NewClient(ctx, opt...)
}

func UnmarshalMetadata(metadata map[string]string, fileData *models.DataFile) error {
	autoDeletedAt, err := strconv.ParseInt(metadata[HeaderAutoDeletedAt], 10, 64)
	if err != nil {
		return err
	}

	privateUrlExpiredAt, err := strconv.Atoi(metadata[HeaderPrivateUrlExpires])
	if err != nil {
		return err
	}

	cutoff := time.Now().Add(604801 * time.Second) // 7 days + 1 second
	if !time.Now().Add(time.Duration(privateUrlExpiredAt) * time.Second).Before(cutoff) {
		return errors.New("expired_url_should_be_within_7_day_from_now")
	}

	// can be boolean string or number string 0 is false, otherwise true
	isPublic, err := strconv.ParseBool(metadata[HeaderIsPublic])
	if err != nil {
		boolInt, err := strconv.Atoi(metadata[HeaderIsPublic])
		if err != nil {
			return err
		}

		switch {
		case boolInt == 0:
			isPublic = false
		default:
			isPublic = true
		}

	}

	fileData.AutoDeletedAt = autoDeletedAt
	fileData.PrivateUrlExpires = privateUrlExpiredAt
	fileData.IsPublic = isPublic

	return nil
}

// Format change url and fileName
func Format(dataFile *models.DataFile) {
	split := strings.SplitN(dataFile.Name, "/", 2)
	if len(split) > 1 {
		userName := split[0]
		fileName := split[1]

		if dataFile.IsPublic {
			dataFile.Url = fmt.Sprintf("%s/files/%s/public/%s", os.Getenv("SERVER_URI"), userName, fileName)
		}

		dataFile.Name = fileName
	}
}

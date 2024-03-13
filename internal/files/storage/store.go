package store

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"cloud.google.com/go/storage"
	"github.com/afifurrohman-id/tempsy/internal/files/models"
	"github.com/gofiber/fiber/v2"
	"google.golang.org/api/option"
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

// Since HTTP 1.1 is case insensitive, but we follow HTTP 2.0 standard which is lowercase
const (
	HeaderAutoDeleteAt      = "file-auto-delete-at"
	HeaderPrivateUrlExpires = "file-private-url-expires"
	HeaderIsPublic          = "file-is-public"
	HeaderFileName          = "file-name"
	DefaultTimeoutCtx       = 25 * time.Second
)

func createClient(ctx context.Context) (*storage.Client, error) {
	serviceAccountByte, err := base64.StdEncoding.DecodeString(os.Getenv("GOOGLE_CLOUD_STORAGE_SERVICE_ACCOUNT"))
	if err != nil {
		return nil, err
	}
	opt := []option.ClientOption{
		option.WithCredentialsJSON(serviceAccountByte),
	}
	if os.Getenv("APP_ENV") != "production" {
		opt = append(opt, option.WithEndpoint(os.Getenv("GOOGLE_CLOUD_STORAGE_EMULATOR_ENDPOINT")))
	}

	return storage.NewClient(ctx, opt...)
}

// should guarantee metadata use lowercase as file header follow HTTP 2.0 standard
func UnmarshalMetadata(metadata map[string]string, fileData *models.DataFile) error {
	autoDeleteAt, err := strconv.ParseInt(metadata[HeaderAutoDeleteAt], 10, 64)
	if err != nil {
		return errors.New("auto_delete_at_must_be_valid_integer")
	}

	privateUrlInt64, err := strconv.ParseInt(metadata[HeaderPrivateUrlExpires], 10, 0)
	if err != nil {
		return errors.New("private_url_expires_must_be_valid_positive_integer")
	}

	// can be boolean string or number string 0 is false, otherwise true
	isPublic, err := strconv.ParseBool(metadata[HeaderIsPublic])
	if err != nil {
		boolInt, err := strconv.Atoi(metadata[HeaderIsPublic])
		if err != nil {
			return errors.New("is_public_must_be_valid_boolean_or_integer")
		}

		switch boolInt {
		case 0:
			isPublic = false
		default:
			isPublic = true
		}

	}

	fileData.AutoDeleteAt = autoDeleteAt
	fileData.PrivateUrlExpires = uint(privateUrlInt64)
	fileData.IsPublic = isPublic

	return nil
}

// Format change url and fileName
func Format(dataFile *models.DataFile) {
	split := strings.SplitN(dataFile.Name, "/", 2)
	if len(split) > 1 {
		fileName := split[1]

		if dataFile.IsPublic {
			dataFile.Url = fmt.Sprintf("%s/files/%s/public/%s", os.Getenv("SERVER_URL"), split[0], fileName)
		}

		dataFile.Name = fileName
	}
}

type FileHeader map[string]string

// map http header to file header
// return new map and key is lowercase
func MapFileHeader(header map[string][]string) FileHeader {
	fileHeader := make(map[string]string)
	for key, value := range header {
		fileHeader[strings.ToLower(key)] = value[0]
	}

	return fileHeader
}

// a safe way to get value
func (fh *FileHeader) Get(key string) string {
	return (*fh)[strings.ToLower(key)]
}

package internal

import (
	"github.com/gofiber/fiber/v2/log"
)

const (
	ErrorTypeFileNotPublic     = "file_not_found_or_not_public"
	ErrorTypeFileNotFound      = "file_not_found"
	ErrorTypeHaveToken         = "already_have_valid_token"
	ErrorTypeInvalidToken      = "invalid_token"
	ErrorTypeEmptyData         = "delete_empty_data"
	ErrorTypeInvalidHeaderFile = "invalid_header_file"
	ErrorTypeEmptyFile         = "invalid_empty_file"
	ErrorTypeMismatchType      = "mismatch_content_type"
	ErrorTypeFileExists        = "file_already_exists"
	ErrorTypeInvalidFileName   = "invalid_file_name"
	ErrorTypeUnsupportedType   = "unsupported_content_type"
)

// Check is a helper function to check error and panic if error is not nil
func Check(err error) {
	if err != nil {
		log.Panic(err)
	}
}

// LogErr is a helper function to log error if error is not nil
func LogErr(err error) {
	if err != nil {
		log.Error(err)
	}
}

package internal

import (
	"log"
	"os"
	"strings"
	"time"
)

func Check(err error) {
	if err != nil {
		log.New(os.Stderr, "ERROR ", log.LstdFlags|log.Lshortfile).Panic(err)
	}
}

// FormatDate date parameter is time in milliseconds
func FormatDate(date int64) string {
	dateWithDot := strings.Split(time.UnixMilli(date).Local().String(), "+")[0]
	return strings.Split(dateWithDot, ".")[0]
}

const (
	HeaderAutoDeletedAt     = "X-File-Auto-Deleted-At"
	HeaderPrivateUrlExpires = "X-File-Private-Url-Expires"
	HeaderIsPublic          = "X-File-Is-Public"
	HeaderFileName          = "X-File-Name"
	HeaderRealIp            = "X-Real-IP"
	BearerPrefix            = "Bearer "
)

package internal

import (
	"encoding/json"
	"errors"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"os"
	"strings"
	"time"
)

var GOAuth2Error = errors.New("oauth2_error_response_code_not_ok")

func New() (*GOAuth2Config, error) {
	oConfJSON := os.Getenv("GOOGLE_OAUTH2_CONFIG")

	oConf := new(GOAuth2Config)

	err := json.Unmarshal([]byte(oConfJSON), &oConf)
	if err != nil {
		return nil, err
	}

	return oConf, nil
}

func GetGoogleAccountInfo(accessToken string) (*GoogleAccountInfo, error) {
	const timeout = 10 * time.Second

	agent := fiber.Get("https://www.googleapis.com/userinfo/v2/me")

	agent.Set(fiber.HeaderAuthorization, BearerPrefix+accessToken)
	// TODO: Add parameter timeout
	agent.Timeout(timeout)

	userinfo := new(GoogleAccountInfo)

	statusCode, body, errs := agent.Struct(&userinfo)
	if len(errs) > 0 {
		return nil, errs[0]
	}

	if statusCode != fiber.StatusOK {
		log.Errorf("response_from_%d_body_%s", statusCode, string(body))
		return nil, GOAuth2Error
	}

	userinfo.UserName = strings.ReplaceAll(strings.Join(strings.SplitN(userinfo.Email, "@", 2), "-"), ".", "-")

	return userinfo, nil
}

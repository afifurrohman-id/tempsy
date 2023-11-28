package oauth2

import (
	"errors"
	"fmt"
	"github.com/afifurrohman-id/tempsy/internal/auth"
	"github.com/afifurrohman-id/tempsy/internal/models"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"os"
	"strings"
	"time"
)

var GOAuth2Error = errors.New("oauth2_error_response_code_not_ok")

func GetAccessToken(refreshToken string) (*models.GOAuth2Token, error) {
	payloadFormUri := fmt.Sprintf("client_secret=%s&grant_type=refresh_token&refresh_token=%s&client_id=%s", os.Getenv("GOOGLE_OAUTH2_CLIENT_SECRET_TEST"), refreshToken, os.Getenv("GOOGLE_OAUTH2_CLIENT_ID_TEST"))

	agent := fiber.Post("https://oauth2.googleapis.com/token")

	agent.Body([]byte(payloadFormUri))
	agent.Set(fiber.HeaderContentType, "application/x-www-form-urlencoded")

	oToken := new(models.GOAuth2Token)
	statusCode, body, errs := agent.Struct(&oToken)
	if len(errs) > 0 {
		return nil, errs[0]
	}

	if statusCode != fiber.StatusOK {
		log.Errorf("access_token_error_not_ok_status_code_%d_body_%s", statusCode, body)
		return nil, GOAuth2Error
	}

	return oToken, nil
}

func GetGoogleAccountInfo(accessToken string) (*models.GoogleAccountInfo, error) {
	const timeout = 10 * time.Second

	agent := fiber.Get("https://www.googleapis.com/userinfo/v2/me")

	agent.Set(fiber.HeaderAuthorization, auth.BearerPrefix+accessToken)
	// TODO: Add parameter timeout
	agent.Timeout(timeout)

	userinfo := new(models.GoogleAccountInfo)

	statusCode, body, errs := agent.Struct(&userinfo)
	if len(errs) > 0 {
		return nil, errs[0]
	}

	if statusCode != fiber.StatusOK {
		log.Errorf("response_from_%d_body_%s", statusCode, string(body))
		return nil, GOAuth2Error
	}

	userinfo.User = &models.User{
		UserName: strings.ReplaceAll(strings.Join(strings.SplitN(userinfo.Email, "@", 2), "-"), ".", "-"),
	}

	return userinfo, nil
}

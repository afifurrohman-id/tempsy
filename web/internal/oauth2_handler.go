package internal

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"strings"
)

// RedirectUrl returns the url for redirecting to consent screen, if the configuration is not valid, it will return empty string
func (gO2Conf *GOAuth2Config) RedirectUrl() string {

	switch {
	case gO2Conf.CallbackURL == "":
		return ""
	case len(gO2Conf.Scopes) == 0:
		return ""
	case gO2Conf.ClientID == "":
		return ""
	}

	return fmt.Sprintf("https://accounts.google.com/o/oauth2/v2/auth?redirect_uri=%s&prompt=consent&response_type=code&client_id=%s&scope=%s&access_type=offline", gO2Conf.CallbackURL, gO2Conf.ClientID, strings.Join(gO2Conf.Scopes, "+"))
}

func (gO2Conf *GOAuth2Config) ExchangeCode(code string) (*GOAuth2Token, error) {
	payloadFormUri := fmt.Sprintf("code=%s&redirect_uri=%s&client_id=%s&client_secret=%s&scope=&grant_type=authorization_code", code, gO2Conf.CallbackURL, gO2Conf.ClientID, gO2Conf.ClientSecret)

	agent := fiber.Post("https://oauth2.googleapis.com/token")
	agent.Body([]byte(payloadFormUri))
	agent.Set(fiber.HeaderContentType, "application/x-www-form-urlencoded")

	oToken := new(GOAuth2Token)
	statusCode, body, errs := agent.Struct(&oToken)
	if len(errs) > 0 {
		return nil, errs[0]
	}

	if statusCode != fiber.StatusOK {
		log.Errorf("exchange_code_error_not_ok_status_code_%d_body_%s", statusCode, body)
		return nil, GOAuth2Error
	}

	return oToken, nil
}
func (gO2Conf *GOAuth2Config) AccessToken(refreshToken string) (*GOAuth2Token, error) {

	payloadFormUri := fmt.Sprintf("client_secret=%s&grant_type=refresh_token&refresh_token=%s&client_id=%s", gO2Conf.ClientSecret, refreshToken, gO2Conf.ClientID)

	agent := fiber.Post("https://oauth2.googleapis.com/token")

	agent.Body([]byte(payloadFormUri))
	agent.Set(fiber.HeaderContentType, "application/x-www-form-urlencoded")

	oToken := new(GOAuth2Token)
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

func (gO2Conf *GOAuth2Config) RevokeAccessToken(accessToken string) error {
	payloadFormUri := fmt.Sprintf("token=%s", accessToken)

	agent := fiber.Post("https://oauth2.googleapis.com/revoke")

	agent.Body([]byte(payloadFormUri))
	agent.Set(fiber.HeaderContentType, "application/x-www-form-urlencoded")

	statusCode, body, errs := agent.String()

	if len(errs) > 0 {
		return errs[0]
	}

	if statusCode != fiber.StatusOK {
		log.Errorf("revoke_access_token_error_not_ok_status_code_%d_body_%s", statusCode, body)
		return GOAuth2Error
	}

	return nil
}

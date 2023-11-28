package guest

import (
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"math/rand"
	"os"
	"strings"
	"time"
)

const UsernamePrefix = "tempsyanonym-"

// GenerateUsername
// Format: tempsyanonym-<unix-milli-expired-in-7-days>-<random-string>
func GenerateUsername() string {
	lettersLower := "abcdefghijklmnopqrstuvwxyz0123456789"

	b := make([]rune, 18) //TODO: increase length as needed
	for i := range b {
		b[i] = rune(lettersLower[rand.Intn(len(lettersLower))])
	}

	return fmt.Sprintf("%s%d-%s", UsernamePrefix, time.Now().Add(168*time.Hour).UnixMilli(), string(b))
}

func CreateToken(username string) (string, error) {
	if !strings.HasPrefix(username, UsernamePrefix) {
		return "", errors.New("invalid_username_must_be_within_format")
	}
	jose := jwt.NewWithClaims(jwt.SigningMethodHS256, &jwt.RegisteredClaims{
		Subject:   "guest",
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(168 * time.Hour)), // 7 days
		ID:        username,
	})

	return jose.SignedString([]byte(os.Getenv("JWT_SECRET_KEY")))
}

func ParseToken(accessToken string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(accessToken, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected_signing_method_%v", token.Header["alg"])
		}
		return []byte(os.Getenv("JWT_SECRET_KEY")), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, err
}

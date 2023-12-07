package oauth2

import (
	"errors"
	"github.com/afifurrohman-id/tempsy/internal"
	"github.com/afifurrohman-id/tempsy/internal/auth"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"math/big"
	"os"
	"path"
	"strings"
	"testing"
	"time"
)

func init() {
	internal.LogErr(godotenv.Load(path.Join("..", "..", "..", "deployments", ".env")))
}

func TestOAuth2(test *testing.T) {
	test.Run("TestAccessToken", func(test *testing.T) {
		test.Run("TestOk", func(test *testing.T) {
			oToken, err := GetAccessToken(os.Getenv("GOOGLE_OAUTH2_REFRESH_TOKEN_TEST"))
			require.NoError(test, err)
			assert.NotEmpty(test, oToken)
			assert.NotEmpty(test, oToken.AccessToken)
			assert.Empty(test, oToken.RefreshToken)
			assert.Equal(test, oToken.TokenType, strings.TrimSpace(auth.BearerPrefix))
			assert.NotEmpty(test, oToken.IdToken)
			assert.Greater(test, time.Now().Add(time.Duration(oToken.ExpiresIn)*time.Second).UnixMilli(), time.Now().UnixMilli())
		})

		test.Run("TestOnInvalidRefreshToken", func(test *testing.T) {
			oToken, err := GetAccessToken("invalid")
			require.Error(test, err)

			assert.True(test, errors.Is(err, GOAuth2Error))
			assert.Nil(test, oToken)
		})
	})
}

func TestGetGoogleAccountInfo(test *testing.T) {
	tokens, err := GetAccessToken(os.Getenv("GOOGLE_OAUTH2_REFRESH_TOKEN_TEST"))
	require.NoError(test, err)

	test.Run("TestOk", func(test *testing.T) {
		accountInfo, err := GetGoogleAccountInfo(tokens.AccessToken)
		require.NoError(test, err)

		assert.NotEmpty(test, accountInfo)
		assert.NotEmpty(test, accountInfo.UserName)
		assert.NotContains(test, accountInfo.UserName, "@")
		assert.True(test, accountInfo.VerifiedEmail)

		numID, ok := new(big.Int).SetString(accountInfo.ID, 10)
		require.True(test, ok)
		assert.NotEmpty(test, numID)
		assert.NotEmpty(test, accountInfo.Picture)
	})

	test.Run("TestOnInvalidAccessToken", func(test *testing.T) {
		accountInfo, err := GetGoogleAccountInfo("invalid")
		require.Error(test, err)
		assert.Nil(test, accountInfo)
		assert.True(test, errors.Is(err, GOAuth2Error))
	})
}

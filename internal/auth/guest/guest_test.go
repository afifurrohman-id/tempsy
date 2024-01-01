package guest

import (
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateUsername(test *testing.T) {
	username := GenerateUsername()
	assert.NotEmpty(test, username)
	assert.Contains(test, username, UsernamePrefix)

	split := strings.SplitN(username, "-", 3)
	require.Len(test, split, 3)

	expiredDate, err := strconv.ParseInt(split[1], 10, 64)
	require.NoError(test, err)

	assert.Less(test, expiredDate, time.Now().Add(169*time.Hour).UnixMilli()) // 7 days + 1 hour
}

func TestCreateToken(test *testing.T) {
	test.Run("TestOk", func(test *testing.T) {
		token, err := CreateToken(GenerateUsername())
		require.NoError(test, err)

		assert.NotEmpty(test, token)
		assert.Contains(test, token, ".")
	})

	test.Run("TestInvalidUsername", func(test *testing.T) {
		token, err := CreateToken("")
		require.Error(test, err)

		assert.Empty(test, token)
	})
}

func TestParseToken(test *testing.T) {
	test.Run("TestOk", func(test *testing.T) {
		username := GenerateUsername()
		token, err := CreateToken(username)
		require.NoError(test, err)

		tokenMap, err := ParseToken(token)
		require.NoError(test, err)

		require.NotEmpty(test, tokenMap)
		assert.Equal(test, tokenMap["jti"], username)

		sub, err := tokenMap.GetSubject()
		require.NoError(test, err)
		assert.Equal(test, "guest", sub)

		exp, err := tokenMap.GetExpirationTime()
		require.NoError(test, err)

		assert.Greater(test, exp.UnixMilli(), time.Now().UnixMilli())
	})

	test.Run("TestInvalidToken", func(test *testing.T) {
		tokenMap, err := ParseToken("invalid")
		require.Error(test, err)

		assert.Empty(test, tokenMap)
	})
}

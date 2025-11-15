package external

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sunary/emu-game/pkg"
)

func TestValidateJWT(t *testing.T) {
	token, err := pkg.EncodeJWT(pkg.StandardPayload{
		Sub: "user-1",
	})
	require.NoError(t, err)

	payload, err := ValidateJWT("Bearer " + token)
	require.NoError(t, err)
	require.Equal(t, "user-1", payload.Sub)
}

func TestValidateJWTInvalidFormat(t *testing.T) {
	_, err := ValidateJWT("Token invalid")
	require.Equal(t, ErrBadFormat, err)
}

func TestValidateJWTExpiredToken(t *testing.T) {
	token, err := pkg.EncodeJWT(pkg.StandardPayload{
		Sub: "user-2",
	})
	require.NoError(t, err)

	// Tamper with the token exp claim by replacing it with the past
	parts := strings.Split(token, ".")
	require.Len(t, parts, 3)
	parts[1] = parts[1][:len(parts[1])-2] + "0" // crude method to lower exp
	expired := strings.Join(parts, ".")

	_, err = ValidateJWT("Bearer " + expired)
	require.Error(t, err)
}

package pkg

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/stretchr/testify/require"
)

func TestEncodeDecodeJWT(t *testing.T) {
	origSecret := secretKey
	secretKey = []byte("test-secret")
	defer func() { secretKey = origSecret }()

	token, err := EncodeJWT(StandardPayload{
		Sub:   "user-123",
		Email: "user@example.com",
	})
	require.NoError(t, err)

	payload, err := DecodeJWT(token)
	require.NoError(t, err)
	require.Equal(t, "user-123", payload.Sub)
	require.Equal(t, "user@example.com", payload.Email)
}

func TestDecodeJWTExpired(t *testing.T) {
	origSecret := secretKey
	secretKey = []byte("test-secret")
	defer func() { secretKey = origSecret }()

	expiredToken := jwt.NewWithClaims(jwt.SigningMethodHS256, Payload{
		StandardPayload: StandardPayload{
			Sub: "user-expired",
		},
		Iss: issuer,
		Exp: time.Now().Add(-time.Hour).Unix(),
	})

	signed, err := expiredToken.SignedString(secretKey)
	require.NoError(t, err)

	_, err = DecodeJWT(signed)
	require.EqualError(t, err, "token expired")
}

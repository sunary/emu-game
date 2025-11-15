package external

import (
	"errors"
	"fmt"
	"strings"

	"github.com/sunary/emu-game/pkg"
)

var (
	ErrEmptyToken = errors.New("authorization token is empty")
	ErrBadFormat  = errors.New("authorization header must be in Bearer format")
)

const bearerPrefix = "Bearer "

func ValidateJWT(authHeader string) (*pkg.Payload, error) {
	raw := strings.TrimSpace(authHeader)
	if raw == "" {
		return nil, ErrEmptyToken
	}

	if strings.HasPrefix(raw, bearerPrefix) {
		raw = strings.TrimSpace(raw[len(bearerPrefix):])
	} else {
		return nil, ErrBadFormat
	}

	payload, err := pkg.DecodeJWT(raw)
	if err != nil {
		return nil, fmt.Errorf("decode jwt: %w", err)
	}

	return payload, nil
}

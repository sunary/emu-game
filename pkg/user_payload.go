package pkg

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt"
)

var (
	secretKey = []byte("keep-it-secret")
)

const (
	issuer = "emu-user"
	expiry = time.Hour * 24
)

type StandardPayload struct {
	jwt.StandardClaims `json:",squash"`
	Email              string `json:"email"`
	Phone              string `json:"phone"`
	Name               string `json:"name"`
	Sub                string `json:"sub"`
}

type Payload struct {
	StandardPayload `json:",squash"`
	Iss             string   `json:"iss"`
	Exp             int64    `json:"exp"`
	Groups          []string `json:"groups"`
}

func EncodeJWT(payload StandardPayload) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Payload{
		StandardPayload: payload,
		Iss:             issuer,
		Exp:             time.Now().Add(expiry).Unix(),
	})
	tokenString, err := token.SignedString(secretKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func DecodeJWT(t string) (*Payload, error) {
	p := Payload{}
	_, err := jwt.ParseWithClaims(t, &p, func(token *jwt.Token) (interface{}, error) {
		return secretKey, nil
	})
	if err != nil {
		return nil, err
	}

	if p.Exp < time.Now().Unix() {
		return nil, errors.New("token expired")
	}

	return &p, nil
}

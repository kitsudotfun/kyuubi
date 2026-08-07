package api

import (
	"errors"
	"io"

	"github.com/golang-jwt/jwt/v5"
	"github.com/syumai/workers/cloudflare/kv"
)

const (
	JwtKeyNamespace = "KITSU_JWT_KEYS"
)

var (
	jwtMethod = jwt.SigningMethodHS256
)

var (
	ErrInvalidToken = errors.New("invalid token")
)

func MustGetJwtKey(id string) []byte {
	keys, err := kv.NewNamespace(JwtKeyNamespace)
	if err != nil {
		panic(err)
	}
	r, err := keys.GetReader(id, nil)
	if err != nil {
		panic(err)
	}
	key, err := io.ReadAll(r)
	if err != nil {
		panic(err)
	}

	return key
}

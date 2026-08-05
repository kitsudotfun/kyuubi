package main

import (
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

func GetJwtKey(id string) ([]byte, error) {
	keys, err := kv.NewNamespace(JwtKeyNamespace)
	if err != nil {
		return nil, err
	}
	r, err := keys.GetReader(id, nil)
	if err != nil {
		return nil, err
	}
	key, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	return key, nil
}

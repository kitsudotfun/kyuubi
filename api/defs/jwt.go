package defs

import (
	"errors"

	"github.com/golang-jwt/jwt/v5"
)

const (
	JwtKeyNamespace = "KITSU_JWT_KEYS"
)

var (
	JwtMethod = jwt.SigningMethodHS256
)

var (
	ErrInvalidToken = errors.New("invalid token")
)

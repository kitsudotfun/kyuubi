package defs

import (
	"net/netip"

	"github.com/golang-jwt/jwt/v5"
)

const (
	NatNegServer = "51.81.22.70:62426"
)

type NatNegClaims struct {
	jwt.RegisteredClaims
	Session `json:"session"`

	Addr netip.AddrPort `json:"addr"`
}

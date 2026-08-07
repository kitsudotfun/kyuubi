package api

import (
	"net/netip"

	"github.com/golang-jwt/jwt/v5"
)

const (
	NatNegServer = "natneg.kitsu.fun:62426"
)

type NatNegClaims struct {
	jwt.RegisteredClaims

	Addr netip.AddrPort `json:"addr"`
}

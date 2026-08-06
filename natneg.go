package main

import (
	"net/netip"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	NatNegServer = "natneg.kitsu.fun:62426"
)

type NatNegTokenRequest struct{}
type NatNegTokenResponse struct {
	Token  string `json:"token"`
	Server string `json:"server"`
}

func NatNegToken(_ NatNegTokenRequest, s Session) (NatNegTokenResponse, error) {
	token, err := jwt.NewWithClaims(jwtMethod, jwt.RegisteredClaims{
		Subject:   s.ID.String(),
		Audience:  jwt.ClaimStrings{"natneg_discover"},
		ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(time.Minute)),
	}).SignedString(MustGetJwtKey("natneg"))
	if err != nil {
		return NatNegTokenResponse{}, err
	}

	return NatNegTokenResponse{
		Token:  token,
		Server: NatNegServer,
	}, nil
}

type NatNegVerifyRequest struct {
	Token string `json:"token"`
}
type NatNegVerifyResponse struct {
	Addr netip.AddrPort `json:"addr"`
}

type NatNegClaims struct {
	jwt.RegisteredClaims

	Addr netip.AddrPort `json:"addr"`
}

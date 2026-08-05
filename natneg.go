package main

import (
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

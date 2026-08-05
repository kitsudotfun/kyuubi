package main

import (
	"errors"
	"net/netip"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/syumai/workers/cloudflare/kv"
)

const (
	NatNegServer = "natneg.kitsu.fun:62426"
)

var (
	ErrGameAddrSet = errors.New("game address already set")
)

type NatNegClaims struct {
	jwt.RegisteredClaims

	Addr netip.AddrPort
}

type NatNegNewRequest struct{}
type NatNegNewResponse struct {
	Token  string `json:"token"`
	Server string `json:"server"`
}

func NatNegNew(_ NatNegNewRequest, s Session) (NatNegNewResponse, error) {
	if s.GameAddr.IsValid() {
		return NatNegNewResponse{}, ErrGameAddrSet
	}

	token, err := jwt.NewWithClaims(jwtMethod, jwt.RegisteredClaims{
		Subject:   s.ID.String(),
		ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(time.Minute)),
	}).SignedString(MustGetJwtKey("natneg"))
	if err != nil {
		return NatNegNewResponse{}, err
	}

	return NatNegNewResponse{
		Token:  token,
		Server: NatNegServer,
	}, nil
}

type NatNegVerifyRequest struct {
	Token string `json:"token"`
}
type NatNegVerifyResponse struct {
	Addr netip.Addr `json:"addr"`
}

func NatNegVerify(req NatNegVerifyRequest, s Session) (NatNegVerifyResponse, error) {
	if s.GameAddr.IsValid() {
		return NatNegVerifyResponse{}, ErrGameAddrSet
	}

	var claims NatNegClaims
	token, err := jwt.ParseWithClaims(req.Token, &claims, func(token *jwt.Token) (any, error) {
		return MustGetJwtKey("natneg"), nil
	})
	if err != nil {
		return NatNegVerifyResponse{}, err
	}
	if !token.Valid {
		return NatNegVerifyResponse{}, ErrInvalidToken
	}
	if !claims.Addr.IsValid() {
		return NatNegVerifyResponse{}, ErrInvalidToken
	}

	s.GameAddr = claims.Addr

	err = PutEncodedKV(s.ID.String(), SessionNamespace, s, &kv.PutOptions{ExpirationTTL: 60 * 60 * 24})
	if err != nil {
		return NatNegVerifyResponse{}, err
	}

	return NatNegVerifyResponse{Addr: s.GameAddr.Addr()}, nil
}

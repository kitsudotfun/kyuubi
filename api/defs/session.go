package defs

import (
	"encoding/base64"
	"errors"

	"github.com/golang-jwt/jwt/v5"
)

const (
	SessionIdLen = 8

	ProofDifficulty = 24 // bits
	ProofSaltLen    = 16
)

type SessionID [SessionIdLen]byte

func (sid SessionID) String() string {
	return base64.RawStdEncoding.EncodeToString(sid[:])
}

func (sid *SessionID) FromString(s string) error {
	b, err := base64.RawStdEncoding.DecodeString(s)
	if err != nil {
		return err
	}

	copy(sid[:], b)

	return nil
}

type Session struct {
	ID     SessionID
	GameID string
}

type SessionClaims struct {
	jwt.RegisteredClaims

	GameID string `json:"game"`

	Difficulty int                `json:"difficulty"`
	Salt       [ProofSaltLen]byte `json:"salt"`
}

var (
	ErrSessionExists = errors.New("session exists")
)

// /session/new
type SessionNewRequest struct {
	GameID string `json:"game"`
}
type SessionNewResponse struct {
	Token string `json:"token"`

	Difficulty int                `json:"difficulty"`
	Salt       [ProofSaltLen]byte `json:"salt"`
}

// /session/verify
type SessionVerifyRequest struct {
	Token string `json:"token"` // proof token instead of session so it goes here instead
	Proof []byte `json:"proof"`
}
type SessionVerifyResponse struct {
	Token        string `json:"token"`
	NatNegServer string `json:"natneg_server"`
}

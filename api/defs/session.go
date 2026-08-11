package defs

import (
	"errors"
	"net/netip"

	"github.com/golang-jwt/jwt/v5"
)

const (
	ProofDifficulty = 24 // bits
	ProofSaltLen    = 16
)

type PeerID uint32

type Session struct {
	ID     PeerID
	GameID string
}

type SessionClaims struct {
	jwt.RegisteredClaims
	Session `json:"session"`

	Difficulty int                `json:"difficulty,omitzero"`
	Salt       [ProofSaltLen]byte `json:"salt,omitzero"`
}

var (
	ErrInvalidProof = errors.New("invalid proof")
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
	Token      string         `json:"token"`
	ID         PeerID         `json:"id"`
	NatNegAddr netip.AddrPort `json:"natneg_addr"`
}

package defs

import (
	"encoding/base64"
	"encoding/json"
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

	sid.FromBytes(b)

	return nil
}

func (sid SessionID) Bytes() []byte {
	return sid[:]
}

func (sid *SessionID) FromBytes(b []byte) {
	copy(sid[:], b)
}

func (sid SessionID) MarshalJSON() ([]byte, error) {
	return json.Marshal(sid[:])
}

func (sid *SessionID) UnmarshalJSON(data []byte) error {
	var b []byte
	err := json.Unmarshal(data, &b)
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
	Token        string    `json:"token"`
	ID           SessionID `json:"id"`
	NatNegServer string    `json:"natneg_server"`
}

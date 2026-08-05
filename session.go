package main

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"math/bits"
	"net/netip"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/syumai/workers/cloudflare/kv"
)

const (
	SessionIdLen = 8

	ProofDifficulty = 24 // bits
	ProofSaltLen    = 16

	SessionNamespace = "KITSU_SESSIONS"
)

type SessionID [SessionIdLen]byte

func (sid SessionID) String() string {
	return base64.RawStdEncoding.EncodeToString(sid[:])
}

type Session struct {
	ID SessionID

	GameID   string
	GameAddr netip.AddrPort
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

type SessionNewRequest struct {
	GameID string `json:"game"`
}
type SessionNewResponse struct {
	Token string `json:"token"`

	Difficulty int                `json:"difficulty"`
	Salt       [ProofSaltLen]byte `json:"salt"`
}

func SessionNew(req SessionNewRequest, _ Session) (SessionNewResponse, error) {
	var game Game
	err := GetEncodedKV(req.GameID, GameNamespace, &game)
	if err != nil {
		return SessionNewResponse{}, ErrUnknownGame
	}

	var sessionID SessionID
	rand.Read(sessionID[:])

	var salt [ProofSaltLen]byte
	rand.Read(salt[:])

	// create jwt
	token, err := jwt.NewWithClaims(jwtMethod, SessionClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   sessionID.String(),
			ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(time.Minute)),
		},
		GameID:     game.ID,
		Difficulty: ProofDifficulty,
		Salt:       salt,
	}).SignedString(MustGetJwtKey("proof"))
	if err != nil {
		return SessionNewResponse{}, err
	}

	return SessionNewResponse{
		Token:      token,
		Difficulty: ProofDifficulty,
		Salt:       salt,
	}, nil
}

type SessionVerifyRequest struct {
	Token string `json:"token"` // proof token instead of session so it goes here instead
	Proof []byte `json:"proof"`
}
type SessionVerifyResponse struct {
	Token string `json:"token"`
}

func SessionVerify(req SessionVerifyRequest, _ Session) (SessionVerifyResponse, error) {
	var claims SessionClaims
	token, err := jwt.ParseWithClaims(req.Token, &claims, func(t *jwt.Token) (any, error) { return MustGetJwtKey("proof"), nil })
	if err != nil {
		return SessionVerifyResponse{}, err
	}
	if !token.Valid {
		return SessionVerifyResponse{}, err
	}

	var session Session
	err = GetEncodedKV(claims.Subject, SessionNamespace, &session)
	if err == nil { // error if it exists
		return SessionVerifyResponse{}, ErrSessionExists
	}

	var game Game
	err = GetEncodedKV(claims.GameID, GameNamespace, &game)
	if err != nil {
		return SessionVerifyResponse{}, ErrUnknownGame
	}

	hash := sha256.New()
	hash.Write([]byte(game.ProofKey))
	hash.Write(claims.Salt[:])
	hash.Write([]byte(req.Proof))
	if bits.LeadingZeros64(binary.BigEndian.Uint64(hash.Sum(nil))) < claims.Difficulty {
		return SessionVerifyResponse{}, err
	}

	id, _ := base64.RawStdEncoding.DecodeString(claims.Subject)

	copy(session.ID[:], id)
	session.GameID = claims.GameID

	err = PutEncodedKV(session.ID.String(), SessionNamespace, session, &kv.PutOptions{ExpirationTTL: 60 * 60 * 24})
	if err != nil {
		return SessionVerifyResponse{}, err
	}

	claims.RegisteredClaims.ExpiresAt = jwt.NewNumericDate(time.Now().UTC().Add(time.Hour * 24))

	tokenStr, err := jwt.NewWithClaims(jwtMethod, claims).SignedString(MustGetJwtKey("session"))
	if err != nil {
		return SessionVerifyResponse{}, err
	}

	return SessionVerifyResponse{
		Token: tokenStr,
	}, nil
}

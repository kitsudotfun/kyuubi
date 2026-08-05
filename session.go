package main

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"encoding/gob"
	"math/bits"
	"net/http"
	"net/netip"
	"time"

	_ "embed"

	"github.com/golang-jwt/jwt/v5"
	"github.com/syumai/workers/cloudflare/kv"
)

const (
	SessionIdLen = 8

	ProofDifficulty = 24 // bits
	ProofSaltLen    = 16

	SessionNamespace = "KITSU_SESSIONS"
)

var (
	//go:embed data/proof.key
	jwtProofKey []byte

	//go:embed data/session.key
	jwtSessionKey []byte

	jwtMethod = jwt.SigningMethodHS256
)

type SessionID [SessionIdLen]byte

func (sid SessionID) String() string {
	return base64.RawStdEncoding.EncodeToString(sid[:])
}

type Session struct {
	ID       SessionID
	GameAddr netip.Addr
}

type SessionClaims struct {
	jwt.RegisteredClaims

	GameID string `json:"game"`

	Difficulty int                `json:"difficulty"`
	Salt       [ProofSaltLen]byte `json:"salt"`
}

type SessionNewRequest struct {
	GameID string `json:"game"`
}

type SessionNewResponse struct {
	Token string `json:"token"`

	Difficulty int                `json:"difficulty"`
	Salt       [ProofSaltLen]byte `json:"salt"`
}

func SessionNew(_ *http.Request, req SessionNewRequest, _ Session) (SessionNewResponse, error) {
	game, err := GetGame(req.GameID)
	if err != nil {
		return SessionNewResponse{}, err
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
	}).SignedString(jwtProofKey)
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
	Proof []byte `json:"proof"`
}

type SessionVerifyResponse struct {
	Token string `json:"token"`
}

func SessionVerify(r *http.Request, req SessionVerifyRequest, _ Session) (SessionVerifyResponse, error) {
	var claims SessionClaims
	token, err := jwt.ParseWithClaims(r.Header.Get("Authorization"), &claims, func(t *jwt.Token) (any, error) { return jwtProofKey, nil })
	if err != nil {
		return SessionVerifyResponse{}, err
	}
	if !token.Valid {
		return SessionVerifyResponse{}, err
	}

	sessions, err := kv.NewNamespace(SessionNamespace)
	if err != nil {
		return SessionVerifyResponse{}, err
	}

	// reject if session already exists
	sessionV, err := sessions.GetString(claims.Subject, nil)
	if err != nil {
		return SessionVerifyResponse{}, err

	}
	if sessionV == "" {
		return SessionVerifyResponse{}, err
	}

	game, _ := GetGame(claims.GameID)

	hash := sha256.New()
	hash.Write([]byte(game.ProofKey))
	hash.Write(claims.Salt[:])
	hash.Write([]byte(req.Proof))
	if bits.LeadingZeros64(binary.BigEndian.Uint64(hash.Sum(nil))) < claims.Difficulty {
		return SessionVerifyResponse{}, err
	}

	id, _ := base64.RawStdEncoding.DecodeString(claims.Subject)

	var session Session
	copy(session.ID[:], id)

	var buf bytes.Buffer
	err = gob.NewEncoder(&buf).Encode(session)
	if err != nil {
		return SessionVerifyResponse{}, err
	}

	err = sessions.PutReader(session.ID.String(), &buf, &kv.PutOptions{ExpirationTTL: 60 * 60 * 24})
	if err != nil {
		return SessionVerifyResponse{}, err
	}

	claims.RegisteredClaims.ExpiresAt = jwt.NewNumericDate(time.Now().UTC().Add(time.Hour * 24))

	tokenStr, err := jwt.NewWithClaims(jwtMethod, claims).SignedString(jwtSessionKey)
	if err != nil {
		return SessionVerifyResponse{}, err
	}

	return SessionVerifyResponse{
		Token: tokenStr,
	}, nil
}

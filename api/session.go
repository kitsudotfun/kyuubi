package api

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"math/bits"
	"time"

	. "github.com/kitsudotfun/kyuubi/api/defs"

	"github.com/golang-jwt/jwt/v5"
)

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
	token, err := jwt.NewWithClaims(JwtMethod, SessionClaims{
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

func SessionVerify(req SessionVerifyRequest, _ Session) (SessionVerifyResponse, error) {
	var claims SessionClaims
	token, err := jwt.ParseWithClaims(req.Token, &claims, func(t *jwt.Token) (any, error) { return MustGetJwtKey("proof"), nil })
	if err != nil {
		return SessionVerifyResponse{}, err
	}
	if !token.Valid {
		return SessionVerifyResponse{}, ErrInvalidToken
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
		return SessionVerifyResponse{}, ErrInvalidProof
	}

	claims.RegisteredClaims.ExpiresAt = jwt.NewNumericDate(time.Now().UTC().Add(time.Hour * 24))

	tokenStr, err := jwt.NewWithClaims(JwtMethod, claims).SignedString(MustGetJwtKey("session"))
	if err != nil {
		return SessionVerifyResponse{}, err
	}

	return SessionVerifyResponse{
		Token:        tokenStr,
		NatNegServer: NatNegServer,
	}, nil
}

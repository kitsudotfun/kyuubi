package api

import (
	"errors"
)

const (
	GameNamespace = "KITSU_GAMES"
)

type Game struct {
	ID       string
	ProofKey []byte
}

var (
	ErrUnknownGame = errors.New("unknown game")
)

package main

import (
	"encoding/gob"
	"errors"

	"github.com/syumai/workers/cloudflare/kv"
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

func GetGame(id string) (Game, error) {
	games, err := kv.NewNamespace(GameNamespace)
	if err != nil {
		return Game{}, err
	}
	r, err := games.GetReader(id, nil)
	if err != nil {
		return Game{}, err
	}

	var game Game
	err = gob.NewDecoder(r).Decode(&game)
	if err != nil {
		return Game{}, err
	}

	return game, nil
}

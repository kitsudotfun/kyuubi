package main

import (
	"bytes"
	"encoding/json"
	"errors"

	_ "embed"
)

//go:embed data/games.json
var gamesJson []byte

type Game struct {
	ID       string `json:"id"`
	ProofKey []byte `json:"key"`
}

var (
	ErrUnknownGame = errors.New("unknown game")
)

func GetGame(id string) (Game, error) {
	var games map[string]Game
	err := json.NewDecoder(bytes.NewReader(gamesJson)).Decode(&games)
	if err != nil {
		return Game{}, err
	}

	game, exists := games[id]
	if !exists {
		return Game{}, ErrUnknownGame
	}

	return game, nil
}

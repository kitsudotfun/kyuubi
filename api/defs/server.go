package defs

import (
	"errors"
	"net/netip"

	"github.com/golang-jwt/jwt/v5"
)

type Server struct {
	ID SessionID `json:"id"`

	GameID string         `json:"-"`
	Addr   netip.AddrPort `json:"-"`

	Name   string `json:"name"`
	Hidden bool   `json:"hidden"`

	Password string `json:"password"`

	Players    int `json:"players"`
	MaxPlayers int `json:"max_players"`

	Data any `json:"data"`
}

func (s Server) Key() string {
	return s.GameID + "|" + s.ID.String()
}

var (
	ErrUnknownGameAddr = errors.New("unknown game address")
	ErrUnknownServer   = errors.New("unknown server")
	ErrBadPassword     = errors.New("bad server password")
)

// /server/heartbeat
type ServerHeartbeatRequest struct {
	Server
	Token string `json:"token"`
}
type ServerHeartbeatResponse struct{}

// /server/delete
type ServerDeleteRequest struct{}
type ServerDeleteResponse struct{}

// /server/list
type ServerListRequest struct{}
type ServerListResponse struct {
	Servers []Server `json:"servers"`
}

// /server/join
type ServerJoinRequest struct {
	ServerID SessionID `json:"server_id"`
	Password string    `json:"password"`
}
type ServerJoinResponse struct {
	Token string `json:"token"`
}

type ServerJoinClaims struct {
	jwt.RegisteredClaims

	ServerID   SessionID      `json:"server_id"`
	ServerAddr netip.AddrPort `json:"server_addr"`
}

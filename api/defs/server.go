package defs

import (
	"errors"
	"fmt"
	"net/netip"

	"github.com/golang-jwt/jwt/v5"
)

type Server struct {
	ID PeerID `json:"id"`

	GameID string         `json:"-"`
	Addr   netip.AddrPort `json:"-"`

	Name     string `json:"name"`
	Password string `json:"password"`

	Hidden bool `json:"hidden"`

	Players    int `json:"players"`
	MaxPlayers int `json:"max_players"`

	Data any `json:"data"`
}

func (s Server) Key() string {
	return fmt.Sprintf("%s|%d", s.GameID, s.ID)
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
	ServerID PeerID `json:"server_id"`
	Password string `json:"password"`
}
type ServerJoinResponse struct {
	Token string `json:"token"`
}

type ServerJoinClaims struct {
	jwt.RegisteredClaims
	Session `json:"session"`

	ServerID   PeerID         `json:"server_id"`
	ServerAddr netip.AddrPort `json:"server_addr"`
}

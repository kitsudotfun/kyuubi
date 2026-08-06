package main

import (
	"errors"
	"net/netip"
	"slices"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Server struct {
	ID SessionID `json:"id"`

	GameID string         `json:"-"`
	Addr   netip.AddrPort `json:"-"`

	Name   string `json:"name"`
	Hidden bool   `json:"hidden"`

	HasPassword bool   `json:"has_password"`
	Password    string `json:"password"`

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

type ServerHeartbeatRequest struct {
	Server
	Token string `json:"token"`
}
type ServerHeartbeatResponse struct{}

func ServerHeartbeat(req ServerHeartbeatRequest, s Session) (ServerHeartbeatResponse, error) {
	var claims NatNegClaims
	token, err := jwt.ParseWithClaims(req.Token, &claims, func(t *jwt.Token) (any, error) { return MustGetJwtKey("natneg"), nil })
	if err != nil {
		return ServerHeartbeatResponse{}, err
	}
	if !token.Valid || !slices.Contains(claims.Audience, "natneg_attest") || claims.Subject != s.ID.String() {
		return ServerHeartbeatResponse{}, ErrInvalidToken
	}

	server := req.Server
	server.ID = s.ID
	server.GameID = s.GameID
	server.Addr = claims.Addr

	err = PutServer(server)
	if err != nil {
		return ServerHeartbeatResponse{}, err
	}

	return ServerHeartbeatResponse{}, nil
}

type ServerDeleteRequest struct{}
type ServerDeleteResponse struct{}

func ServerDelete(req ServerDeleteRequest, s Session) (ServerDeleteResponse, error) {
	err := DeleteServer(s.ID, s.GameID)
	if err != nil {
		return ServerDeleteResponse{}, err
	}

	return ServerDeleteResponse{}, nil
}

type ServerListRequest struct{}
type ServerListResponse struct {
	Servers []Server `json:"servers"`
}

func ServerList(req ServerListRequest, s Session) (ServerListResponse, error) {
	var resp ServerListResponse
	var err error
	resp.Servers, err = GetServers(s.GameID)
	if err != nil {
		return ServerListResponse{}, err
	}

	return resp, nil
}

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

func ServerJoin(req ServerJoinRequest, s Session) (ServerJoinResponse, error) {
	server, err := GetServer(req.ServerID, s.GameID)
	if err != nil {
		return ServerJoinResponse{}, ErrUnknownServer
	}
	if server.Password != req.Password {
		return ServerJoinResponse{}, ErrBadPassword
	}

	token, err := jwt.NewWithClaims(jwtMethod, ServerJoinClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   s.ID.String(),
			Audience:  jwt.ClaimStrings{"natneg_join"},
			ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(time.Minute)),
		},
		ServerID:   server.ID,
		ServerAddr: server.Addr,
	}).SignedString(MustGetJwtKey("natneg"))
	if err != nil {
		return ServerJoinResponse{}, err
	}

	return ServerJoinResponse{
		Token: token,
	}, nil
}

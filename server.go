package main

import (
	"errors"
	"net/netip"
	"slices"

	"github.com/syumai/workers/cloudflare/kv"
)

const (
	ServerNamespace      = "KITSU_SERVERS"
	ReservationNamespace = "KITSU_RESERVATIONS"
)

type Server struct {
	ID SessionID `json:"id"`

	GameID string         `json:"-"`
	Addr   netip.AddrPort `json:"-"`

	Name   string `json:"name"`
	Public bool   `json:"public"`

	HasPassword bool   `json:"has_password"`
	Password    string `json:"password"`

	Players    int `json:"players"`
	MaxPlayers int `json:"max_players"`
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
}
type ServerHeartbeatResponse struct{}

func ServerHeartbeat(req ServerHeartbeatRequest, s Session) (ServerHeartbeatResponse, error) {
	if !s.GameAddr.IsValid() {
		return ServerHeartbeatResponse{}, ErrUnknownGameAddr
	}

	server := req.Server
	server.ID = s.ID
	server.GameID = s.GameID
	server.Addr = s.GameAddr

	var stored Server
	GetEncodedKV(server.Key(), ServerNamespace, &stored)
	if stored != server {
		err := PutEncodedKV(server.Key(), ServerNamespace, server, &kv.PutOptions{ExpirationTTL: 60 * 60}) // 1 hour
		if err != nil {
			return ServerHeartbeatResponse{}, err
		}
	}

	return ServerHeartbeatResponse{}, nil
}

type ServerDeleteRequest struct{}
type ServerDeleteResponse struct{}

func ServerDelete(req ServerDeleteRequest, s Session) (ServerDeleteResponse, error) {
	if !s.GameAddr.IsValid() {
		return ServerDeleteResponse{}, ErrUnknownGameAddr
	}

	servers, err := kv.NewNamespace(ServerNamespace)
	if err != nil {
		return ServerDeleteResponse{}, err
	}
	err = servers.Delete(s.GameID + "|" + s.ID.String())
	if err != nil {
		return ServerDeleteResponse{}, err
	}

	return ServerDeleteResponse{}, nil
}

type ServerListRequest struct {
	Limit  int    `json:"limit"`
	Cursor string `json:"cursor"`
}
type ServerListResponse struct {
	Cursor  string   `json:"cursor"`
	Servers []Server `json:"servers"`
}

func ServerList(req ServerListRequest, s Session) (ServerListResponse, error) {
	if !s.GameAddr.IsValid() {
		return ServerListResponse{}, ErrUnknownGameAddr
	}

	servers, err := kv.NewNamespace(ServerNamespace)
	if err != nil {
		return ServerListResponse{}, err
	}
	list, err := servers.List(&kv.ListOptions{
		Limit:  req.Limit,
		Prefix: s.GameID + "|",
		Cursor: req.Cursor,
	})
	if err != nil {
		return ServerListResponse{}, err
	}

	var resp ServerListResponse
	if !list.ListComplete {
		resp.Cursor = list.Cursor
	}
	for _, item := range list.Keys {
		var server Server
		err = GetEncodedKV(item.Name, ServerNamespace, &server)
		if err != nil {
			continue
		}

		// skip non-public servers
		if !server.Public {
			continue
		}

		// hide password
		if server.Password != "" {
			server.HasPassword = true
			server.Password = ""
		}

		resp.Servers = append(resp.Servers, server)
	}

	return resp, nil
}

type ServerJoinRequest struct {
	ID       string `json:"id"`
	Password string `json:"password"`
}
type ServerJoinResponse struct {
	Addr netip.AddrPort `json:"addr"`
}

type Reservation struct {
	ID   SessionID      `json:"id"`
	Addr netip.AddrPort `json:"addr"`
}

func ServerJoin(req ServerJoinRequest, s Session) (ServerJoinResponse, error) {
	if !s.GameAddr.IsValid() {
		return ServerJoinResponse{}, ErrUnknownGameAddr
	}

	var server Server
	err := GetEncodedKV(s.GameID+"|"+req.ID, ServerNamespace, &server)
	if err != nil {
		return ServerJoinResponse{}, ErrUnknownServer
	}
	if server.Password != req.Password {
		return ServerJoinResponse{}, ErrBadPassword
	}

	resv := Reservation{
		ID:   s.ID,
		Addr: s.GameAddr,
	}

	var reservations []Reservation
	GetEncodedKV(server.Key(), ReservationNamespace, &reservations)
	if !slices.Contains(reservations, resv) {
		err = PutEncodedKV(server.Key(), ReservationNamespace, append(reservations, resv), &kv.PutOptions{ExpirationTTL: 60})
		if err != nil {
			return ServerJoinResponse{}, err
		}
	}

	return ServerJoinResponse{
		Addr: server.Addr,
	}, nil
}

type ServerResvListRequest struct{}
type ServerResvListResponse struct {
	Reservations []Reservation `json:"reservations"`
}

func ServerResvList(_ ServerResvListRequest, s Session) (ServerResvListResponse, error) {
	if !s.GameAddr.IsValid() {
		return ServerResvListResponse{}, ErrUnknownGameAddr
	}

	var resp ServerResvListResponse
	GetEncodedKV(s.GameID+"|"+s.ID.String(), ReservationNamespace, &resp.Reservations)

	return resp, nil
}

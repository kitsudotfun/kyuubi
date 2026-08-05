package main

import (
	"bytes"
	"encoding/gob"
	"errors"
	"net/http"
	"net/netip"

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

	Name     string `json:"name"`
	Public   bool   `json:"public"`
	Password string `json:"password"`

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

func GetServer(id string) (Server, error) {
	servers, err := kv.NewNamespace(ServerNamespace)
	if err != nil {
		return Server{}, err
	}
	r, err := servers.GetReader(id, nil)
	if err != nil {
		return Server{}, err
	}

	var server Server
	err = gob.NewDecoder(r).Decode(&server)
	if err != nil {
		return Server{}, err
	}

	return server, nil
}

type ServerHeartbeatRequest struct {
	Server
}
type ServerHeartbeatResponse struct{}

func ServerHeartbeat(_ *http.Request, req ServerHeartbeatRequest, s Session) (ServerHeartbeatResponse, error) {
	if !s.GameAddr.IsValid() {
		return ServerHeartbeatResponse{}, ErrUnknownGameAddr
	}

	server := req.Server
	server.ID = s.ID
	server.GameID = s.GameID
	server.Addr = s.GameAddr

	var buf bytes.Buffer
	err := gob.NewEncoder(&buf).Encode(server)
	if err != nil {
		return ServerHeartbeatResponse{}, err
	}

	servers, err := kv.NewNamespace(ServerNamespace)
	if err != nil {
		return ServerHeartbeatResponse{}, err
	}
	err = servers.PutReader(server.Key(), &buf, &kv.PutOptions{ExpirationTTL: 60 * 5})
	if err != nil {
		return ServerHeartbeatResponse{}, err
	}

	return ServerHeartbeatResponse{}, nil
}

type ServerDeleteRequest struct{}
type ServerDeleteResponse struct{}

func ServerDelete(_ *http.Request, req ServerDeleteRequest, s Session) (ServerDeleteResponse, error) {
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

func ServerList(_ *http.Request, req ServerListRequest, s Session) (ServerListResponse, error) {
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
		r, err := servers.GetReader(item.Name, nil)
		if err != nil {
			continue
		}

		var server Server
		err = gob.NewDecoder(r).Decode(&server)
		if err != nil {
			continue
		}

		// skip non-public servers
		if !server.Public {
			continue
		}

		// hide password
		server.Password = ""

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

func ServerJoin(_ *http.Request, req ServerJoinRequest, s Session) (ServerJoinResponse, error) {
	if !s.GameAddr.IsValid() {
		return ServerJoinResponse{}, ErrUnknownGameAddr
	}

	server, err := GetServer(s.GameID + "|" + req.ID)
	if err != nil {
		return ServerJoinResponse{}, err
	}
	if server.Password != req.Password {
		return ServerJoinResponse{}, ErrBadPassword
	}

	reservations, err := kv.NewNamespace(ReservationNamespace)
	if err != nil {
		return ServerJoinResponse{}, err
	}

	var buf bytes.Buffer
	err = gob.NewEncoder(&buf).Encode(Reservation{
		ID:   s.ID,
		Addr: s.GameAddr,
	})
	if err != nil {
		return ServerJoinResponse{}, err
	}
	err = reservations.PutReader(s.GameID+"|"+req.ID+"|"+s.ID.String(), &buf, &kv.PutOptions{ExpirationTTL: 60})
	if err != nil {
		return ServerJoinResponse{}, err
	}

	return ServerJoinResponse{
		Addr: server.Addr,
	}, nil
}

type ServerResvListRequest struct{}
type ServerResvListResponse struct {
	Reservations []Reservation `json:"reservations"`
}

func ServerResvList(_ *http.Request, _ ServerResvListRequest, s Session) (ServerResvListResponse, error) {
	if !s.GameAddr.IsValid() {
		return ServerResvListResponse{}, ErrUnknownGameAddr
	}

	reservations, err := kv.NewNamespace(ReservationNamespace)
	if err != nil {
		return ServerResvListResponse{}, err
	}

	var resp ServerResvListResponse
	list, err := reservations.List(&kv.ListOptions{Prefix: s.GameID + "|" + s.ID.String() + "|"})
	if err != nil {
		return ServerResvListResponse{}, err
	}
	for _, item := range list.Keys {
		r, err := reservations.GetReader(item.Name, nil)
		if err != nil {
			continue
		}

		var resv Reservation
		err = gob.NewDecoder(r).Decode(&resv)
		if err != nil {
			continue
		}

		resp.Reservations = append(resp.Reservations, resv)
	}

	return resp, nil
}

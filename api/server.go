package api

import (
	"slices"
	"time"

	. "github.com/kitsudotfun/kyuubi/api/defs"

	"github.com/golang-jwt/jwt/v5"
)

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

func ServerDelete(req ServerDeleteRequest, s Session) (ServerDeleteResponse, error) {
	err := DeleteServer(s.ID, s.GameID)
	if err != nil {
		return ServerDeleteResponse{}, err
	}

	return ServerDeleteResponse{}, nil
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

func ServerJoin(req ServerJoinRequest, s Session) (ServerJoinResponse, error) {
	server, err := GetServer(req.ServerID, s.GameID)
	if err != nil {
		return ServerJoinResponse{}, ErrUnknownServer
	}
	if server.Password != req.Password {
		return ServerJoinResponse{}, ErrBadPassword
	}

	token, err := jwt.NewWithClaims(JwtMethod, ServerJoinClaims{
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

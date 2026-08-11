package api

import (
	"database/sql"
	"encoding/json"
	"net/netip"

	. "github.com/kitsudotfun/kyuubi/api/defs"

	_ "github.com/syumai/workers/cloudflare/d1"
)

func MustGetDB() *sql.DB {
	db, err := sql.Open("d1", DatabaseName)
	if err != nil {
		panic(err)
	}

	return db
}

func PutServer(server Server) error {
	data, err := json.Marshal(server.Data)
	if err != nil {
		return err
	}

	_, err = MustGetDB().Exec(PutServerQuery, server.ID, server.GameID, server.Addr.String(), server.Name, server.Hidden, server.Password, server.Players, server.MaxPlayers, string(data))
	if err != nil {
		return err
	}

	return nil
}

func DeleteServer(id PeerID, game string) error {
	_, err := MustGetDB().Exec(DeleteServerQuery, game, id)
	if err != nil {
		return err
	}

	return nil
}

func GetServer(id PeerID, game string) (Server, error) {
	var server Server
	var addr, data string
	err := MustGetDB().QueryRow(GetServerQuery, game, id).Scan(&server.ID, &server.GameID, &addr, &server.Name, &server.Hidden, &server.Password, &server.Players, &server.MaxPlayers, &data)
	if err != nil {
		return Server{}, err
	}

	server.Addr, err = netip.ParseAddrPort(addr)
	if err != nil {
		return Server{}, err
	}

	err = json.Unmarshal([]byte(data), &server.Data)
	if err != nil {
		return Server{}, err
	}

	return server, nil
}

func GetServers(game string) ([]Server, error) {
	rows, err := MustGetDB().Query(GetServersQuery, game)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var servers []Server
	for rows.Next() {
		var server Server
		var addr, data string
		err = rows.Scan(&server.ID, &server.GameID, &addr, &server.Name, &server.Password, &server.Players, &server.MaxPlayers, &data)
		if err != nil {
			return nil, err
		}

		server.Addr, err = netip.ParseAddrPort(addr)
		if err != nil {
			return nil, err
		}

		err = json.Unmarshal([]byte(data), &server.Data)
		if err != nil {
			return nil, err
		}

		// mask password
		if server.Password != "" {
			server.Password = "<protected>"
		}

		servers = append(servers, server)
	}
	if rows.Err() != nil {
		return nil, err
	}

	return servers, nil
}

func CleanServers() error {
	_, err := MustGetDB().Exec(CleanServersQuery)
	if err != nil {
		return err
	}

	return nil
}

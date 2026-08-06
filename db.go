package main

import (
	"database/sql"
	"encoding/json"
	"net/netip"

	_ "github.com/syumai/workers/cloudflare/d1"
)

const DatabaseName = "KITSU_DB"

func MustGetDB() *sql.DB {
	db, err := sql.Open("d1", DatabaseName)
	if err != nil {
		panic(err)
	}

	return db
}

const (
	putServer = `
	INSERT OR REPLACE INTO servers 
	(id, game, addr, name, hidden, password, players, max_players, data)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`

	deleteServer = `
	DELETE FROM servers 
	WHERE game = ? 
	AND id = ?`

	getServers = `
	SELECT id, game, addr, name, hidden, password, players, max_players, data 
	FROM servers 
	WHERE game = ? 
	AND updated > DATETIME('now', '-5 minutes')`
)

func PutServer(server Server) error {
	data, err := json.Marshal(server.Data)
	if err != nil {
		return err
	}

	_, err = MustGetDB().Exec(putServer, server.ID, server.GameID, server.Addr, server.Name, server.Hidden, server.Password, server.Players, server.MaxPlayers, string(data))
	if err != nil {
		return err
	}

	return nil
}

func DeleteServer(id SessionID, game string) error {
	_, err := MustGetDB().Exec(deleteServer, game, id)
	if err != nil {
		return err
	}

	return nil
}

func GetServer(id SessionID, game string) (Server, error) {
	var server Server
	var addr, data string
	err := MustGetDB().QueryRow(getServers+" AND id = ?", game, id).Scan(&server.ID, &server.GameID, &addr, &server.Name, &server.Hidden, &server.Password, &server.Players, &server.MaxPlayers, &data)
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
	rows, err := MustGetDB().Query(getServers, game)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var servers []Server
	for rows.Next() {
		var server Server
		var addr, data string
		err = rows.Scan(&server.ID, &server.GameID, &addr, &server.Name, &server.Hidden, &server.Password, &server.Players, &server.MaxPlayers, &data)
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

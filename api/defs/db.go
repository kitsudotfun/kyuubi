package defs

const DatabaseName = "KITSU_DB"

const (
	PutServerQuery = `
	INSERT OR REPLACE INTO servers 
	(id, game, addr, name, hidden, password, players, max_players, data)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`

	DeleteServerQuery = `
	DELETE FROM servers 
	WHERE game = ? 
	AND id = ?`

	GetServerQuery = `
	SELECT id, game, addr, name, hidden, password, players, max_players, data 
	FROM servers 
	WHERE game = ? 
	AND id = ?
	AND updated > DATETIME('now', '-5 minutes')`

	GetServersQuery = `
	SELECT id, game, addr, name, password, players, max_players, data 
	FROM servers 
	WHERE game = ? 
	AND hidden = 0
	AND updated > DATETIME('now', '-5 minutes')`

	CleanServersQuery = `
	DELETE FROM servers
	WHERE updated < DATETIME('now', '-5 minutes')`
)

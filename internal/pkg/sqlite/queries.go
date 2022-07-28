package sqlite

var (
	queryCreate = `CREATE TABLE IF NOT EXISTS secret(
		key     TEXT UNIQUE,
		content TEXT
	)`
	queryInsert = `INSERT INTO secret(key, content) VALUES($1, $2)`
	querySelect = `SELECT content FROM secret WHERE key=$1`
	queryDelete = `DELETE FROM secret WHERE key=$1`
)

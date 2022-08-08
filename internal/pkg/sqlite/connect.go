package sqlite

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/Nunoki/onetimesharer/internal/pkg/crypter"
	_ "github.com/mattn/go-sqlite3"
)

type sqliteStore struct {
	Context  context.Context
	Crypter  crypter.Crypter
	Database *sql.DB
}

var (
	filename = "secrets.db"
)

// New returns a new instance of the sqliteStore with an initialized database connection.
func New(ctx context.Context, cr crypter.Crypter) (sqliteStore, error) {
	db, err := sql.Open("sqlite3", filename)
	if err != nil {
		return sqliteStore{}, fmt.Errorf("can't open connection to database: %w", err)
	}

	if err = db.PingContext(ctx); err != nil {
		return sqliteStore{}, fmt.Errorf("failed connection to database: %w", err)
	}

	db.ExecContext(
		ctx,
		queryCreate,
	)
	if err != nil {
		return sqliteStore{}, fmt.Errorf("failed to create database table: %w", err)
	}

	return sqliteStore{
			Context:  ctx,
			Crypter:  cr,
			Database: db,
		},
		nil
}

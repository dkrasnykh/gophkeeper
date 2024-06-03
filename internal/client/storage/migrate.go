package storage

import (
	"database/sql"
	"embed"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
	"github.com/pressly/goose/v3"
)

//go:embed migrations
var migrations embed.FS

func migrate(db *sql.DB, version int64) error {
	goose.SetBaseFS(migrations)

	if err := goose.SetDialect("sqlite3"); err != nil {
		return fmt.Errorf("sqlite3 migrate set dialect sqlite3: %w", err)
	}

	if err := goose.UpTo(db, "migrations", version); err != nil {
		return fmt.Errorf("sqlite3 migrate up: %w", err)
	}

	return nil
}

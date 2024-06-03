package storage

import (
	"database/sql"
	"errors"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

var (
	ErrInternal     = errors.New("internal error")
	ErrItemNotFound = errors.New("item not found")
)

func New(storagePath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", storagePath)
	if err != nil {
		return nil, fmt.Errorf("failed connect to database %w", ErrInternal)
	}

	err = migrate(db, 1)
	if err != nil {
		return nil, fmt.Errorf("failed migrate database schema %w", ErrInternal)
	}

	return db, nil
}

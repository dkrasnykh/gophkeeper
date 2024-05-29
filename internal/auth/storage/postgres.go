package storage

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrUserExists   = errors.New("user already exists")
	ErrUserNotFound = errors.New("user not found")
	ErrAppNotFound  = errors.New("app not found")
)

func New(databaseURL string, timeout time.Duration) (*pgxpool.Pool, error) {
	newCtx, cancel := context.WithTimeout(context.Background(), timeout*time.Second)
	defer cancel()

	pool, err := pgxpool.New(newCtx, databaseURL)
	if err != nil {
		return nil, fmt.Errorf("postgres connection error: %w", err)
	}
	if err = migrate(pool, 1); err != nil {
		return nil, fmt.Errorf("postgres migration error: %w", err)
	}
	return pool, nil
}

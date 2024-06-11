package storage

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/dkrasnykh/gophkeeper/pkg/models"
)

// AppPostgres implements AppProvider interface.
// auth service may be used by users of other applications.
type AppPostgres struct {
	db      *pgxpool.Pool
	timeout time.Duration
}

func NewAppPostgres(db *pgxpool.Pool, timeout time.Duration) *AppPostgres {
	return &AppPostgres{
		db:      db,
		timeout: timeout,
	}
}

// App returns application information by application id.
func (s *AppPostgres) App(ctx context.Context, id int) (models.App, error) {
	const op = "storage.postgres.App"

	newCtx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	var app models.App
	var err error

	row := s.db.QueryRow(newCtx, "SELECT id, name, secret FROM apps WHERE id = $1", id)
	err = row.Scan(&app.ID, &app.Name, &app.Secret)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.App{}, fmt.Errorf("%s: %w", op, ErrAppNotFound)
		}
		return models.App{}, fmt.Errorf("%s: %w", op, err)
	}
	return app, nil
}

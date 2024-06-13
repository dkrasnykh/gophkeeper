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

func NewAppPostgres(databaseURL string, timeout time.Duration) (*AppPostgres, error) {
	pool, err := newPool(databaseURL, timeout)
	if err != nil {
		return nil, err
	}
	return &AppPostgres{
		db:      pool,
		timeout: timeout,
	}, nil
}

// App returns application information by application id.
func (s *AppPostgres) App(ctx context.Context, id int) (models.App, error) {
	const op = "storage.postgres.App"

	newCtx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	rows, err := s.db.Query(newCtx, "SELECT (id, name, secret) FROM apps WHERE id = $1", id)
	if err != nil {
		return models.App{}, fmt.Errorf("%s: %w", op, err)
	}
	app, err := pgx.CollectExactlyOneRow(rows, pgx.RowTo[models.App])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.App{}, fmt.Errorf("%s: %w", op, ErrAppNotFound)
		}
		return models.App{}, fmt.Errorf("%s: %w", op, err)
	}
	return app, nil
}

func (s *AppPostgres) Close() {
	s.db.Close()
}

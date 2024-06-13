package storage

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/dkrasnykh/gophkeeper/pkg/models"
)

// UserPostgres implements UserProvider interface.
type UserPostgres struct {
	db      *pgxpool.Pool
	timeout time.Duration
}

func NewUserPostgres(databaseURL string, timeout time.Duration) (*UserPostgres, error) {
	pool, err := newPool(databaseURL, timeout)
	if err != nil {
		return nil, err
	}
	return &UserPostgres{
		db:      pool,
		timeout: timeout,
	}, nil
}

// SaveUser saved new user into database.
// It returns ErrUserExists error, if user alredy inserted into database.
func (s *UserPostgres) SaveUser(ctx context.Context, email string, passHash []byte) (int64, error) {
	const op = "storage.postgres.SaveUser"
	newCtx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	var id int64
	var err error

	row := s.db.QueryRow(newCtx, "INSERT INTO users (login, password_hash) values ($1, $2) RETURNING id", email, passHash)
	err = row.Scan(&id)
	if err != nil {
		if isLoginExistError(err) {
			return 0, fmt.Errorf("%s: %w", op, ErrUserExists)
		}
		return 0, fmt.Errorf("%s: %w", op, err)
	}
	return id, nil
}

// User returns user credentials by email.
// It returns ErrUserNotFound error, if user with email does not registered.
func (s *UserPostgres) User(ctx context.Context, email string) (models.User, error) {
	const op = "storage.postgres.User"

	newCtx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	rows, err := s.db.Query(newCtx, "select (id, login, password_hash) from users where login = $1", email)
	if err != nil {
		return models.User{}, fmt.Errorf("%s: %w", op, err)
	}
	user, err := pgx.CollectExactlyOneRow(rows, pgx.RowTo[models.User])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.User{}, fmt.Errorf("%s: %w", op, ErrUserNotFound)
		}
		return models.User{}, fmt.Errorf("%s: %w", op, err)
	}
	return user, nil
}

func (s *UserPostgres) Close() {
	s.db.Close()
}

func isLoginExistError(err error) bool {
	pgxErr, ok := err.(*pgconn.PgError)
	if ok && pgxErr.Code == "23505" {
		return true
	}
	return false
}

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

type UserPostgres struct {
	db      *pgxpool.Pool
	timeout time.Duration
}

func NewUserPostgres(db *pgxpool.Pool, timeout time.Duration) *UserPostgres {
	return &UserPostgres{
		db:      db,
		timeout: timeout,
	}
}

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

func (s *UserPostgres) User(ctx context.Context, email string) (models.User, error) {
	const op = "storage.postgres.User"
	newCtx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	var user models.User
	var err error

	row := s.db.QueryRow(newCtx, "select id, login, password_hash from users where login = $1", email)
	err = row.Scan(&user.ID, &user.Email, &user.PassHash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.User{}, fmt.Errorf("%s: %w", op, ErrUserNotFound)
		}
		return models.User{}, fmt.Errorf("%s: %w", op, err)
	}
	return user, nil
}

func isLoginExistError(err error) bool {
	pgxErr, ok := err.(*pgconn.PgError)
	if ok && pgxErr.Code == "23505" {
		return true
	}
	return false
}

package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/dkrasnykh/gophkeeper/pkg/models"
)

type CredentialsSqlite struct {
	db      *sql.DB
	timeout time.Duration
}

func NewCredentialsSqlite(storagePath string, timeout time.Duration) (*CredentialsSqlite, error) {
	db, err := newSQLDB(storagePath)
	if err != nil {
		return nil, err
	}
	return &CredentialsSqlite{
		db:      db,
		timeout: timeout,
	}, nil
}

func (s *CredentialsSqlite) All(ctx context.Context) ([]models.Credentials, error) {
	const op = "storage.sqlite.Credentials.All"

	newCtx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	stmt, err := s.db.Prepare("SELECT tag, login, password, comment, created_at FROM credentials")
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	rows, err := stmt.QueryContext(newCtx)

	res := []models.Credentials{}

	for rows.Next() {
		cred := models.Credentials{Type: models.CredItem}
		err = rows.Scan(&cred.Tag, &cred.Login, &cred.Password, &cred.Comment, &cred.Created)
		if err != nil {
			continue
		}
		res = append(res, cred)
	}
	return res, nil
}

func (s *CredentialsSqlite) ByLogin(ctx context.Context, login string) (models.Credentials, error) {
	const op = "storage.sqlite.Credentials.ByLogin"

	newCtx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	stmt, err := s.db.Prepare("SELECT tag, login, password, comment, created_at FROM credentials WHERE login = ?")
	if err != nil {
		return models.Credentials{}, fmt.Errorf("%s: %w", op, err)
	}

	row := stmt.QueryRowContext(newCtx, login)

	cred := models.Credentials{Type: models.CredItem}
	err = row.Scan(&cred.Tag, &cred.Login, &cred.Password, &cred.Comment, &cred.Created)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.Credentials{}, fmt.Errorf("%s, %w", op, ErrItemNotFound)
		}

		return models.Credentials{}, fmt.Errorf("%s: %w", op, err)
	}

	return cred, nil
}

func (s *CredentialsSqlite) Save(ctx context.Context, cred models.Credentials) error {
	const op = "storage.sqlite.Credentials.Save"

	newCtx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	stmt, err := s.db.Prepare("INSERT INTO credentials(tag, login, password, comment, created_at) VALUES(?, ?, ?, ?, ?)")
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	_, err = stmt.ExecContext(newCtx, cred.Tag, cred.Login, cred.Password, cred.Comment, cred.Created)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (s *CredentialsSqlite) Update(ctx context.Context, cred models.Credentials) error {
	const op = "storage.sqlite.Credentials.Update"

	newCtx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	stmt, err := s.db.Prepare("UPDATE credentials SET tag=?, password=?, comment=?, created_at=? WHERE login=?")
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	_, err = stmt.ExecContext(newCtx, cred.Tag, cred.Password, cred.Comment, cred.Created, cred.Login)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (s *CredentialsSqlite) Close() error {
	if err := s.db.Close(); err != nil {
		return ErrInternal
	}
	return nil
}

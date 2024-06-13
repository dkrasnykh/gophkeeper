package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/dkrasnykh/gophkeeper/pkg/models"
)

type TextSqlite struct {
	db      *sql.DB
	timeout time.Duration
}

func NewTextSqlite(storagePath string, timeout time.Duration) (*TextSqlite, error) {
	db, err := newSQLDB(storagePath)
	if err != nil {
		return nil, err
	}
	return &TextSqlite{
		db:      db,
		timeout: timeout,
	}, nil
}

func (s *TextSqlite) All(ctx context.Context) ([]models.Text, error) {
	const op = "storage.sqlite.Text.All"

	newCtx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	stmt, err := s.db.Prepare("SELECT tag, key, value, comment, created_at FROM text")
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	rows, err := stmt.QueryContext(newCtx)
	res := make([]models.Text, 0)
	for rows.Next() {
		text := models.Text{Type: models.TextItem}
		err = rows.Scan(&text.Tag, &text.Key, &text.Value, &text.Comment, &text.Created)
		if err != nil {
			continue
		}
		res = append(res, text)
	}
	return res, nil
}

func (s *TextSqlite) ByKey(ctx context.Context, key string) (models.Text, error) {
	const op = "storage.sqlite.Text.ByKey"

	newCtx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	stmt, err := s.db.Prepare("SELECT tag, key, value, comment, created_at FROM text WHERE key = ?")
	if err != nil {
		return models.Text{}, fmt.Errorf("%s: %w", op, err)
	}

	row := stmt.QueryRowContext(newCtx, key)

	text := models.Text{Type: models.TextItem}
	err = row.Scan(&text.Tag, &text.Key, &text.Value, &text.Comment, &text.Created)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.Text{}, fmt.Errorf("%s: %w", op, ErrItemNotFound)
		}
		return models.Text{}, fmt.Errorf("%s, %w", op, err)
	}

	return text, nil
}

func (s *TextSqlite) Save(ctx context.Context, text models.Text) error {
	const op = "storage.sqlite.Text.Save"

	newCtx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	stmt, err := s.db.Prepare("INSERT INTO text(tag, key, value, comment, created_at) VALUES(?, ?, ?, ?, ?)")
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	_, err = stmt.ExecContext(newCtx, text.Tag, text.Key, text.Value, text.Comment, text.Created)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (s *TextSqlite) Update(ctx context.Context, text models.Text) error {
	const op = "storage.sqlite.Text.Update"

	newCtx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	stmt, err := s.db.Prepare("UPDATE text SET tag = ?, value=?, comment=?, created_at=? WHERE key=?")
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	_, err = stmt.ExecContext(newCtx, text.Tag, text.Value, text.Comment, text.Created, text.Key)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (s *TextSqlite) Close() error {
	if err := s.db.Close(); err != nil {
		return ErrInternal
	}
	return nil
}

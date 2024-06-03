package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/dkrasnykh/gophkeeper/pkg/models"
	"time"
)

type BinarySqlite struct {
	db      *sql.DB
	timeout time.Duration
}

func NewBinarySqlite(db *sql.DB, timeout time.Duration) *BinarySqlite {
	return &BinarySqlite{
		db:      db,
		timeout: timeout,
	}
}

func (s *BinarySqlite) All(ctx context.Context) ([]models.Binary, error) {
	const op = "storage.sqlite.Binary.All"

	newCtx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	stmt, err := s.db.Prepare("SELECT tag, key, value, comment, created_at FROM binary")
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	rows, err := stmt.QueryContext(newCtx)

	bins := []models.Binary{}
	for rows.Next() {
		bin := models.Binary{Type: "bin"}
		err = rows.Scan(&bin.Tag, &bin.Key, &bin.Value, &bin.Comment, &bin.Created)
		if err != nil {
			continue
		}
		bins = append(bins, bin)
	}

	return bins, nil
}

func (s *BinarySqlite) ByKey(ctx context.Context, key string) (models.Binary, error) {
	const op = "storage.sqlite.Binary.ByKey"

	newCtx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	stmt, err := s.db.Prepare("SELECT tag, key, value, comment, created_at FROM binary WHERE key = ?")
	if err != nil {
		return models.Binary{}, fmt.Errorf("%s: %w", op, err)
	}

	row := stmt.QueryRowContext(newCtx, key)

	var bin models.Binary
	err = row.Scan(&bin.Tag, &bin.Key, &bin.Value, &bin.Comment, &bin.Created)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.Binary{}, fmt.Errorf("%s: %w", op, ErrItemNotFound)
		}

		return models.Binary{}, fmt.Errorf("%s: %w", op, err)
	}

	return bin, nil
}

func (s *BinarySqlite) Save(ctx context.Context, bin models.Binary) error {
	const op = "storage.sqlite.Binary.Save"

	newCtx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	stmt, err := s.db.Prepare("INSERT INTO binary(tag, key, value, comment, created_at) VALUES(?, ?, ?, ?, ?)")
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	_, err = stmt.ExecContext(newCtx, bin.Tag, bin.Key, bin.Value, bin.Comment, bin.Created)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *BinarySqlite) Update(ctx context.Context, bin models.Binary) error {
	const op = "storage.sqlite.Binary.Save"

	newCtx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	stmt, err := s.db.Prepare("UPDATE binary SET tag = ?, value=?, comment=?, created_at=? WHERE key=?")
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	_, err = stmt.ExecContext(newCtx, bin.Tag, bin.Value, bin.Comment, bin.Created, bin.Key)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

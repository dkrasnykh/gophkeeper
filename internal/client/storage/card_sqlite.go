package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/dkrasnykh/gophkeeper/pkg/models"
	"time"
)

type CardSqlite struct {
	db      *sql.DB
	timeout time.Duration
}

func NewCardSqlite(db *sql.DB, timeout time.Duration) *CardSqlite {
	return &CardSqlite{
		db:      db,
		timeout: timeout,
	}
}

func (s *CardSqlite) All(ctx context.Context) ([]models.Card, error) {
	const op = "storage.sqlite.Card.All"

	newCtx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	stmt, err := s.db.Prepare("SELECT tag, number, exp, cvv, comment, created_at FROM card")
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	rows, err := stmt.QueryContext(newCtx)

	cards := []models.Card{}

	for rows.Next() {
		card := models.Card{Type: "card"}
		err = rows.Scan(&card.Tag, &card.Number, &card.Exp, &card.CVV, &card.Comment, &card.Created)
		if err != nil {
			continue
		}
		cards = append(cards, card)
	}

	return cards, nil
}

func (s *CardSqlite) ByNumber(ctx context.Context, number string) (models.Card, error) {
	const op = "storage.sqlite.Card.ByNumber"

	newCtx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	stmt, err := s.db.Prepare("SELECT tag, number, exp, cvv, comment, created_at FROM card WHERE number = ?")
	if err != nil {
		return models.Card{}, fmt.Errorf("%s: %w", op, err)
	}

	row := stmt.QueryRowContext(newCtx, number)

	var card models.Card
	err = row.Scan(&card.Tag, &card.Number, &card.Exp, &card.CVV, &card.Comment, &card.Created)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.Card{}, fmt.Errorf("%s: %w", op, ErrItemNotFound)
		}

		return models.Card{}, fmt.Errorf("%s: %w", op, err)
	}
	return card, nil
}

func (s *CardSqlite) Save(ctx context.Context, card models.Card) error {
	const op = "storage.sqlite.Card.Save"

	newCtx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	stmt, err := s.db.Prepare("INSERT INTO card(tag, number, exp, cvv, comment, created_at) VALUES(?, ?, ?, ?, ?, ?)")
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)

	}
	_, err = stmt.ExecContext(newCtx, card.Tag, card.Number, card.Exp, card.CVV, card.Comment, card.Created)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (s *CardSqlite) Update(ctx context.Context, card models.Card) error {
	const op = "storage.sqlite.Card.Save"

	newCtx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	stmt, err := s.db.Prepare("UPDATE card SET tag=?, exp=?, cvv=?, comment=?, created_at=? WHERE number=?")
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	_, err = stmt.ExecContext(newCtx, card.Tag, card.Exp, card.CVV, card.Comment, card.Created, card.Number)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

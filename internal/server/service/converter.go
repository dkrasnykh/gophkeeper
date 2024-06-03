package service

import (
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/dkrasnykh/gophkeeper/internal/server/storage"
	"github.com/dkrasnykh/gophkeeper/pkg/logger/sl"
	"github.com/dkrasnykh/gophkeeper/pkg/models"
)

func (s *Service) Validate(msg models.Message) (models.Message, error) {
	const op = "servicekeeper.Validate"
	log := s.log.With(
		slog.String("op", op),
	)

	var kind struct{ Type string }
	err := json.Unmarshal(msg.Value, &kind)
	if err != nil {
		log.Error(
			"failed extract value from message",
			slog.String(`msg should contains "Type" field`, string(msg.Value)),
			sl.Err(err),
		)
		return models.Message{}, fmt.Errorf("%op: %w", op, ErrInvalidMessage)
	}
	switch kind.Type {
	case "cred":
		var cred models.Credentials
		err := json.Unmarshal(msg.Value, &cred)
		if err != nil {
			log.Error(
				`failed unmarshal message value into "Credentials" model`,
				slog.String(`msg`, string(msg.Value)),
				sl.Err(err),
			)
			return models.Message{}, fmt.Errorf("%s: %w", op, ErrInvalidMessage)
		}
	case "text":
		var text models.Text
		err = json.Unmarshal(msg.Value, &text)
		if err != nil {
			log.Error(
				`failed unmarshal message value into "Text" model`,
				slog.String(`msg`, string(msg.Value)),
				sl.Err(err),
			)
			return models.Message{}, fmt.Errorf("%s: %w", op, ErrInvalidMessage)
		}
	case "bin":
		var bin models.Binary
		err = json.Unmarshal(msg.Value, &bin)
		if err != nil {
			log.Error(
				`failed unmarshal message value into "Binary" model`,
				slog.String(`msg`, string(msg.Value)),
				sl.Err(err),
			)
			return models.Message{}, fmt.Errorf("%s: %w", op, ErrInvalidMessage)
		}
	case "card":
		var card models.Card
		err = json.Unmarshal(msg.Value, &card)
		if err != nil {
			log.Error(
				`failed unmarshal message value into "Card" model`,
				slog.String(`msg`, string(msg.Value)),
				sl.Err(err),
			)
			return models.Message{}, fmt.Errorf("%s: %w", op, ErrInvalidMessage)
		}
	default:
		log.Error(
			`failed unmarshal message into existing models`,
			slog.String(`unknown message type`, kind.Type),
		)
		return models.Message{}, fmt.Errorf("%s: %w", op, ErrInvalidMessage)
	}
	return models.Message{Type: "update", Value: msg.Value}, nil
}

func (s *Service) convertItemListToMessage(items []storage.Item) models.Message {
	const op = "servicekeeper.ConvertItemListToMessage"
	log := s.log.With(
		slog.String("op", op),
	)

	values := make([][]byte, 0, len(items))
	for _, item := range items {
		values = append(values, item.Data)
	}
	msg, _ := json.Marshal(values)
	log.Info(
		"items converted",
		slog.Int("number of added items into message", len(items)),
	)

	return models.Message{Type: "snapshot", Value: msg}
}

func (s *Service) convertMessageToItem(userID int64, msg models.Message) storage.Item {
	const op = "servicekeeper.ConvertItemListToMessage"
	log := s.log.With(
		slog.String("op", op),
		slog.Int64("user_id", userID),
	)

	var item storage.Item
	item.Data = msg.Value
	item.UserID = userID

	var kind struct{ Type string }
	_ = json.Unmarshal(msg.Value, &kind)

	switch kind.Type {
	case "cred":
		var cred models.Credentials
		_ = json.Unmarshal(msg.Value, &cred)

		item.Kind = "cred"
		item.Key = cred.Login
		item.CreatedAt = cred.Created

	case "text":
		var text models.Text
		_ = json.Unmarshal(msg.Value, &text)

		item.Kind = "text"
		item.Key = text.Key
		item.CreatedAt = text.Created

	case "bin":
		var bin models.Binary
		_ = json.Unmarshal(msg.Value, &bin)

		item.Kind = "bin"
		item.Key = bin.Key
		item.CreatedAt = bin.Created

	case "card":
		var card models.Card
		_ = json.Unmarshal(msg.Value, &card)

		item.Kind = "card"
		item.Key = card.Number
		item.CreatedAt = card.Created
	}

	log.Info(
		"message converted",
		slog.String("msg value", string(msg.Value)),
	)

	return item
}

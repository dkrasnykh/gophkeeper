package service

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/dkrasnykh/gophkeeper/pkg/logger/sl"
	"github.com/dkrasnykh/gophkeeper/pkg/models"
)

type CredentialsStorager interface {
	All(ctx context.Context) ([]models.Credentials, error)
	ByLogin(ctx context.Context, login string) (models.Credentials, error)
	Save(ctx context.Context, cred models.Credentials) error
	Update(ctx context.Context, cred models.Credentials) error
}

type TextStorager interface {
	All(ctx context.Context) ([]models.Text, error)
	ByKey(ctx context.Context, key string) (models.Text, error)
	Save(ctx context.Context, text models.Text) error
	Update(ctx context.Context, text models.Text) error
}

type BinaryStorager interface {
	All(ctx context.Context) ([]models.Binary, error)
	ByKey(ctx context.Context, key string) (models.Binary, error)
	Save(ctx context.Context, bin models.Binary) error
	Update(ctx context.Context, bin models.Binary) error
}

type CardStorager interface {
	All(ctx context.Context) ([]models.Card, error)
	ByNumber(ctx context.Context, number string) (models.Card, error)
	Save(ctx context.Context, card models.Card) error
	Update(ctx context.Context, card models.Card) error
}

type Keeper struct {
	log       *slog.Logger
	ch        chan models.Message
	credStore CredentialsStorager
	textStore TextStorager
	binStore  BinaryStorager
	cardStore CardStorager
}

func NewKeeper(log *slog.Logger, ch chan models.Message, credStore CredentialsStorager,
	textStore TextStorager, binStore BinaryStorager, cardStore CardStorager) *Keeper {

	return &Keeper{
		log:       log,
		ch:        ch,
		credStore: credStore,
		textStore: textStore,
		binStore:  binStore,
		cardStore: cardStore,
	}
}

func (s *Keeper) ApplyMessage(ctx context.Context, msg models.Message) {
	switch msg.Type {
	case "update":
		s.apply(ctx, msg.Value)
	case "snapshot":
		var values [][]byte
		_ = json.Unmarshal(msg.Value, &values)

		for _, value := range values {
			s.apply(ctx, value)
		}
	}
}

func (s *Keeper) apply(ctx context.Context, value []byte) {
	const op = "service.Keeper.ApplyMessage"
	log := s.log.With(
		slog.String("op", op),
	)

	var header struct{ Type string }
	_ = json.Unmarshal(value, &header)

	switch header.Type {
	case "cred":
		var cred models.Credentials
		_ = json.Unmarshal(value, &cred)
		if err := s.saveCredentials(ctx, cred); err != nil {
			log.Error("apply credentials message error", sl.Err(err))
		}

	case "text":
		var text models.Text
		_ = json.Unmarshal(value, &text)
		if err := s.saveText(ctx, text); err != nil {
			log.Error("apply text message error", sl.Err(err))
		}

	case "bin":
		var bin models.Binary
		_ = json.Unmarshal(value, &bin)
		if err := s.saveBinary(ctx, bin); err != nil {
			log.Error("apply binary message error", sl.Err(err))
		}

	case "card":
		var card models.Card
		_ = json.Unmarshal(value, &card)
		if err := s.saveCard(ctx, card); err != nil {
			log.Error("apply card message error", sl.Err(err))
		}
	}
}

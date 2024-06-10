package service

import (
	"log/slog"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dkrasnykh/gophkeeper/internal/server/storage"
	"github.com/dkrasnykh/gophkeeper/pkg/hash"
	"github.com/dkrasnykh/gophkeeper/pkg/models"
)

func TestValidateOK(t *testing.T) {
	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	s := Service{log: log}
	data := []byte(`{"type":"text","tag":"tag1","key":"key1","value":"value 1","comment":"comment","created":1}`)
	msg := models.Message{Type: models.New, Value: data}

	validated, err := s.Validate(msg)

	require.NoError(t, err)
	assert.Equal(t, models.Update, validated.Type)
	assert.Equal(t, string(data), string(validated.Value))
}

func TestValidateFailCases(t *testing.T) {
	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	s := Service{log: log}
	tests := []struct {
		name        string
		data        []byte
		ecpectedMsg models.Message
		expectedErr error
	}{
		{
			name:        "invalid message type",
			data:        []byte(`{"type":"undefined","tag":"tag1","key":"key1","value":"value 1","comment":"comment","created":1}`),
			ecpectedMsg: models.Message{},
			expectedErr: ErrInvalidMessage,
		},
		{
			name:        "invalid message value",
			data:        []byte(`{"type":"text","tag":"tag1","key":5,"value":"value 1","comment":"comment","created":1}`),
			ecpectedMsg: models.Message{},
			expectedErr: ErrInvalidMessage,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validated, err := s.Validate(models.Message{Value: tt.data})

			require.Error(t, err)
			assert.ErrorIs(t, err, tt.expectedErr)
			assert.Equal(t, tt.ecpectedMsg, validated)
		})
	}
}

func TestConvertMessageToItemOK(t *testing.T) {
	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	key := "s5as4d5a#$%#%s6ad545##$%#4353KSFjH"
	s := Service{log: log, key: key}

	data := []byte(`{"type":"text","tag":"tag1","key":"key1","value":"value 1","comment":"comment","created":1}`)
	msg := models.Message{Value: data}
	expected := storage.Item{UserID: 1, Kind: hash.EncodeMsg([]byte("text"), key), Key: hash.EncodeMsg([]byte("key1"), key), Data: []byte(hash.EncodeMsg(msg.Value, key)), CreatedAt: 1}
	converted := s.convertMessageToItem(1, msg)

	assert.Equal(t, expected, converted)
}

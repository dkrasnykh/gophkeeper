package service

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

	"github.com/dkrasnykh/gophkeeper/internal/server/storage"
	mock_storage "github.com/dkrasnykh/gophkeeper/internal/server/storage/mocks"
	"github.com/dkrasnykh/gophkeeper/pkg/models"
)

func TestSnapshotOK(t *testing.T) {
	type mockBehavior func(r *mock_storage.MockStorager, userID int64)

	c := gomock.NewController(t)
	defer c.Finish()

	key := "s5as4d5a#$%#%s6ad545##$%#4353KSFjH"

	var behavior mockBehavior
	behavior = func(r *mock_storage.MockStorager, userID int64) {
		item1 := storage.Item{UserID: 1, Kind: "80c3314f39f125002c7c3274a7df6c0a747a7bb0", Key: "9fc3300add723e7db0dd2482907fa14c7e5be6de", Data: []byte("8f843d426a06cba17a7b7d0713ec9dbbcad7461f278302ecef9cb9bebd1365c3072d150d41d5653af14972db6bbde20d74e88d2025d30b3be9c053308c4c877e8b50d2399cc22916b28ece6a06ffd07b4864bd128708dfc74517aa74fef3ae1182cf80e25e924d8d0f1fc508d88d645cfa05b61d"), CreatedAt: 1717748173}
		item2 := storage.Item{UserID: 1, Kind: "97d42c5f258fc90849ea6210cf2d44fa4d5b9270", Key: "98c92e527452ad98dd7bd056253875e29de43b3eb37a", Data: []byte("8f843d426a06cba17a6c6a1a03ec9dbbcad7461f278302ecef9cb9bebd146fdd4c79155c06c03b7fb40535982beaf74e3db98c2e3bc24c30e99c1020905685698115d22f9cc22916b28ece6a15e7de3a447bb5169d198ac75304fe37aca5ee4cd7da89ae9c6f78aed30a2ac58b110a5488418bf74a7feda6ac39fbea0bb6"), CreatedAt: 1717748206}
		r.EXPECT().Snapshot(context.Background(), int64(1)).Return([]storage.Item{item1, item2}, nil)
	}

	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	repo := mock_storage.NewMockStorager(c)
	s := Service{log: log, key: key, storage: repo}
	behavior(repo, int64(1))

	msg, err := s.Snapshot(context.Background(), int64(1))
	require.NoError(t, err)
	require.Equal(t, models.Snapshot, msg.Type)
}

func TestSave(t *testing.T) {
	c := gomock.NewController(t)
	defer c.Finish()

	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	key := "s5as4d5a#$%#%s6ad545##$%#4353KSFjH"
	repo := mock_storage.NewMockStorager(c)
	s := Service{log: log, key: key, storage: repo}
	userID := int64(1)

	behavior := func(r *mock_storage.MockStorager, userID int64, msg models.Message) {
		item := s.convertMessageToItem(userID, msg)
		r.EXPECT().Save(context.Background(), item).Return(nil)
	}

	msg := models.Message{Type: models.New, Value: []byte(`{"type":"text","tag":"tag1","key":"key1","value":"value 1","comment":"comment","created":1}`)}

	behavior(repo, userID, msg)

	err := s.Save(context.Background(), userID, msg)
	require.NoError(t, err)
}

func TestSaveError(t *testing.T) {
	c := gomock.NewController(t)
	defer c.Finish()

	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	key := "key"
	repo := mock_storage.NewMockStorager(c)
	s := Service{log: log, key: key, storage: repo}
	userID := int64(1)

	behavior := func(r *mock_storage.MockStorager, userID int64, msg models.Message) {
		item := s.convertMessageToItem(userID, msg)
		r.EXPECT().Save(context.Background(), item).Return(errors.New("saving db error"))
	}

	msg := models.Message{Type: models.New, Value: []byte(`{"type":"text","tag":"tag1","key":"key1","value":"value 1","comment":"comment","created":1}`)}

	behavior(repo, userID, msg)

	err := s.Save(context.Background(), userID, msg)
	require.ErrorIs(t, err, ErrInternal)
}

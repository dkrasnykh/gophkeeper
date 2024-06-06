package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/dkrasnykh/gophkeeper/pkg/models"
)

var (
	text1 = models.Text{Type: models.TextItem, Tag: "tag1", Key: "key1", Value: "value 1", Comment: "comment", Created: 1}
	text2 = models.Text{Type: models.TextItem, Tag: "tag1", Key: "key1", Value: "value 2", Comment: "comment 2", Created: 2}
	cred1 = models.Credentials{Type: models.CredItem, Tag: "tag1", Login: "login1", Password: "pass1", Comment: "credentials comment", Created: 1}
	cred2 = models.Credentials{Type: models.CredItem, Tag: "tag1", Login: "login1", Password: "pass2", Comment: "credentials comment 2", Created: 2}
)

type Storager interface {
	Snapshot(ctx context.Context, userID int64) ([]Item, error)
	Save(ctx context.Context, item Item) error
}

type testStorager interface {
	Storager
	clean(ctx context.Context) error
}

type PostgresTestSuite struct {
	suite.Suite
	testStorager

	tc *tcpostgres.PostgresContainer
}

func (ts *PostgresTestSuite) SetupSuite() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	pgc, err := tcpostgres.RunContainer(ctx,
		testcontainers.WithImage("docker.io/postgres:latest"),
		tcpostgres.WithDatabase("testdb"),
		tcpostgres.WithUsername("postgres"),
		tcpostgres.WithPassword("postgres"),
		tcpostgres.WithInitScripts(),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(10*time.Second),
		),
	)

	require.NoError(ts.T(), err)

	host, err := pgc.Host(ctx)
	require.NoError(ts.T(), err)

	port, err := pgc.MappedPort(ctx, "5432")
	require.NoError(ts.T(), err)

	ts.tc = pgc
	databaseURL := fmt.Sprintf("postgres://postgres:postgres@%s:%s/testdb?sslmode=disable", host, port.Port())

	db, err := New(databaseURL, time.Second*10)
	storage := NewKeeperPostgres(db, time.Second*10)
	require.NoError(ts.T(), err)

	ts.testStorager = storage

	ts.T().Logf("stared postgres at %s:%d", host, port.Int())
}

func (s *KeeperPostgres) clean(ctx context.Context) error {
	newCtx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()

	_, err := s.db.Exec(newCtx, "DELETE FROM store")
	return err
}

func (ts *PostgresTestSuite) TearDownSuite() {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	require.NoError(ts.T(), ts.tc.Terminate(ctx))
}

func TestPostgres(t *testing.T) {
	suite.Run(t, new(PostgresTestSuite))
}

func (ts *PostgresTestSuite) SetupTest() {
	ts.Require().NoError(ts.clean(context.Background()))
}

func (ts *PostgresTestSuite) TearDownTest() {
	ts.Require().NoError(ts.clean(context.Background()))
}

func (ts *PostgresTestSuite) TestSave() {
	userID := int64(1)
	data, _ := json.Marshal(text1)
	itemToSave := Item{UserID: userID, Kind: "text", Key: "key1", Data: data, CreatedAt: 1}

	ts.NoError(ts.Save(context.Background(), itemToSave))

	savedItems, err := ts.Snapshot(context.Background(), userID)
	ts.NoError(err)
	ts.Equal(len(savedItems), 1)
	ts.Equal(itemToSave, savedItems[0])
}

func (ts *PostgresTestSuite) TestSnapshotWithAnotherUserID() {
	userID1, userID2 := int64(1), int64(2)
	data, _ := json.Marshal(text1)
	ts.NoError(ts.Save(context.Background(), Item{UserID: userID1, Kind: text1.Type.String(), Key: text1.Key, Data: data, CreatedAt: text1.Created}))
	data, _ = json.Marshal(cred1)
	ts.NoError(ts.Save(context.Background(), Item{UserID: userID1, Kind: cred1.Type.String(), Key: cred1.Login, Data: data, CreatedAt: cred1.Created}))

	savedItems, err := ts.Snapshot(context.Background(), userID2)
	ts.NoError(err)
	ts.Equal(len(savedItems), 0)
}

func (ts *PostgresTestSuite) TestSnapshot() {
	userID := int64(1)
	data, _ := json.Marshal(text2)
	itemText2 := Item{UserID: userID, Kind: string(text2.Type), Key: text2.Key, Data: data, CreatedAt: text2.Created}
	ts.NoError(ts.Save(context.Background(), itemText2))
	data, _ = json.Marshal(text1)
	ts.NoError(ts.Save(context.Background(), Item{UserID: userID, Kind: text1.Type.String(), Key: text1.Key, Data: data, CreatedAt: text1.Created}))
	data, _ = json.Marshal(cred1)
	ts.NoError(ts.Save(context.Background(), Item{UserID: userID, Kind: cred1.Type.String(), Key: cred1.Login, Data: data, CreatedAt: cred1.Created}))
	data, _ = json.Marshal(cred2)
	itemCred2 := Item{UserID: userID, Kind: cred2.Type.String(), Key: cred2.Login, Data: data, CreatedAt: cred2.Created}
	ts.NoError(ts.Save(context.Background(), itemCred2))

	savedItems, err := ts.Snapshot(context.Background(), userID)
	ts.NoError(err)
	ts.Equal(len(savedItems), 2)
	ts.True(contains(itemText2, savedItems))
	ts.True(contains(itemCred2, savedItems))
}

func contains(target Item, items []Item) bool {
	for _, item := range items {
		if target.Kind == item.Kind && target.UserID == target.UserID &&
			target.Key == item.Key && string(target.Data) == string(item.Data) &&
			target.CreatedAt == item.CreatedAt {
			return true
		}
	}
	return false
}

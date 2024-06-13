package storage

import (
	"context"
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

type UserProvider interface {
	SaveUser(ctx context.Context, email string, passHash []byte) (int64, error)
	User(ctx context.Context, email string) (models.User, error)
}

type testUserProvider interface {
	UserProvider
	clean(ctx context.Context) error
}

type UserPostgresTestSuite struct {
	suite.Suite
	testUserProvider

	tc *tcpostgres.PostgresContainer
}

func (ts *UserPostgresTestSuite) SetupSuite() {
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

	err = Migrate(databaseURL, time.Second*10)
	require.NoError(ts.T(), err)
	storage, err := NewUserPostgres(databaseURL, time.Second*10)
	require.NoError(ts.T(), err)

	ts.testUserProvider = storage

	ts.T().Logf("stared postgres at %s:%d", host, port.Int())
}

func (s *UserPostgres) clean(ctx context.Context) error {
	newCtx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()

	_, err := s.db.Exec(newCtx, "DELETE FROM users")
	return err
}

func (ts *UserPostgresTestSuite) TearDownSuite() {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	require.NoError(ts.T(), ts.tc.Terminate(ctx))
}

func TestUserPostgres(t *testing.T) {
	suite.Run(t, new(UserPostgresTestSuite))
}

func (ts *UserPostgresTestSuite) SetupTest() {
	ts.Require().NoError(ts.clean(context.Background()))
}

func (ts *UserPostgresTestSuite) TearDownTest() {
	ts.Require().NoError(ts.clean(context.Background()))
}

func (ts *UserPostgresTestSuite) TestSaveUser() {
	email, passHash := "name@example.com", []byte("hash")

	userID, err := ts.SaveUser(context.Background(), email, passHash)
	ts.NoError(err)

	saved, err := ts.User(context.Background(), email)
	ts.NoError(err)
	ts.Equal(userID, saved.ID)
	ts.Equal(email, saved.Email)
	ts.Equal(string(passHash), string(saved.PassHash))
}

func (ts *UserPostgresTestSuite) TestSaveUserDuplicateEmail() {
	email, passHash1 := "name@example.com", []byte("hash")
	passHash2 := []byte("hash2")

	_, err := ts.SaveUser(context.Background(), email, passHash1)
	ts.NoError(err)
	_, err = ts.SaveUser(context.Background(), email, passHash2)
	ts.ErrorIs(err, ErrUserExists)
}

func (ts *UserPostgresTestSuite) TestUserNotFound() {
	email := "name@example.com"

	saved, err := ts.User(context.Background(), email)
	ts.ErrorIs(err, ErrUserNotFound)
	ts.Equal(models.User{}, saved)
}

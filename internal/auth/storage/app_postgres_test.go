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

type AppProvider interface {
	App(ctx context.Context, id int) (models.App, error)
}

type PostgresTestSuite struct {
	suite.Suite
	AppProvider

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

	err = Migrate(databaseURL, time.Second*10)
	require.NoError(ts.T(), err)
	storage, err := NewAppPostgres(databaseURL, time.Second*10)
	require.NoError(ts.T(), err)

	ts.AppProvider = storage

	ts.T().Logf("stared postgres at %s:%d", host, port.Int())
}

func TestPostgres(t *testing.T) {
	suite.Run(t, new(PostgresTestSuite))
}

func (ts *PostgresTestSuite) TestApp() {
	appID := 1
	app, err := ts.App(context.Background(), appID)
	ts.NoError(err)
	ts.Equal("gophkeeper", app.Name)
}

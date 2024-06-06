package auth

import (
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	grpcapp "github.com/dkrasnykh/gophkeeper/internal/auth/grpc"
	"github.com/dkrasnykh/gophkeeper/internal/auth/service"
	"github.com/dkrasnykh/gophkeeper/internal/auth/storage"
)

type App struct {
	grpcApp *grpcapp.App
	db      *pgxpool.Pool
}

func New(log *slog.Logger, grpcPort int, databaseURL string, tokenTTL time.Duration, timeout time.Duration, certFile string, keyFile string) *App {
	db, err := storage.New(databaseURL, timeout)
	if err != nil {
		panic(err)
	}

	userStorage := storage.NewUserPostgres(db, timeout)
	appStorage := storage.NewAppPostgres(db, timeout)
	authService := service.New(log, userStorage, appStorage, tokenTTL)

	grpcApp, err := grpcapp.New(log, authService, grpcPort, certFile, keyFile)
	if err != nil {
		panic(err)
	}

	return &App{
		grpcApp: grpcApp,
		db:      db,
	}
}

func (app *App) MustRun() {
	app.grpcApp.MustRun()
}

func (app *App) Stop() {
	app.grpcApp.Stop()
	app.db.Close()
}

package auth

import (
	"log/slog"

	"github.com/dkrasnykh/gophkeeper/internal/auth/config"
	grpcapp "github.com/dkrasnykh/gophkeeper/internal/auth/grpc"
	"github.com/dkrasnykh/gophkeeper/internal/auth/service"
	"github.com/dkrasnykh/gophkeeper/internal/auth/storage"
)

// App with GRPC server and database connections pool.
// Provides start/stop methods.
type App struct {
	grpcApp *grpcapp.App
}

func New(log *slog.Logger, cfg *config.Config) (*App, error) {
	err := storage.Migrate(cfg.DatabaseURL, cfg.ConnectTimeout)
	if err != nil {
		return nil, err
	}

	userStorage, err := storage.NewUserPostgres(cfg.DatabaseURL, cfg.ConnectTimeout)
	if err != nil {
		return nil, err
	}
	appStorage, err := storage.NewAppPostgres(cfg.DatabaseURL, cfg.ConnectTimeout)
	if err != nil {
		return nil, err
	}
	authService := service.New(log, userStorage, appStorage, cfg.TokenTTL)

	grpcApp, err := grpcapp.New(log, authService, cfg)
	if err != nil {
		return nil, err
	}

	return &App{
		grpcApp: grpcApp,
	}, nil
}

func (app *App) MustRun() {
	app.grpcApp.MustRun()
}

func (app *App) Stop() {
	app.grpcApp.Stop()
}

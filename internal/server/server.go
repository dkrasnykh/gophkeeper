package server

import (
	"log/slog"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/dkrasnykh/gophkeeper/internal/server/clients"
	"github.com/dkrasnykh/gophkeeper/internal/server/config"
	"github.com/dkrasnykh/gophkeeper/internal/server/handler"
	"github.com/dkrasnykh/gophkeeper/internal/server/service"
	"github.com/dkrasnykh/gophkeeper/internal/server/storage"
)

type App struct {
	db *pgxpool.Pool
}

func Run(log *slog.Logger, cfg *config.Config) {
	// TODO gracefull shutdown
	db, err := storage.New(cfg.DatabaseURL, cfg.QueryTimeout)
	if err != nil {
		panic(err)
	}
	storageKeeper := storage.NewKeeperPostgres(db, cfg.QueryTimeout)
	serviceKeeper := service.New(log, storageKeeper, cfg.Key)
	conns := clients.NewUserWSConnMap()
	h := handler.NewHandler(log, serviceKeeper, conns)

	http.HandleFunc("/ws", h.Handle)

	err = http.ListenAndServeTLS(cfg.WS.Address, cfg.CertFile, cfg.KeyFile, nil)
	if err != nil {
		panic(err)
	}
}

package server

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/dkrasnykh/gophkeeper/internal/server/clients"
	"github.com/dkrasnykh/gophkeeper/internal/server/handler"
	"github.com/dkrasnykh/gophkeeper/internal/server/service"
	"github.com/dkrasnykh/gophkeeper/internal/server/storage"
)

type App struct {
	db *pgxpool.Pool
}

func Run(log *slog.Logger, wsAddress string, databaseURL string, timeout time.Duration) {
	// TODO gracefull shutdown
	db, err := storage.New(databaseURL, timeout)
	if err != nil {
		panic(err)
	}
	storageKeeper := storage.NewKeeperPostgres(db, timeout)
	serviceKeeper := service.New(log, storageKeeper)
	conns := clients.NewUserWSConnMap()
	h := handler.NewHandler(log, serviceKeeper, conns)

	http.HandleFunc("/ws", h.Handle)

	err = http.ListenAndServe(wsAddress, nil)
	if err != nil {
		panic(err)
	}
}

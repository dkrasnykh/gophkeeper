package main

import (
	"log/slog"
	"os"

	"github.com/dkrasnykh/gophkeeper/internal/server"
	"github.com/dkrasnykh/gophkeeper/internal/server/config"
)

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	cfg := config.MustLoad()
	log.Debug("starting application", slog.Any("config", cfg))

	server.Run(log, cfg)

	// TODO gracefull shutdown
}

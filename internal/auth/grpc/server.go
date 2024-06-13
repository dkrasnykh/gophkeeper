// grpcapp provides GRPC server configuration and start/shutdown methods.
package grpcapp

import (
	"errors"
	"fmt"
	"log/slog"
	"net"

	"google.golang.org/grpc"

	"github.com/dkrasnykh/gophkeeper/internal/auth/config"
	"github.com/dkrasnykh/gophkeeper/internal/auth/tls"
	"github.com/dkrasnykh/gophkeeper/pkg/logger/sl"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
)

type App struct {
	log        *slog.Logger
	gRPCServer *grpc.Server
	service    Auth
	port       int
}

func New(log *slog.Logger, authService Auth, cfg *config.Config) (*App, error) {
	tlsCredentials, err := tls.LoadTLSCredentials(cfg.CertFile, cfg.KeyFile)
	if err != nil {
		log.Error(
			"failed to load cert and key from files",
			sl.Err(err),
			slog.String("cert file path", cfg.CertFile),
			slog.String("key file path", cfg.KeyFile),
		)
		return nil, errors.New("get TLS credentials error")
	}

	gRPCServer := grpc.NewServer(
		grpc.UnaryInterceptor(logging.UnaryServerInterceptor(&rpcLogger{log: log})),
		grpc.Creds(tlsCredentials),
	)
	Register(gRPCServer, authService)
	return &App{
		log:        log,
		gRPCServer: gRPCServer,
		service:    authService,
		port:       cfg.GRPC.Port,
	}, nil
}

func (a *App) MustRun() {
	if err := a.Run(); err != nil {
		a.log.Error(
			"error running GRPC server",
			sl.Err(err),
		)
		return
	}
}

func (a *App) Run() error {
	const op = "grpcapp.Run"
	log := a.log.With(
		slog.String("op", op),
		slog.Int("port", a.port),
	)

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", a.port))
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	log.Info("gRPC server is running", slog.String("addr", listener.Addr().String()))

	if err := a.gRPCServer.Serve(listener); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

// Stop graceful stopped GRPC server
func (a *App) Stop() {
	const op = "grpcapp.Stop"

	a.log.With(slog.String("op", op)).
		Info("stopping gRPC server", slog.Int("port", a.port))

	a.gRPCServer.GracefulStop()
	a.service.Close()
}

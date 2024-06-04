package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/dkrasnykh/gophkeeper/internal/auth/storage"
	"github.com/dkrasnykh/gophkeeper/pkg/jwt"
	"github.com/dkrasnykh/gophkeeper/pkg/logger/sl"
	"github.com/dkrasnykh/gophkeeper/pkg/models"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserExists         = errors.New("user already exists")
)

type UserProvider interface {
	SaveUser(ctx context.Context, email string, passHash []byte) (int64, error)
	User(ctx context.Context, email string) (models.User, error)
}

type AppProvider interface {
	App(ctx context.Context, id int) (models.App, error)
}

type Auth struct {
	log          *slog.Logger
	userProvider UserProvider
	appProvider  AppProvider
	tokenTTL     time.Duration
}

func New(log *slog.Logger, userProvider UserProvider, appProvider AppProvider, tokenTTL time.Duration) *Auth {
	return &Auth{
		log:          log,
		userProvider: userProvider,
		appProvider:  appProvider,
		tokenTTL:     tokenTTL,
	}
}

func (a *Auth) Register(ctx context.Context, email string, password string) (userID int64, err error) {
	const op = "auth.Register"
	log := a.log.With(
		slog.String("op", op),
		slog.String("email", email),
	)

	log.Info("registering user")

	passHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Error("failed to generate password", sl.Err(err))
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	id, err := a.userProvider.SaveUser(ctx, email, passHash)
	if err != nil {
		if errors.Is(err, storage.ErrUserExists) {
			log.Error("user already exists", sl.Err(err))
			return 0, fmt.Errorf("%s: %w", op, ErrUserExists)
		}
		log.Error("failed to save user", sl.Err(err))
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("user registered")
	return id, nil
}

func (a *Auth) Login(ctx context.Context, email string, password string, appID int) (string, error) {
	const op = "auth.Login"
	log := a.log.With(
		slog.String("op", op),
		slog.String("username", email),
	)

	log.Info("attempting to login user")

	user, err := a.userProvider.User(ctx, email)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			log.Error("user not found", sl.Err(err))
			return "", fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
		}
		log.Error("failed to get user", sl.Err(err))
		return "", fmt.Errorf("%s: %w", op, err)
	}

	if err := bcrypt.CompareHashAndPassword(user.PassHash, []byte(password)); err != nil {
		log.Info("invalid credentials", sl.Err(err))
		return "", fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
	}

	app, err := a.appProvider.App(ctx, appID)
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}
	log.Info("user logged in successfully")

	token, err := jwt.NewToken(user, app, a.tokenTTL)
	if err != nil {
		log.Error("failed to generate token", sl.Err(err))
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return token, nil
}

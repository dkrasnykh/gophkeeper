package client

import (
	"context"
	"database/sql"
	"log/slog"
	"os"
	"strconv"
	"syscall"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	viewaddbinary "github.com/dkrasnykh/gophkeeper/internal/client/cli/view_add_binary"
	viewaddcard "github.com/dkrasnykh/gophkeeper/internal/client/cli/view_add_card"
	viewaddcredentials "github.com/dkrasnykh/gophkeeper/internal/client/cli/view_add_credentials"
	viewaddtext "github.com/dkrasnykh/gophkeeper/internal/client/cli/view_add_text"
	viewauth "github.com/dkrasnykh/gophkeeper/internal/client/cli/view_auth"
	"github.com/dkrasnykh/gophkeeper/internal/client/cli/view_command_list"
	viewlist "github.com/dkrasnykh/gophkeeper/internal/client/cli/view_list"
	viewlogin "github.com/dkrasnykh/gophkeeper/internal/client/cli/view_login"
	viewregister "github.com/dkrasnykh/gophkeeper/internal/client/cli/view_register"
	"github.com/dkrasnykh/gophkeeper/internal/client/grpcclient"
	"github.com/dkrasnykh/gophkeeper/internal/client/service"
	"github.com/dkrasnykh/gophkeeper/internal/client/storage"
	"github.com/dkrasnykh/gophkeeper/internal/client/ws"
	"github.com/dkrasnykh/gophkeeper/pkg/logger/sl"
	"github.com/dkrasnykh/gophkeeper/pkg/models"
)

type AppClient struct {
	db           *sql.DB
	ch           chan models.Message
	log          *slog.Logger
	grpcClient   *grpcclient.GRPCClient
	storagePath  string
	grpcAddress  string
	WSURL        string
	queryTimeout time.Duration
	caCertFile   string
}

func NewAppClient(log *slog.Logger, storagePath string, grpcAddress string,
	WSURL string, queryTimeout time.Duration, caCertFile string) *AppClient {
	return &AppClient{
		log:          log,
		storagePath:  storagePath,
		grpcAddress:  grpcAddress,
		WSURL:        WSURL,
		queryTimeout: queryTimeout,
		caCertFile:   caCertFile,
	}
}

func (app *AppClient) Stop() {
	err := app.db.Close()
	if err != nil {
		return
	}
	close(app.ch)
	app.grpcClient.Stop()
}

func (app *AppClient) Run(ctx context.Context, stop chan os.Signal) {
	const op = "client.Run"
	log := app.log.With(
		slog.String("op", op),
	)

	app.ch = make(chan models.Message)

	var err error
	app.db, err = storage.New(app.storagePath)
	if err != nil {
		log.Error(
			"failed open database connection",
			sl.Err(err),
		)
		stop <- syscall.SIGTERM
		return
	}

	dbCred := storage.NewCredentialsSqlite(app.db, app.queryTimeout)
	dbText := storage.NewTextSqlite(app.db, app.queryTimeout)
	dbBin := storage.NewBinarySqlite(app.db, app.queryTimeout)
	dbCard := storage.NewCardSqlite(app.db, app.queryTimeout)

	keeper := service.NewKeeper(log, app.ch, dbCred, dbText, dbBin, dbCard)

	app.grpcClient, err = grpcclient.NewGRPCClient(app.grpcAddress, app.caCertFile)
	if err != nil {
		log.Error(
			"failed connect to GRPC auth server",
			slog.String("GRPC address", app.grpcAddress),
			sl.Err(err),
		)
		stop <- syscall.SIGTERM
		return
	}

	p := tea.NewProgram(viewauth.Model{})
	m, _ := p.Run()

	modelAuth, _ := m.(viewauth.Model)
	if modelAuth.Choice == "" {
		// user stopped execution in UI (q, ctrl+C, esc)
		log.Info("user stopped execution (q, ctrl+C, esc)")
		stop <- syscall.SIGTERM
		return
	}
	if modelAuth.Choice == "Register" {
	loop:
		for {
			select {
			case <-ctx.Done():
				return
			default:
				p = tea.NewProgram(viewregister.InitialModel(app.grpcClient))
				m, _ = p.Run()

				modelRegister, _ := m.(viewregister.Model)

				if modelRegister.State == "" {
					// user stopped execution in UI (q, ctrl+C, esc)
					log.Info("user stopped execution (q, ctrl+C, esc)")
					stop <- syscall.SIGTERM
					return
				}
				if modelRegister.State == "again" {
					continue
				}
				if modelRegister.State == "error" {
					log.Error("registration failed")
					stop <- syscall.SIGTERM
					return
				}
				break loop
			}
		}

	}
	var token string
	//login
	for {
		select {
		case <-ctx.Done():
			return
		default:
			p = tea.NewProgram(viewlogin.InitialModel(app.grpcClient))
			m, _ = p.Run()

			modelLogin, _ := m.(viewlogin.Model)

			if modelLogin.State == "" {
				// user stopped execution in UI (q, ctrl+C, esc)
				log.Info("user stopped execution (q, ctrl+C, esc)")
				stop <- syscall.SIGTERM
				return
			}
			if modelLogin.State == "again" {
				continue
			}
			if modelLogin.State == "error" {
				log.Error("")
				return
			}
			token = modelLogin.Token
			break
		}
		if token != "" {
			break
		}
	}

	if token == "" {
		log.Error("token is empty, stopping execution")
		stop <- syscall.SIGTERM
		return
	}

	wsClient := ws.NewWSClient(log, app.ch, keeper, app.WSURL)

	// interrupt - chan for receiving signal from the websocket connection (receive error message from server)
	interrupt := make(chan struct{})
	go func(interrupt chan struct{}) {
		<-interrupt
		stop <- syscall.SIGTERM
	}(interrupt)

	wsClient.Run(ctx, interrupt, token)

	for {
		select {
		case <-ctx.Done():
			return
		default:
			p := tea.NewProgram(view_command_list.Model{})
			m, _ := p.Run()

			modelComandList, _ := m.(view_command_list.Model)

			if modelComandList.Choice == "" {
				// user stopped execution in UI (q, ctrl+C, esc)
				log.Info("user stopped execution (q, ctrl+C, esc)")
				stop <- syscall.SIGTERM
				return
			}
			switch modelComandList.Choice {
			case "Get all secrets":
				creds, err := keeper.AllCredentials(ctx)
				if err != nil {
					log.Error("query all credentials error", sl.Err(err))
				}
				texts, err := keeper.AllText(ctx)
				if err != nil {
					log.Error("query all text data error", sl.Err(err))
				}
				bins, err := keeper.AllBinary(ctx)
				if err != nil {
					log.Error("query all binary data error", sl.Err(err))
				}
				cards, err := keeper.AllCard(ctx)
				if err != nil {
					log.Error("query all cards error", sl.Err(err))
				}
				// view result
				p := tea.NewProgram(viewlist.Model{Msg: viewlist.Convert(creds, texts, bins, cards)})
				_, _ = p.Run()

			case "Add credentials":
				p = tea.NewProgram(viewaddcredentials.InitialModel())
				m, _ = p.Run()

				modelAddCredentials, _ := m.(viewaddcredentials.Model)

				if modelAddCredentials.State == "quit" {
					// user stopped execution in UI (q, ctrl+C, esc)
					log.Info("user stopped execution (q, ctrl+C, esc)")
					stop <- syscall.SIGTERM
					return
				}
				cred := models.Credentials{
					Type:     "cred",
					Tag:      modelAddCredentials.Inputs[0].Value(),
					Login:    modelAddCredentials.Inputs[1].Value(),
					Password: modelAddCredentials.Inputs[2].Value(),
					Comment:  modelAddCredentials.Inputs[3].Value(),
					Created:  time.Now().Unix(),
				}
				// TODO validate item
				err := keeper.SendSaveCredentials(ctx, cred)
				if err != nil {
					// TODO view result
					log.Error("saving credentials error", sl.Err(err))
				}

			case "Add text data":
				p = tea.NewProgram(viewaddtext.InitialModel())
				m, _ = p.Run()

				modelAddText, _ := m.(viewaddtext.Model)

				if modelAddText.State == "quit" {
					// user stopped execution in UI (q, ctrl+C, esc)
					log.Info("user stopped execution (q, ctrl+C, esc)")
					stop <- syscall.SIGTERM
					return
				}
				text := models.Text{
					Type:    "text",
					Tag:     modelAddText.Inputs[0].Value(),
					Key:     modelAddText.Inputs[1].Value(),
					Value:   modelAddText.Inputs[2].Value(),
					Comment: modelAddText.Inputs[3].Value(),
					Created: time.Now().Unix(),
				}
				// TODO validate item
				err := keeper.SendSaveText(ctx, text)
				if err != nil {
					// TODO view result
					log.Error("saving text error", sl.Err(err))
				}
			case "Add binary data":
				p = tea.NewProgram(viewaddbinary.InitialModel())
				m, _ = p.Run()

				modelAddBinary, _ := m.(viewaddbinary.Model)

				if modelAddBinary.State == "quit" {
					// user stopped execution in UI (q, ctrl+C, esc)
					log.Info("user stopped execution (q, ctrl+C, esc)")
					stop <- syscall.SIGTERM
					return
				}
				path := modelAddBinary.Inputs[1].Value()
				fileName, data, err := keeper.ExtractDataFromFile(path)
				if err != nil {
					log.Error(
						"failed extract data from file",
						slog.String("path", path),
						sl.Err(err),
					)
					continue
				}
				bin := models.Binary{
					Type:    "bin",
					Tag:     modelAddBinary.Inputs[0].Value(),
					Key:     fileName,
					Value:   data,
					Comment: modelAddBinary.Inputs[2].Value(),
					Created: time.Now().Unix(),
				}
				// TODO validate binary item
				err = keeper.SendSaveBinary(ctx, bin)
				if err != nil {
					// TODO view result
					log.Error("saving binary data error", sl.Err(err))
				}
			case "Add card data":
				p = tea.NewProgram(viewaddcard.InitialModel())
				m, _ = p.Run()

				modelAddCard, _ := m.(viewaddcard.Model)

				if modelAddCard.State == "quit" {
					// user stopped execution in UI (q, ctrl+C, esc)
					log.Info("user stopped execution (q, ctrl+C, esc)")
					stop <- syscall.SIGTERM
					return
				}
				cvv, _ := strconv.Atoi(modelAddCard.Inputs[3].Value())
				card := models.Card{
					Type:    "card",
					Tag:     modelAddCard.Inputs[0].Value(),
					Number:  modelAddCard.Inputs[1].Value(),
					Exp:     modelAddCard.Inputs[2].Value(),
					CVV:     int32(cvv),
					Comment: modelAddCard.Inputs[4].Value(),
					Created: time.Now().Unix(),
				}
				// TODO validate card item
				err := keeper.SendSaveCard(ctx, card)
				if err != nil {
					// TODO view result
					log.Error("saving card data error", sl.Err(err))
				}
			}
		}
	}
}

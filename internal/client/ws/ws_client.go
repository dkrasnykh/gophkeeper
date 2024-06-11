// ws module establish websocket connection with server
// It starts two gorutine for reading and writing messages.
// When the client just establishes a connection, client received from server actual data snapshot.
// If user saved new private data, ws sends to the server update.
// If same user used other client and makes changes, then current client receives update message.
package ws

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"log/slog"

	"github.com/gorilla/websocket"

	"github.com/dkrasnykh/gophkeeper/pkg/logger/sl"
	"github.com/dkrasnykh/gophkeeper/pkg/models"
)

var (
	ErrConnectToServer = errors.New("failed establish websocket connection")
)

type MessageService interface {
	ApplyMessage(ctx context.Context, msg models.Message)
}

type WSClient struct {
	log  *slog.Logger
	conn *websocket.Conn
	ch   chan models.Message
	s    MessageService
	url  string
}

func NewWSClient(log *slog.Logger, ch chan models.Message, s MessageService, url string) *WSClient {
	return &WSClient{
		log: log,
		ch:  ch,
		s:   s,
		url: url,
	}
}

func (ws *WSClient) Run(ctx context.Context, interrupt chan struct{}, token string) {
	const op = "ws.Run"
	log := ws.log.With(
		slog.String("op", op),
	)

	dialer := *websocket.DefaultDialer
	dialer.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	headers := make(map[string][]string)
	headers["token"] = append(headers["token"], token)

	var err error
	ws.conn, _, err = dialer.DialContext(ctx, ws.url, headers)

	if err != nil {
		log.Error(
			"failed establish websocket connection",
			sl.Err(err),
		)
		interrupt <- struct{}{}
		return
	}

	go ws.read(ctx, interrupt)
	go ws.write(ctx, token, interrupt)
}

func (ws *WSClient) read(ctx context.Context, interrupt chan struct{}) {
	op := "ws.Run.read"
	log := ws.log.With(
		slog.String("op", op),
	)

	for {
		select {
		case <-ctx.Done():
			log.Info(
				"receive context done message",
			)
			return

		default:
			var header struct{ Type string }
			mt, data, err := ws.conn.ReadMessage()
			if err != nil {
				// TODO implement restoring connection to the server
				log.Error(
					"error receiving message from server",
					slog.String("address", ws.conn.RemoteAddr().String()),
					sl.Err(err),
				)
				return
			}
			if mt != websocket.TextMessage {
				continue
			}
			err = json.Unmarshal(data, &header)
			if err != nil || (header.Type != "update" && header.Type != "snapshot" && header.Type != "error") {
				continue
			}
			var msg models.Message
			err = json.Unmarshal(data, &msg)
			if err != nil {
				log.Warn(
					"receiving unexpected message from server",
					slog.String("message value", string(msg.Value)),
					sl.Err(err),
				)
				continue
			}
			if msg.Type == "error" && string(msg.Value) == "invalid token" {
				interrupt <- struct{}{}
				return
			}

			ws.s.ApplyMessage(ctx, msg)
		}
	}
}

func (ws *WSClient) write(ctx context.Context, token string, interrupt chan struct{}) {
	op := "ws.Run.write"
	log := ws.log.With(
		slog.String("op", op),
	)

	for {
		select {
		case <-ctx.Done():
			log.Info("recieve context done message")

			return
		case msg := <-ws.ch:
			msg.Token = token
			data, _ := json.Marshal(msg)
			err := ws.conn.WriteMessage(websocket.TextMessage, data)
			if err != nil {
				// TODO implement restoring connection to the server
				interrupt <- struct{}{}
				return
			}

		}
	}
}

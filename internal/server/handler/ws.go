package handler

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/gorilla/websocket"

	"github.com/dkrasnykh/gophkeeper/internal/server/clients"
	"github.com/dkrasnykh/gophkeeper/internal/server/lib"
	"github.com/dkrasnykh/gophkeeper/pkg/logger/sl"
	"github.com/dkrasnykh/gophkeeper/pkg/models"
)

type IService interface {
	Snapshot(ctx context.Context, userID int64) (models.Message, error)
	Save(ctx context.Context, userID int64, msg models.Message) error
	Validate(msg models.Message) (models.Message, error)
}

type Handler struct {
	log        *slog.Logger
	service    IService
	wsUpgrader *websocket.Upgrader
	conns      *clients.UserWSConnMap
}

func NewHandler(log *slog.Logger, s IService, conns *clients.UserWSConnMap) *Handler {
	return &Handler{
		log:        log,
		service:    s,
		wsUpgrader: &websocket.Upgrader{},
		conns:      conns,
	}
}

func (h *Handler) Handle(w http.ResponseWriter, r *http.Request) {
	op := "ws.Handle"
	log := h.log.With(
		slog.String("op", op),
	)

	ctx := r.Context()

	conn, err := h.wsUpgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Error(
			"failed establish websocket connection",
			sl.Err(err),
		)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	token := r.Header.Get("token")

	if token == "" {
		log.Error(
			"request does not contains token, interupt websocket connection",
			sl.Err(err),
		)
		w.WriteHeader(http.StatusBadRequest)
		_ = conn.Close()
		return
	}

	userID, err := lib.ParseToken(token)
	if err != nil {
		log.Error(
			"invalid token",
			slog.String("token", token),
			sl.Err(err),
		)
		w.WriteHeader(http.StatusBadRequest)
		_ = conn.Close()
		return
	}

	h.conns.Put(userID, conn)
	snapshot, err := h.service.Snapshot(ctx, userID)
	if err != nil {
		log.Error(
			"failed collect init snapshot data for user",
			slog.Int64("user_id", userID),
			sl.Err(err),
		)
		errMsg := models.Message{Type: "error", Value: []byte("failed collect init snapshot data")}
		errMsgText, _ := json.Marshal(errMsg)
		err = conn.WriteMessage(websocket.TextMessage, errMsgText)

		if err != nil {
			// TODO handle interrupted connection with client
			log.Error(
				"error sending message to user",
				slog.Int64("user_id", userID),
				slog.String("address", conn.RemoteAddr().String()),
				sl.Err(err),
			)
		}
	}
	msg, _ := json.Marshal(snapshot)
	err = conn.WriteMessage(websocket.TextMessage, msg)
	if err != nil {
		// TODO handle interrupted connection with client
		log.Error(
			"error sending message to user",
			slog.Int64("user_id", userID),
			slog.String("address", conn.RemoteAddr().String()),
			sl.Err(err),
		)
	}

	for {
		select {
		case <-ctx.Done():
			log.Info("client logged out")
			// TODO clear user_id - conn map
			return
		default:
			mt, data, err := conn.ReadMessage()
			if err != nil {
				// TODO handle interrupted connection with client
				log.Error(
					"error listening client connection",
					slog.Int64("user_id", userID),
					slog.String("address", conn.RemoteAddr().String()),
					sl.Err(err),
				)
				continue
			}
			if mt != websocket.TextMessage {
				log.Info(
					"unexpected ws message type",
					slog.Int64("user_id", userID),
					slog.Int("websocket message type", mt),
				)
				continue
			}

			var mesg models.Message
			if err := json.Unmarshal(data, &mesg); err != nil {
				log.Info(
					"message cannot be converted into models.Message",
					slog.Int64("user_id", userID),
					slog.String("message", string(data)),
					sl.Err(err),
				)
				continue
			}

			_, err = lib.ParseToken(mesg.Token)
			if err != nil {
				log.Error(
					"invalid token",
					slog.String("token", mesg.Token),
					sl.Err(err),
				)
				errMsg, _ := json.Marshal(models.Message{Type: "error", Value: []byte("invalid token")})
				_ = conn.WriteMessage(websocket.TextMessage, errMsg)
				// TODO clear user_id - conn map
				return
			}

			updateMsg, err := h.service.Validate(mesg)
			if err != nil {
				log.Error(
					"invalid message",
					slog.Int64("user_id", userID),
					slog.String("message", string(mesg.Value)),
					sl.Err(err),
				)
				continue
			}
			err = h.service.Save(ctx, userID, mesg)
			if err != nil {
				log.Error(
					"error saving message into database",
					slog.Int64("user_id", userID),
					slog.String("message", string(mesg.Value)),
					sl.Err(err),
				)
				continue
			}

			go h.sendUpdates(userID, updateMsg)
			/*
				update, err := json.Marshal(updateMsg)
				for _, c := range h.conns.UserConns(userID) {
					err := c.WriteMessage(websocket.TextMessage, update)
					if err != nil {
						/// что делать, если сообщение не может быть отправлено?
						// удалять из мапы соединение
						//conn.PingHandler()
					}
				}

			*/
		}
	}

}

func (h *Handler) sendUpdates(userID int64, msg models.Message) {
	update, _ := json.Marshal(msg)
	for _, c := range h.conns.UserConns(userID) {
		err := c.WriteMessage(websocket.TextMessage, update)
		if err != nil {
			// TODO ? clear user_id - conn map
			h.log.Error(
				"error listening client connection",
				slog.Int64("user_id", userID),
				slog.String("address", c.RemoteAddr().String()),
				sl.Err(err),
			)
			continue
		}
	}
}

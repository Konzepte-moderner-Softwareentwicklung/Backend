package gateway

import (
	"encoding/json"
	"fmt"
	"net/http"

	repoangebot "github.com/Konzepte-moderner-Softwareentwicklung/Backend/internal/http/angebotservice/service/repo_angebot"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/nats-io/nats.go"
)

const (
	UserIdHeader = "UserId"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func (s *Service) WSHandler(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	id, err := s.DecodeUUID(token)
	if err != nil {
		s.Error(w, "Invalid token", http.StatusBadRequest)
		return
	}
	s.GetLogger().Info().Str("user_id", id.String()).Msg("WebSocket connection requested")
	conn, err := upgrader.Upgrade(w, r, nil)

	if err != nil {
		s.Error(w, "WebSocket upgrade failed", http.StatusInternalServerError)
		return
	}
	defer func() {
		if err := conn.Close(); err != nil {
			s.GetLogger().Err(err).Msg("Failed to close WebSocket connection")
		}
	}()

	subject := fmt.Sprintf("user.%s", id.String())

	writeChan := make(chan []byte)
	done := make(chan struct{})

	sub, err := s.NR.Subscribe(subject, func(m *nats.Msg) {
		select {
		case writeChan <- m.Data:
		case <-done:
		}
	})
	if err != nil {
		s.Error(w, "Failed to subscribe to subject", http.StatusInternalServerError)
		return
	}
	defer func() {
		if err := sub.Unsubscribe(); err != nil {
			s.GetLogger().Err(err).Msg("Failed to unsubscribe from NATS subject")
		}
	}()

	go func() {
		for {
			select {
			case msg := <-writeChan:
				if err := conn.WriteMessage(websocket.TextMessage, msg); err != nil {
					s.GetLogger().Err(err)
					close(done)
					return
				}
				s.GetLogger().Debug().Msgf("Sent to [%s], message: %s", id.String(), msg)
			case <-done:
				return
			}
		}
	}()

	// Reader-Loop (nur um Verbindung offen zu halten)
	for {
		if _, data, err := conn.ReadMessage(); err != nil {
			s.GetLogger().Err(err)
			close(done)
			break
		} else {
			s.GetLogger().Debug().Msgf("Received from [%s], message: %s", id.String(), data)
		}
	}
}

func (s *Service) HandleChatWS(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	token := r.URL.Query().Get("token")
	userId, err := s.DecodeUUID(token)
	if err != nil {
		s.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	chatid, err := uuid.Parse(vars["chatId"])
	if err != nil {
		s.Error(w, "Invalid chatId", http.StatusBadRequest)
		return
	}

	if chatid == uuid.Nil {
		s.Error(w, "Invalid chatId", http.StatusBadRequest)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		s.Error(w, "WebSocket upgrade failed", http.StatusInternalServerError)
		return
	}
	defer func() {
		if err := conn.Close(); err != nil {
			s.GetLogger().Err(err).Msg("Failed to close WebSocket connection")
		}
	}()

	subject := fmt.Sprintf("chat.message.%s", chatid.String())
	_, err = s.NR.Subscribe(subject, func(msg *nats.Msg) {
		s.GetLogger().Debug().Msgf("Received message: %s", msg.Data)
		if err := conn.WriteMessage(websocket.TextMessage, msg.Data); err != nil {
			s.GetLogger().Err(err).Msg("Failed to write message to WebSocket")
		}
	})
	if err != nil {
		s.Error(w, "Failed to subscribe to NATS", http.StatusInternalServerError)
		return
	}

	for {
		if _, data, err := conn.ReadMessage(); err != nil {
			s.GetLogger().Err(err)
			break
		} else {
			s.GetLogger().Debug().Msgf("Received from [%s], message: %s", userId.String(), data)
		}
	}
}

type TrackingRequest struct {
	Location repoangebot.Location `json:"location" bson:"location"`
}

func (s *Service) HandleTracking(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	id, err := s.DecodeUUID(token)
	if err != nil {
		s.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		s.Error(w, "WebSocket upgrade failed", http.StatusInternalServerError)
		return
	}
	defer func() {
		if err := conn.Close(); err != nil {
			s.GetLogger().Err(err).Msg("Failed to close WebSocket connection")
		}
	}()

	subject := fmt.Sprintf("tracking.user.%s", id)
	var trackingRequest TrackingRequest

	for {
		err := conn.ReadJSON(&trackingRequest)

		if err != nil {
			s.GetLogger().Err(err).Msg("WebSocket read error or closed")
			break
		}
		// Ensure that the request is valid
		// TODO: Check if the location is valid

		jsonData, err := json.Marshal(trackingRequest)
		if err != nil {
			s.GetLogger().Err(err).Msg("Failed to marshal tracking request")
			s.Error(w, "Failed to marshal tracking request", http.StatusInternalServerError)
			return
		}

		err = s.NR.Publish(subject, jsonData)
		if err != nil {
			s.GetLogger().Err(err).Msg("Failed to publish tracking request")
			s.Error(w, "Failed to publish tracking request", http.StatusInternalServerError)
			return
		}

	}
}

package gateway

import (
	"fmt"
	"net/http"

	"github.com/google/uuid"
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
	id := r.Header.Get(UserIdHeader)
	uid, err := uuid.Parse(id)
	if err != nil {
		s.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)

	if err != nil {
		s.Error(w, "WebSocket upgrade failed", http.StatusInternalServerError)
		return
	}
	defer conn.Close()

	subject := fmt.Sprintf("user.%s", uid.String())

	// Optional: channel für thread-sicheres Schreiben
	writeChan := make(chan []byte)
	done := make(chan struct{})

	// NATS-Subscription
	sub, err := s.Subscribe(subject, func(m *nats.Msg) {
		select {
		case writeChan <- m.Data:
		case <-done:
		}
	})
	if err != nil {
		s.Error(w, "Failed to subscribe to subject", http.StatusInternalServerError)
		return
	}
	defer sub.Unsubscribe()

	// Writer-Goroutine
	go func() {
		for {
			select {
			case msg := <-writeChan:
				if err := conn.WriteMessage(websocket.TextMessage, msg); err != nil {
					s.GetLogger().Err(err)
					close(done)
					return
				}
			case <-done:
				return
			}
		}
	}()

	// Reader-Loop (nur um Verbindung offen zu halten)
	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			s.GetLogger().Err(err)
			close(done)
			break
		}
	}
}

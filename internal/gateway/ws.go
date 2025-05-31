package gateway

import (
	"fmt"
	"net/http"

	"github.com/Konzepte-moderner-Softwareentwicklung/Backend/internal/server"
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

// sends messages to the websocket connection from the NATS server (subject = user.<uuid>)
func (s *Service) WSHandler(w http.ResponseWriter, r *http.Request) {
	id := r.Header.Get(UserIdHeader)
	var (
		uid uuid.UUID
		err error
	)
	if uid, err = uuid.Parse(id); err != nil {
		server.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}
	conn, err := upgrader.Upgrade(w, r, nil)
	subject := fmt.Sprintf("user.%s", uid.String())

	s.Subscribe(subject, func(m *nats.Msg) {
		conn.WriteMessage(websocket.TextMessage, m.Data)
	})

	if err != nil {
		s.GetLogger().Err(err)
		return
	}

	defer conn.Close()
}

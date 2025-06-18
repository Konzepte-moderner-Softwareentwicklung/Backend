package ratingservice

import (
	"encoding/json"

	"github.com/Konzepte-moderner-Softwareentwicklung/Backend/internal/server"
	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
)

type Service struct {
	server.Server
	*nats.Conn
	Repository
}

func NewService(natsUrl string, repo Repository) *Service {
	conn, err := nats.Connect(natsUrl)
	if err != nil {
		panic(err)
	}

	svc := &Service{
		Server:     *server.NewServer(),
		Conn:       conn,
		Repository: repo,
	}

	return svc
}

type Rating struct {
	UserIDFrom uuid.UUID `json:"user_id_from" bson:"user_id_from"`
	UserIDTo   uuid.UUID `json:"user_id_to" bson:"user_id_to"`
	Value      int       `json:"value" bson:"value"`
	Content    string    `json:"content" bson:"content"`
}

func (svc *Service) StartNats(done <-chan struct{}) {
	subject := "ratings."

	sub, _ := svc.Subscribe(subject+"*", func(msg *nats.Msg) {
		// TODO: validate user id
		var (
			rating Rating
			userId uuid.UUID
			err    error
		)
		user := msg.Subject[len(subject):]
		if userId, err = uuid.Parse(user); err != nil {
			svc.GetLogger().Err(err)
			return
		}

		if err = json.Unmarshal(msg.Data, &rating); err != nil {
			svc.GetLogger().Err(err)
			return
		}
		rating.UserIDFrom = userId
		if err := svc.CreateRating(&rating); err != nil {
			svc.GetLogger().Err(err)
			return
		}
	})

	select {
	case <-done:
		svc.GetLogger().Info().Msg("Shutting down NATS service")
		sub.Unsubscribe()
		svc.Conn.Close()
	}
}

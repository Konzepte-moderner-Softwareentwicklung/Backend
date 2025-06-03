package natsreciver

import "github.com/nats-io/nats.go"

type Receiver struct {
	*nats.Conn
}

func New(url string) (*Receiver, error) {
	nc, err := nats.Connect(url)
	if err != nil {
		return nil, err
	}
	return &Receiver{
		Conn: nc,
	}, nil
}

func (r *Receiver) Close() error {
	return r.Drain()
}

package events

import (
	"bytes"
	"context"
	"encoding/gob"

	"github.com/go/events-cqrs-go/models"
	"github.com/nats-io/nats.go"
)

type NatsEventStore struct {
	conn            *nats.Conn         // Parametro donde declaro la conexion con nat
	feedCreatedSub  *nats.Subscription // Esta va a ser la suscripcion para conectarce a un evento cuando a sido creado
	feedCreatedChan chan CreatedFeedMessage
}

func NewNats(url string) (*NatsEventStore, error) {
	conn, err := nats.Connect(url)
	if err != nil {
		return nil, err
	}
	return &NatsEventStore{
		conn: conn,
	}, nil
}

func (n *NatsEventStore) Close() {
	if n.conn != nil {
		n.conn.Close()
	}

	if n.feedCreatedSub != nil {
		n.feedCreatedSub.Unsubscribe()
	}

	close(n.feedCreatedChan)
}

func (n *NatsEventStore) encodeMessage(m Message) ([]byte, error) {
	b := bytes.Buffer{}
	err := gob.NewEncoder(&b).Encode(m)
	if err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

// Metodo para avisar a todos los microservicios que otro feed se conecto
func (n *NatsEventStore) PublishCreatedFeed(ctx context.Context, feed *models.Feed) error {
	msg := CreatedFeedMessage{
		ID:          feed.ID,
		Title:       feed.Title,
		Description: feed.Description,
		CreatedAt:   feed.CreatedAt,
	}

	data, err := n.encodeMessage(msg)

	if err != nil {
		return err
	}

	return n.conn.Publish(msg.Type(), data)
}

func (n *NatsEventStore) decodeMessage(data []byte, m any) error {
	b := bytes.Buffer{}
	b.Write(data)
	return gob.NewDecoder(&b).Decode(m)
}

func (n *NatsEventStore) OnCreateFeed(f func(CreatedFeedMessage)) (err error) {
	msg := CreatedFeedMessage{}
	n.feedCreatedSub, err = n.conn.Subscribe(msg.Type(), func(m *nats.Msg) {
		n.decodeMessage(m.Data, &msg)
		f(msg)
	})
	return
}

func (n *NatsEventStore) SubscribeCreatedFeed(ctx context.Context) (<-chan CreatedFeedMessage, error) {
	m := CreatedFeedMessage{}
	n.feedCreatedChan = make(chan CreatedFeedMessage, 64)
	ch := make(chan *nats.Msg, 64)

	var err error
	n.feedCreatedSub, err = n.conn.ChanSubscribe(m.Type(), ch)

	if err != nil {
		return nil, err
	}

	go func() {
		for {
			select {
			case msg := <-ch: // una vez que ch reciba data, va entrar a este caso
				n.decodeMessage(msg.Data, &m)
				n.feedCreatedChan <- m // transmitimos m a n.feedCreatedChan
			}
		}
	}()
	return (<-chan CreatedFeedMessage)(n.feedCreatedChan), nil
}

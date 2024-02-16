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

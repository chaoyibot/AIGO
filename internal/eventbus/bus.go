package eventbus

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

type Event struct {
	ID        string          `json:"id"`
	Type      string          `json:"type"`
	CreatedAt time.Time       `json:"created_at"`
	Data      json.RawMessage `json:"data"`
}

type Bus struct {
	nc  *nats.Conn
	js  jetstream.JetStream
	ctx context.Context
}

func New(natsURL string) (*Bus, error) {
	nc, err := nats.Connect(natsURL)
	if err != nil {
		return nil, fmt.Errorf("connect nats: %w", err)
	}
	js, err := jetstream.New(nc)
	if err != nil {
		return nil, fmt.Errorf("jetstream: %w", err)
	}
	ctx := context.Background()

	// Create stream if not exists
	_, err = js.CreateOrUpdateStream(ctx, jetstream.StreamConfig{
		Name:     "AIGO_EVENTS",
		Subjects: []string{"aigo.events.>"},
		Storage:  jetstream.MemoryStorage,
	})
	if err != nil {
		return nil, fmt.Errorf("create stream: %w", err)
	}

	return &Bus{nc: nc, js: js, ctx: ctx}, nil
}

func (b *Bus) Publish(eventType string, data interface{}) error {
	payload, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("marshal event data: %w", err)
	}

	event := Event{
		ID:        uuid.New().String(),
		Type:      eventType,
		CreatedAt: time.Now(),
		Data:      payload,
	}

	eventJSON, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal event: %w", err)
	}

	_, err = b.js.Publish(b.ctx, "aigo.events."+eventType, eventJSON)
	return err
}

func (b *Bus) Subscribe(eventType string, handler func(event Event)) (*jetstream.Consumer, error) {
	cons, err := b.js.CreateOrUpdateConsumer(b.ctx, "AIGO_EVENTS", jetstream.ConsumerConfig{
		Name:          "consumer-" + eventType,
		FilterSubject: "aigo.events." + eventType,
		AckPolicy:     jetstream.AckExplicitPolicy,
	})
	if err != nil {
		return nil, fmt.Errorf("create consumer: %w", err)
	}

	iter, err := cons.Messages()
	if err != nil {
		return nil, fmt.Errorf("consumer messages: %w", err)
	}

	go func() {
		for {
			msg, err := iter.Next()
			if err != nil {
				return
			}
			var event Event
			if err := json.Unmarshal(msg.Data(), &event); err == nil {
				handler(event)
			}
			msg.Ack()
		}
	}()

	return &cons, nil
}

func (b *Bus) Close() {
	b.nc.Close()
}

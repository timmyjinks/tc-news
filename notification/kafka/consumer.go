package kafka

import (
	"context"
	"encoding/json"

	"github.com/segmentio/kafka-go"
)

type Consumer struct {
	reader *kafka.Reader
}

func NewConsumer(r *kafka.Reader) *Consumer {
	return &Consumer{
		reader: r,
	}
}

func (c *Consumer) Read(ctx context.Context) (Message, error) {
	var msg Message
	m, err := c.reader.ReadMessage(ctx)
	if err != nil {
		return Message{}, err
	}

	if err := json.Unmarshal(m.Value, &msg); err != nil {
		return Message{}, err
	}
	return msg, nil
}

package kafka

import (
	"context"
	"encoding/json"

	"github.com/segmentio/kafka-go"
)

type Producer struct {
	writer *kafka.Writer
}

func NewProducer(w *kafka.Writer) *Producer {
	return &Producer{
		writer: w,
	}
}

func (p *Producer) Send(ctx context.Context, topic string, msg Message) error {
	b, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	if err := p.writer.WriteMessages(ctx, kafka.Message{
		Value:     b,
		Partition: 0,
	}); err != nil {
		return err
	}

	return nil
}

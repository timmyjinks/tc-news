package kafka

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/segmentio/kafka-go"
)

type KafkaService struct {
	Consumer *Consumer
}

func NewKafkaService(topic string) *KafkaService {
	dialer := kafka.Dialer{
		Timeout:   10 * time.Second,
		DualStack: true,
	}

	_, err := dialer.DialLeader(context.Background(), "tcp", "kafka-service:9092", topic, 0)
	if err != nil {
		log.Println("[WARN]" + err.Error())
	}

	c := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     []string{"kafka-service:9092"},
		Partition:   0,
		GroupID:     fmt.Sprintf("%s-group", topic),
		Topic:       topic,
		StartOffset: kafka.LastOffset,
	},
	)

	return &KafkaService{
		Consumer: NewConsumer(c),
	}
}

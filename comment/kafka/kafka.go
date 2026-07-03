package kafka

import (
	"context"
	"log"
	"time"

	"github.com/segmentio/kafka-go"
)

type KafkaService struct {
	Producer *Producer
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

	p := &kafka.Writer{
		Addr:  kafka.TCP("kafka-service:9092"),
		Topic: topic,
	}

	return &KafkaService{
		Producer: NewProducer(p),
	}
}

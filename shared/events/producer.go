package events

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
)

// Producer wraps a kafka-go writer.
// Each service creates one Producer and uses it to publish events.
//
// Why one writer per service instead of one per topic?
// kafka-go's writer can handle multiple topics. Keeping one writer
// means one connection pool, less overhead.
type Producer struct {
	writer *kafka.Writer
}

// NewProducer creates a producer connected to the given brokers.
// brokers is a comma-separated list e.g. "kafka:9092"
func NewProducer(brokers []string) *Producer {
	w := &kafka.Writer{
		// Addr is where Kafka is running
		Addr: kafka.TCP(brokers...),

		// Balancer decides which Kafka partition to write to.
		// LeastBytes sends to the partition with least pending data — good default.
		Balancer: &kafka.LeastBytes{},

		// RequiredAcks = 1 means the leader partition must acknowledge the write.
		// This balances durability vs speed. For financial data you'd use WriterConfig.RequiredAcks = -1 (all replicas).
		RequiredAcks: kafka.RequireOne,

		// If the topic doesn't exist yet, create it automatically.
		// In production you'd pre-create topics with specific partition counts.
		AllowAutoTopicCreation: true,

		// How long to wait before flushing a batch of messages.
		// 10ms batching improves throughput without noticeable latency.
		BatchTimeout: 10 * time.Millisecond,
	}
	return &Producer{writer: w}
}

// Publish serializes an event to JSON and writes it to the given topic.
// The event must be one of the event structs defined in events.go.
//
// The key is the orderID or userID — Kafka uses it to decide which partition
// to send to. Messages with the same key always go to the same partition,
// which guarantees ordering for a specific order's events.
func (p *Producer) Publish(ctx context.Context, topic string, key uuid.UUID, event any) error {
	// serialize the event to JSON — this is what gets stored in Kafka
	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("producer: failed to marshal event: %w", err)
	}

	msg := kafka.Message{
		Topic: topic,
		Key:   []byte(key.String()), // partition key
		Value: payload,              // the actual event data
		Time:  time.Now(),
	}

	if err := p.writer.WriteMessages(ctx, msg); err != nil {
		return fmt.Errorf("producer: failed to write to topic %s: %w", topic, err)
	}

	slog.Info("event published", "topic", topic, "key", key)
	return nil
}

// Close shuts down the writer cleanly — call this on service shutdown.
func (p *Producer) Close() error {
	return p.writer.Close()
}
package events

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/segmentio/kafka-go"
)

// Handler is the function signature every consumer implements.
// It receives the raw JSON bytes and returns an error if processing failed.
// If it returns an error, the message will NOT be committed — Kafka will
// redeliver it on the next attempt.
type Handler func(ctx context.Context, payload []byte) error

// Consumer wraps a kafka-go reader for a single topic + consumer group.
// Each service creates one Consumer per topic it wants to listen to.
type Consumer struct {
	reader *kafka.Reader
	topic  string
}

// NewConsumer creates a consumer for a specific topic and group.
//
// groupID is critical — it determines which consumer group this belongs to.
// Example: "order-service" or "notification-service"
// Kafka tracks the offset (position) per group independently.
func NewConsumer(brokers []string, topic, groupID string) *Consumer {
	r := kafka.NewReader(kafka.ReaderConfig{
		// where Kafka is running
		Brokers: brokers,

		// which topic to read from
		Topic: topic,

		// which consumer group this belongs to
		// Kafka uses this to track how far this service has read
		GroupID: groupID,

		// MinBytes/MaxBytes control how much data to fetch per request.
		// MinBytes=1 means fetch as soon as there's any data — low latency.
		// MaxBytes=10MB is the upper limit per fetch.
		MinBytes: 1,
		MaxBytes: 10e6,

		// StartOffset = -2 means start from the beginning if this group
		// has never consumed this topic before.
		// Use -1 (LastOffset) if you only want new messages.
		StartOffset: kafka.FirstOffset,
	})

	return &Consumer{reader: r, topic: topic}
}

// Listen starts reading messages in a blocking loop.
// Call this in a goroutine — it runs forever until ctx is cancelled.
//
// For each message it:
// 1. Reads from Kafka (blocks until a message arrives)
// 2. Calls your handler with the raw JSON bytes
// 3. If handler succeeds → commits the offset (marks message as processed)
// 4. If handler fails → logs the error but still commits to avoid infinite retry loops
//    In production you'd send failed messages to a dead letter queue instead.
func (c *Consumer) Listen(ctx context.Context, handler Handler) {
	slog.Info("consumer started", "topic", c.topic)

	for {
		// ReadMessage blocks here until a new message arrives in Kafka.
		// When ctx is cancelled (service shutting down), this returns an error.
		msg, err := c.reader.ReadMessage(ctx)
		if err != nil {
			// ctx cancelled = clean shutdown, not an error
			if ctx.Err() != nil {
				slog.Info("consumer shutting down", "topic", c.topic)
				return
			}
			slog.Error("consumer read error", "topic", c.topic, "err", err)
			continue
		}

		slog.Info("event received",
			"topic", c.topic,
			"key", string(msg.Key),
			"offset", msg.Offset,
		)

		// call the handler with the raw JSON payload
		if err := handler(ctx, msg.Value); err != nil {
			slog.Error("handler failed",
				"topic", c.topic,
				"offset", msg.Offset,
				"err", err,
			)
			// TODO: send to dead letter queue instead of skipping
			// For now we log and move on to avoid getting stuck
		}
	}
}

// Unmarshal is a helper for handlers to decode the raw payload into a struct.
// Usage: events.Unmarshal(payload, &myEvent)
func Unmarshal[T any](payload []byte, dst *T) error {
	if err := json.Unmarshal(payload, dst); err != nil {
		return fmt.Errorf("consumer: failed to unmarshal payload: %w", err)
	}
	return nil
}

// Close shuts down the reader — call on service shutdown.
func (c *Consumer) Close() error {
	return c.reader.Close()
}
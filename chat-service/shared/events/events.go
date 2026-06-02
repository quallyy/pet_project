package events

import (
	"time"

	"github.com/google/uuid"
)

// Every event has a base with metadata useful for debugging and tracing.
// Every consumer can log or trace using these fields without parsing the payload.
type Base struct {
	EventID   uuid.UUID `json:"event_id"`   // unique ID for this event instance
	OccuredAt time.Time `json:"occured_at"` // when the event happened
}

// ── ORDER EVENTS ──────────────────────────────────────────────────────────────

// OrderCreatedEvent is published by order-service when a customer creates an order.
// notification-service consumes this to notify all dealers.
type OrderCreatedEvent struct {
	Base
	OrderID     uuid.UUID `json:"order_id"`
	CustomerID  uuid.UUID `json:"customer_id"`
	Description string    `json:"description"`
	PhotoPath   string    `json:"photo_path"`
	ExpiresAt   time.Time `json:"expires_at"`
}

// OrderAgreedEvent is published by chat-service when a customer confirms a deal.
// order-service consumes this to update status and reject other bids.
// notification-service consumes this to notify losing dealers.
type OrderAgreedEvent struct {
	Base
	OrderID    uuid.UUID `json:"order_id"`
	ChatID     uuid.UUID `json:"chat_id"`
	BidID      uuid.UUID `json:"bid_id"`
	CustomerID uuid.UUID `json:"customer_id"`
	DealerID   uuid.UUID `json:"dealer_id"` // the winning dealer
}

// OrderCancelledEvent is published by order-service when a customer cancels.
// notification-service consumes this to notify dealers who placed bids.
type OrderCancelledEvent struct {
	Base
	OrderID    uuid.UUID `json:"order_id"`
	CustomerID uuid.UUID `json:"customer_id"`
}

// ── BID EVENTS ────────────────────────────────────────────────────────────────

// BidPlacedEvent is published by order-service when a dealer places a bid.
// notification-service consumes this to notify the customer.
type BidPlacedEvent struct {
	Base
	BidID         uuid.UUID `json:"bid_id"`
	OrderID       uuid.UUID `json:"order_id"`
	DealerID      uuid.UUID `json:"dealer_id"`
	CustomerID    uuid.UUID `json:"customer_id"`
	Price         float64   `json:"price"`
	EstimatedDays int       `json:"estimated_days"`
}

// ── CHAT EVENTS ───────────────────────────────────────────────────────────────

// MessageSentEvent is published by chat-service when a message is sent.
// notification-service consumes this to push to offline users.
type MessageSentEvent struct {
	Base
	MessageID  uuid.UUID `json:"message_id"`
	ChatID     uuid.UUID `json:"chat_id"`
	SenderID   uuid.UUID `json:"sender_id"`
	ReceiverID uuid.UUID `json:"receiver_id"`
	Body       string    `json:"body"`
}
// Package events defines the Kafka topics and event types shared
// across all services. Both producers and consumers import from here
// so topic names and payload shapes never get out of sync.
package events

// Topics — every Kafka topic used on the platform.
// Services publish to and consume from these names.
const (
	// TopicOrderCreated fires when a customer submits a new order.
	// Consumers: notification-service (tell all dealers)
	TopicOrderCreated = "order.created"

	// TopicOrderAgreed fires when a customer confirms a deal with a dealer.
	// Consumers: order-service (update status, reject other bids)
	//            notification-service (tell other dealers the order is filled)
	TopicOrderAgreed = "order.agreed"

	// TopicOrderCancelled fires when a customer cancels before agreement.
	// Consumers: notification-service (tell any dealers who bid)
	TopicOrderCancelled = "order.cancelled"

	// TopicBidPlaced fires when a dealer places a bid on an order.
	// Consumers: notification-service (tell the customer)
	TopicBidPlaced = "bid.placed"

	// TopicMessageSent fires when a chat message is sent.
	// Consumers: notification-service (push to offline users)
	TopicMessageSent = "message.sent"
)
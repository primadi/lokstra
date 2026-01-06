package serviceapi

import "context"

// EventType represents the type of event
type EventType string

// SubscriptionID is a unique identifier for a subscription
type SubscriptionID uint64

// Event represents a generic event
type Event struct {
	Type    EventType
	Payload any
}

// EventHandler is a function that handles an event
type EventHandler func(ctx context.Context, event Event) error

type EventBus interface {
	Subscribe(eventType EventType, handler EventHandler) SubscriptionID
	Publish(ctx context.Context, event Event) error
	PublishAsync(ctx context.Context, event Event)
	Unsubscribe(subID SubscriptionID) bool
	UnsubscribeAll(eventType EventType) int
	HandlerCount(eventType EventType) int
}

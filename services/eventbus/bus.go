package eventbus

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/primadi/lokstra/serviceapi"
)

// subscription represents a registered handler with its ID
type subscription struct {
	id      serviceapi.SubscriptionID
	handler serviceapi.EventHandler
}

// Bus is a simple in-memory event bus
type Bus struct {
	handlers  map[serviceapi.EventType][]subscription
	mu        sync.RWMutex
	nextSubID serviceapi.SubscriptionID
}

// NewBus creates a new event bus
func NewBus() *Bus {
	return &Bus{
		handlers: make(map[serviceapi.EventType][]subscription),
	}
}

// Subscribe registers a handler for a specific event type
// Returns a SubscriptionID that can be used to unsubscribe later
func (b *Bus) Subscribe(eventType serviceapi.EventType, handler serviceapi.EventHandler) serviceapi.SubscriptionID {
	b.mu.Lock()
	defer b.mu.Unlock()

	// Generate unique subscription ID
	subID := serviceapi.SubscriptionID(atomic.AddUint64((*uint64)(&b.nextSubID), 1))

	// Add subscription
	b.handlers[eventType] = append(b.handlers[eventType], subscription{
		id:      subID,
		handler: handler,
	})

	return subID
}

// Publish publishes an event to all registered handlers
// Executes handlers synchronously in order they were registered
func (b *Bus) Publish(ctx context.Context, event serviceapi.Event) error {
	b.mu.RLock()
	subs := b.handlers[event.Type]
	b.mu.RUnlock()

	for i, sub := range subs {
		if err := sub.handler(ctx, event); err != nil {
			return fmt.Errorf("handler %d (id=%d) for event %s failed: %w", i, sub.id, event.Type, err)
		}
	}

	return nil
}

// PublishAsync publishes an event asynchronously to all registered handlers
// Each handler runs in its own goroutine, errors are logged but don't block
func (b *Bus) PublishAsync(ctx context.Context, event serviceapi.Event) {
	b.mu.RLock()
	subs := b.handlers[event.Type]
	b.mu.RUnlock()

	var wg sync.WaitGroup
	for i, sub := range subs {
		wg.Add(1)
		go func(idx int, s subscription) {
			defer wg.Done()
			if err := s.handler(ctx, event); err != nil {
				// TODO: Use proper logger
				fmt.Printf("async handler %d (id=%d) for event %s failed: %v\n", idx, s.id, event.Type, err)
			}
		}(i, sub)
	}

	// Optional: wait for all handlers to complete
	// Remove this if you want fire-and-forget behavior
	wg.Wait()
}

// Unsubscribe removes a specific handler by its subscription ID
// Returns true if the subscription was found and removed, false otherwise
func (b *Bus) Unsubscribe(subID serviceapi.SubscriptionID) bool {
	b.mu.Lock()
	defer b.mu.Unlock()

	// Search through all event types
	for eventType, subs := range b.handlers {
		for i, sub := range subs {
			if sub.id == subID {
				// Remove this subscription by slicing it out
				b.handlers[eventType] = append(subs[:i], subs[i+1:]...)

				// Clean up empty handler lists
				if len(b.handlers[eventType]) == 0 {
					delete(b.handlers, eventType)
				}

				return true
			}
		}
	}

	return false
}

// UnsubscribeAll removes all handlers for a specific event type
// Returns the number of handlers that were removed
func (b *Bus) UnsubscribeAll(eventType serviceapi.EventType) int {
	b.mu.Lock()
	defer b.mu.Unlock()

	count := len(b.handlers[eventType])
	delete(b.handlers, eventType)
	return count
}

// HandlerCount returns the number of handlers registered for an event type
func (b *Bus) HandlerCount(eventType serviceapi.EventType) int {
	b.mu.RLock()
	defer b.mu.RUnlock()

	return len(b.handlers[eventType])
}

var _ serviceapi.EventBus = (*Bus)(nil)

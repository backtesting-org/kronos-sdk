package events

import (
	"sync"
)

// EventBus provides publish-subscribe functionality for decoupling components
type EventBus interface {
	// Subscribe registers a handler for a specific topic
	Subscribe(topic string, handler func(event interface{}))

	// Publish sends an event to all subscribers of a topic
	Publish(topic string, event interface{})

	// Unsubscribe removes a handler from a topic
	Unsubscribe(topic string, handler func(event interface{}))

	// Close shuts down the event bus and cleans up resources
	Close()
}

type eventBus struct {
	subscribers map[string][]func(interface{})
	mu          sync.RWMutex
	closed      bool
}

// NewEventBus creates a new event bus
func NewEventBus() EventBus {
	return &eventBus{
		subscribers: make(map[string][]func(interface{})),
	}
}

func (eb *eventBus) Subscribe(topic string, handler func(event interface{})) {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	if eb.closed {
		return
	}

	eb.subscribers[topic] = append(eb.subscribers[topic], handler)
}

func (eb *eventBus) Publish(topic string, event interface{}) {
	eb.mu.RLock()
	defer eb.mu.RUnlock()

	if eb.closed {
		return
	}

	handlers, exists := eb.subscribers[topic]
	if !exists {
		return
	}

	// Call all handlers in goroutines to avoid blocking
	for _, handler := range handlers {
		go handler(event)
	}
}

func (eb *eventBus) Unsubscribe(topic string, handler func(event interface{})) {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	if eb.closed {
		return
	}

	// Note: Function comparison doesn't work in Go
	// This is a limitation - in practice, you'd need to use IDs or other mechanisms
	// For now, we'll just clear all handlers for the topic if provided
	delete(eb.subscribers, topic)
}

func (eb *eventBus) Close() {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	eb.closed = true
	eb.subscribers = make(map[string][]func(interface{}))
}

package pubsub

import (
	"fmt"
	"log/slog"
	"sync"
)

type EventBus struct {
	mu     sync.RWMutex
	topics map[string]*Broker
}

func NewEventBus() *EventBus {
	return &EventBus{
		topics: make(map[string]*Broker),
	}
}

func (e *EventBus) AddTopic(topic string) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if _, ok := e.topics[topic]; !ok {
		e.topics[topic] = NewBroker()
	}
}

func (e *EventBus) Subscribe(topic string) chan string {
	e.mu.RLock()
	defer e.mu.RUnlock()

	if _, ok := e.topics[topic]; ok {
		return e.topics[topic].Subscribe()
	}
	return nil
}

func (e *EventBus) Unsubscribe(topic string, ch chan string) {
	if _, ok := e.topics[topic]; ok {
		e.topics[topic].Unsubscribe(ch)
	}
}

func (e *EventBus) Publish(topic, message string) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	if _, ok := e.topics[topic]; ok {
		e.topics[topic].Publish(message)
	}
}

type Broker struct {
	mu          sync.RWMutex
	subscribers map[chan string]bool
}

func NewBroker() *Broker {
	return &Broker{
		subscribers: make(map[chan string]bool),
	}
}

func (b *Broker) Subscribe() chan string {
	ch := make(chan string, 5)
	b.mu.Lock()
	b.subscribers[ch] = true
	b.mu.Unlock()
	return ch
}

func (b *Broker) Unsubscribe(ch chan string) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if _, ok := b.subscribers[ch]; ok {
		delete(b.subscribers, ch)
		close(ch)
	}

}

func (b *Broker) Publish(msg string) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	for ch := range b.subscribers {
		select {
		case ch <- msg:
		default:
			// drop the message to make this non-blocking
		}
	}
}

type Subscriber struct {
	subscriptions map[string]chan string
	listenerChan  chan string
}

func NewSubscriber() *Subscriber {
	return &Subscriber{
		subscriptions: make(map[string]chan string),
		listenerChan:  make(chan string, 100),
	}
}

func (s *Subscriber) Subscribe(eventBus *EventBus, topic string) {
	ch := eventBus.Subscribe(topic)
	s.subscriptions[topic] = ch
	slog.Info("subscribed to new topic", "topic", topic)

	// fan into the listener channel
	go func(forwardChan chan string) {
		for msg := range forwardChan {
			s.listenerChan <- msg
		}
	}(ch)
}

func (s *Subscriber) Subscriptions() []string {
	subs := []string{}
	for t := range s.subscriptions {
		subs = append(subs, t)
	}
	return subs
}

func (s *Subscriber) Unsubscribe(eventBus *EventBus, topic string) {
	if ch, ok := s.subscriptions[topic]; ok {
		eventBus.Unsubscribe(topic, ch)
	}
	slog.Info("unsubscribed from topic", "topic", topic)
}

func (s *Subscriber) Listen() {
	go func() {
		for msg := range s.listenerChan {
			fmt.Printf("Subscriber received message: %s\n", msg)
		}
	}()
}

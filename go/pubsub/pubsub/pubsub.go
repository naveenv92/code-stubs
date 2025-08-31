package pubsub

import (
	"fmt"
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
	if _, ok := e.topics[topic]; ok {
		return e.topics[topic].Subscribe()
	}
	return nil
}

func (e *EventBus) Publish(topic, message string) {
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
	delete(b.subscribers, ch)
	close(ch)
	b.mu.Unlock()
}

func (b *Broker) Publish(msg string) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	for ch := range b.subscribers {
		ch <- msg
	}
}

type Subscriber struct {
	subscriptions map[string]chan string
}

func NewSubscriber() *Subscriber {
	return &Subscriber{
		subscriptions: make(map[string]chan string),
	}
}

func (s *Subscriber) Subscribe(eventBus *EventBus, topic string) {
	ch := eventBus.Subscribe(topic)
	s.subscriptions[topic] = ch
}

func (s *Subscriber) Subscriptions() []string {
	subs := []string{}
	for t := range s.subscriptions {
		subs = append(subs, t)
	}
	return subs
}

func (s *Subscriber) Listen() {
	for _, sub := range s.subscriptions {
		go func() {
			for msg := range sub {
				fmt.Printf("Subscriber received message: %s\n", msg)
			}
		}()
	}
}

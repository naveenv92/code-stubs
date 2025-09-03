package pubsub

import (
	"errors"
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

func (e *EventBus) Subscribe(topic string) (chan string, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	if _, ok := e.topics[topic]; ok {
		return e.topics[topic].Subscribe(), nil
	}
	return nil, errors.New("ERROR: topic does not exist")
}

func (e *EventBus) Unsubscribe(topic string, ch chan string) error {
	e.mu.RLock()
	defer e.mu.RUnlock()

	if _, ok := e.topics[topic]; ok {
		return e.topics[topic].Unsubscribe(ch)
	}

	return errors.New("ERROR: topic does not exist")
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

func (b *Broker) Unsubscribe(ch chan string) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if _, ok := b.subscribers[ch]; ok {
		delete(b.subscribers, ch)
		close(ch)
		return nil
	}

	return errors.New("ERROR: channel was not subscribed to this topic")

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
	mu            sync.RWMutex
	subscriptions map[string]*subscription
	listenerChan  chan string
}

type subscription struct {
	topic string
	ch    chan string
	done  chan struct{}
}

func NewSubscriber() *Subscriber {
	return &Subscriber{
		subscriptions: make(map[string]*subscription),
		listenerChan:  make(chan string, 100),
	}
}

func (s *Subscriber) Subscribe(eventBus *EventBus, topic string) error {
	ch, err := eventBus.Subscribe(topic)
	if err != nil {
		return fmt.Errorf("failed to subscribe: topic %s does not exist", topic)
	}

	sub := &subscription{
		topic: topic,
		ch:    ch,
		done:  make(chan struct{}),
	}

	s.mu.Lock()
	s.subscriptions[topic] = sub
	s.mu.Unlock()

	slog.Info("subscribed to new topic", "topic", topic)

	// fan into the listener channel
	go func() {
		for {
			select {
			case msg := <-sub.ch:
				s.listenerChan <- msg
			case <-sub.done:
				slog.Info("shutting down fan-in for topic", "topic", sub.topic)
				return
			}
		}
	}()

	return nil
}

func (s *Subscriber) Subscriptions() []string {
	subs := []string{}
	for t := range s.subscriptions {
		subs = append(subs, t)
	}
	return subs
}

func (s *Subscriber) Unsubscribe(eventBus *EventBus, topic string) error {
	s.mu.Lock()

	if sub, ok := s.subscriptions[topic]; ok {
		delete(s.subscriptions, sub.topic)
		s.mu.Unlock()

		// close the done channel
		close(sub.done)

		// unsubscribe from the event bus
		eventBus.Unsubscribe(topic, sub.ch)

		slog.Info("unsubscribed from topic", "topic", topic)
		return nil
	} else {
		s.mu.Unlock()
		return errors.New("ERROR: not subscribed to topic")
	}
}

func (s *Subscriber) Listen() {
	go func() {
		for msg := range s.listenerChan {
			fmt.Printf("Subscriber received message: %s\n", msg)
		}
	}()
}

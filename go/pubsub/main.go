package main

import (
	"time"

	"github.com/naveenv92/code-stubs/go/pubsub/pubsub"
)

func main() {
	// create event bus
	eventBus := pubsub.NewEventBus()

	// add two topics
	eventBus.AddTopic("topic_1")
	eventBus.AddTopic("topic_2")

	// create subscribers
	s1 := pubsub.NewSubscriber()
	s2 := pubsub.NewSubscriber()

	// subscribe to topics
	s1.Subscribe(eventBus, "topic_1")
	s1.Subscribe(eventBus, "topic_2")

	s2.Subscribe(eventBus, "topic_1")

	// start listeners
	s1.Listen()
	s2.Listen()

	// publish to topic 1 and wait
	eventBus.Publish("topic_1", "hello on topic 1!")
	time.Sleep(1 * time.Second)

	// publish to topic 2 and wait
	eventBus.Publish("topic_2", "hello on topic 2!")
	time.Sleep(1 * time.Second)

	// unsubscribe from topic 1
	s1.Unsubscribe(eventBus, "topic_1")

	// publish to topic 1 and wait
	eventBus.Publish("topic_1", "hello again on topic 1!")
	time.Sleep(1 * time.Second)

}

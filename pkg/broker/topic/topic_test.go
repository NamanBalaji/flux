package topic

import (
	"context"
	"testing"
	"time"

	"github.com/NamanBalaji/flux/pkg/broker/subscriber"
	"github.com/NamanBalaji/flux/pkg/config"
	"github.com/NamanBalaji/flux/pkg/message"
)

// Helper function to create a default config
func defaultConfig() config.Config {
	return config.Config{
		Subscriber: config.Subscriber{
			InactiveTime: 1,
		},
		Message: config.Message{
			TTL: 1,
		},
	}
}

func TestCreateTopic(t *testing.T) {
	topic := CreateTopic("testTopic", 100)
	if topic.Name != "testTopic" {
		t.Errorf("Expected topic name testTopic, got %s", topic.Name)
	}
	if cap(topic.MessageChan) != 100 {
		t.Errorf("Expected channel capacity 100, got %d", cap(topic.MessageChan))
	}
}

// add message
func TestAddMessage(t *testing.T) {
	topic := CreateTopic("testTopic", 100)
	msg := &message.Message{Id: "1", Payload: "Hello World"}
	topic.AddMessage(msg)
	defer close(topic.MessageChan)

	if _, exists := topic.MessageSet[msg.Id]; !exists {
		t.Error("Message ID should exist in MessageSet")
	}

	retrievedMessage := topic.MessageQueue.Peek()
	if retrievedMessage != msg {
		t.Errorf("Expected message to be %v, got %v", msg, retrievedMessage)
	}
}

func TestDeliverToSubscriber(t *testing.T) {
	topic := CreateTopic("testTopic", 100)
	msg := message.NewMessage("1", "Hello World")

	sub := subscriber.NewSubscriber("localhost:6969")
	topic.Subscribers = append(topic.Subscribers, sub)

	topic.AddMessage(msg)
	defer close(topic.MessageChan)

	time.Sleep(100 * time.Millisecond)

	sub.Lock.Lock()
	if sub.MessageQueue.Len() != 1 {
		t.Errorf("Expected message queue length of 1, got %d", sub.MessageQueue.Len())
	}

	retrievedMessage := sub.MessageQueue.Peek()
	if retrievedMessage != msg {
		t.Errorf("Expected message to be %v, got %v", msg, retrievedMessage)
	}
	sub.Lock.Unlock()

	msg.Lock.Lock()
	defer msg.Lock.Unlock()

	if _, ok := msg.Delivered[sub.Addr]; !ok {
		t.Error("Message not tracking subscriber")
	}
}

func TestShouldEnqueue(t *testing.T) {
	topic := CreateTopic("testTopic", 100)
	msg := message.NewMessage("1", "Hello World")

	if !topic.ShouldEnqueue(msg) {
		t.Errorf("Should enqueue message %v", msg)
	}

	topic.MessageQueue.Enqueue(msg)
	if !topic.ShouldEnqueue(msg) {
		t.Errorf("Should enqueue message %v", msg)
	}
}

func TestSubscribe_ReadOldFalse(t *testing.T) {
	topic := CreateTopic("testTopic", 100)
	defer close(topic.MessageChan)

	msg := message.NewMessage("1", "Hello World")
	topic.MessageQueue.Enqueue(msg)
	topic.MessageSet[msg.Id] = struct{}{}

	topic.Subscribe(context.Background(), defaultConfig(), "localhost:6969", false)

	if len(topic.Subscribers) != 1 {
		t.Error("Should be 1 subscriber in topic")
	}

	sub := topic.Subscribers[0]
	sub.CancelFunc()

	if sub.MessageQueue.Len() != 0 {
		t.Errorf("Expected message queue length of 0, got %d", sub.MessageQueue.Len())
	}

}

func TestSubscribe_ReadOldTrue(t *testing.T) {
	topic := CreateTopic("testTopic", 100)
	defer close(topic.MessageChan)

	msg := message.NewMessage("1", "Hello World")
	topic.MessageQueue.Enqueue(msg)
	topic.MessageSet[msg.Id] = struct{}{}

	topic.Subscribe(context.Background(), defaultConfig(), "localhost:6969", true)

	if len(topic.Subscribers) != 1 {
		t.Error("Should be 1 subscriber in topic")
	}

	sub := topic.Subscribers[0]
	sub.CancelFunc()

	if sub.MessageQueue.Len() != 1 {
		t.Errorf("Expected message queue length of 1, got %d", sub.MessageQueue.Len())
	}

	retrievedMessage := sub.MessageQueue.Peek()
	if retrievedMessage != msg {
		t.Errorf("Expected message to be %v, got %v", msg, retrievedMessage)
	}
}

func TestSubscribe_ReactivateOld(t *testing.T) {
	topic := CreateTopic("testTopic", 100)
	defer close(topic.MessageChan)

	msg := message.NewMessage("1", "Hello World")
	topic.MessageQueue.Enqueue(msg)
	topic.MessageSet[msg.Id] = struct{}{}

	sub := subscriber.NewSubscriber("localhost:6969")
	_, cancel := context.WithCancel(context.Background())
	sub.CancelFunc = cancel
	sub.IsActive = false

	topic.Subscribers = append(topic.Subscribers, sub)

	topic.Subscribe(context.Background(), defaultConfig(), "localhost:6969", true)

	if len(topic.Subscribers) != 1 {
		t.Error("Should be 1 subscriber in topic")
	}

	sub.CancelFunc()

	if !sub.IsActive {
		t.Error("Subscriber should have been activated")
	}
}

func TestUnsubscribe(t *testing.T) {
	topic := CreateTopic("testTopic", 100)
	defer close(topic.MessageChan)

	sub := subscriber.NewSubscriber("localhost:6969")
	sub.IsActive = true

	topic.Subscribers = append(topic.Subscribers, sub)

	err := topic.Unsubscribe("localhost:6969")

	if err != nil {
		t.Error("Unsubscribe should not return an error")
	}

	if len(topic.Subscribers) != 1 {
		t.Error("Should be 1 subscriber in topic")
	}

	if sub.IsActive {
		t.Error("Subscriber should have been inactive")
	}
}

func TestCleanupSubscribers_Remove(t *testing.T) {
	topic := CreateTopic("testTopic", 100)
	defer close(topic.MessageChan)

	sub := subscriber.NewSubscriber("localhost:6969")
	sub.IsActive = false
	sub.LastActive = time.Now().Add(-time.Hour)

	topic.CleanupSubscribers(defaultConfig())
	if len(topic.Subscribers) != 0 {
		t.Error("Subscribers should have been deleted")
	}
}

func TestCleanupSubscribers_DontRemove(t *testing.T) {
	topic := CreateTopic("testTopic", 100)
	defer close(topic.MessageChan)

	sub := subscriber.NewSubscriber("localhost:6969")
	sub.IsActive = true
	sub.LastActive = time.Now().Add(-time.Hour)

	topic.CleanupSubscribers(defaultConfig())
	if len(topic.Subscribers) == 1 {
		t.Error("Subscribers should still be there")
	}
}

func TestCleanupMessages_Remove(t *testing.T) {
	topic := CreateTopic("testTopic", 100)
	defer close(topic.MessageChan)

	msg := message.NewMessage("1", "Hello World")
	msg.AddedAt = time.Now().Add(-time.Hour)

	topic.MessageQueue.Enqueue(msg)
	topic.MessageSet[msg.Id] = struct{}{}

	topic.CleanupMessages(defaultConfig())
	if topic.MessageQueue.Len() != 0 {
		t.Error("Messages should have been deleted")
	}
	if len(topic.MessageSet) != 1 {
		t.Error("Messages should still be in the set")
	}
}

func TestCleanupMessages_DontRemove(t *testing.T) {
	topic := CreateTopic("testTopic", 100)
	defer close(topic.MessageChan)

	msg := message.NewMessage("1", "Hello World")
	msg.AddedAt = time.Now().Add(time.Hour)

	topic.MessageQueue.Enqueue(msg)
	topic.MessageSet[msg.Id] = struct{}{}

	topic.CleanupMessages(defaultConfig())
	if topic.MessageQueue.Len() != 1 {
		t.Error("Messages should still be in the queue")
	}
}

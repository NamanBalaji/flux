package message

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/NamanBalaji/flux/pkg/config"
)

func TestNewMessage(t *testing.T) {
	id := "msg1"
	payload := "Hello, World!"
	msg := NewMessage(id, payload)

	if msg.Id != id || msg.Payload != payload {
		t.Errorf("Message fields not initialized correctly. Got id %s and payload %s", msg.Id, msg.Payload)
	}

	if len(msg.Delivered) != 0 {
		t.Errorf("Delivered map should be initialized empty. Got size %d", len(msg.Delivered))
	}
}

func TestAck(t *testing.T) {
	msg := NewMessage("msg1", "test payload")
	subscriber := "sub1"
	msg.AddSubscriber(subscriber)
	msg.Ack(subscriber)

	if !msg.Delivered[subscriber] {
		t.Errorf("Ack failed to set delivery status to true for subscriber %s", subscriber)
	}
}

func TestConcurrentAcks(t *testing.T) {
	msg := NewMessage("msg1", "test payload")
	subscriberCount := 100

	var wg sync.WaitGroup
	wg.Add(subscriberCount)

	for i := 0; i < subscriberCount; i++ {
		go func(subID int) {
			defer wg.Done()
			subscriberAddress := fmt.Sprintf("subscriber%d", subID)
			msg.AddSubscriber(subscriberAddress)
			msg.Ack(subscriberAddress)
		}(i)
	}

	wg.Wait()

	if len(msg.Delivered) != subscriberCount {
		t.Errorf("Expected %d delivered entries, got %d", subscriberCount, len(msg.Delivered))
	}

	for _, delivered := range msg.Delivered {
		if !delivered {
			t.Errorf("Expected all messages to be acknowledged")
		}
	}
}

func TestAddAndRemoveSubscriber(t *testing.T) {
	msg := NewMessage("msg1", "test payload")
	subscriber := "sub1"
	msg.AddSubscriber(subscriber)

	if _, ok := msg.Delivered[subscriber]; !ok {
		t.Errorf("AddSubscriber failed to add subscriber %s", subscriber)
	}

	msg.RemoveSubscriber(subscriber)
	if _, ok := msg.Delivered[subscriber]; ok {
		t.Errorf("RemoveSubscriber failed to remove subscriber %s", subscriber)
	}
}

func TestConcurrentAddRemoveSubscribersIndependent(t *testing.T) {
	msg := NewMessage("msg1", "test payload")
	subscriberCount := 100

	var wg sync.WaitGroup
	wg.Add(subscriberCount)

	// Each goroutine adds and then removes a subscriber
	for i := 0; i < subscriberCount; i++ {
		go func(subID int) {
			subscriberAddress := fmt.Sprintf("subscriber%d", subID)
			msg.AddSubscriber(subscriberAddress)
			msg.RemoveSubscriber(subscriberAddress)
			wg.Done()
		}(i)
	}

	wg.Wait()

	if len(msg.Delivered) != 0 {
		t.Errorf("Expected 0 subscribers after all operations, got %d", len(msg.Delivered))
	}
}

func TestSafeToDelete(t *testing.T) {
	msg := NewMessage("msg1", "test payload")
	msg.AddedAt = time.Now().Add(-time.Hour)
	msg.AddSubscriber("sub1")
	msg.Ack("sub1")

	cfg := config.Config{
		Message: config.Message{TTL: 1},
	}

	if !msg.SafeToDelete(cfg) {
		t.Errorf("SafeToDelete incorrectly returned false")
	}
}

func TestNotSafeToDelete_MessageUnAcked(t *testing.T) {
	msg := NewMessage("msg1", "test payload")
	msg.AddSubscriber("sub1")

	cfg := config.Config{
		Message: config.Message{TTL: 1},
	}

	if msg.SafeToDelete(cfg) {
		t.Errorf("SafeToDelete incorrectly returned true for undelivered message")
	}
}

func TestNotSafeToDelete_MessageNotExpired(t *testing.T) {
	msg := NewMessage("msg1", "test payload")
	msg.AddSubscriber("sub1")
	msg.Ack("sub1")

	cfg := config.Config{
		Message: config.Message{TTL: 40},
	}

	if msg.SafeToDelete(cfg) {
		t.Errorf("SafeToDelete incorrectly returned true for undelivered message")
	}
}

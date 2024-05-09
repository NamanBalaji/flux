package subscriber

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/NamanBalaji/flux/pkg/config"
	"github.com/NamanBalaji/flux/pkg/message"
	"github.com/NamanBalaji/flux/pkg/request"
)

func setupMockServer() *httptest.Server {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var reqBody request.PollMessage
		if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
			http.Error(w, "Invalid request", 400)
			return
		}
		// Simulate a successful ack response
		if reqBody.Id != "fail" {
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
	})
	return httptest.NewServer(handler)
}

func TestAddMessage(t *testing.T) {
	sub := NewSubscriber("http://example.com")
	msg := &message.Message{Id: "1", Payload: "data"}
	sub.AddMessage(msg)

	if sub.MessageQueue.Len() != 1 {
		t.Errorf("Expected 1 message in the queue, got %d", sub.MessageQueue.Len())
	}
}

func TestHandleQueue(t *testing.T) {
	server := setupMockServer()
	defer server.Close()

	cfg := config.Config{
		Subscriber: config.Subscriber{Timeout: 1, RetryCount: 1},
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sub := NewSubscriber(server.URL)
	sub.IsActive = true
	msg := message.NewMessage("1", "data")
	sub.AddMessage(msg)
	msg.AddSubscriber(server.URL)

	go sub.HandleQueue(ctx, cfg, "test-topic")

	time.Sleep(1 * time.Second)

	sub.Lock.Lock()
	defer sub.Lock.Unlock()

	if sub.IsActive != true {
		t.Errorf("Subscriber should still be active")
	}

	if sub.MessageQueue.Len() != 0 {
		t.Errorf("Expected 0 messages in the queue, got %d", sub.MessageQueue.Len())
	}
}

func TestHandleQueue_PushFail(t *testing.T) {
	server := setupMockServer()
	defer server.Close()

	cfg := config.Config{
		Subscriber: config.Subscriber{Timeout: 1, RetryCount: 1},
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sub := NewSubscriber(server.URL)
	sub.IsActive = true
	msg := message.NewMessage("fail", "data")
	sub.AddMessage(msg)
	msg.AddSubscriber(server.URL)

	go sub.HandleQueue(ctx, cfg, "test-topic")

	time.Sleep(1 * time.Second)

	sub.Lock.Lock()
	defer sub.Lock.Unlock()

	if sub.IsActive != false {
		t.Errorf("Subscriber should be inactive")
	}

	if sub.MessageQueue.Len() != 1 {
		t.Errorf("Expected 1 messages in the queue, got %d", sub.MessageQueue.Len())
	}
}

func TestPushMessageFailure(t *testing.T) {
	server := setupMockServer()
	defer server.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	cfg := config.Config{
		Subscriber: config.Subscriber{Timeout: 50, RetryCount: 2},
	}

	sub := NewSubscriber(server.URL)
	sub.IsActive = true
	msg := message.NewMessage("fail", "data")
	sub.AddMessage(msg)
	msg.AddSubscriber(server.URL)

	err := sub.pushMessage(ctx, cfg, msg, "test-topic")
	if err == nil {
		t.Errorf("Expected error from pushMessage, got nil")
	}
}

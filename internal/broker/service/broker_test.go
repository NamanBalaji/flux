package service

import (
	"context"
	"testing"
	"time"

	"github.com/NamanBalaji/flux/pkg/broker/subscriber"
	topicPkg "github.com/NamanBalaji/flux/pkg/broker/topic"
	"github.com/NamanBalaji/flux/pkg/config"
	"github.com/NamanBalaji/flux/pkg/message"
)

func setupBrokerAndConfig() (*Broker, config.Config) {
	cfg := config.Config{
		Topic: config.Topic{
			Buffer: 10,
		},
		Subscriber: config.Subscriber{
			InactiveTime: 1,
		},
		Message: config.Message{
			TTL: 1,
		},
	}

	broker := NewBroker()
	return broker, cfg
}

func assert(t *testing.T, condition bool, msg string) {
	if !condition {
		t.Error(msg)
	}
}

func TestPublishMessage(t *testing.T) {
	broker, cfg := setupBrokerAndConfig()
	broker.Topics["testTopic"] = topicPkg.CreateTopic("testTopic", 10)
	msg := message.NewMessage("id", "payload")
	broker.publishMessage(cfg, "testTopic", msg)

	topic := broker.Topics["testTopic"]

	_, ok := topic.MessageSet["id"]
	l := topic.MessageQueue.Len()

	assert(t, ok, "message should be in topic")
	assert(t, l == 1, "message should be in topic queue")
}

func TestCreateTopicAndPublishMessage(t *testing.T) {
	broker, cfg := setupBrokerAndConfig()
	msg := message.NewMessage("id", "payload")
	broker.publishMessage(cfg, "testTopic", msg)

	topic, exist := broker.Topics["testTopic"]
	assert(t, exist, "topic should have been created")

	_, ok := topic.MessageSet["id"]
	l := topic.MessageQueue.Len()

	assert(t, ok, "message should be in topic")
	assert(t, l == 1, "message should be in topic queue")
}

func TestValidateTopics(t *testing.T) {
	broker, _ := setupBrokerAndConfig()
	broker.Topics["exists"] = topicPkg.CreateTopic("exists", 10)

	err := broker.ValidateTopics([]string{"exists"})
	assert(t, err == nil, "Validation should pass for existing topic")

	err = broker.ValidateTopics([]string{"nonexistent"})
	assert(t, err != nil, "Validation should fail for nonexistent topic")
}

func TestSubscribeAndUnsubscribe(t *testing.T) {
	broker, cfg := setupBrokerAndConfig()
	broker.Topics["exists"] = topicPkg.CreateTopic("newTopic", 10)

	ctx := context.Background()

	broker.Subscribe(ctx, cfg, "newTopic", "newAddress", false)

	topic := broker.Topics["newTopic"]

	var subscriber *subscriber.Subscriber = nil
	for _, sub := range topic.Subscribers {
		if sub.Addr == "newAddress" {
			subscriber = sub
			break
		}
	}
	assert(t, subscriber != nil, "subscriber not subscribed")

	broker.Unsubscribe("newTopic", "newAddress")

	assert(t, !subscriber.IsActive, "subscriber not unsubscribed")
}

func TestCleanupSubscribers(t *testing.T) {
	broker, cfg := setupBrokerAndConfig()
	topic := topicPkg.CreateTopic("testTopic", 10)
	broker.Topics["testTopic"] = topic

	ctx := context.Background()
	_, cancel := context.WithCancel(ctx)

	// Add an inactive subscriber
	sub := subscriber.NewSubscriber("inactiveAddress")
	sub.IsActive = false
	sub.LastActive = time.Now().Add(-time.Hour)
	sub.CancelFunc = cancel

	topic.Subscribers = append(topic.Subscribers, sub)

	broker.CleanSubscribers(cfg)
	assert(t, len(topic.Subscribers) == 0, "All inactive subscribers should be cleaned up")
}

func TestCleanupMessages(t *testing.T) {
	broker, cfg := setupBrokerAndConfig()
	topic := topicPkg.CreateTopic("testTopic", 10)
	broker.Topics["testTopic"] = topic

	// Add an old message that should be deleted
	msg := &message.Message{Id: "oldMsg", Payload: "old", AddedAt: time.Now().Add(-24 * time.Hour)}
	topic.MessageQueue.Enqueue(msg)

	broker.CleanupMessages(cfg)
	assert(t, topic.MessageQueue.Len() == 0, "Old messages should be cleaned up")
}

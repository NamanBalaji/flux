package service

import (
	"context"
	"fmt"
	"log"
	"sync"

	topicPkg "github.com/NamanBalaji/flux/pkg/broker/topic"
	"github.com/NamanBalaji/flux/pkg/config"
	"github.com/NamanBalaji/flux/pkg/message"
)

type Broker struct {
	mu     sync.Mutex
	Topics topicPkg.Topics
	// other brokers
}

func NewBroker() *Broker {
	return &Broker{
		Topics: topicPkg.CreateTopics(),
	}
}

func (b *Broker) PublishMessage(cfg config.Config, topicName string, msg *message.Message) bool {
	log.Printf("Trying to publish message with id: %s \n", msg.Id)
	b.mu.Lock()

	topic, ok := b.Topics[topicName]
	if !ok {
		topic = topicPkg.CreateTopic(topicName, cfg.Topic.Buffer)
		b.Topics[topicName] = topic

		log.Println("Created topic: ", topicName)
	}
	defer b.mu.Unlock()

	return topic.AddMessage(msg)
}

func (b *Broker) ValidateTopics(topics []string) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	for _, topic := range topics {
		_, ok := b.Topics[topic]
		if !ok {
			return fmt.Errorf("topic %s does not exist", topic)
		}
	}

	return nil
}

func (b *Broker) Subscribe(ctx context.Context, cfg config.Config, topicName string, address string, readOld bool) {
	b.mu.Lock()
	topic := b.Topics[topicName]
	b.mu.Unlock()

	log.Printf("Subscriber[Address: %s] trying to subscribe to the topic %s \n", address, topicName)
	topic.Subscribe(ctx, cfg, address, readOld)
}

func (b *Broker) Unsubscribe(topicName string, address string) error {
	b.mu.Lock()
	topic, ok := b.Topics[topicName]
	b.mu.Unlock()

	if !ok {
		return fmt.Errorf("no topic with the name %s exist", topicName)
	}

	log.Printf("Subscriber[Address: %s] trying to unsubscribe from the topic %s \n", address, topicName)
	return topic.Unsubscribe(address)
}

func (b *Broker) CleanSubscribers(cfg config.Config) {
	var topicsToClean []*topicPkg.Topic

	b.mu.Lock()
	for _, topic := range b.Topics {
		topicsToClean = append(topicsToClean, topic)
	}
	b.mu.Unlock()

	var wg sync.WaitGroup
	for _, topic := range topicsToClean {
		wg.Add(1)
		go func(t *topicPkg.Topic) {
			log.Println("Cleaning subscriber for topic ", t.Name)
			defer wg.Done()
			t.CleanupSubscribers(cfg)
		}(topic)
	}
	wg.Wait()
}

func (b *Broker) CleanupMessages(cfg config.Config) {
	var topicsToClean []*topicPkg.Topic

	b.mu.Lock()
	for _, topic := range b.Topics {
		topicsToClean = append(topicsToClean, topic)
	}
	b.mu.Unlock()

	var wg sync.WaitGroup
	for _, topic := range topicsToClean {
		wg.Add(1)
		go func(cfg config.Config, t *topicPkg.Topic) {
			log.Println("Cleaning messages for topic ", t.Name)
			defer wg.Done()
			t.CleanupMessages(cfg)
		}(cfg, topic)
	}
	wg.Wait()
}

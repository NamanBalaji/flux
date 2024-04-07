package service

import (
	"fmt"
	"github.com/NamanBalaji/flux/pkg/message"
	"github.com/NamanBalaji/flux/pkg/subscriber"
	topickPkg "github.com/NamanBalaji/flux/pkg/topic"
)

type Broker struct {
	Topics      []topickPkg.Topic
	Messages    message.MessagesByTopic
	Subscribers subscriber.SubscribersByTopic
	// other brokers
}

func NewBroker() *Broker {
	topics := make([]topickPkg.Topic, 0)
	messages := make(map[string][]message.Message)

	return &Broker{
		Topics:   topics,
		Messages: messages,
	}
}

func (b *Broker) GetOrCreateTopic(topicName string) *topickPkg.Topic {
	topic := b.FindTopic(topicName)
	if topic != nil {
		return topic
	}

	topic = topickPkg.CreateTopic(topicName)
	b.Topics = append(b.Topics, *topic)

	return topic
}

func (b *Broker) FindTopic(topicName string) *topickPkg.Topic {
	for _, topic := range b.Topics {
		if topic.Name == topicName {
			return &topic
		}
	}

	return nil
}

func (b *Broker) AddMessage(topicName string, message message.Message) (*message.Message, error) {
	topic := b.FindTopic(topicName)
	if topic == nil {
		return nil, fmt.Errorf("no topic with name %s exists", topicName)
	}

	b.Messages.AddMessage(topicName, message)

	return &message, nil
}

func (b *Broker) AddSubscriber(topicName string, subscriber subscriber.Subscriber) (*subscriber.Subscriber, error) {
	topic := b.FindTopic(topicName)
	if topic == nil {
		return nil, fmt.Errorf("no topic with name %s exists", topicName)
	}

	b.Subscribers.AddSubscriber(topicName, subscriber)

	return &subscriber, nil
}

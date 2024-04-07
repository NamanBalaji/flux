package service

import (
	"fmt"
	"github.com/NamanBalaji/flux/pkg/model"
)

type Broker struct {
	Topics   []model.Topic
	Messages model.MessagesByTopic
	// subscribers ?
	// other brokers
}

func NewBroker() *Broker {
	topics := make([]model.Topic, 0)
	messages := make(map[string][]model.Message)

	return &Broker{
		Topics:   topics,
		Messages: messages,
	}
}

func (b *Broker) GetOrCreateTopic(topicName string) *model.Topic {
	topic := b.findTopic(topicName)
	if topic != nil {
		return topic
	}

	topic = model.CreateTopic(topicName)
	b.Topics = append(b.Topics, *topic)

	return topic
}

func (b *Broker) findTopic(topicName string) *model.Topic {
	for _, topic := range b.Topics {
		if topic.Name == topicName {
			return &topic
		}
	}

	return nil
}

func (b *Broker) AddMessage(topicName string, message model.Message) (*model.Message, error) {
	topic := b.findTopic(topicName)
	if topic == nil {
		return nil, fmt.Errorf("no topic with name %s exists", topicName)
	}

	b.Messages.AddMessage(topicName, message)

	return &message, nil
}

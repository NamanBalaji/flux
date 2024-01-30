package service

import "github.com/NamanBalaji/flux/internal/broker/model"

type Broker struct {
	DefaultPartitions int
	Topics            map[string]*model.Topic
}

func NewBroker(partitions int) *Broker {
	return &Broker{
		DefaultPartitions: partitions,
		Topics:            make(map[string]*model.Topic),
	}
}

func (b *Broker) GetOrCreateTopic(name string) *model.Topic {
	if topic, ok := b.Topics[name]; ok {
		return topic
	}

	return b.createTopic(name)
}

func (b *Broker) createTopic(name string) *model.Topic {
	partitions := make([]model.Partition, b.DefaultPartitions)
	topic := &model.Topic{Partitions: partitions}
	b.Topics[name] = topic

	return topic
}

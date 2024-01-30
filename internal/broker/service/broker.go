package service

import (
	"fmt"
	"github.com/NamanBalaji/flux/internal/broker/model"
)

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

func (b *Broker) ReadMessage(topicName string, partitionIndex int, offset int) ([]model.Message, error) {
	topic, exist := b.Topics[topicName]
	if !exist {
		return nil, fmt.Errorf("no topic called %s exists", topicName)
	}

	if partitionIndex >= len(topic.Partitions) {
		return nil, fmt.Errorf("partition with index %d not found for topic %s", partitionIndex, topicName)
	}

	partition := topic.Partitions[partitionIndex]
	if offset >= len(partition.Log) {
		return nil, fmt.Errorf("offset %d out of range for partition %d in topic %s", offset, partitionIndex, topicName)
	}

	return partition.Log[offset:], nil
}

package model

import (
	"fmt"
	"hash/fnv"
)

type Topic struct {
	Partitions []Partition
}

func (t *Topic) AssignPartition(key string) (*Partition, error) {
	if len(t.Partitions) == 0 {
		return nil, fmt.Errorf("no partitions available for topic")
	}

	hashFn := fnv.New32a()
	_, _ = hashFn.Write([]byte(key))
	hash := hashFn.Sum32()

	// Use the modulus operator to find the partition index
	partitionIndex := int(hash) % len(t.Partitions)

	return &t.Partitions[partitionIndex], nil
}

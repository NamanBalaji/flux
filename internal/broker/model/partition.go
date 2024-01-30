package model

type Message struct {
	Key    string `json:"key"`
	Value  string `json:"value"`
	Offset int    `json:"offset"`
}

type Partition struct {
	Log []Message
}

func (p *Partition) AppendMessageToPartition(value string, key string) int {
	offset := len(p.Log)
	p.Log = append(p.Log, Message{
		Key:    key,
		Value:  value,
		Offset: offset,
	})

	return offset
}

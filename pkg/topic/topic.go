package topic

import "github.com/NamanBalaji/flux/pkg/message"

type Topics map[string]chan message.Message

func CreateTopics() Topics {
	return make(map[string]chan message.Message)
}

package topic

import (
	"github.com/NamanBalaji/flux/pkg/broker/subscriber"
	"github.com/NamanBalaji/flux/pkg/message"
	"github.com/NamanBalaji/flux/pkg/queue"
	"sync"
)

type Topic struct {
	lock         sync.RWMutex
	Name         string
	MessageChan  chan message.Message
	MessageQueue *queue.Queue
	Subscribers  []*subscriber.Subscriber
}

type Topics map[string]*Topic

func CreateTopics() Topics {
	return Topics{}
}

func CreateTopic(name string) *Topic {
	msgQueue := queue.NewQueue()
	topic := &Topic{
		Name:         name,
		MessageChan:  make(chan message.Message),
		MessageQueue: msgQueue,
	}

	go topic.ManageTopic()

	return topic
}

func (t *Topic) AddMessage(msg message.Message) {
	t.lock.RLock()
	defer t.lock.RUnlock()
	t.MessageQueue.Enqueue(msg)
}

func (t *Topic) ManageTopic() {
	for msg := range t.MessageChan {
		t.deliverMessageToSubscribers(msg)
	}
}

func (t *Topic) deliverMessageToSubscribers(msg message.Message) {
	t.lock.Lock()
	subscribers := t.Subscribers

	subsCopy := make([]*subscriber.Subscriber, len(subscribers))
	for _, sub := range subscribers {
		subsCopy = append(subsCopy, sub)
	}
	t.lock.Unlock()

	for _, sub := range subsCopy {
		sub.AddMessage(msg)
	}
}

func (t *Topic) Subscribe(address string, readOld bool) {
	t.lock.Lock()
	// if old subscriber trying to join back
	for _, sub := range t.Subscribers {
		if sub.Addr == address {
			sub.IsActive = true
			go sub.HandleQueue(t.Name)
			t.lock.Unlock()

			return
		}
	}

	sub := subscriber.NewSubscriber(address)
	if readOld {
		sub = sub.WithQueue(t.MessageQueue.GetCopy())
	}
	t.Subscribers = append(t.Subscribers, sub)
	t.lock.Unlock()

	go sub.HandleQueue(t.Name)
}

func (t *Topic) Unsubscribe(addr string) {
	t.lock.Lock()
	defer t.lock.Unlock()

	for _, s := range t.Subscribers {
		if s.Addr == addr {
			s.IsActive = false
			break
		}
	}
}

func (t *Topic) CleanupSubscribers() {
	t.lock.Lock()
	defer t.lock.Unlock()

	var activeSubscribers []*subscriber.Subscriber
	for _, sub := range t.Subscribers {
		if sub.IsActive {
			activeSubscribers = append(activeSubscribers, sub)
		}
	}

	t.Subscribers = activeSubscribers
}

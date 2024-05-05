package topic

import (
	"context"
	"errors"
	"github.com/NamanBalaji/flux/pkg/broker/subscriber"
	"github.com/NamanBalaji/flux/pkg/config"
	"github.com/NamanBalaji/flux/pkg/message"
	"github.com/NamanBalaji/flux/pkg/queue"
	"sync"
)

type Topic struct {
	lock         sync.Mutex
	Name         string
	MessageChan  chan *message.Message
	MessageQueue *queue.Queue
	MessageSet   map[string]struct{}
	Subscribers  []*subscriber.Subscriber
}

type Topics map[string]*Topic

func CreateTopics() Topics {
	return Topics{}
}

func CreateTopic(name string, bufferSize int) *Topic {
	msgQueue := queue.NewQueue()
	topic := &Topic{
		Name:         name,
		MessageChan:  make(chan *message.Message, bufferSize),
		MessageQueue: msgQueue,
		MessageSet:   make(map[string]struct{}),
	}

	go topic.ManageTopic()

	return topic
}

func (t *Topic) AddMessage(msg *message.Message) bool {
	t.lock.Lock()
	defer t.lock.Unlock()
	if _, ok := t.MessageSet[msg.Id]; ok {
		return false
	}

	t.MessageSet[msg.Id] = struct{}{}
	t.MessageQueue.Enqueue(msg)

	return true
}

func (t *Topic) ManageTopic() {
	for msg := range t.MessageChan {
		t.deliverMessageToSubscribers(msg)
	}
}

func (t *Topic) deliverMessageToSubscribers(msg *message.Message) {
	t.lock.Lock()
	subscribers := t.Subscribers

	subsCopy := make([]*subscriber.Subscriber, len(subscribers))
	for _, sub := range subscribers {
		subsCopy = append(subsCopy, sub)
	}
	t.lock.Unlock()

	for _, sub := range subsCopy {
		sub.AddMessage(msg)
		msg.Ack(sub.Addr)
	}
}

func (t *Topic) Subscribe(ctx context.Context, cfg config.Config, address string, readOld bool) {
	ctx, cancel := context.WithCancel(ctx)

	t.lock.Lock()
	// if old subscriber trying to join back
	for _, sub := range t.Subscribers {
		if sub.Addr == address {
			sub.IsActive = true
			sub.CancelFunc = cancel
			t.lock.Unlock()
			go sub.HandleQueue(ctx, cfg, t.Name)

			return
		}
	}

	sub := subscriber.NewSubscriber(ctx, address)
	if readOld {
		totalMessages := t.MessageQueue.Len()
		for i := 0; i < totalMessages; i++ {
			msg := t.MessageQueue.GetAt(i)
			sub.AddMessage(msg)
			msg.AddSubscriber(sub.Addr)
		}
	}

	t.Subscribers = append(t.Subscribers, sub)
	t.lock.Unlock()

	go sub.HandleQueue(ctx, cfg, t.Name)
}

func (t *Topic) Unsubscribe(addr string) error {
	t.lock.Lock()
	defer t.lock.Unlock()

	for _, s := range t.Subscribers {
		if s.Addr == addr {
			s.IsActive = false

			return nil
		}
	}

	return errors.New("no such topic: " + t.Name)
}

func (t *Topic) CleanupSubscribers() {
	t.lock.Lock()
	defer t.lock.Unlock()

	var activeSubscribers []*subscriber.Subscriber
	for _, sub := range t.Subscribers {
		if !sub.IsActive {
			sub.CancelFunc()

			totalMessages := t.MessageQueue.Len()
			for i := 0; i < totalMessages; i++ {
				msg := t.MessageQueue.GetAt(i)
				msg.RemoveSubscriber(sub.Addr)
			}

			continue
		}

		activeSubscribers = append(activeSubscribers, sub)
	}

	t.Subscribers = activeSubscribers
}

func (t *Topic) CleanupMessages(cfg config.Config) {
	t.lock.Lock()
	defer t.lock.Unlock()
	totalMessages := t.MessageQueue.Len()
	for i := 0; i < totalMessages; i++ {
		if t.MessageQueue.GetAt(i).SafeToDelete(cfg) {
			msg := t.MessageQueue.DeleteAtIndex(i)
			delete(t.MessageSet, msg.Id)
		}
	}
}

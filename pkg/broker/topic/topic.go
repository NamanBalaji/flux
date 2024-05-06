package topic

import (
	"context"
	"errors"
	"log"
	"sync"
	"time"

	"github.com/NamanBalaji/flux/pkg/broker/subscriber"
	"github.com/NamanBalaji/flux/pkg/config"
	"github.com/NamanBalaji/flux/pkg/message"
	"github.com/NamanBalaji/flux/pkg/queue"
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
	if _, ok := t.MessageSet[msg.Id]; ok {
		log.Printf("Topic %s already has a message with id %s \n", t.Name, msg.Id)
		t.lock.Unlock()

		return true
	}
	t.lock.Unlock()

	select {
	case t.MessageChan <- msg:
		t.lock.Lock()
		t.MessageSet[msg.Id] = struct{}{}
		t.MessageQueue.Enqueue(msg)
		t.lock.Unlock()

		log.Printf("Published message with id %s to topic: %s \n", msg.Id, t.Name)
		return true
	default:
		log.Printf("Could not publish message with id %s to topic %s as topic channel is full\n", msg.Id, t.Name)
		return false
	}
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
	for i, sub := range subscribers {
		subsCopy[i] = sub
	}
	t.lock.Unlock()

	for _, sub := range subsCopy {
		sub.AddMessage(msg)
		msg.AddSubscriber(sub.Addr)
	}
}

func (t *Topic) Subscribe(ctx context.Context, cfg config.Config, address string, readOld bool) {
	t.lock.Lock()
	defer t.lock.Unlock()

	// Check if the subscriber already exists and reactivate them
	for _, sub := range t.Subscribers {
		if sub.Addr == address {
			sub.Lock.Lock()
			if !sub.IsActive {
				log.Printf("Subscriber[Address: %s] already exists, reactivating \n", address)

				sub.IsActive = true
				sub.CancelFunc()

				newCtx, cancel := context.WithCancel(ctx)
				sub.CancelFunc = cancel

				sub.Lock.Unlock()

				go sub.HandleQueue(newCtx, cfg, t.Name)

				return
			}

			sub.Lock.Unlock()

			return
		}
	}

	// Creating new subscriber if it does not exist
	newCtx, cancel := context.WithCancel(ctx)
	sub := subscriber.NewSubscriber(newCtx, address)
	sub.CancelFunc = cancel

	if readOld {
		log.Printf("Enqueing old topic %s messages for Subscriber[Address: %s] .\n", t.Name, address)
		totalMessages := t.MessageQueue.Len()
		for i := 0; i < totalMessages; i++ {
			msg := t.MessageQueue.GetAt(i)
			sub.AddMessage(msg)
			msg.AddSubscriber(sub.Addr)
		}
	}

	t.Subscribers = append(t.Subscribers, sub)

	go sub.HandleQueue(newCtx, cfg, t.Name)

	log.Printf("Successfully subscribed Subscriber[Address: %s] to the topic %s \n", address, t.Name)
}

func (t *Topic) Unsubscribe(addr string) error {
	t.lock.Lock()
	defer t.lock.Unlock()

	for _, s := range t.Subscribers {
		s.Lock.Lock()
		if s.Addr == addr {
			s.IsActive = false
			s.Lock.Unlock()
			log.Printf("Subscriber[Address: %s] subscribed to the topic %s set to inactive \n", addr, t.Name)

			return nil
		}
	}

	return errors.New("no such topic: " + t.Name)
}

func (t *Topic) CleanupSubscribers(cfg config.Config) {
	t.lock.Lock()
	defer t.lock.Unlock()

	var activeSubscribers []*subscriber.Subscriber

	for _, sub := range t.Subscribers {
		sub.Lock.Lock()
		if !sub.IsActive && time.Now().Sub(sub.LastActive) > time.Duration(cfg.Subscriber.InactiveTime)*time.Second {
			sub.CancelFunc()
			sub.Lock.Unlock()

			totalMessages := t.MessageQueue.Len()
			var wg sync.WaitGroup
			for i := 0; i < totalMessages; i++ {
				wg.Add(1)
				go func(index int) {
					defer wg.Done()
					msg := t.MessageQueue.GetAt(index)
					msg.RemoveSubscriber(sub.Addr)
				}(i)
			}
			wg.Wait()

			log.Printf("Subscriber[Address: %s] has been deleted from the topic %s\n", sub.Addr, t.Name)
			continue
		}
		sub.Lock.Unlock()

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

			log.Printf("Message with id %s has been deleted from the topic %s \n", msg.Id, t.Name)
		}
	}
}

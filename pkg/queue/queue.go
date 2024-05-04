package queue

import (
	"github.com/NamanBalaji/flux/pkg/message"
	"sync"
)

type Queue struct {
	messages []message.Message
	lock     sync.Mutex
	cond     *sync.Cond
}

func NewQueue() *Queue {
	q := &Queue{}
	q.cond = sync.NewCond(&q.lock)

	return q
}

func (q *Queue) Enqueue(msg message.Message) {
	q.lock.Lock()
	defer q.lock.Unlock()

	q.messages = append(q.messages, msg)
	q.cond.Signal()
}

func (q *Queue) Dequeue() message.Message {
	q.lock.Lock()
	defer q.lock.Unlock()

	for len(q.messages) == 0 {
		q.cond.Wait()
	}

	msg := q.messages[0]
	q.messages = q.messages[1:]

	return msg
}

func (q *Queue) Peek() message.Message {
	q.lock.Lock()
	defer q.lock.Unlock()

	for len(q.messages) == 0 {
		q.cond.Wait()
	}

	msg := q.messages[0]

	return msg
}

func (q *Queue) GetCopy() *Queue {
	return &Queue{
		messages: q.messages,
	}
}

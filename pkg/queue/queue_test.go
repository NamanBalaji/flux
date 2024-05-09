package queue

import (
	"sync"
	"testing"

	"github.com/NamanBalaji/flux/pkg/message"
)

func TestEnqueueAndDequeue(t *testing.T) {
	q := NewQueue()
	msg1 := &message.Message{Id: "1", Payload: "first"}
	msg2 := &message.Message{Id: "2", Payload: "second"}

	q.Enqueue(msg1)
	q.Enqueue(msg2)

	if q.Len() != 2 {
		t.Errorf("Expected queue length of 2, got %d", q.Len())
	}

	deqMsg1 := q.Dequeue()
	if deqMsg1 != msg1 {
		t.Errorf("Expected first dequeued message to be msg1, got %+v", deqMsg1)
	}

	deqMsg2 := q.Dequeue()
	if deqMsg2 != msg2 {
		t.Errorf("Expected second dequeued message to be msg2, got %+v", deqMsg2)
	}

	if q.Len() != 0 {
		t.Errorf("Expected queue length of 0 after dequeuing, got %d", q.Len())
	}
}

func TestPeek(t *testing.T) {
	q := NewQueue()
	msg := &message.Message{Id: "1", Payload: "test message"}

	q.Enqueue(msg)
	peekedMsg := q.Peek()

	if peekedMsg != msg {
		t.Errorf("Peeked message does not match enqueued message")
	}

	if q.Len() != 1 {
		t.Errorf("Queue length should be 1 after peeking, got %d", q.Len())
	}
}

func TestGetAtAndDeleteAtIndex(t *testing.T) {
	q := NewQueue()
	msg1 := &message.Message{Id: "1", Payload: "first"}
	msg2 := &message.Message{Id: "2", Payload: "second"}
	msg3 := &message.Message{Id: "3", Payload: "third"}

	q.Enqueue(msg1)
	q.Enqueue(msg2)
	q.Enqueue(msg3)

	if q.GetAt(1) != msg2 {
		t.Errorf("GetAt failed to retrieve the correct message at index 1")
	}

	deletedMsg := q.DeleteAtIndex(1)
	if deletedMsg != msg2 {
		t.Errorf("DeleteAtIndex did not delete the correct message")
	}

	if q.Len() != 2 {
		t.Errorf("Expected queue length of 2 after deletion, got %d", q.Len())
	}

	if q.GetAt(1) != msg3 {
		t.Errorf("Expected msg3 at index 1 after deletion, got %+v", q.GetAt(1))
	}
}

func TestQueueConcurrency(t *testing.T) {
	q := NewQueue()
	start := make(chan struct{})

	wg := sync.WaitGroup{}
	wg.Add(2)

	go func() {
		<-start
		for i := 0; i < 100; i++ {
			q.Enqueue(&message.Message{Id: string(rune(i)), Payload: "payload"})
		}

		wg.Done()
	}()

	go func() {
		<-start
		for i := 0; i < 100; i++ {
			q.Dequeue()
		}

		wg.Done()
	}()

	close(start)
	wg.Wait()

	if q.Len() != 0 {
		t.Errorf("Expected empty queue after concurrent enqueue and dequeue, got length %d", q.Len())
	}
}

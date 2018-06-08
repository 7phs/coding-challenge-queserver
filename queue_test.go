package main

import (
	"testing"
	"container/heap"
	"reflect"
	"time"
)

func TestMsgHeap(t *testing.T) {
	msgHeap := MsgHeap{}

	for _, sequenceId := range []int64{
		1024,
		1000,
		900,
		950,
		980,
		500,
	} {
		heap.Push(&msgHeap, &Message{
			sequenceId: sequenceId,
		})
	}

	exist := []int64{}
	expected := []int64{500, 900, 950, 980, 1000, 1024}
	for msgHeap.Len() > 0 {
		v := heap.Pop(&msgHeap)
		if v == nil {
			break
		}

		exist = append(exist, v.(*Message).sequenceId)
	}

	if !reflect.DeepEqual(exist, expected) {
		t.Error("failed to pop messages in order. Got ", exist, ", but expected is ", expected)
	}
}

type TestQueue struct {
	sequencesId []int64
}

func (o *TestQueue) PushMessage(msg *Message) {
	o.sequencesId = append(o.sequencesId, msg.sequenceId)
}

func TestNewQueue(t *testing.T) {
	testQueue := TestQueue{}

	queue := NewQueue(&Config{
		queueTTL:   24 * 60 * 1000,
		queueLimit: 0,
	}, &testQueue)
	defer queue.Shutdown()

	queue.Run()

	expectedCount := int64(10)
	expectedIds := []int64{}

	queue.PushMessage(&Message{
		sequenceId: 0,
		typ:        MESSAGE_UNKNOWN,
	})

	for i := int64(1); i <= expectedCount; i++ {
		expectedIds = append(expectedIds, i)

		queue.PushMessage(&Message{
			sequenceId: i,
			typ:        MESSAGE_PRIVATE_MSG,
		})
	}

	time.Sleep(50 * time.Millisecond)

	// last msg is waiting for a signal, but without timer or the next message never pull
	if exist := len(testQueue.sequencesId); exist != int(expectedCount-1) {
		t.Error("failed to get all pushed messages. Got ", exist, ", but expected is ", expectedCount-1)
	}

	if !reflect.DeepEqual(testQueue.sequencesId, expectedIds[:expectedCount-1]) {
		t.Error("failed to get all pushed messages. Got ", testQueue.sequencesId, ", but expected is ", expectedIds[:expectedCount-1])
	}

}

func TestNewQueue_PullByTTL(t *testing.T) {
	testQueue := TestQueue{}

	queue := NewQueue(&Config{
		queueTTL:   5,
		queueLimit: 1000,
	}, &testQueue)
	defer queue.Shutdown()

	queue.Run()

	expectedCount := int64(10)
	expectedIds := []int64{}

	for i := int64(1); i <= expectedCount; i++ {
		expectedIds = append(expectedIds, i)

		queue.PushMessage(&Message{
			sequenceId: i,
			typ:        MESSAGE_PRIVATE_MSG,
		})
	}

	time.Sleep(50 * time.Millisecond)

	// last msg is waiting for a signal, but without timer or the next message never pull
	if exist := len(testQueue.sequencesId); exist != int(expectedCount) {
		t.Error("failed to get all pushed messages. Got ", exist, ", but expected is ", expectedCount)
	}

	if !reflect.DeepEqual(testQueue.sequencesId, expectedIds) {
		t.Error("failed to get all pushed messages. Got ", testQueue.sequencesId, ", but expected is ", expectedIds)
	}

}

package main

import (
	"sync"
	"container/heap"
	"time"
	"github.com/7phs/coding-challenge-queserver/logger"
)

// An IntHeap is a min-heap of ints.
type MsgHeap []*Message

func (h MsgHeap) Len() int           { return len(h) }
func (h MsgHeap) Less(i, j int) bool { return h[i].sequenceId < h[j].sequenceId }
func (h MsgHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }

func (h *MsgHeap) Push(x interface{}) {
	// Push and Pop use pointer receivers because they modify the slice's length,
	// not just its contents.
	*h = append(*h, x.(*Message))
}

func (h *MsgHeap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}

func (h *MsgHeap) Peek() *Message {
	return (*h)[0]
}

type Queue struct {
	sync.Mutex

	queue  MsgHeap
	pullCh chan int64

	queueLimit int64
	queueTTL   time.Duration

	chain MessageQueue

	shutdown chan struct{}
	wait     sync.WaitGroup
}

func NewQueue(config *Config, queue MessageQueue) *Queue {
	q := &Queue{
		chain:  queue,
		pullCh: make(chan int64),

		shutdown: make(chan struct{}),

		queueLimit: config.QueueLimit(),
		queueTTL:   config.QueueTTL(),
	}

	heap.Init(&q.queue)

	return q
}

func (o *Queue) PushMessage(msg *Message) {
	if !msg.IsValid() {
		logger.Debug("[QUEUE]: push invalid message ", msg.payload)

		return
	}

	logger.Debug("[QUEUE]: push message ", msg.payload)

	// better use lock free queue. Now, it is trade-off
	peakSequenceId := func () int64 {
		o.Lock()
		defer o.Unlock()

		heap.Push(&o.queue, msg)

		return o.queue.Peek().sequenceId
	}()

	if msg.sequenceId>o.queueLimit && msg.sequenceId - peakSequenceId > o.queueLimit {
		limitId := msg.sequenceId-o.queueLimit

		go func() { o.pullCh <- limitId }()
	}
}

func (o *Queue) Run() {
	o.wait.Add(1)

	go func() {
		logger.Info("[QUEUE]: start working goroutin")

		for {
			// check shutdown
			select {
			// start to send all msg older than limitId (by sequenceId)
			case limitId := <-o.pullCh:
				o.pullByLimit(limitId)

			// start to send all msg stored older than queueTTL
			case <-time.After(o.queueTTL):
				o.pullByTTL()

			case <-o.shutdown:
				logger.Info("[QUEUE]: shutdown working goroutin")
				o.wait.Done()
				return
			}
		}
	}()
}

func (o *Queue) pullByLimit(limitId int64) {
	if o.queueLimit>10 {
		limitId -= limitId%(o.queueLimit/10)
	}

	logger.Debug("[QUEUE]: pull messages by limit ", limitId)

	pullQueues := func () []*Message {
		o.Lock()
		defer o.Unlock()

		result := make([]*Message, 0, MaxInt64(16, limitId - o.queue.Peek().sequenceId))

		for o.queue.Len()>0 && o.queue.Peek().sequenceId<limitId {
			result = append(result, heap.Pop(&o.queue).(*Message))
		}

		return result
	}()

	o.pullMessages(pullQueues)
}

func (o *Queue) pullByTTL() {
	limit := time.Now().Add(-o.queueTTL)

	logger.Debug("[QUEUE]: pull messages by TTL ", limit)

	pullQueues := func () []*Message {
		o.Lock()
		defer o.Unlock()

		result := make([]*Message, 0, 128)

		for o.queue.Len() > 0 && o.queue.Peek().created.Before(limit) {
			result = append(result, heap.Pop(&o.queue).(*Message))
		}

		return result
	}()

	o.pullMessages(pullQueues)
}

func (o *Queue) pullMessages(msgs []*Message) {
	for _, msg := range msgs {
		logger.Debug("[QUEUE]: pull message ", msg)

		o.chain.PushMessage(msg)
	}
}

func (o *Queue) Shutdown() {
	logger.Info("[QUEUE]: shutdown")

	close(o.shutdown)

	o.wait.Wait()
}

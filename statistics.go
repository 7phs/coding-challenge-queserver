package main

import (
	"sync"
	"sync/atomic"
	"time"
	"bytes"
	"fmt"
	"github.com/7phs/coding-challenge-queserver/logger"
)

const (
	STATISTICS_DUMP_INTERVAL = 2 * time.Second
)

var (
	STATISITICS_DUMP_INFO = []MessageType{
		MESSAGE_BROADCAST,
		MESSAGE_FOLLOW,
		MESSAGE_STATUS_UPDATE,
		MESSAGE_UNFOLLOW,
		MESSAGE_PRIVATE_MSG,
	}
)

type Direction int

const (
	MESSAGE_RECIEVE Direction = iota + 1
	MESSAGE_SEND
)

type Statistics struct {
	runDumping sync.Once

	received      []uint64
	receivedTotal uint64
	sent          []uint64
	sentTotal     uint64

	shutdown chan struct{}
	wait     sync.WaitGroup
}

func NewStatistics() *Statistics {
	return &Statistics{
		received: make([]uint64, MESSAGE_UNKNOWN),
		sent:     make([]uint64, MESSAGE_UNKNOWN),
		shutdown: make(chan struct{}),
	}
}

func (o *Statistics) Add(direction Direction, messageType MessageType) {
	go func() {
		o.runDumping.Do(func() {
			o.wait.Add(1)
			go o.Working()
		})

		switch direction {
		case MESSAGE_RECIEVE:
			atomic.AddUint64(&o.received[messageType], 1)
			atomic.AddUint64(&o.receivedTotal, 1)
		case MESSAGE_SEND:
			atomic.AddUint64(&o.sent[messageType], 1)
			atomic.AddUint64(&o.sentTotal, 1)
		}
	}()
}

func (o *Statistics) Working() {
	logger.Info("[STATISTICS]: start working goroutin")

	for {
		select {
		case <-time.After(STATISTICS_DUMP_INTERVAL):
			logger.Info("[STATISTICS]: " + o.DumpState())
		case <-o.shutdown:
			logger.Info("[STATISTICS]: stop working goroutin")
			o.wait.Done()
			return
		}
	}
}

func (o *Statistics) DumpState() string {
	line := bytes.NewBufferString("Received/sent: ")
	add := 0

	for _, messageType := range STATISITICS_DUMP_INFO {
		received := atomic.LoadUint64(&o.received[messageType])
		sent := atomic.LoadUint64(&o.sent[messageType])

		if received > 0 || sent > 0 {
			if add > 0 {
				line.WriteString(", ")
			}

			line.WriteString(fmt.Sprint(messageType.String(), " -> ", received, "/", sent))

			add++
		}
	}

	if add > 0 {
		line.WriteString(", ")
	}

	received := atomic.LoadUint64(&o.receivedTotal)
	sent := atomic.LoadUint64(&o.sentTotal)

	line.WriteString(fmt.Sprint("total -> ", received, "/", sent))

	return line.String()
}

func (o *Statistics) Shutdown() {
	logger.Info("[STATISTICS]: shutdown")

	close(o.shutdown)

	o.wait.Wait()
}

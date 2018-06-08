package main

import (
	"net"
	"bufio"
	"sync"
	"github.com/7phs/coding-challenge-queserver/logger"
)

const (
	EVENT_LOG_INTERVAL = 100000
)

type EventSource struct {
	addr string

	listener   net.Listener
	queue      MessageQueue
	statistics *Statistics

	shutdown chan struct{}
	wait     sync.WaitGroup
}

func NewEventSource(config *Config, queue MessageQueue, statistics *Statistics) (*EventSource, error) {
	return (&EventSource{
		queue:      queue,
		statistics: statistics,
		addr:       config.EventSource(),
		shutdown:   make(chan struct{}),
	}).Listen()
}

func (o *EventSource) Listen() (s *EventSource, err error) {
	s = o

	logger.Info("[EVENT_SOURCE]: listen ", o.addr)

	o.listener, err = net.Listen("tcp", o.addr)

	return
}

func (o *EventSource) Run() {
	o.wait.Add(1)

	go func() {
		logger.Info("[EVENT_SOURCE]: start working goroutin")

		acceptCh := make(chan interface{})

		for {
			// read message
			go func() {
				conn, err := o.listener.Accept()

				if err != nil {
					acceptCh <- err
				} else {
					acceptCh <- conn
				}
			}()

			// check shutdown
			select {
			case v := <-acceptCh:
				switch i := v.(type) {
				case error:
					logger.Error("[EVENT_SOURCE]: error while accept connection messages: ", i)

				case net.Conn:
					logger.Debug("[EVENT_SOURCE]: accept a connection")
					go o.handleConnection(i)
				}

			case <-o.shutdown:
				logger.Info("[EVENT_SOURCE]: shutdown working goroutin")
				o.wait.Done()
				return
			}

		}
	}()
}

func (o *EventSource) handleConnection(conn net.Conn) {
	o.wait.Add(1)
	go func() {
		logger.Info("[EVENT_SOURCE]: start processing goroutin")

		reader := bufio.NewReader(conn)
		readCh := make(chan interface{})
		for {
			// read message
			go func() {
				line, _, err := reader.ReadLine()
				if err != nil {
					readCh <- err
				} else {
					readCh <- line
				}
			}()

			// check shutdown
			select {
			case v := <-readCh:
				switch i := v.(type) {
				case error:
					logger.Error("[EVENT_SOURCE]: error while receive messages: ", i)
					logger.Info("[EVENT_SOURCE]: stop processing goroutin")
					o.wait.Done()
					return

				case []byte:
					msg := NewMessage(string(i))
					// push message
					logger.Debug("[EVENT_SOURCE]: receive a message: ", msg)

					o.statistics.Add(MESSAGE_RECIEVE, msg.typ)
					o.queue.PushMessage(msg)
				}

			case <-o.shutdown:
				logger.Info("[EVENT_SOURCE]: shutdown processing goroutin")
				o.wait.Done()
				return
			}
		}
	}()
}

func (o *EventSource) Shutdown() {
	logger.Info("[EVENT_SOURCE]: shutdown")

	close(o.shutdown)

	o.wait.Wait()

	if o.listener != nil {
		o.listener.Close()
	}
}

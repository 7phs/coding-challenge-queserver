package main

import (
	"testing"
	"net"
	"math/rand"
	"fmt"
	"sync"
	"bufio"
)

const (
	TEST_REGISTER_INIT         = iota
	TEST_REGISTER_REGISTERED
	TEST_REGISTER_UNREGISTERED
)

type TestClientRouter struct {
	ch chan *Message

	userId     int64
	registered int
}

func NewTestClientRouter() *TestClientRouter {
	return &TestClientRouter{
		ch:         make(chan *Message),
		registered: TEST_REGISTER_INIT,
	}
}

func (o *TestClientRouter) PushMessage(msg *Message) {
	o.ch <- msg
}

func (o *TestClientRouter) RegisterClient(userId int64) <-chan *Message {
	o.userId = userId
	o.registered = TEST_REGISTER_REGISTERED

	return o.ch
}

func (o *TestClientRouter) UnregisterClient(userId int64) {
	if o.userId == userId {
		o.registered = TEST_REGISTER_UNREGISTERED
	}
}

func TestNewClient(t *testing.T) {
	randPort := fmt.Sprintf(":%d", 16000+rand.Intn(60000-16000))
	userId := int64(1000 + rand.Intn(60000))

	listener, err := net.Listen("tcp", randPort)
	if err != nil {
		t.Error("failed to start listening client ", randPort, " with error ", err)
	}
	defer listener.Close()

	testRouter := NewTestClientRouter()
	statistics := NewStatistics()
	shutdown := make(chan struct{})

	var client *Client

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				// TODO: log
				return
			}

			select {
			case <-shutdown:
				return
			default:
			}

			client = NewClient(conn, testRouter, statistics, shutdown)

			go client.Run()
		}
	}()

	connection, err := net.Dial("tcp", randPort)
	if err != nil {
		t.Error("failed to connect as a client to ", randPort, " with error: ", err)
		close(shutdown)
		return
	}

	connection.Write([]byte(fmt.Sprintf("%d\r\n", userId)))

	var (
		wait     sync.WaitGroup
		exist    []byte
		expected = "666|F|60|50"
	)

	wait.Add(1)
	go func() {
		testRouter.PushMessage(NewMessage(expected))

		wait.Done()
	}()

	wait.Add(1)
	go func() {
		reader := bufio.NewReader(connection)

		exist, _, err = reader.ReadLine()
		if err != nil {
			t.Error("failed to read data from connection")
		}

		wait.Done()
	}()

	wait.Wait()

	if err := client.HasError(); err != nil {
		t.Error("failed to init client with error ", err)
	}

	if string(exist) != expected {
		t.Error("failed to read pushed messages. Got '", string(exist), "', but expected is '", expected, "'")
	}

	if testRouter.userId != userId {
		t.Error("error register user id. Got ", userId, ", but expected is ", testRouter.userId)
	}

	if testRouter.registered != TEST_REGISTER_REGISTERED {
		t.Error("failed to register user id. Got status ", testRouter.registered, ", but expected is ", TEST_REGISTER_REGISTERED)
	}

	connection.Close()
	client.Unregister()

	if testRouter.registered != TEST_REGISTER_UNREGISTERED {
		t.Error("failed to unregister user id. Got status ", testRouter.registered, ", but expected is ", TEST_REGISTER_UNREGISTERED)
	}

	close(shutdown)
}

func TestNewClient_Err(t *testing.T) {
	randPort := fmt.Sprintf(":%d", 16000+rand.Intn(60000-16000))

	listener, err := net.Listen("tcp", randPort)
	if err != nil {
		t.Error("failed to start listening client ", randPort, " with error ", err)
	}
	defer listener.Close()

	testRouter := NewTestClientRouter()
	statistics := NewStatistics()
	shutdown := make(chan struct{})

	var client *Client

	var wait sync.WaitGroup

	wait.Add(1)
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				// TODO: log
				return
			}

			select {
			case <-shutdown:
				return
			default:
			}

			client = NewClient(conn, testRouter, statistics, shutdown)

			wait.Done()

			go client.Run()
		}
	}()

	connection, err := net.Dial("tcp", randPort)
	if err != nil {
		t.Error("failed to connect as a client to ", randPort, " with error: ", err)
		close(shutdown)
		return
	}

	connection.Write([]byte("unknown\r\n"))

	wait.Wait()

	if client.HasError() == nil {
		t.Error("failed to catch an error")
	}

	close(shutdown)
}

package main

import (
	"fmt"
	"math/rand"
	"net"
	"sync"
	"bufio"
	"testing"
)

func TestNewServer(t *testing.T) {
	randPort := fmt.Sprintf(":%d", 16000+rand.Intn(60000-16000))
	userId := int64(1000 + rand.Intn(60000))
	testRouter := NewTestClientRouter()
	statisitics := NewStatistics()

	server, err := NewServer(&Config{
		client: randPort,
	}, testRouter, statisitics)

	if err!=nil {
		t.Error("failed to implement a server with err: ", err)
		return
	}
	defer server.Shutdown()

	go server.Run()

	connection, err := net.Dial("tcp", randPort)
	if err != nil {
		t.Error("failed to connect as a client to ", randPort, " with error: ", err)
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

	if testRouter.userId != userId {
		t.Error("error register user id. Got ", userId, ", but expected is ", testRouter.userId)
	}

	if testRouter.registered != TEST_REGISTER_REGISTERED {
		t.Error("failed to register user id. Got status ", testRouter.registered, ", but expected is ", TEST_REGISTER_REGISTERED)
	}

	if string(exist)!=expected {
		t.Error("failed to read a messag. Got '", string(exist), "', but expected is '", expected, "'")
	}
}

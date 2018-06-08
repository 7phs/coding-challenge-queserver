package main

import (
	"testing"
	"fmt"
	"math/rand"
	"net"
	"sync"
	"reflect"
	"sort"
	"bufio"
)

func TestNewEventSource(t *testing.T) {
	randPort := fmt.Sprintf(":%d", 16000+rand.Intn(60000-16000))
	testRouter := NewTestClientRouter()
	statistics := NewStatistics()

	eventSource, err := NewEventSource(&Config{
		eventSource: randPort,
	}, testRouter, statistics)

	if err != nil {
		t.Error("failed to implement an event source with err: ", err)
		return
	}
	defer eventSource.Shutdown()

	go eventSource.Run()

	connection, err := net.Dial("tcp", randPort)
	if err != nil {
		t.Error("failed to connect as a client to ", randPort, " with error: ", err)
		return
	}

	var (
		flow     = []byte("666|F|60|50\r\n1|U|12|9\r\n542532|B\r\n43|P|32|56\r\n")
		expexted = []string{
			"666|F|60|50",
			"1|U|12|9",
			"542532|B",
			"43|P|32|56",
		}
		exist = []string{}
		wait  sync.WaitGroup
		shutdown = make(chan struct{})
	)

	wait.Add(1)
	go func() {
		writer := bufio.NewWriter(connection)

		_, err = writer.Write(flow)
		if err != nil {
			t.Error("failed to write data to connection: ", err)
		}

		writer.Flush()

		wait.Done()
	}()

	wait.Add(1)
	go func() {
		counter := 0

		for {
			select {
			case msg := <-testRouter.ch:
				counter++
				exist = append(exist, msg.payload)

				if counter == len(expexted) {
					wait.Done()
				}

			case <-shutdown:
				return
			}
		}
	}()
	wait.Wait()

	close(shutdown)

	sort.StringSlice(exist).Sort()
	sort.StringSlice(expexted).Sort()

	if !reflect.DeepEqual(exist, expexted) {
		t.Error("failed to read data from event sources. Got ", exist, ", but expected is ", expexted)
	}

}

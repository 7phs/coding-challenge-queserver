package main

import (
	"testing"
	"reflect"
	"time"
)

func TestNewStatistics(t *testing.T) {
	types := []MessageType{MESSAGE_BROADCAST, MESSAGE_FOLLOW, MESSAGE_STATUS_UPDATE, MESSAGE_UNFOLLOW, MESSAGE_PRIVATE_MSG}

	statistics := NewStatistics()
	defer statistics.Shutdown()

	for j, d := range []Direction{MESSAGE_SEND, MESSAGE_RECIEVE} {
		for _, v := range types {
			for i := 0; i < j+1; i++ {
				statistics.Add(d, v)
			}
		}
	}
	// wait for go routing
	time.Sleep(100 * time.Millisecond)

	expected := []uint64{0, 2, 2, 2, 2, 2}
	if !reflect.DeepEqual(statistics.received, expected) {
		t.Error("failed to calc all recieved items. Got ", statistics.received, ", but expected is ", expected)
	}
	if statistics.receivedTotal!=10 {
		t.Error("failed to calc total recieved items. Got ", statistics.receivedTotal, ", but expected is ", 10)
	}

	expected = []uint64{0, 1, 1, 1, 1, 1}
	if !reflect.DeepEqual(statistics.sent, expected) {
		t.Error("failed to calc all sent items. Got ", statistics.sent, ", but expected is ", expected)
	}
	if statistics.sentTotal!=5 {
		t.Error("failed to calc total recieved items. Got ", statistics.sentTotal, ", but expected is ", 5)
	}

	expectedStr := "Received/sent: Broadcast -> 2/1, Follow -> 2/1, StatusUpdate -> 2/1, Unfollow -> 2/1, Private -> 2/1, total -> 10/5"
	exist := statistics.DumpState()
	if exist!=expectedStr {
		t.Error("failed to get state. Got '", exist, "', but expected is '", expectedStr, "'")
	}
}

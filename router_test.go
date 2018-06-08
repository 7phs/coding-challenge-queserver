package main

import (
	"testing"
	"time"
	"sync/atomic"
)

const (
	TEST_ID_PARTITION = 100
)

func TestUser_dumpSubscriptions(t *testing.T) {
	userInfo := UserInfo{}

	exist := userInfo.dumpSubscriptions()
	expected := ""
	if exist!=expected {
		t.Error("failed to dump empty subscriptions")
	}

	userInfo.Follow(12)
	userInfo.Follow(375)

	exist = userInfo.dumpSubscriptions()
	expected = "12, 375"
	expected2 := "375, 12"
	if exist!=expected && exist!=expected2 {
		t.Error("failed to dump user subscriptions. Got '", exist, "', but expected is '", expected, "'")
	}

	userInfo.Unfollow(375)

	exist = userInfo.dumpSubscriptions()
	expected = "12"
	if exist!=expected {
		t.Error("failed to dump user subscriptions after unfollow. Got '", exist, "', but expected is '", expected, "'")
	}
}

func TestNewRouter(t *testing.T) {
	router := NewRouter()

	usersId := []int64{
		2*TEST_ID_PARTITION + 56,
		2*TEST_ID_PARTITION + 87,
		3*TEST_ID_PARTITION + 12,
		3*TEST_ID_PARTITION + 56,
	}

	for _, userId := range usersId {
		router.RegisterClient(userId)
	}

	expectedCount := 4
	existCount := 0

	router.clients.Range(func(_, _ interface{}) bool {
		existCount++

		return true
	})

	if existCount != expectedCount {
		t.Error("failed to count registered user. Got ", existCount, ", but expected is ", expectedCount)
	}
}

func TestNewRouter_RegisterUnregister(t *testing.T) {
	router := NewRouter()

	usersId := []int64{
		2*TEST_ID_PARTITION + 56,
		2*TEST_ID_PARTITION + 87,
		3*TEST_ID_PARTITION + 12,
		3*TEST_ID_PARTITION + 56,
	}

	unknownUsersId := []int64{
		57,
		2*TEST_ID_PARTITION + 57,
		2*TEST_ID_PARTITION + 88,
		3*TEST_ID_PARTITION + 13,
		3*TEST_ID_PARTITION + 57,
	}

	for _, userId := range usersId {
		router.RegisterClient(userId)
	}

	expectedCount := 9
	existCount := 0

	calc := func() {
		for _, userId := range usersId {
			if router.getOrAddUserInfo(userId) != nil {
				existCount++
			}
		}

		for _, userId := range unknownUsersId {
			if router.getOrAddUserInfo(userId) != nil {
				existCount++
			}
		}
	}

	calc()

	if existCount != expectedCount {
		t.Error("failed to count of registered user. Got ", existCount, ", but expected is ", expectedCount)
	}

	expectedCount = 9
	existCount = 0

	router.UnregisterClient(usersId[0])
	router.UnregisterClient(unknownUsersId[0])

	calc()

	if existCount != expectedCount {
		t.Error("failed to count of registered user after unregister one. Got ", existCount, ", but expected is ", expectedCount)
	}
}

func TestRouter_PushMessage(t *testing.T) {
	router := NewRouter()

	shutdown := make(chan struct{})

	usersId := map[int64]*struct {
		ch      <-chan *Message
		counter int32
	}{
		2*TEST_ID_PARTITION + 1: {},
		2*TEST_ID_PARTITION + 2: {},
		3*TEST_ID_PARTITION + 1: {},
		3*TEST_ID_PARTITION + 2: {},
		3*TEST_ID_PARTITION + 3: {},
		3*TEST_ID_PARTITION + 4: {},
		3*TEST_ID_PARTITION + 5: {},
		3*TEST_ID_PARTITION + 6: {},
	}

	for userId := range usersId {
		usersId[userId].ch = router.RegisterClient(userId)

		go func(userId int64) {
			for {
				select {
				case <-usersId[userId].ch:
					atomic.AddInt32(&usersId[userId].counter, 1)

				case <-shutdown:
					return
				}
			}
		}(userId)
	}

	checkCounters := func(title string, expected map[int64]int32) {
		for userId, expectedCounter := range expected {
			if _, ok:=usersId[userId]; !ok {
				t.Error(title, ": failed to check expected counters for user #", userId, " - it's not exists.")
				continue
			}

			if exist:=atomic.LoadInt32(&usersId[userId].counter); exist != expectedCounter {
				t.Error(title, ": failed to receive message all expected counters for user #", userId, ". Got ", exist, ", but expected is ", expectedCounter)
			}
		}
	}

	router.PushMessage(&Message{
		payload: "msg1",
		typ: MESSAGE_BROADCAST,
	})

	time.Sleep(50 * time.Millisecond)

	checkCounters("Broadcast", map[int64]int32 {
		2*TEST_ID_PARTITION + 1: 1,
		2*TEST_ID_PARTITION + 2: 1,
		3*TEST_ID_PARTITION + 1: 1,
		3*TEST_ID_PARTITION + 2: 1,
		3*TEST_ID_PARTITION + 3: 1,
		3*TEST_ID_PARTITION + 4: 1,
		3*TEST_ID_PARTITION + 5: 1,
		3*TEST_ID_PARTITION + 6: 1,
	})

	dstId := int64(2*TEST_ID_PARTITION + 1)

	router.PushMessage(&Message{
		payload: "msg1",
		typ: MESSAGE_FOLLOW,
		from: 3*TEST_ID_PARTITION + 1,
		to: dstId,
	})

	router.PushMessage(&Message{
		payload: "msg1",
		typ: MESSAGE_FOLLOW,
		from: 3*TEST_ID_PARTITION + 2,
		to: dstId,
	})

	router.PushMessage(&Message{
		payload: "msg1",
		typ: MESSAGE_FOLLOW,
		from: 3*TEST_ID_PARTITION + 3,
		to: dstId,
	})

	dstId = int64(2*TEST_ID_PARTITION + 2)

	router.PushMessage(&Message{
		payload: "msg1",
		typ: MESSAGE_FOLLOW,
		from: 3*TEST_ID_PARTITION + 4,
		to: dstId,
	})

	router.PushMessage(&Message{
		payload: "msg1",
		typ: MESSAGE_FOLLOW,
		from: 3*TEST_ID_PARTITION + 5,
		to: dstId,
	})

	router.PushMessage(&Message{
		payload: "msg1",
		typ: MESSAGE_FOLLOW,
		from: 3*TEST_ID_PARTITION + 6,
		to: dstId,
	})

	time.Sleep(50 * time.Millisecond)

	checkCounters("Follow", map[int64]int32 {
		2*TEST_ID_PARTITION + 1: 4,
		2*TEST_ID_PARTITION + 2: 4,
		3*TEST_ID_PARTITION + 1: 1,
		3*TEST_ID_PARTITION + 2: 1,
		3*TEST_ID_PARTITION + 3: 1,
		3*TEST_ID_PARTITION + 4: 1,
		3*TEST_ID_PARTITION + 5: 1,
		3*TEST_ID_PARTITION + 6: 1,
	})

	router.PushMessage(&Message{
		payload: "msg1",
		typ: MESSAGE_PRIVATE_MSG,
		from: 3*TEST_ID_PARTITION + 5,
		to: 3*TEST_ID_PARTITION + 6,
	})

	time.Sleep(50 * time.Millisecond)

	checkCounters("Private", map[int64]int32 {
		2*TEST_ID_PARTITION + 1: 4,
		2*TEST_ID_PARTITION + 2: 4,
		3*TEST_ID_PARTITION + 1: 1,
		3*TEST_ID_PARTITION + 2: 1,
		3*TEST_ID_PARTITION + 3: 1,
		3*TEST_ID_PARTITION + 4: 1,
		3*TEST_ID_PARTITION + 5: 1,
		3*TEST_ID_PARTITION + 6: 2,
	})

	router.PushMessage(&Message{
		payload: "msg1",
		typ: MESSAGE_STATUS_UPDATE,
		from: int64(2*TEST_ID_PARTITION + 1),
	})

	time.Sleep(50 * time.Millisecond)

	checkCounters("Status update", map[int64]int32 {
		2*TEST_ID_PARTITION + 1: 4,
		2*TEST_ID_PARTITION + 2: 4,
		3*TEST_ID_PARTITION + 1: 2,
		3*TEST_ID_PARTITION + 2: 2,
		3*TEST_ID_PARTITION + 3: 2,
		3*TEST_ID_PARTITION + 4: 1,
		3*TEST_ID_PARTITION + 5: 1,
		3*TEST_ID_PARTITION + 6: 2,
	})

	router.PushMessage(&Message{
		payload: "msg1",
		typ: MESSAGE_UNFOLLOW,
		from: 3*TEST_ID_PARTITION + 6,
		to: 2*TEST_ID_PARTITION + 2,
	})

	time.Sleep(50 * time.Millisecond)

	checkCounters("Unfollow", map[int64]int32 {
		2*TEST_ID_PARTITION + 1: 4,
		2*TEST_ID_PARTITION + 2: 4,
		3*TEST_ID_PARTITION + 1: 2,
		3*TEST_ID_PARTITION + 2: 2,
		3*TEST_ID_PARTITION + 3: 2,
		3*TEST_ID_PARTITION + 4: 1,
		3*TEST_ID_PARTITION + 5: 1,
		3*TEST_ID_PARTITION + 6: 2,
	})

	router.PushMessage(&Message{
		payload: "msg1",
		typ: MESSAGE_STATUS_UPDATE,
		from: int64(2*TEST_ID_PARTITION + 2),
	})

	time.Sleep(50 * time.Millisecond)

	checkCounters("Status update after Unfollow", map[int64]int32 {
		2*TEST_ID_PARTITION + 1: 4,
		2*TEST_ID_PARTITION + 2: 4,
		3*TEST_ID_PARTITION + 1: 2,
		3*TEST_ID_PARTITION + 2: 2,
		3*TEST_ID_PARTITION + 3: 2,
		3*TEST_ID_PARTITION + 4: 2,
		3*TEST_ID_PARTITION + 5: 2,
		3*TEST_ID_PARTITION + 6: 2,
	})

	router.PushMessage(&Message{
		payload: "msg1",
		typ: MESSAGE_UNKNOWN,
		from: int64(2*TEST_ID_PARTITION + 2),
		to: int64(2*TEST_ID_PARTITION + 1),
	})

	time.Sleep(50 * time.Millisecond)

	checkCounters("Unknown", map[int64]int32 {
		2*TEST_ID_PARTITION + 1: 4,
		2*TEST_ID_PARTITION + 2: 4,
		3*TEST_ID_PARTITION + 1: 2,
		3*TEST_ID_PARTITION + 2: 2,
		3*TEST_ID_PARTITION + 3: 2,
		3*TEST_ID_PARTITION + 4: 2,
		3*TEST_ID_PARTITION + 5: 2,
		3*TEST_ID_PARTITION + 6: 2,
	})

	router.PushMessage(&Message{
		payload: "msg1",
		typ: MESSAGE_FOLLOW,
		from: int64(1),
		to: int64(2),
	})

	router.PushMessage(&Message{
		payload: "msg1",
		typ: MESSAGE_PRIVATE_MSG,
		from: int64(1),
		to: int64(2),
	})

	router.PushMessage(&Message{
		payload: "msg1",
		typ: MESSAGE_UNFOLLOW,
		from: int64(3),
		to: int64(4),
	})

	router.PushMessage(&Message{
		payload: "msg1",
		typ: MESSAGE_STATUS_UPDATE,
		from: int64(3),
	})

	time.Sleep(50 * time.Millisecond)

	checkCounters("NewId", map[int64]int32 {
		2*TEST_ID_PARTITION + 1: 4,
		2*TEST_ID_PARTITION + 2: 4,
		3*TEST_ID_PARTITION + 1: 2,
		3*TEST_ID_PARTITION + 2: 2,
		3*TEST_ID_PARTITION + 3: 2,
		3*TEST_ID_PARTITION + 4: 2,
		3*TEST_ID_PARTITION + 5: 2,
		3*TEST_ID_PARTITION + 6: 2,
	})
}

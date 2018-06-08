package main

import (
	"testing"
	"reflect"
	"os"
)

func TestNewMessage(t *testing.T) {
	testSuites := []*struct {
		payload     string
		expected    *Message
		expectedErr bool
	}{
		{payload: "666|F|60|50", expected: &Message{
			payload:    "666|F|60|50",
			sequenceId: 666,
			typ:        MESSAGE_FOLLOW,
			from:       60,
			to:         50,
		}},
		{payload: "1|U|12|9", expected: &Message{
			payload:    "1|U|12|9",
			sequenceId: 1,
			typ:        MESSAGE_UNFOLLOW,
			from:       12,
			to:         9,
		}},
		{payload: "542532|B", expected: &Message{
			payload:    "542532|B",
			sequenceId: 542532,
			typ:        MESSAGE_BROADCAST,
		}},
		{payload: "43|P|32|56", expected: &Message{
			payload:    "43|P|32|56",
			sequenceId: 43,
			typ:        MESSAGE_PRIVATE_MSG,
			from:       32,
			to:         56,
		}},
		{payload: "634|S|32", expected: &Message{
			payload:    "634|S|32",
			sequenceId: 634,
			typ:        MESSAGE_STATUS_UPDATE,
			from:       32,
		}},
		{payload: "43|P|32", expectedErr: true},
		{payload: "43|P|32|45|56", expectedErr: true},
		{payload: "43|P|abc|45", expectedErr: true},
		{payload: "43|P|32|abc", expectedErr: true},
		{payload: "634|S|abc", expectedErr: true},
		{payload: "634|S", expectedErr: true},
		{payload: "634|S|34|56", expectedErr: true},
		{payload: "634|J|34|56", expectedErr: true},
	}

	for _, test := range testSuites {
		exist := NewMessage(test.payload)

		if test.expectedErr {
			if exist.HasError() == nil {
				t.Error("failed to catch an error for payload '", test.payload, "'")
			}
		} else if err := exist.HasError(); err != nil {
			t.Error("failed to parse payload '", test.payload, "' with error ", err)
		} else {
			// mock the field value related time
			test.expected.created = exist.created

			if !reflect.DeepEqual(exist, test.expected) {
				t.Error("failed to parse payload '", test.payload, "'. Got ", exist, ", but expected is ", test.expected)
			}
		}
	}
}

func TestNewMessage_IsValid(t *testing.T) {
	testSuites := []*struct {
		message  *Message
		expected bool
	}{
		{
			message: &Message{
				payload:    "1",
				sequenceId: 9889,
				typ:        MESSAGE_PRIVATE_MSG,
			},
			expected: true,
		},
		{
			message: &Message{
				payload:    "2",
				sequenceId: 0,
				typ:        MESSAGE_PRIVATE_MSG,
			},
			expected: false,
		},
		{
			message: &Message{
				payload:    "3",
				sequenceId: 9889,
				typ:        MESSAGE_UNKNOWN,
			},
			expected: false,
		},
		{
			message: &Message{
				payload:    "4",
				sequenceId: 9889,
				typ:        MESSAGE_PRIVATE_MSG,
				err:        os.ErrInvalid,
			},
			expected: false,
		},
	}

	for _, test := range testSuites {
		exist := test.message.IsValid()
		if exist != test.expected {
			t.Error("failed to check valid of message '", test.message, "'. Got ", exist, ", but expected is ", test.expected)
		}
	}
}

func TestNewMessage_FromTo(t *testing.T) {
	payload := "666|F|60|50"
	message := NewMessage(payload)

	expectedFrom, expectedTo := int64(60), int64(50)
	from, to := message.FromTo()

	if expectedFrom != from {
		t.Error("failed to get from. Got ", expectedFrom, ", but expected is ", from)
	}

	if expectedTo != to {
		t.Error("failed to get to. Got ", expectedFrom, ", but expected is ", from)
	}
}

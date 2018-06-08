package main

import (
	"strings"
	"errors"
	"strconv"
	"time"
)

type MessageType int

const (
	MESSAGE_BROADCAST     MessageType = iota + 1
	MESSAGE_FOLLOW
	MESSAGE_STATUS_UPDATE
	MESSAGE_UNFOLLOW
	MESSAGE_PRIVATE_MSG

	MESSAGE_UNKNOWN
)

func (o MessageType) String() string {
	switch o {
	case MESSAGE_BROADCAST:
		return "Broadcast"
	case MESSAGE_FOLLOW:
		return "Follow"
	case MESSAGE_STATUS_UPDATE:
		return "StatusUpdate"
	case MESSAGE_UNFOLLOW:
		return "Unfollow"
	case MESSAGE_PRIVATE_MSG:
		return "Private"
	default:
		return "Unknown"
	}
}

func FromString(typ string) MessageType {
	switch typ {
	case "B":
		return MESSAGE_BROADCAST
	case "F":
		return MESSAGE_FOLLOW
	case "S":
		return MESSAGE_STATUS_UPDATE
	case "U":
		return MESSAGE_UNFOLLOW
	case "P":
		return MESSAGE_PRIVATE_MSG
	default:
		return MESSAGE_UNKNOWN
	}
}

var (
	invalidMessageErr = errors.New("invalid message format")
)

type Message struct {
	payload    string
	sequenceId int64
	typ        MessageType
	from       int64
	to         int64
	err        error
	created    time.Time
}

func NewMessage(payload string) *Message {
	parts := strings.Split(payload, "|")

	return (&Message{
		payload: payload,
		typ:     FromString(parts[1]),
		created: time.Now(),
	}).Parse(parts)
}

func (o *Message) String() string {
	return o.payload
}

func (o *Message) Parse(parts []string) *Message {
	var err error

	o.sequenceId, _ = strconv.ParseInt(parts[0], 10, 64)

	switch o.typ {
	case MESSAGE_FOLLOW, MESSAGE_PRIVATE_MSG, MESSAGE_UNFOLLOW:
		if len(parts) == 4 {
			o.from, err = strconv.ParseInt(parts[2], 10, 64)
			if err != nil {
				o.err = err
			}

			o.to, err = strconv.ParseInt(parts[3], 10, 64)
			if err != nil {
				o.err = err
			}
		} else {
			o.err = invalidMessageErr
		}
	case MESSAGE_STATUS_UPDATE:
		if len(parts) == 3 {
			o.from, o.err = strconv.ParseInt(parts[2], 10, 64)
		} else {
			o.err = invalidMessageErr
		}
	case MESSAGE_UNKNOWN:
		o.err = invalidMessageErr
	}

	return o
}

func (o *Message) FromTo() (int64, int64) {
	return o.from, o.to
}

func (o *Message) HasError() error {
	return o.err
}

func (o *Message) IsValid() bool {
	return o.sequenceId > 0 && o.typ != MESSAGE_UNKNOWN && o.err == nil
}

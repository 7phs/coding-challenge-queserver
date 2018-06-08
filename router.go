package main

import (
	"sync"
	"sync/atomic"
	"strings"
	"strconv"
	"github.com/7phs/coding-challenge-queserver/logger"
)

type MessageQueue interface {
	PushMessage(*Message)
}

type EventRouter interface {
	RegisterClient(int64) <-chan *Message
	UnregisterClient(int64)
}

type UserInfo struct {
	userId        int64
	ch            chan *Message
	subscriptions sync.Map

	registered int32
}

func (o *UserInfo) Follow(userId int64) {
	o.subscriptions.Store(userId, true)
}

func (o *UserInfo) Unfollow(userId int64) {
	o.subscriptions.Delete(userId)
}

func (o *UserInfo) Range(f func(key, value interface{}) bool) {
	o.subscriptions.Range(f)
}

func (o *UserInfo) SetRegister(registered bool) {
	v := int32(0)
	if registered {
		v = 1
	}

	atomic.StoreInt32(&o.registered, v)
}

func (o *UserInfo) IsRegistered() bool {
	return atomic.LoadInt32(&o.registered) != 0
}

func (o *UserInfo) dumpSubscriptions() string {
	result := []string{}

	o.subscriptions.Range(func(key, _ interface{}) bool {
		result = append(result, strconv.FormatInt(key.(int64), 10))

		return true
	})

	return strings.Join(result, ", ")
}

type Router struct {
	clients sync.Map
}

func NewRouter() *Router {
	return &Router{}
}

func (o *Router) PushMessage(msg *Message) {
	logger.Debug("[ROUTER]: push message ", msg.payload)

	switch msg.typ {
	case MESSAGE_BROADCAST:
		o.handleBroadcast(msg)
	case MESSAGE_FOLLOW:
		o.handleFollow(msg)
	case MESSAGE_STATUS_UPDATE:
		o.handleStatusUpdate(msg)
	case MESSAGE_UNFOLLOW:
		o.handleUnfollow(msg)
	case MESSAGE_PRIVATE_MSG:
		o.handlePrivateMsg(msg)
	default:
		logger.Warning("[ROUTER]: processed a message with unknown type: ", msg)
	}
}

func (o *Router) handleFollow(msg *Message) {
	userInfo := o.getOrAddUserInfo(msg.to)
	if userInfo != nil {
		userInfo.Follow(msg.from)
		o.sendMessage(userInfo, msg)
	}
}

func (o *Router) handleUnfollow(msg *Message) {
	userInfo := o.getOrAddUserInfo(msg.to)
	if userInfo != nil {
		userInfo.Unfollow(msg.from)
	}
}

func (o *Router) handleBroadcast(msg *Message) {
	o.sendBroadcast(msg)
}

func (o *Router) handlePrivateMsg(msg *Message) {
	userInfo := o.getOrAddUserInfo(msg.to)
	if userInfo != nil {
		o.sendMessage(userInfo, msg)
	}
}

func (o *Router) handleStatusUpdate(msg *Message) {
	userInfo := o.getOrAddUserInfo(msg.from)
	if userInfo != nil {
		userInfo.Range(func(key, _ interface{}) bool {
			// go func(userId int64) {
			userInfo := o.getOrAddUserInfo(key.(int64))
			if userInfo != nil {
				o.sendMessage(userInfo, msg)
			}

			// }(key.(int64))

			return true
		})
	}
}

func (o *Router) sendBroadcast(msg *Message) {
	// visit each partition
	o.clients.Range(func(_, userInfo interface{}) bool {
		// send message to user
		o.sendMessage(userInfo.(*UserInfo), msg)

		return true
	})
}

func (o *Router) sendMessage(userInfo *UserInfo, msg *Message) {
	if !userInfo.IsRegistered() {
		return
	}

	logger.Debug("[ROUTER]: send message ", msg.payload, " -> ", userInfo.userId)

	userInfo.ch <- msg
}

func (o *Router) RegisterClient(userId int64) <-chan *Message {
	logger.Info("[ROUTER]: register client, user id #", userId)

	userInfo := o.getOrAddUserInfo(userId)
	if userInfo == nil {
		return nil
	}

	userInfo.SetRegister(true)
	return userInfo.ch
}

func (o *Router) UnregisterClient(userId int64) {
	logger.Info("[ROUTER]: unregister client, user id #", userId)

	userInfo := o.getOrAddUserInfo(userId)
	if userInfo == nil {
		return
	}

	userInfo.SetRegister(false)
}

func (o *Router) getOrAddUserInfo(userId int64) *UserInfo {
	userInfo, _ := o.clients.LoadOrStore(userId, &UserInfo{
		userId: userId,
		ch:     make(chan *Message),
	})

	return userInfo.(*UserInfo)
}

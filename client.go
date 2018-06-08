package main

import (
	"net"
	"bufio"
	"strconv"
	"github.com/7phs/coding-challenge-queserver/logger"
)

type Client struct {
	conn       net.Conn
	router     EventRouter
	statistics *Statistics
	err        error

	userId   int64
	ch       <-chan *Message
	shutdown chan struct{}
}

func NewClient(conn net.Conn, router EventRouter, statistics *Statistics, shutdown chan struct{}) *Client {
	return (&Client{
		conn:       conn,
		router:     router,
		statistics: statistics,
		shutdown:   shutdown,
	}).
		Handshake().
		Register()
}

func (o *Client) HasError() error {
	return o.err
}

func (o *Client) Handshake() (c *Client) {
	c = o

	if o.HasError() != nil {
		logger.Debug("[CLIENT]: handshaking, skip for error")
		return
	}

	logger.Debug("[CLIENT]: handshaking, start")

	line, _, err := bufio.NewReader(o.conn).ReadLine()
	if err != nil {
		logger.Warning("[CLIENT]: handshaking, error while read line with id: ", err)

		o.err = err
		return
	}

	o.userId, err = strconv.ParseInt(string(line), 10, 64)
	if err != nil {
		logger.Warning("[CLIENT]: handshaking, error while parse user id: ", err)

		o.err = err
		return
	}

	logger.Debug("[CLIENT]: handshaking, got user id #", o.userId)

	return
}

func (o *Client) Register() (c *Client) {
	c = o

	if o.HasError() != nil {
		logger.Debug("[CLIENT]: register #", o.userId, ", skip for error")

		return
	}

	logger.Debug("[CLIENT]: register #", o.userId)

	o.ch = o.router.RegisterClient(o.userId)

	return
}

func (o *Client) Run() {
	if o.HasError() != nil {
		logger.Debug("[CLIENT]: #", o.userId, " run, skip for error")
		return
	}

	go func() {
		work := true

		logger.Debug("[CLIENT]: #", o.userId, ", start working goroutin")

		for work {
			select {
			case msg := <-o.ch:
				o.statistics.Add(MESSAGE_SEND, msg.typ)

				n, err := o.conn.Write([]byte(msg.payload + "\r\n"))
				if err != nil {
					logger.Warning("[CLIENT]: #", o.userId, ", got error while write data: ", err)

					o.Unregister()
					work = false
				}

				logger.Debug("[CLIENT]: write '", msg.payload, "':", n)

			case <-o.shutdown:
				work = false
			}
		}

		logger.Debug("[CLIENT]: #", o.userId, ", stop working goroutin and close connection")

		o.conn.Close()
	}()
}

func (o *Client) Unregister() {
	logger.Debug("[CLIENT]: unregister #", o.userId)

	o.router.UnregisterClient(o.userId)
}

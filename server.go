package main

import (
	"net"
	"sync"
	"github.com/7phs/coding-challenge-queserver/logger"
)

type Server struct {
	addr string

	listener   net.Listener
	router     EventRouter
	statistics *Statistics

	shutdown chan struct{}

	wait sync.WaitGroup
}

func NewServer(config *Config, router EventRouter, statistics *Statistics) (*Server, error) {
	return (&Server{
		addr:       config.Client(),
		router:     router,
		statistics: statistics,
		shutdown:   make(chan struct{}),
	}).Listen()
}

func (o *Server) Listen() (s *Server, err error) {
	s = o

	logger.Info("[SERVER]: listen ", o.addr)

	o.listener, err = net.Listen("tcp", o.addr)

	return
}

func (o *Server) Run() {
	o.wait.Add(1)

	go func() {
		logger.Info("[SERVER]: start working goroutin")

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
					logger.Error("[SERVER]: error while accept connection messages: ", i)

				case net.Conn:
					logger.Debug("[SERVER]: accept a connection")
					go NewClient(i, o.router, o.statistics, o.shutdown).Run()
				}

			case <-o.shutdown:
				logger.Info("[SERVER]: shutdown working goroutin")
				o.wait.Done()
				return
			}
		}
	}()
}

func (o *Server) Shutdown() {
	logger.Info("[SERVER]: shutdown")

	close(o.shutdown)

	o.wait.Wait()

	if o.listener != nil {
		o.listener.Close()
	}
}

package main

import (
	"sync"
	"os/signal"
	"os"
	"github.com/7phs/coding-challenge-queserver/logger"
)

type Shutdowned interface {
	Shutdown()
}

type ShutdownQueue struct {
	shutdownList []Shutdowned
}

func (o *ShutdownQueue) Add(item Shutdowned) {
	// in reverse order
	o.shutdownList = append([]Shutdowned{item}, o.shutdownList...)
}

func (o *ShutdownQueue) Shutdown() {
	for _, item := range o.shutdownList {
		item.Shutdown()
	}
}

func main() {
	shutdownQueue := &ShutdownQueue{}

	logger.SetFlags(logger.ALL)

	logger.Info("[SOUNDSERVER]: starting")
	logger.Info("[SOUNDSERVER]: read config")
	config, err := ParseConfig()
	if err != nil {
		logger.Error("[SOUNDSERVER]: failed to get configuration parameters: ", err)
		return
	}

	logger.SetFlags(config.LogLevel())

	logger.Info("[SOUNDSERVER]: config parameters - ",
		"EVENT_SOURCE=", config.EventSource(),
		"; CLIENT=", config.Client(),
		"; QUEUE_LIMIT=", config.QueueLimit(),
		"; QUEUE_TTL=", config.QueueTTL(),
		"; LOG_LEVEL=", logger.LevelToString(config.LogLevel()))

	logger.Info("[SOUNDSERVER]: create a statistics")
	statistics := NewStatistics()
	shutdownQueue.Add(statistics)

	logger.Info("[SOUNDSERVER]: create a router")
	router := NewRouter()

	logger.Info("[SOUNDSERVER]: create a queue")
	queue := NewQueue(config, router)
	shutdownQueue.Add(queue)

	logger.Info("[SOUNDSERVER]: create an event source")
	eventSource, err := NewEventSource(config, queue, statistics)
	if err != nil {
		logger.Error("[SOUNDSERVER]: failed to init an event source server: ", err)

		shutdownQueue.Shutdown()
		return
	}
	shutdownQueue.Add(eventSource)

	logger.Info("[SOUNDSERVER]: create a server for a client")
	server, err := NewServer(config, router, statistics)
	if err != nil {
		logger.Error("[SOUNDSERVER]: failed to init a client server: ", err)

		shutdownQueue.Shutdown()
		return
	}
	shutdownQueue.Add(server)

	var wait sync.WaitGroup

	// main work
	wait.Add(1)
	go func() {
		// start all services
		queue.Run()
		eventSource.Run()
		server.Run()
		// wait for Ctrl+C
		interrupt := make(chan os.Signal, 2)
		signal.Notify(interrupt, os.Interrupt) // CTRL-C
		<-interrupt
		// shut down all services
		shutdownQueue.Shutdown()
		// dump final statistics
		logger.Info("[SOUNDSERVER]:", statistics.DumpState())

		wait.Done()
	}()

	wait.Wait()

	logger.Info("[SOUNDSERVER]: finished")
}

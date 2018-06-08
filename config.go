package main

import (
	"os"
	"errors"
	"time"
	"github.com/7phs/coding-challenge-queserver/logger"
)

const (
	DEFAULT_EVENT_SOURCE   = ":9090"
	DEFAULT_CLIENT         = ":9099"
	DEFAULT_QUEUE_LIMIT    = 1000
	DEFAULT_QUEUE_TTL      = 500 // milliseconds
	DEFAULT_LOG_LEVEL      = logger.INFO

	CONFIG_EVENT_SOURCE = "EVENT_SOURCE"
	CONFIG_CLIENT       = "CLIENT"
	CONFIG_QUEUE_LIMIT  = "QUEUE_LIMIT"
	CONFIG_QUEUE_TTL    = "QUEUE_TTL"
	CONFIG_LOG_LEVEL    = "LOG_LEVEL"
)

type Config struct {
	eventSource   string
	client        string
	queueLimit    int64
	queueTTL      int64
	logLevel      int
}

func (o *Config) EventSource() string {
	return o.eventSource
}

func (o *Config) Client() string {
	return o.client
}

func (o *Config) QueueLimit() int64 {
	return o.queueLimit
}

func (o *Config) QueueTTL() time.Duration {
	return time.Duration(o.queueTTL) * time.Millisecond
}

func (o *Config) LogLevel() int {
	return o.logLevel
}

func ParseConfig() (*Config, error) {
	eventSource, err := ParseAddress(os.Getenv(CONFIG_EVENT_SOURCE), DEFAULT_EVENT_SOURCE)
	if err != nil {
		return nil, errors.New("failed to parse an event source config parameter: " + err.Error())
	}

	port, err := ParseAddress(os.Getenv(CONFIG_CLIENT), DEFAULT_CLIENT)
	if err != nil {
		return nil, errors.New("failed to parse a client config parameter: " + err.Error())
	}

	queueLimit := ParseInt64(os.Getenv(CONFIG_QUEUE_LIMIT), DEFAULT_QUEUE_LIMIT)
	queueTTL := ParseInt64(os.Getenv(CONFIG_QUEUE_TTL), DEFAULT_QUEUE_TTL)
	logLevel := logger.ParseLevel(os.Getenv(CONFIG_LOG_LEVEL), DEFAULT_LOG_LEVEL)

	return &Config{
		eventSource:   eventSource,
		client:        port,
		queueLimit:    queueLimit,
		queueTTL:      queueTTL,
		logLevel:      logLevel,
	}, nil
}

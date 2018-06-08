package main

import (
	"os"
	"testing"
	"time"
	"github.com/7phs/coding-challenge-queserver/logger"
)

func SetUpParseCofigParameter() func() {
	prev := map[string]string{}

	for _, name := range []string{
		CONFIG_CLIENT, CONFIG_EVENT_SOURCE,
		CONFIG_QUEUE_LIMIT, CONFIG_QUEUE_TTL, CONFIG_LOG_LEVEL} {
		prev[name] = os.Getenv(name)
	}

	return func() {
		for name, value := range prev {
			os.Setenv(name, value)
		}
	}
}

func TestParseConfig(t *testing.T) {
	defer SetUpParseCofigParameter()()

	testSuites := []*struct {
		client             string
		expectedClient     string
		eventSource        string
		expectedEventSrc   string
		queueLimit         string
		expectedQueueLimit int64
		queueTTL           string
		expectedQueueTTL   time.Duration
		logLevel           string
		expectedLogLevel   int
	}{
		{
			expectedClient:     DEFAULT_CLIENT, expectedEventSrc: DEFAULT_EVENT_SOURCE,
			expectedQueueLimit: DEFAULT_QUEUE_LIMIT, expectedQueueTTL: DEFAULT_QUEUE_TTL * time.Millisecond, expectedLogLevel: logger.CalcLevel(DEFAULT_LOG_LEVEL),
		},
		{
			client:      ":sdfsdf", expectedClient: DEFAULT_CLIENT,
			eventSource: ":dasdas", expectedEventSrc: DEFAULT_EVENT_SOURCE,
			queueLimit:  "dkjadslj", expectedQueueLimit: DEFAULT_QUEUE_LIMIT,
			queueTTL:    "asdasd", expectedQueueTTL: DEFAULT_QUEUE_TTL * time.Millisecond,
			logLevel:    "unksljkdf", expectedLogLevel: logger.CalcLevel(DEFAULT_LOG_LEVEL),
		},
		{
			client:      ":9090", expectedClient: ":9090",
			eventSource: ":7777", expectedEventSrc: ":7777",
			queueLimit:  "8888", expectedQueueLimit: 8888,
			queueTTL:    "700", expectedQueueTTL: 700 * time.Millisecond,
			logLevel:    "error", expectedLogLevel: logger.ERROR,
		},
	}

	for i, test := range testSuites {
		os.Setenv(CONFIG_CLIENT, test.client)
		os.Setenv(CONFIG_EVENT_SOURCE, test.eventSource)
		os.Setenv(CONFIG_QUEUE_LIMIT, test.queueLimit)
		os.Setenv(CONFIG_QUEUE_TTL, test.queueTTL)
		os.Setenv(CONFIG_LOG_LEVEL, test.logLevel)

		params, err := ParseConfig()

		if err != nil {
			t.Error("failed to parse config")
		}

		if exist := params.EventSource(); exist != test.expectedEventSrc {
			t.Error(i, ": failed to parse environment param for event source. Got '", exist, "', but expected is '", test.expectedEventSrc, "'")
		}

		if exist := params.Client(); exist != test.expectedClient {
			t.Error(i, ": failed to parse environment param for client. Got '", exist, "', but expected is '", test.expectedClient, "'")
		}

		if exist := params.QueueLimit(); exist != test.expectedQueueLimit {
			t.Error(i, ": failed to parse environment param for queue limit. Got '", exist, "', but expected is '", test.expectedQueueLimit, "'")
		}

		if exist := params.QueueTTL(); exist != test.expectedQueueTTL {
			t.Error(i, ": failed to parse environment param for queue TTL. Got '", exist, "', but expected is '", test.expectedQueueTTL, "'")
		}

		if exist := params.LogLevel(); exist != test.expectedLogLevel {
			t.Error(i, ": failed to parse environment param for log level. Got '", exist, "', but expected is '", test.expectedLogLevel, "'")
		}
	}
}

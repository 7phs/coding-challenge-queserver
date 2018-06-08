package logger

import (
	"testing"
)

func TestParseLogLevel(t *testing.T) {
	defaultIn := WARNING

	testSuites := []*struct {
		in       string
		expected int
	}{
		{in: "Debug", expected: DEBUG | INFO | WARNING | ERROR,},
		{in: "InFo", expected: INFO | WARNING | ERROR,},
		{in: "warninG", expected: WARNING | ERROR,},
		{in: "error", expected: ERROR,},
		{expected: WARNING | ERROR,},
		{in: "Unknown", expected: WARNING | ERROR,},
	}

	for _, test := range testSuites {
		exist := ParseLevel(test.in, defaultIn)
		if exist!=test.expected {
			t.Error("failed to parse log level '", test.in, "'. Got ", exist, ", but ecpected is ", test.expected)
		}
	}
}

func TestLevelToString(t *testing.T) {
	testSuites := []*struct {
		in       int
		expected string
	}{
		{in: DEBUG | INFO | WARNING | ERROR, expected: "Debug",},
		{in: INFO | WARNING | ERROR, expected: "Info",},
		{in: WARNING | ERROR, expected: "Warning",},
		{in: ERROR, expected: "Error",},
	}

	for _, test := range testSuites {
		exist := LevelToString(test.in)

		if exist!=test.expected {
			t.Error("failed to parse log level ", test.in, " to string. Got '", exist, "', but ecpected is '", test.expected, "'")
		}
	}
}
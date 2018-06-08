package logger

import (
	"log"
	"strings"
)

const (
	DEBUG   = 0x1
	INFO    = 0x2
	WARNING = 0x4
	ERROR   = 0x8

	ALL = DEBUG | INFO | WARNING | ERROR
)

var (
	logFlags = ALL
)

// Parse a string with log level name
func ParseLevel(level string, edgeLevel int) int {
	switch strings.ToLower(level) {
	case "debug":
		edgeLevel = DEBUG
	case "info":
		edgeLevel = INFO
	case "warning":
		edgeLevel = WARNING
	case "error":
		edgeLevel = ERROR
	}

	return CalcLevel(edgeLevel)
}

func CalcLevel(edgeLevel int) int {
	result := 0
	start := false

	for _, v := range []int{DEBUG, INFO, WARNING, ERROR} {
		if v == edgeLevel {
			start = true
		}

		if start {
			result |= v
		}
	}

	return result
}

func LevelToString(level int) string {
	minimum := func() int {
		for _, v := range []int{DEBUG, INFO, WARNING, ERROR} {
			if level&v!=0 {
				return v
			}
		}

		return 0
	}()

	switch minimum {
	case DEBUG:
		return "Debug"
	case INFO:
		return "Info"
	case WARNING:
		return "Warning"
	case ERROR:
		return "Error"
	}

	return "Unknown"
}

// set logging flags, but only before start a server to prevent sync problems
func SetFlags(flags int) {
	logFlags = flags
}

func Debug(msgs ... interface{}) {
	if logFlags&DEBUG==0 {
		return
	}

	log.Println(msgs...)
}

func Info(msgs ... interface{}) {
	if logFlags&INFO==0 {
		return
	}

	log.Println(msgs...)
}

func Warning(msgs ... interface{}) {
	if logFlags&WARNING==0 {
		return
	}

	log.Println(msgs...)
}

func Error(msgs ... interface{}) {
	if logFlags&ERROR==0 {
		return
	}

	log.Println(msgs...)
}

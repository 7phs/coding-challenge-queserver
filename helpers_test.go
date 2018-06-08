package main

import (
	"strings"
	"testing"
)

func TestParseAddress(t *testing.T) {
	testSuites := []*struct {
		addr        string
		defaultAddr string
		expected    string
		expectedErr bool
	}{
		{expectedErr: true,},
		{addr: "3456", expectedErr: true,},
		{addr: ":3456", expected: ":3456",},
		{defaultAddr: ":3456", expected: ":3456",},
		{addr: "10.0.0.1:3456", expected: "10.0.0.1:3456",},
		{defaultAddr: "10.0.0.1:3456", expected: "10.0.0.1:3456",},
		{addr: "10.0.0.1", expectedErr: true,},
		{addr: "10.0.0.1", defaultAddr: ":9090", expected: "10.0.0.1:9090",},
		{addr: ":9090", defaultAddr: "10.0.0.1", expected: "10.0.0.1:9090",},
		{addr: ":unknown", expectedErr: true,},
		{defaultAddr: ":unknown", expectedErr: true,},
	}

	for _, test := range testSuites {
		exist, err := ParseAddress(test.addr, test.defaultAddr)
		if test.expectedErr && err == nil {
			t.Error("failed to catch an error for: '", test.addr, "' and '", test.defaultAddr, "'")
			continue
		}

		if err != nil {
			if !test.expectedErr {
				t.Error("failed to parse '", test.addr, "' and '", test.defaultAddr, "' with error: ", err)
			}
		} else if strings.Compare(exist, test.expected) != 0 {
			t.Error("failed to parse '", test.addr, "' and '", test.defaultAddr, "'. Got '", exist, "', but expected is '", test.expected, "")
		}
	}
}

func TestParseInt64(t *testing.T) {
	defaultIn := int64(100)

	testSuites := []*struct {
		in       string
		expected int64
	}{
		{in: "12938721897", expected: 12938721897},
		{in: "asdsa721897", expected: defaultIn},
	}

	for _, test := range testSuites {
		exist := ParseInt64(test.in, defaultIn)
		if exist != test.expected {
			t.Error("failed to parse '", test.in, "' to int64. Got ", exist, ", but expected is ", test.expected)
		}
	}
}

func TestMaxInt64(t *testing.T) {
	testSuites := []*struct {
		in1      int64
		in2      int64
		expected int64
	}{
		{in1: 300, in2: 600, expected: 600},
		{in1: 900, in2: 100, expected: 900},
		{in1: 1900, in2: 1900, expected: 1900},
	}

	for _, test := range testSuites {
		exist := MaxInt64(test.in1, test.in2)
		if exist != test.expected {
			t.Error("failed to get maximum from ", test.in1, " and ", test.in2, ". Got ", exist, ", but expected is ", test.expected)
		}
	}
}

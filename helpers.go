package main

import (
	"strings"
	"net"
	"errors"
	"strconv"
)

func ParseAddress(addr, defaultAddr string) (string, error) {
	if addr == "" && defaultAddr == "" {
		return "", errors.New("empty an address and a default address")
	}

	var (
		host        string
		port        string
		defaultHost string
		defaultPort string
		defaultErr  error
	)

	if addr != "" {
		parsedAddr := addr
		if strings.Index(parsedAddr,":")<0 {
			parsedAddr += ":"
		}

		if host, port, _ = net.SplitHostPort(parsedAddr); host != "" && port != "" {
			return host + ":" + port, nil
		}

		if _, err := strconv.Atoi(strings.TrimPrefix(port, ":")); err!=nil {
			port = ""
		}
	}

	if host == "" || port == "" && defaultAddr != "" {
		parsedAddr := defaultAddr
		if strings.Index(parsedAddr,":")<0 {
			parsedAddr += ":"
		}

		if defaultHost, defaultPort, defaultErr = net.SplitHostPort(parsedAddr); defaultErr == nil {
			if _, err := strconv.Atoi(strings.TrimPrefix(defaultPort, ":")); err!=nil {
				defaultPort = ""
			}

			if host == "" {
				host = defaultHost
			}

			if port == "" {
				port = defaultPort
			}
		}
	}

	if port == "" {
		return "", errors.New("failed to parse address, client has to have a value, but it is empty")
	}

	return host + ":" + port, nil
}

func ParseInt64(v string, defaultV int64) int64 {
	result, err := strconv.ParseInt(v, 10, 64)
	if err != nil || v == "" {
		result = defaultV
	}

	return result
}

func MaxInt64(v, v1 int64) int64 {
	if v>=v1 {
		return v
	}

	return v1
}
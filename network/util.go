package network

import (
	"errors"
	"strconv"
	"strings"
)

// SplitHost dissects a string of fotm "host:port" string into components.
func SplitHost(host string) (addr string, port int, err error) {
	idx := strings.Index(host, ":")
	if idx == -1 {
		err = errors.New("Invalid host definition")
		return
	}
	addr = host[:idx]
	port, err = strconv.Atoi(host[idx+1:])
	if err != nil || port < 1 || port > 65535 {
		err = errors.New("Invalid host definition")
	}
	return
}

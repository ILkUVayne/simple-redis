package utils

import (
	"strconv"
	"strings"
)

// StrToHost string type host to []byte host
// e.g. "127.0.0.1" -> []byte{127,0,0,1}
func StrToHost(host string) []byte {
	hosts := strings.Split(host, ".")
	if len(hosts) != 4 {
		Error("str2host error: host is bad, host == ", host)
	}
	h := make([]byte, 4)
	for idx, v := range hosts {
		i, err := strconv.Atoi(v)
		if err != nil {
			Error("str2host error: host is bad: ", host)
		}
		h[idx] = uint8(i)
	}
	return h
}

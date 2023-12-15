package utils

import (
	"strconv"
	"strings"
)

// StrToHost string type host to []byte host
// e.g. "127.0.0.1" -> []byte{127,0,0,1}
func StrToHost(host string) [4]byte {
	hosts := strings.Split(host, ".")
	if len(hosts) != 4 {
		Error("str2host error: host is bad, host == ", host)
	}
	var h [4]byte
	for idx, v := range hosts {
		i, err := strconv.Atoi(v)
		if err != nil {
			Error("str2host error: host is bad: ", host)
		}
		h[idx] = uint8(i)
	}
	return h
}

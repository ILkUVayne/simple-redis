// Package src 自定义错误库
package src

import (
	"strconv"
)

type srError int

func (e srError) Error() string {
	if 0 <= int(e) && int(e) < len(srErrors) {
		s := srErrors[e]
		if s != "" {
			return s
		}
	}
	return "errno " + strconv.Itoa(int(e))
}

var srErrors = [...]string{
	0x01: "connect disconnected",
}

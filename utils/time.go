package utils

import "time"

func GetMsTime() int64 {
	return time.Now().UnixNano() / 1e6
}

package utils

import "time"

func GetMsTime() int64 {
	return time.Now().UnixNano() / 1e6
}

func GetMsTimeByTime(t *time.Time) int64 {
	return t.UnixNano() / 1e6
}

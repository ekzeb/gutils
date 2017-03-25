package util

import (
	"time"
	"strconv"
)

func NowLocalAsMs() int64 {
	return TimeToMs(time.Now())
}

func NowUTCAsMs() int64 {
	return TimeToMs(time.Now().UTC())
}

func TimeToMs(t time.Time) int64 {
	return t.UnixNano() / int64(time.Millisecond)
}

func MsToTime(ms string) (time.Time, error) {

	msInt, err := strconv.ParseInt(ms, 10, 64)

	if err != nil {
		return time.Time{}, err
	}

	return time.Unix(0, msInt*int64(time.Millisecond)), nil
}
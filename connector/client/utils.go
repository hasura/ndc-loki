package client

import (
	"errors"
	"strconv"
	"time"
)

var (
	errQueryRequired     = errors.New("query parameter must not be empty")
	errLabelNameRequired = errors.New("label name must not be empty")
)

// FormatUnixTimestamp formats the time instance to unix timestamp
func FormatUnixTimestamp(ts time.Time) string {
	return strconv.FormatInt(ts.UnixNano(), 10)
}

func noopCancel() {}

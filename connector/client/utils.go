package client

import (
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"time"

	"github.com/hasura/ndc-sdk-go/scalar"
)

var (
	errQueryRequired     = errors.New("query parameter must not be empty")
	errLabelNameRequired = errors.New("label name must not be empty")
)

// FormatUnixNanoTimestamp formats the time instance to unix timestamp in nanoseconds
func FormatUnixNanoTimestamp(ts time.Time) string {
	return strconv.FormatInt(ts.UnixNano(), 10)
}

// validate and set time range query parameters
func applyQueryTimeRange(q *url.Values, start *time.Time, end *time.Time, since *scalar.Duration, maxTimeRange time.Duration) error {
	if start == nil && end == nil && since == nil {
		if maxTimeRange > time.Minute {
			q.Set("since", (maxTimeRange - time.Minute).String())
		}

		return nil
	}

	endTime := time.Now()
	if end != nil {
		endTime = *end
		q.Set("end", FormatUnixNanoTimestamp(endTime))
	}

	if since != nil {
		if maxTimeRange > 0 && since.Duration > maxTimeRange {
			return fmt.Errorf("the query time range exceeds the limit (query length: %s, limit: %s)", since.String(), maxTimeRange.String())
		}

		q.Set("since", since.String())
	}

	if start != nil {
		offset := endTime.Sub(*start)
		if maxTimeRange > 0 && offset >= maxTimeRange {
			return fmt.Errorf("the query time range exceeds the limit (query length: %s, limit: %s)", offset.String(), maxTimeRange.String())
		}
		q.Set("start", FormatUnixNanoTimestamp(*start))
	}

	return nil
}

func noopCancel() {}

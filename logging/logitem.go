package logging

import (
	"net/http"
	"net/url"
	"time"
)

type LogContent struct {
	Header http.Header
	Body   []byte
}

// LogItem records information about one HTTP round-trip.
type LogItem struct {
	Method     string
	URL        *url.URL
	StatusCode int
	Request    LogContent
	Response   LogContent
	Err        error
	Start      time.Time
	Duration   time.Duration
	Level      Level
}

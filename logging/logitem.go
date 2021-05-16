package logging

import (
	"github.com/rickb777/httpclient/body"
	"net/http"
	"net/url"
	"time"
)

type LogContent struct {
	Header http.Header
	Body   *body.Body
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

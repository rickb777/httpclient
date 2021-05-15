package loggingtransport

import (
	"github.com/rickb777/httpclient/internal"
	"github.com/rickb777/httpclient/logging"
	"net/http"
)

// LoggingTransport is a http.RoundTripper with a pluggable logger.
type LoggingTransport struct {
	upstream http.RoundTripper
	log      logging.Logger
	filter   logging.Filter
}

// Wrap a client and logs all requests made to it.
func Wrap(client *http.Client, logger logging.Logger, level logging.Level) *http.Client {
	return WrapWithFilter(client, logger, logging.FixedLevel(level))
}

func WrapWithFilter(client *http.Client, logger logging.Logger, filter logging.Filter) *http.Client {
	upstream := http.DefaultTransport
	if client.Transport != nil {
		upstream = client.Transport
	}

	client.Transport = NewWithFilter(upstream, logger, filter)
	return client
}

// New wraps an upstream client and logs all requests made to it.
func New(upstream http.RoundTripper, logger logging.Logger, level logging.Level) http.RoundTripper {
	return NewWithFilter(upstream, logger, logging.FixedLevel(level))
}

func NewWithFilter(upstream http.RoundTripper, logger logging.Logger, filter logging.Filter) http.RoundTripper {
	if upstream == nil || logger == nil {
		panic("Incorrect setup")
	}
	return &LoggingTransport{
		upstream: upstream,
		log:      logger,
		filter:   filter,
	}
}

func (lt *LoggingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	level := lt.filter.Level(req)
	if level == logging.Off {
		return lt.upstream.RoundTrip(req)
	}

	item := internal.PrepareTheLogItem(req, level)
	res, err := lt.upstream.RoundTrip(req)
	return internal.CompleteTheLoggging(res, err, item, lt.log, level)
}

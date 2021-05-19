package loggingtransport

import (
	. "github.com/rickb777/httpclient/internal"
	"github.com/rickb777/httpclient/logging"
	"github.com/rickb777/httpclient/logging/logger"
	"net/http"
)

// LoggingTransport is a http.RoundTripper with a pluggable logger.
type LoggingTransport struct {
	upstream http.RoundTripper
	log      logger.Logger
	filter   logging.Filter
}

// Wrap a client and logs all requests made.
func Wrap(client *http.Client, logger logger.Logger, level logging.Level) *http.Client {
	return WrapWithFilter(client, logger, logging.FixedLevel(level))
}

// WrapWithFilter wraps a client and logs requests made according to the filter.
func WrapWithFilter(client *http.Client, logger logger.Logger, filter logging.Filter) *http.Client {
	upstream := http.DefaultTransport
	if client.Transport != nil {
		upstream = client.Transport
	}

	client.Transport = NewWithFilter(upstream, logger, filter)
	return client
}

// New wraps an upstream client and logs all requests made.
func New(upstream http.RoundTripper, logger logger.Logger, level logging.Level) http.RoundTripper {
	return NewWithFilter(upstream, logger, logging.FixedLevel(level))
}

// NewWithFilter wraps an upstream client and logs requests made according to the filter.
func NewWithFilter(upstream http.RoundTripper, logger logger.Logger, filter logging.Filter) http.RoundTripper {
	if upstream == nil || logger == nil {
		panic("Incorrect setup")
	}
	return &LoggingTransport{
		upstream: upstream,
		log:      logger,
		filter:   filter,
	}
}

// RoundTrip implements http.RoundTripper.
func (lt *LoggingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	level := lt.filter.Level(req)
	if level == logging.Off {
		return lt.upstream.RoundTrip(req)
	}

	item := PrepareTheLogItem(req, level)
	res, err := lt.upstream.RoundTrip(req)
	return CompleteTheLogging(res, err, item, ILogger(lt.log), level)
}

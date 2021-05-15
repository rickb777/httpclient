package loggingtransport

import (
	"github.com/rickb777/httpclient/internal"
	"github.com/rickb777/httpclient/logging"
	"net/http"
	"sync"
)

// LoggingTransport is a http.RoundTripper with a pluggable logger.
type LoggingTransport struct {
	upstream http.RoundTripper
	log      logging.Logger
	level    logging.Level
	mu       sync.RWMutex
}

// Wrap a client and logs all requests made to it.
func Wrap(client *http.Client, logger logging.Logger, level logging.Level) *http.Client {
	upstream := http.DefaultTransport
	if client.Transport != nil {
		upstream = client.Transport
	}

	client.Transport = New(upstream, logger, level)
	return client
}

// New wraps an upstream client and logs all requests made to it.
func New(upstream http.RoundTripper, logger logging.Logger, level logging.Level) *LoggingTransport {
	if upstream == nil || logger == nil {
		panic("Incorrect setup")
	}
	return &LoggingTransport{
		upstream: upstream,
		log:      logger,
		level:    level,
		mu:       sync.RWMutex{},
	}
}

func (lt *LoggingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	level := lt.getLevel()
	if level == logging.Off {
		return lt.upstream.RoundTrip(req)
	}

	item := internal.PrepareTheLogItem(req, level)
	res, err := lt.upstream.RoundTrip(req)
	return internal.CompleteTheLoggging(res, err, item, lt.log, level)
}

func (lt *LoggingTransport) getLevel() logging.Level {
	lt.mu.RLock()
	defer lt.mu.RUnlock()
	l := lt.level
	return l
}

// SetLevel alters the logging level. This can be called concurrently
// from any goroutine.
func (lt *LoggingTransport) SetLevel(newLevel logging.Level) {
	lt.mu.Lock()
	defer lt.mu.Unlock()
	lt.level = newLevel
}

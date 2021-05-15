package loggingclient

import (
	"github.com/rickb777/httpclient"
	"github.com/rickb777/httpclient/internal"
	"github.com/rickb777/httpclient/logging"
	"net/http"
	"sync"
)

// LoggingClient is a HttpClient with a pluggable logger.
type LoggingClient struct {
	upstream httpclient.HttpClient
	log      logging.Logger
	level    logging.Level
	mu       sync.RWMutex
}

// New wraps an upstream client and logs all requests made to it.
func New(upstream httpclient.HttpClient, logger logging.Logger, level logging.Level) *LoggingClient {
	if upstream == nil || logger == nil {
		panic("Incorrect setup")
	}
	return &LoggingClient{
		upstream: upstream,
		log:      logger,
		level:    level,
		mu:       sync.RWMutex{},
	}
}

func (lc *LoggingClient) SetCheckRedirect(fn func(req *http.Request, via []*http.Request) error) {
	if hc, ok := lc.upstream.(*http.Client); ok {
		hc.CheckRedirect = fn
	} else if cr, ok := lc.upstream.(httpclient.ControlledRedirectClient); ok {
		cr.SetCheckRedirect(fn)
	}
}

func (lc *LoggingClient) Do(req *http.Request) (*http.Response, error) {
	level := lc.getLevel()
	if level == logging.Off {
		return lc.upstream.Do(req)
	}

	item := internal.PrepareTheLogItem(req, level)
	res, err := lc.upstream.Do(req)
	return internal.CompleteTheLoggging(res, err, item, lc.log, level)
}

func (lc *LoggingClient) getLevel() logging.Level {
	lc.mu.RLock()
	defer lc.mu.RUnlock()
	l := lc.level
	return l
}

// SetLevel alters the logging level. This can be called concurrently
// from any goroutine.
func (lc *LoggingClient) SetLevel(newLevel logging.Level) {
	lc.mu.Lock()
	defer lc.mu.Unlock()
	lc.level = newLevel
}

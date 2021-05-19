package loggingclient

import (
	"github.com/rickb777/httpclient"
	. "github.com/rickb777/httpclient/internal"
	"github.com/rickb777/httpclient/logging"
	"github.com/rickb777/httpclient/logging/logger"
	"net/http"
)

// LoggingClient is a HttpClient with a pluggable logger.
type LoggingClient struct {
	upstream httpclient.HttpClient
	log      logger.Logger
	filter   logging.Filter
}

// New wraps an upstream client and logs all requests made to it.
func New(upstream httpclient.HttpClient, logger logger.Logger, level logging.Level) httpclient.HttpClient {
	return NewWithFilter(upstream, logger, logging.FixedLevel(level))
}

func NewWithFilter(upstream httpclient.HttpClient, logger logger.Logger, filter logging.Filter) httpclient.HttpClient {
	if upstream == nil || logger == nil {
		panic("Incorrect setup")
	}
	return &LoggingClient{
		upstream: upstream,
		log:      logger,
		filter:   filter,
	}
}

// SetCheckRedirect provides access to the http.Client.CheckRedirect field.
func (lc *LoggingClient) SetCheckRedirect(fn func(req *http.Request, via []*http.Request) error) {
	if hc, ok := lc.upstream.(*http.Client); ok {
		hc.CheckRedirect = fn
	} else if cr, ok := lc.upstream.(httpclient.ControlledRedirectClient); ok {
		cr.SetCheckRedirect(fn)
	}
}

func (lc *LoggingClient) Do(req *http.Request) (*http.Response, error) {
	level := lc.filter.Level(req)
	if level == logging.Off {
		return lc.upstream.Do(req)
	}

	item := PrepareTheLogItem(req, level)
	res, err := lc.upstream.Do(req)
	return CompleteTheLogging(res, err, item, ILogger(lc.log), level)
}

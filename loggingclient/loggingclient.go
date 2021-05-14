package loggingclient

import (
	"bytes"
	"github.com/rickb777/httpclient"
	"github.com/rickb777/httpclient/logging"
	"io"
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
func New(upstream httpclient.HttpClient, logger logging.Logger, level logging.Level) httpclient.HttpClient {
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
	return lc.loggingDo(req, level)
}

func (lc *LoggingClient) loggingDo(req *http.Request, level logging.Level) (*http.Response, error) {
	item := &logging.LogItem{
		Method: req.Method,
		URL:    req.URL,
		Level:  level,
	}

	if level <= logging.Discrete {
		u2 := *req.URL
		u2.RawQuery = ""
		item.URL = &u2
	}

	item.Request.Header = req.Header

	if level == logging.WithHeadersAndBodies {
		if req.Body != nil && req.Body != http.NoBody {
			buf, _ := readIntoBuffer(req.Body)
			item.Request.Body = buf.Bytes()
			req.Body = io.NopCloser(bytes.NewBuffer(item.Request.Body))
			req.GetBody = func() (io.ReadCloser, error) {
				return io.NopCloser(bytes.NewBuffer(item.Request.Body)), nil
			}

		} else if req.GetBody != nil {
			rdr, _ := req.GetBody()
			buf, _ := readIntoBuffer(rdr)
			item.Request.Body = buf.Bytes()
		}
	}

	item.Start = logging.Now()

	res, err := lc.upstream.Do(req)

	item.Duration = logging.Now().Sub(item.Start)

	if res != nil {
		item.StatusCode = res.StatusCode
	}

	if err != nil {
		item.Err = err
		lc.log(item)
		return res, err
	}

	if level >= logging.WithHeaders {
		item.Response.Header = res.Header
	}

	if level == logging.WithHeadersAndBodies {
		item.Response.Body, err = captureBytes(res.Body)
		if err != nil {
			return nil, err
		}
		res.Body = io.NopCloser(bytes.NewBuffer(item.Response.Body))
	}

	lc.log(item)
	return res, err
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

func captureBytes(in io.ReadCloser) ([]byte, error) {
	defer in.Close()
	buf, err := readIntoBuffer(in)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func readIntoBuffer(in io.Reader) (*bytes.Buffer, error) {
	buf := &bytes.Buffer{}
	_, err := buf.ReadFrom(in)
	return buf, err
}

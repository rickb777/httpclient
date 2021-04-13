package loggingclient

import (
	"bytes"
	"github.com/rickb777/httpclient"
	"io"
	"net/http"
	"strings"
	"time"
)

// Level allows control of the level of detail in log messages.
type Level int

const (
	// Off turns logging off.
	Off Level = iota

	// Discrete log messages contain only a summary of the request and response.
	// No query parameters are printed in order to hide potential personal information.
	Discrete

	// Summary log messages contain only a summary of the request and response,
	// including the full target URL.
	Summary

	// WithHeaders log messages contain a summary and the request/response headers
	WithHeaders

	// WithHeadersAndBodies log messages contain a summary and the request/response headers and bodies
	// Textual bodies are included in the log; for binary content, the size is shown instead.
	WithHeadersAndBodies
)

// LoggingClient is a HttpClient with a pluggable logger.
type LoggingClient struct {
	upstream httpclient.HttpClient
	log      Logger
	level    Level
}

// New wraps an upstream client and logs all requests made to it.
func New(upstream httpclient.HttpClient, logger Logger, level Level) httpclient.HttpClient {
	if upstream == nil || logger == nil {
		panic("Incorrect setup")
	}
	return &LoggingClient{
		upstream: upstream,
		log:      logger,
		level:    level,
	}
}

func (l *LoggingClient) Do(req *http.Request) (*http.Response, error) {
	if l.level == Off {
		return l.upstream.Do(req)
	}
	return l.loggingDo(req)
}

func (l *LoggingClient) loggingDo(req *http.Request) (*http.Response, error) {
	item := &LogItem{
		Method: req.Method,
		URL:    req.URL.String(),
		Level:  l.level,
	}

	if l.level <= Discrete {
		parts := strings.SplitN(item.URL, "?", 2)
		item.URL = parts[0]
	}

	item.Request.Header = req.Header

	if l.level == WithHeadersAndBodies {
		if req.Body != nil && req.Body != http.NoBody {
			buf, _ := readIntoBuffer(req.Body)
			item.Request.Body = buf.Bytes()
		} else if req.GetBody != nil {
			rdr, _ := req.GetBody()
			buf, _ := readIntoBuffer(rdr)
			item.Request.Body = buf.Bytes()
		}
	}

	t0 := time.Now().UTC()
	res, err := l.upstream.Do(req)
	item.Duration = time.Now().UTC().Sub(t0)

	if res != nil {
		item.StatusCode = res.StatusCode
	}

	if err != nil {
		item.Err = err
		l.log(item)
		return res, err
	}

	if l.level >= WithHeaders {
		item.Response.Header = res.Header
	}

	if l.level == WithHeadersAndBodies {
		item.Response.Body, err = captureBytes(res.Body)
		if err != nil {
			return nil, err
		}
		res.Body = io.NopCloser(bytes.NewBuffer(item.Response.Body))
	}

	l.log(item)
	return res, err
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
	_, err := io.Copy(buf, in)
	return buf, err
}

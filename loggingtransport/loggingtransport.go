package loggingtransport

import (
	"bytes"
	"github.com/rickb777/httpclient/logging"
	"io"
	"net/http"
)

// LoggingTransport is a http.RoundTripper with a pluggable logger.
type LoggingTransport struct {
	upstream http.RoundTripper
	log      logging.Logger
	level    logging.Level
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
func New(upstream http.RoundTripper, logger logging.Logger, level logging.Level) http.RoundTripper {
	if upstream == nil || logger == nil {
		panic("Incorrect setup")
	}
	return &LoggingTransport{
		upstream: upstream,
		log:      logger,
		level:    level,
	}
}

func (lt *LoggingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if lt.level == logging.Off {
		return lt.upstream.RoundTrip(req)
	}
	return lt.loggingDo(req)
}

func (lt *LoggingTransport) loggingDo(req *http.Request) (*http.Response, error) {
	item := &logging.LogItem{
		Method: req.Method,
		URL:    req.URL,
		Level:  lt.level,
	}

	if lt.level <= logging.Discrete {
		u2 := *req.URL
		u2.RawQuery = ""
		item.URL = &u2
	}

	item.Request.Header = req.Header

	if lt.level == logging.WithHeadersAndBodies {
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

	res, err := lt.upstream.RoundTrip(req)

	item.Duration = logging.Now().Sub(item.Start)

	if res != nil {
		item.StatusCode = res.StatusCode
	}

	if err != nil {
		item.Err = err
		lt.log(item)
		return res, err
	}

	if lt.level >= logging.WithHeaders {
		item.Response.Header = res.Header
	}

	if lt.level == logging.WithHeadersAndBodies {
		item.Response.Body, err = captureBytes(res.Body)
		if err != nil {
			return nil, err
		}
		res.Body = io.NopCloser(bytes.NewBuffer(item.Response.Body))
	}

	lt.log(item)
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
	_, err := buf.ReadFrom(in)
	return buf, err
}

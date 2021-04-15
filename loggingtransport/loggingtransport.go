package loggingtransport

import (
	"bytes"
	"github.com/rickb777/httpclient/loggingclient"
	"io"
	"net/http"
	"strings"
	"time"
)

// LoggingTransport is a http.RoundTripper with a pluggable logger.
type LoggingTransport struct {
	upstream http.RoundTripper
	log      loggingclient.Logger
	level    loggingclient.Level
}

// New wraps an upstream client and logs all requests made to it.
func New(upstream http.RoundTripper, logger loggingclient.Logger, level loggingclient.Level) http.RoundTripper {
	if upstream == nil || logger == nil {
		panic("Incorrect setup")
	}
	return &LoggingTransport{
		upstream: upstream,
		log:      logger,
		level:    level,
	}
}

func (t *LoggingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if t.level == loggingclient.Off {
		return t.upstream.RoundTrip(req)
	}
	return t.loggingDo(req)
}

func (t *LoggingTransport) loggingDo(req *http.Request) (*http.Response, error) {
	item := &loggingclient.LogItem{
		Method: req.Method,
		URL:    req.URL.String(),
		Level:  t.level,
	}

	if t.level <= loggingclient.Discrete {
		parts := strings.SplitN(item.URL, "?", 2)
		item.URL = parts[0]
	}

	item.Request.Header = req.Header

	if t.level == loggingclient.WithHeadersAndBodies {
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
	res, err := t.upstream.RoundTrip(req)
	item.Duration = time.Now().UTC().Sub(t0)

	if res != nil {
		item.StatusCode = res.StatusCode
	}

	if err != nil {
		item.Err = err
		t.log(item)
		return res, err
	}

	if t.level >= loggingclient.WithHeaders {
		item.Response.Header = res.Header
	}

	if t.level == loggingclient.WithHeadersAndBodies {
		item.Response.Body, err = captureBytes(res.Body)
		if err != nil {
			return nil, err
		}
		res.Body = io.NopCloser(bytes.NewBuffer(item.Response.Body))
	}

	t.log(item)
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

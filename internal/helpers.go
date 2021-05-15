package internal

import (
	"bytes"
	"github.com/rickb777/httpclient/logging"
	"io"
	"net/http"
)

func PrepareTheLogItem(req *http.Request, level logging.Level) *logging.LogItem {
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
	return item
}

func CompleteTheLoggging(res *http.Response, err error, item *logging.LogItem, log logging.Logger, level logging.Level) (*http.Response, error) {
	item.Duration = logging.Now().Sub(item.Start)

	if res != nil {
		item.StatusCode = res.StatusCode
	}

	if err != nil {
		item.Err = err
		log(item)
		return res, err
	}

	if res == nil {
		panic(item) // not expected ever
	}

	if level >= logging.WithHeaders {
		item.Response.Header = res.Header
	}

	if level == logging.WithHeadersAndBodies {
		item.Response.Body, err = captureBytes(res.Body)
		if err != nil {
			log(item)
			return nil, err
		}
		res.Body = io.NopCloser(bytes.NewBuffer(item.Response.Body))
	}

	log(item)
	return res, nil
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

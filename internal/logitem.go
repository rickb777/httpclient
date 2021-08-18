package internal

import (
	"github.com/rickb777/httpclient/body"
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
			item.Request.Body, _ = body.Copy(req.Body)
			req.Body = item.Request.Body
			req.GetBody = func() (io.ReadCloser, error) {
				return item.Request.Body, nil
			}

		} else if req.GetBody != nil {
			rdr, _ := req.GetBody()
			item.Request.Body, _ = body.Copy(rdr)
		}
	}

	item.Start = logging.Now()
	return item
}

type ILogger func(item *logging.LogItem)

func CompleteTheLogging(res *http.Response, err error, item *logging.LogItem, log ILogger, level logging.Level) (*http.Response, error) {
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
		item.Response.Body, err = body.Copy(res.Body)
		if err != nil {
			log(item)
			return nil, err
		}
		res.Body = item.Response.Body
	}

	log(item)
	return res, nil
}

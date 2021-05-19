package logger

import (
	"bytes"
	"fmt"
	"github.com/onsi/gomega"
	"github.com/rickb777/httpclient/body"
	"github.com/rickb777/httpclient/logging"
	"github.com/spf13/afero"
	"net/http"
	"net/url"
	"testing"
	"time"
)

const longJSON = `{"alpha":"some text","beta":"some more text","gamma":"this might drag on","delta":"and on past the 80 char threshold"}` + "\n"

var t0 = time.Date(2021, 04, 01, 10, 11, 12, 0, time.UTC)

func TestLogWriter_typical_GET_terse(t *testing.T) {
	g := gomega.NewWithT(t)

	u, _ := url.Parse("http://somewhere.com/a/b/c")
	reqHeader := make(http.Header)
	reqHeader.Set("Accept", "application/json")
	reqHeader.Set("Cookie", "a=123")
	reqHeader.Add("Cookie", "b=4556")

	resHeader := make(http.Header)
	resHeader.Set("Content-Type", "application/json; charset=UTF-8")
	resHeader.Set("Content-Length", "18")

	buf := &bytes.Buffer{}
	log := LogWriter(buf, afero.NewMemMapFs())
	log(&logging.LogItem{
		Method:     "GET",
		URL:        u,
		StatusCode: 200,
		Request:    logging.LogContent{Header: reqHeader},
		Response:   logging.LogContent{Header: resHeader},
		Err:        nil,
		Start:      t0,
		Duration:   time.Millisecond,
		Level:      0,
	})

	g.Expect(buf.String()).To(gomega.Equal("10:11:12 GET      http://somewhere.com/a/b/c 200 1ms\n"), buf.String())
}

func TestLogWriter_typical_GET_JSON_short_content(t *testing.T) {
	g := gomega.NewWithT(t)

	u, _ := url.Parse("http://somewhere.com/a/b/c?foo=1")
	reqHeader := make(http.Header)
	reqHeader.Set("Accept", "application/json")
	reqHeader.Set("Cookie", "a=123")
	reqHeader.Add("Cookie", "b=4556")

	resHeader := make(http.Header)
	resHeader.Set("Content-Type", "application/json; charset=UTF-8")
	resHeader.Set("Content-Length", "18")

	buf := &bytes.Buffer{}
	log := LogWriter(buf, afero.NewMemMapFs())
	log(&logging.LogItem{
		Method:     "GET",
		URL:        u,
		StatusCode: 200,
		Request: logging.LogContent{
			Header: reqHeader,
		},
		Response: logging.LogContent{
			Header: resHeader,
			Body:   body.NewBodyString(`{"A":"foo","B":7}` + "\n"),
		},
		Err:      nil,
		Start:    t0,
		Duration: time.Millisecond,
		Level:    logging.WithHeadersAndBodies,
	})

	g.Expect(buf.String()).To(gomega.Equal(
		`10:11:12 GET      http://somewhere.com/a/b/c?foo=1 200 1ms
--> Accept:          application/json
--> Cookie:          a=123
-->                  b=4556
<-- Content-Length:  18
<-- Content-Type:    application/json; charset=UTF-8
{"A":"foo","B":7}
---
`), buf.String())
}

func TestLogWriter_typical_GET_JSON_long_content(t *testing.T) {
	g := gomega.NewWithT(t)

	u, _ := url.Parse("http://somewhere.com/a/b/c?foo=1")
	reqHeader := make(http.Header)
	reqHeader.Set("Accept", "application/json")
	reqHeader.Set("Cookie", "a=123")
	reqHeader.Add("Cookie", "b=4556")

	resHeader := make(http.Header)
	resHeader.Set("Content-Type", "application/json; charset=UTF-8")
	resHeader.Set("Content-Length", "18")

	buf := &bytes.Buffer{}
	log := LogWriter(buf, afero.NewMemMapFs())
	log(&logging.LogItem{
		Method:     "GET",
		URL:        u,
		StatusCode: 200,
		Request: logging.LogContent{
			Header: reqHeader,
		},
		Response: logging.LogContent{
			Header: resHeader,
			Body:   body.NewBodyString(longJSON),
		},
		Err:      nil,
		Start:    t0,
		Duration: time.Millisecond,
		Level:    logging.WithHeadersAndBodies,
	})

	g.Expect(buf.String()).To(gomega.Equal(
		`10:11:12 GET      http://somewhere.com/a/b/c?foo=1 200 1ms
--> Accept:          application/json
--> Cookie:          a=123
-->                  b=4556
<-- Content-Length:  18
<-- Content-Type:    application/json; charset=UTF-8
see 2021-04-01_10-11-12_GET_a_b_c_resp.json
---
`), buf.String())
}

func TestLogWriter_typical_GET_text_long_content(t *testing.T) {
	g := gomega.NewWithT(t)

	u, _ := url.Parse("http://somewhere.com/a/b/c?foo=1")
	reqHeader := make(http.Header)
	reqHeader.Set("Accept", "text/*")
	reqHeader.Set("Cookie", "a=123")
	reqHeader.Add("Cookie", "b=4556")

	resHeader := make(http.Header)
	resHeader.Set("Content-Type", "text/plain; charset=UTF-8")
	resHeader.Set("Content-Length", "18")

	buf := &bytes.Buffer{}
	log := LogWriter(buf, afero.NewMemMapFs())
	log(&logging.LogItem{
		Method:     "GET",
		URL:        u,
		StatusCode: 200,
		Request: logging.LogContent{
			Header: reqHeader,
		},
		Response: logging.LogContent{
			Header: resHeader,
			Body: body.NewBodyString("So shaken as we are, so wan with care\n" +
				"Find we a time for frighted peace to pant\n" +
				"And breathe short-winded accents of new broils\n" +
				"To be commenced in strands afar remote.\n"),
		},
		Err:      nil,
		Start:    t0,
		Duration: time.Millisecond,
		Level:    logging.WithHeadersAndBodies,
	})

	g.Expect(buf.String()).To(gomega.Equal(
		`10:11:12 GET      http://somewhere.com/a/b/c?foo=1 200 1ms
--> Accept:          text/*
--> Cookie:          a=123
-->                  b=4556
<-- Content-Length:  18
<-- Content-Type:    text/plain; charset=UTF-8
see 2021-04-01_10-11-12_GET_a_b_c_resp.txt
---
`), buf.String())
}

func TestLogWriter_typical_GET_XML_long_content(t *testing.T) {
	g := gomega.NewWithT(t)

	u, _ := url.Parse("http://somewhere.com/a/b/c?foo=1")
	reqHeader := make(http.Header)
	reqHeader.Set("Accept", "application/xml")
	reqHeader.Set("Cookie", "a=123")
	reqHeader.Add("Cookie", "b=4556")

	resHeader := make(http.Header)
	resHeader.Set("Content-Type", "application/xml; charset=UTF-8")
	resHeader.Set("Content-Length", "18")

	buf := &bytes.Buffer{}
	log := LogWriter(buf, afero.NewMemMapFs())
	log(&logging.LogItem{
		Method:     "GET",
		URL:        u,
		StatusCode: 200,
		Request: logging.LogContent{
			Header: reqHeader,
		},
		Response: logging.LogContent{
			Header: resHeader,
			Body: body.NewBodyString(`<xml>
<alpha>some text</alpha>
<beta>some more text</beta>
<gamma>this might drag on past the 80 char threshold</gamma>
</xml>` + "\n"),
		},
		Err:      nil,
		Start:    t0,
		Duration: time.Millisecond,
		Level:    logging.WithHeadersAndBodies,
	})

	g.Expect(buf.String()).To(gomega.Equal(
		`10:11:12 GET      http://somewhere.com/a/b/c?foo=1 200 1ms
--> Accept:          application/xml
--> Cookie:          a=123
-->                  b=4556
<-- Content-Length:  18
<-- Content-Type:    application/xml; charset=UTF-8
see 2021-04-01_10-11-12_GET_a_b_c_resp.xml
---
`), buf.String())
}

func TestLogWriter_typical_GET_binary(t *testing.T) {
	g := gomega.NewWithT(t)

	u, _ := url.Parse("http://somewhere.com/a/b/c")
	reqHeader := make(http.Header)
	reqHeader.Set("Accept", "application/*")

	resHeader := make(http.Header)
	resHeader.Set("Content-Type", "application/octet-stream")
	resHeader.Set("Content-Length", "3")

	buf := &bytes.Buffer{}
	log := LogWriter(buf, afero.NewMemMapFs())
	log(&logging.LogItem{
		Method:     "GET",
		URL:        u,
		StatusCode: 200,
		Request: logging.LogContent{
			Header: reqHeader,
		},
		Response: logging.LogContent{
			Header: resHeader,
			Body:   body.NewBodyString("{}\n"),
		},
		Err:      nil,
		Start:    t0,
		Duration: time.Millisecond,
		Level:    logging.WithHeadersAndBodies,
	})

	g.Expect(buf.String()).To(gomega.Equal(
		`10:11:12 GET      http://somewhere.com/a/b/c 200 1ms
--> Accept:          application/*
<-- Content-Length:  3
<-- Content-Type:    application/octet-stream
<-- binary content [3]byte
---
`), buf.String())
}

func TestLogWriter_typical_PUT_headers_only_with_error(t *testing.T) {
	g := gomega.NewWithT(t)

	u, _ := url.Parse("http://somewhere.com/a/b/c")
	reqHeader := make(http.Header)
	reqHeader.Set("Content-Type", "application/json; charset=UTF-8")
	reqHeader.Set("Content-Length", "18")

	buf := &bytes.Buffer{}
	log := LogWriter(buf, afero.NewMemMapFs())
	log(&logging.LogItem{
		Method:     "PUT",
		URL:        u,
		StatusCode: 0,
		Request: logging.LogContent{
			Header: reqHeader,
			Body:   body.NewBodyString(`{"A":"foo","B":7}` + "\n"),
		},
		Start:    t0,
		Duration: 123456,
		Level:    logging.WithHeaders,
		Err:      fmt.Errorf("Bang!"),
	})

	g.Expect(buf.String()).To(gomega.Equal(
		`10:11:12 PUT      http://somewhere.com/a/b/c 0 123µs Bang!
--> Content-Length:  18
--> Content-Type:    application/json; charset=UTF-8
<-- no headers
---
`), buf.String())
}

func TestLogWriter_typical_PUT_short_content(t *testing.T) {
	g := gomega.NewWithT(t)

	u, _ := url.Parse("http://somewhere.com/a/b/c")
	reqHeader := make(http.Header)
	reqHeader.Set("Content-Type", "application/json; charset=UTF-8")
	reqHeader.Set("Content-Length", "18")

	buf := &bytes.Buffer{}
	log := LogWriter(buf, afero.NewMemMapFs())
	log(&logging.LogItem{
		Method:     "PUT",
		URL:        u,
		StatusCode: 204,
		Request: logging.LogContent{
			Header: reqHeader,
			Body:   body.NewBodyString(`{"A":"foo","B":7}` + "\n"),
		},
		Start:    t0,
		Duration: time.Millisecond,
		Level:    logging.WithHeadersAndBodies,
	})

	g.Expect(buf.String()).To(gomega.Equal(
		`10:11:12 PUT      http://somewhere.com/a/b/c 204 1ms
--> Content-Length:  18
--> Content-Type:    application/json; charset=UTF-8
{"A":"foo","B":7}
<-- no headers
---
`), buf.String())
}

func TestLogWriter_typical_PUT_long_content(t *testing.T) {
	g := gomega.NewWithT(t)

	u, _ := url.Parse("http://somewhere.com/a/b/c")
	reqHeader := make(http.Header)
	reqHeader.Set("Content-Type", "application/json; charset=UTF-8")
	reqHeader.Set("Content-Length", "18")

	buf := &bytes.Buffer{}
	log := LogWriter(buf, afero.NewMemMapFs())
	log(&logging.LogItem{
		Method:     "PUT",
		URL:        u,
		StatusCode: 204,
		Request: logging.LogContent{
			Header: reqHeader,
			Body:   body.NewBodyString(longJSON),
		},
		Start:    t0,
		Duration: time.Millisecond,
		Level:    logging.WithHeadersAndBodies,
	})

	g.Expect(buf.String()).To(gomega.Equal(
		`10:11:12 PUT      http://somewhere.com/a/b/c 204 1ms
--> Content-Length:  18
--> Content-Type:    application/json; charset=UTF-8
see 2021-04-01_10-11-12_PUT_a_b_c_req.json
<-- no headers
---
`), buf.String())
}

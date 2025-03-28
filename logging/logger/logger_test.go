package logger

import (
	"bytes"
	"fmt"
	"github.com/rickb777/expect"
	"github.com/rickb777/httpclient/body"
	"github.com/rickb777/httpclient/logging"
	"github.com/spf13/afero"
	"io"
	"net/http"
	"net/url"
	"testing"
	"time"
)

const longJSON = `{"alpha":"some text","beta":"some more text","gamma":"this might drag on","delta":"and on past the 80 char threshold"}` + "\n"

var t0 = time.Date(2021, 04, 01, 10, 11, 12, 1000000, time.UTC)

func TestLogWriter_typical_GET_terse(t *testing.T) {
	u, _ := url.Parse("http://somewhere.com/a/b/c")
	reqHeader := make(http.Header)
	reqHeader.Set("Host", "somewhere.com")
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

	expect.String(buf.String()).ToBe(t, "10:11:12 GET      http://somewhere.com/a/b/c 200 1ms\n")
}

func TestLogWriter_typical_GET_JSON_short_content(t *testing.T) {
	u, _ := url.Parse("http://somewhere.com/a/b/c?foo=1")
	reqHeader := make(http.Header)
	reqHeader.Set("Host", "somewhere.com")
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

	expect.String(buf.String()).ToBe(t,
		`10:11:12 GET      http://somewhere.com/a/b/c?foo=1 200 1ms
--> Accept:          application/json
--> Cookie:          a=123
-->                  b=4556
--> Host:            somewhere.com
<-- Content-Length:  18
<-- Content-Type:    application/json; charset=UTF-8
{"A":"foo","B":7}
---
`)
}

func TestLogWriter_typical_GET_JSON_long_content(t *testing.T) {
	u, _ := url.Parse("http://somewhere.com/a/b/c?foo=1")
	reqHeader := make(http.Header)
	reqHeader.Set("Host", "somewhere.com")
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

	expect.String(buf.String()).ToBe(t,
		`10:11:12 GET      http://somewhere.com/a/b/c?foo=1 200 1ms
--> Accept:          application/json
--> Cookie:          a=123
-->                  b=4556
--> Host:            somewhere.com
<-- Content-Length:  18
<-- Content-Type:    application/json; charset=UTF-8
see 2021-04-01_10-11-12-001_GET_somewhere.com_a_b_c_resp.json
---
`)
}

func TestLogWriter_typical_GET_text_long_content(t *testing.T) {
	u, _ := url.Parse("http://somewhere.com/a/b/c?foo=1")
	reqHeader := make(http.Header)
	reqHeader.Set("Host", "somewhere.com")
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

	expect.String(buf.String()).ToBe(t,
		`10:11:12 GET      http://somewhere.com/a/b/c?foo=1 200 1ms
--> Accept:          text/*
--> Cookie:          a=123
-->                  b=4556
--> Host:            somewhere.com
<-- Content-Length:  18
<-- Content-Type:    text/plain; charset=UTF-8
see 2021-04-01_10-11-12-001_GET_somewhere.com_a_b_c_resp.txt
---
`)
}

func TestLogWriter_typical_GET_XML_long_content(t *testing.T) {
	u, _ := url.Parse("http://somewhere.com/a/b/c?foo=1")
	reqHeader := make(http.Header)
	reqHeader.Set("Host", "somewhere.com")
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

	expect.String(buf.String()).ToBe(t,
		`10:11:12 GET      http://somewhere.com/a/b/c?foo=1 200 1ms
--> Accept:          application/xml
--> Cookie:          a=123
-->                  b=4556
--> Host:            somewhere.com
<-- Content-Length:  18
<-- Content-Type:    application/xml; charset=UTF-8
see 2021-04-01_10-11-12-001_GET_somewhere.com_a_b_c_resp.xml
---
`)
}

func TestLogWriter_typical_GET_binary(t *testing.T) {
	u, _ := url.Parse("http://somewhere.com/a/b/c")
	reqHeader := make(http.Header)
	reqHeader.Set("Host", "somewhere.com")
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

	expect.String(buf.String()).ToBe(t,
		`10:11:12 GET      http://somewhere.com/a/b/c 200 1ms
--> Accept:          application/*
--> Host:            somewhere.com
<-- Content-Length:  3
<-- Content-Type:    application/octet-stream
<-- binary content [3]byte
---
`)
}

func TestLogWriter_typical_PUT_headers_only_with_error(t *testing.T) {
	u, _ := url.Parse("http://somewhere.com/a/b/c")
	reqHeader := make(http.Header)
	reqHeader.Set("Host", "somewhere.com")
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

	expect.String(buf.String()).ToBe(t,
		`10:11:12 PUT      http://somewhere.com/a/b/c 0 123µs Bang!
--> Content-Length:  18
--> Content-Type:    application/json; charset=UTF-8
--> Host:            somewhere.com
<-- no headers
---
`)
}

func TestLogWriter_typical_PUT_short_content(t *testing.T) {
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

	expect.String(buf.String()).ToBe(t,
		`10:11:12 PUT      http://somewhere.com/a/b/c 204 1ms
--> Content-Length:  18
--> Content-Type:    application/json; charset=UTF-8
{"A":"foo","B":7}
<-- no headers
---
`)
}

func TestLogWriter_typical_PUT_long_content(t *testing.T) {
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

	expect.String(buf.String()).ToBe(t,
		`10:11:12 PUT      http://somewhere.com/a/b/c 204 1ms
--> Content-Length:  18
--> Content-Type:    application/json; charset=UTF-8
see 2021-04-01_10-11-12-001_PUT_a_b_c_req.json
<-- no headers
---
`)
}

func BenchmarkLogWriter_terse(b *testing.B) {
	u, _ := url.Parse("http://somewhere.com/a/b/c")

	reqHeader := make(http.Header)
	reqHeader.Set("Host", "somewhere.com")
	reqHeader.Set("Accept", "application/json")
	reqHeader.Set("Cookie", "a=123")
	reqHeader.Add("Cookie", "b=4556")

	resHeader := make(http.Header)
	resHeader.Set("Content-Type", "application/json; charset=UTF-8")
	resHeader.Set("Content-Length", "18")

	log := LogWriter(io.Discard, afero.NewMemMapFs())

	item := logging.LogItem{
		Method:     "GET",
		URL:        u,
		StatusCode: 200,
		Request:    logging.LogContent{Header: reqHeader},
		Response:   logging.LogContent{Header: resHeader},
		Err:        nil,
		Start:      t0,
		Duration:   time.Millisecond,
	}

	b.Run("off", func(b *testing.B) {
		i1 := item
		i1.Level = logging.Off

		for i := 0; i < b.N; i++ {
			log(&i1)
		}
	})

	b.Run("summary", func(b *testing.B) {
		i1 := item
		i1.Level = logging.Summary

		for i := 0; i < b.N; i++ {
			log(&i1)
		}
	})

	b.Run("withHeaders", func(b *testing.B) {
		i1 := item
		i1.Level = logging.WithHeaders

		for i := 0; i < b.N; i++ {
			log(&i1)
		}
	})
}

func BenchmarkLogWriter_long_response(b *testing.B) {
	u, _ := url.Parse("http://somewhere.com/a/b/c")

	reqHeader := make(http.Header)
	reqHeader.Set("Host", "somewhere.com")
	reqHeader.Set("Accept", "application/json")
	reqHeader.Set("Cookie", "a=123")
	reqHeader.Add("Cookie", "b=4556")

	resHeader := make(http.Header)
	resHeader.Set("Content-Type", "application/json; charset=UTF-8")
	resHeader.Set("Content-Length", "18")

	log := LogWriter(io.Discard, afero.NewMemMapFs())

	item := logging.LogItem{
		Method:     "GET",
		URL:        u,
		StatusCode: 200,
		Request:    logging.LogContent{Header: reqHeader},
		Response: logging.LogContent{
			Header: resHeader,
			Body:   body.NewBodyString(longJSON),
		},
		Err:      nil,
		Start:    t0,
		Duration: time.Millisecond,
	}

	b.Run("off", func(b *testing.B) {
		i1 := item
		i1.Level = logging.Off

		for i := 0; i < b.N; i++ {
			log(&i1)
		}
	})

	b.Run("summary", func(b *testing.B) {
		i1 := item
		i1.Level = logging.Summary

		for i := 0; i < b.N; i++ {
			log(&i1)
		}
	})

	b.Run("withHeaders", func(b *testing.B) {
		i1 := item
		i1.Level = logging.WithHeaders

		for i := 0; i < b.N; i++ {
			log(&i1)
		}
	})
}

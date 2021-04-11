package httpclient

import (
	"bytes"
	"github.com/onsi/gomega"
	"net/http"
	"testing"
	"time"
)

func TestLogWriter_typical_GET_terse(t *testing.T) {
	g := gomega.NewWithT(t)

	u := "http://somewhere.com/a/b/c"
	reqHeader := make(http.Header)
	reqHeader.Set("Accept", "application/json")
	reqHeader.Set("Cookie", "a=123")
	reqHeader.Add("Cookie", "b=4556")

	resHeader := make(http.Header)
	resHeader.Set("Content-Type", "application/json; charset=UTF-8")
	resHeader.Set("Content-Length", "18")

	buf := &bytes.Buffer{}
	log := LogWriter(buf)
	log(&LogItem{
		Method:     "GET",
		URL:        u,
		StatusCode: 200,
		Request:    LogContent{Header: reqHeader},
		Response:   LogContent{Header: resHeader},
		Err:        nil,
		Duration:   time.Millisecond,
		Level:      0,
	})

	g.Expect(buf.String()).To(gomega.Equal("GET      http://somewhere.com/a/b/c 200 1ms\n"), buf.String())
}

func TestLogWriter_typical_GET_JSON(t *testing.T) {
	g := gomega.NewWithT(t)

	u := "http://somewhere.com/a/b/c?foo=1"
	reqHeader := make(http.Header)
	reqHeader.Set("Accept", "application/json")
	reqHeader.Set("Cookie", "a=123")
	reqHeader.Add("Cookie", "b=4556")

	resHeader := make(http.Header)
	resHeader.Set("Content-Type", "application/json; charset=UTF-8")
	resHeader.Set("Content-Length", "18")

	buf := &bytes.Buffer{}
	log := LogWriter(buf)
	log(&LogItem{
		Method:     "GET",
		URL:        u,
		StatusCode: 200,
		Request: LogContent{
			Header: reqHeader,
		},
		Response: LogContent{
			Header: resHeader,
			Body:   []byte(`{"A":"foo","B":7}` + "\n"),
		},
		Err:      nil,
		Duration: time.Millisecond,
		Level:    WithHeadersAndBodies,
	})

	g.Expect(buf.String()).To(gomega.Equal(
		`GET      http://somewhere.com/a/b/c?foo=1 200 1ms
--> Accept:         application/json
--> Cookie:         a=123
-->                 b=4556

<-- Content-Length: 18
<-- Content-Type:   application/json; charset=UTF-8

{"A":"foo","B":7}

---
`), buf.String())
}

func TestLogWriter_typical_GET_binary(t *testing.T) {
	g := gomega.NewWithT(t)

	u := "http://somewhere.com/a/b/c"
	reqHeader := make(http.Header)
	reqHeader.Set("Accept", "application/*")

	resHeader := make(http.Header)
	resHeader.Set("Content-Type", "application/octet-stream")
	resHeader.Set("Content-Length", "3")

	buf := &bytes.Buffer{}
	log := LogWriter(buf)
	log(&LogItem{
		Method:     "GET",
		URL:        u,
		StatusCode: 200,
		Request: LogContent{
			Header: reqHeader,
		},
		Response: LogContent{
			Header: resHeader,
			Body:   []byte("{}\n"),
		},
		Err:      nil,
		Duration: time.Millisecond,
		Level:    WithHeadersAndBodies,
	})

	g.Expect(buf.String()).To(gomega.Equal(
		`GET      http://somewhere.com/a/b/c 200 1ms
--> Accept:         application/*

<-- Content-Length: 3
<-- Content-Type:   application/octet-stream
<-- binary content [3]byte

---
`), buf.String())
}

func TestLogWriter_typical_PUT(t *testing.T) {
	g := gomega.NewWithT(t)

	u := "http://somewhere.com/a/b/c"
	reqHeader := make(http.Header)
	reqHeader.Set("Content-Type", "application/json; charset=UTF-8")
	reqHeader.Set("Content-Length", "18")

	buf := &bytes.Buffer{}
	log := LogWriter(buf)
	log(&LogItem{
		Method:     "PUT",
		URL:        u,
		StatusCode: 204,
		Request: LogContent{
			Header: reqHeader,
			Body:   []byte(`{"A":"foo","B":7}` + "\n"),
		},
		Duration: time.Millisecond,
		Level:    WithHeadersAndBodies,
	})

	g.Expect(buf.String()).To(gomega.Equal(
		`PUT      http://somewhere.com/a/b/c 204 1ms
--> Content-Length: 18
--> Content-Type:   application/json; charset=UTF-8

{"A":"foo","B":7}

<-- no headers

---
`), buf.String())
}

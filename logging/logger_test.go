package logging

import (
	"bytes"
	"github.com/onsi/gomega"
	"net/http"
	"net/url"
	"testing"
	"time"
)

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
	log := LogWriter(buf, ".")
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
	log := LogWriter(buf, ".")
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
		Start:    time.Date(2021, 04, 01, 10, 11, 12, 0, time.UTC),
		Duration: time.Millisecond,
		Level:    WithHeadersAndBodies,
	})

	g.Expect(buf.String()).To(gomega.Equal(
		`GET      http://somewhere.com/a/b/c?foo=1 200 1ms
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
	log := LogWriter(buf, ".")
	log(&LogItem{
		Method:     "GET",
		URL:        u,
		StatusCode: 200,
		Request: LogContent{
			Header: reqHeader,
		},
		Response: LogContent{
			Header: resHeader,
			Body:   []byte(`{"alpha":"some text","beta":"some more text","gamma":"this might drag on","delta":"and on past the 80 char threshold"}` + "\n"),
		},
		Err:      nil,
		Start:    time.Date(2021, 04, 01, 10, 11, 12, 0, time.UTC),
		Duration: time.Millisecond,
		Level:    WithHeadersAndBodies,
	})

	g.Expect(buf.String()).To(gomega.Equal(
		`GET      http://somewhere.com/a/b/c?foo=1 200 1ms
--> Accept:          application/json
--> Cookie:          a=123
-->                  b=4556

<-- Content-Length:  18
<-- Content-Type:    application/json; charset=UTF-8
see ./2021-04-01_10-11-12_GET_a_b_c_res.json

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
	log := LogWriter(buf, ".")
	log(&LogItem{
		Method:     "GET",
		URL:        u,
		StatusCode: 200,
		Request: LogContent{
			Header: reqHeader,
		},
		Response: LogContent{
			Header: resHeader,
			Body: []byte(`<xml>
<alpha>some text</alpha>
<beta>some more text</beta>
<gamma>this might drag on past the 80 char threshold</gamma>
</xml>` + "\n"),
		},
		Err:      nil,
		Start:    time.Date(2021, 04, 01, 10, 11, 12, 0, time.UTC),
		Duration: time.Millisecond,
		Level:    WithHeadersAndBodies,
	})

	g.Expect(buf.String()).To(gomega.Equal(
		`GET      http://somewhere.com/a/b/c?foo=1 200 1ms
--> Accept:          application/xml
--> Cookie:          a=123
-->                  b=4556

<-- Content-Length:  18
<-- Content-Type:    application/xml; charset=UTF-8
see ./2021-04-01_10-11-12_GET_a_b_c_res.xml

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
	log := LogWriter(buf, ".")
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
--> Accept:          application/*

<-- Content-Length:  3
<-- Content-Type:    application/octet-stream
<-- binary content [3]byte

---
`), buf.String())
}

func TestLogWriter_typical_PUT(t *testing.T) {
	g := gomega.NewWithT(t)

	u, _ := url.Parse("http://somewhere.com/a/b/c")
	reqHeader := make(http.Header)
	reqHeader.Set("Content-Type", "application/json; charset=UTF-8")
	reqHeader.Set("Content-Length", "18")

	buf := &bytes.Buffer{}
	log := LogWriter(buf, ".")
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
--> Content-Length:  18
--> Content-Type:    application/json; charset=UTF-8

{"A":"foo","B":7}

<-- no headers

---
`), buf.String())
}

func TestUrlToFilename(t *testing.T) {
	g := gomega.NewWithT(t)

	cases := map[string]string{
		"":             "",
		"/":            "",
		"/aaa/bbb/ccc": "aaa_bbb_ccc",
		`/A!B"C#D$E%F&G'H(I)J*K+L,/a:b;c<d=e>f&g[h\i]j^k` + "`/A{B|C}D~": "A-B-C-D-E-F-G-H-I-J-K-L-_a-b-c-d-e-f-g-h-i-j-k-_A-B-C-D-",
	}
	for in, exp := range cases {
		act := urlToFilename(in)
		g.Expect(act).To(gomega.Equal(exp))
	}
}

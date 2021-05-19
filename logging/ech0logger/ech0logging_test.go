package ech0logger

import (
	"fmt"
	"github.com/onsi/gomega"
	"github.com/rickb777/ech0/v3/testlogger"
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

	lgr := testlogger.NewWithConsoleLogger()
	log := LogWriter(lgr.Timestamp(), afero.NewMemMapFs())
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

	g.Expect(lgr.Infos.Len()).To(gomega.Equal(1))
	//t.Log(lgr.Infos.First())
	g.Expect(lgr.Infos.First().Key).To(gomega.Equal("t"))
	g.Expect(lgr.Infos.First().FindByKey("t").Val.(time.Time).Equal(t0)).To(gomega.BeTrue())
	g.Expect(lgr.Infos.First().FindByKey("method").Val).To(gomega.Equal("GET"))
	g.Expect(lgr.Infos.First().FindByKey("url").Val).To(gomega.Equal(u))
	g.Expect(lgr.Infos.First().FindByKey("duration").Val).To(gomega.Equal(time.Millisecond))
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

	lgr := testlogger.NewWithConsoleLogger()
	log := LogWriter(lgr.Timestamp(), afero.NewMemMapFs())
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

	g.Expect(lgr.Infos.Len()).To(gomega.Equal(1))
	//t.Log(lgr.Infos.First())
	g.Expect(lgr.Infos.First().Key).To(gomega.Equal("t"))
	g.Expect(lgr.Infos.First().FindByKey("t").Val.(time.Time).Equal(t0)).To(gomega.BeTrue())
	g.Expect(lgr.Infos.First().FindByKey("method").Val).To(gomega.Equal("GET"))
	g.Expect(lgr.Infos.First().FindByKey("url").Val).To(gomega.Equal(u))
	g.Expect(lgr.Infos.First().FindByKey("duration").Val).To(gomega.Equal(time.Millisecond))
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

	lgr := testlogger.NewWithConsoleLogger()
	log := LogWriter(lgr.Timestamp(), afero.NewMemMapFs())
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

	g.Expect(lgr.Infos.Len()).To(gomega.Equal(1))
	//t.Log(lgr.Infos.First())
	g.Expect(lgr.Infos.First().Key).To(gomega.Equal("t"))
	g.Expect(lgr.Infos.First().FindByKey("t").Val.(time.Time).Equal(t0)).To(gomega.BeTrue())
	g.Expect(lgr.Infos.First().FindByKey("method").Val).To(gomega.Equal("GET"))
	g.Expect(lgr.Infos.First().FindByKey("url").Val).To(gomega.Equal(u))
	g.Expect(lgr.Infos.First().FindByKey("duration").Val).To(gomega.Equal(time.Millisecond))
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

	lgr := testlogger.NewWithConsoleLogger()
	log := LogWriter(lgr.Timestamp(), afero.NewMemMapFs())
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

	g.Expect(lgr.Infos.Len()).To(gomega.Equal(1))
	//t.Log(lgr.Infos.First())
	g.Expect(lgr.Infos.First().Key).To(gomega.Equal("t"))
	g.Expect(lgr.Infos.First().FindByKey("t").Val.(time.Time).Equal(t0)).To(gomega.BeTrue())
	g.Expect(lgr.Infos.First().FindByKey("method").Val).To(gomega.Equal("GET"))
	g.Expect(lgr.Infos.First().FindByKey("url").Val).To(gomega.Equal(u))
	g.Expect(lgr.Infos.First().FindByKey("duration").Val).To(gomega.Equal(time.Millisecond))
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

	lgr := testlogger.NewWithConsoleLogger()
	log := LogWriter(lgr.Timestamp(), afero.NewMemMapFs())
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

	g.Expect(lgr.Infos.Len()).To(gomega.Equal(1))
	//t.Log(lgr.Infos.First())
	g.Expect(lgr.Infos.First().Key).To(gomega.Equal("t"))
	g.Expect(lgr.Infos.First().FindByKey("t").Val.(time.Time).Equal(t0)).To(gomega.BeTrue())
	g.Expect(lgr.Infos.First().FindByKey("method").Val).To(gomega.Equal("GET"))
	g.Expect(lgr.Infos.First().FindByKey("url").Val).To(gomega.Equal(u))
	g.Expect(lgr.Infos.First().FindByKey("duration").Val).To(gomega.Equal(time.Millisecond))
}

func TestLogWriter_typical_GET_binary(t *testing.T) {
	g := gomega.NewWithT(t)

	u, _ := url.Parse("http://somewhere.com/a/b/c")
	reqHeader := make(http.Header)
	reqHeader.Set("Accept", "application/*")

	resHeader := make(http.Header)
	resHeader.Set("Content-Type", "application/octet-stream")
	resHeader.Set("Content-Length", "3")

	lgr := testlogger.NewWithConsoleLogger()
	log := LogWriter(lgr.Timestamp(), afero.NewMemMapFs())
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

	g.Expect(lgr.Infos.Len()).To(gomega.Equal(1))
	//t.Log(lgr.Infos.First())
	g.Expect(lgr.Infos.First().Key).To(gomega.Equal("t"))
	g.Expect(lgr.Infos.First().FindByKey("t").Val.(time.Time).Equal(t0)).To(gomega.BeTrue())
	g.Expect(lgr.Infos.First().FindByKey("method").Val).To(gomega.Equal("GET"))
	g.Expect(lgr.Infos.First().FindByKey("url").Val).To(gomega.Equal(u))
	g.Expect(lgr.Infos.First().FindByKey("duration").Val).To(gomega.Equal(time.Millisecond))
}

func TestLogWriter_typical_PUT_headers_only_with_error(t *testing.T) {
	g := gomega.NewWithT(t)

	u, _ := url.Parse("http://somewhere.com/a/b/c")
	reqHeader := make(http.Header)
	reqHeader.Set("Content-Type", "application/json; charset=UTF-8")
	reqHeader.Set("Content-Length", "18")

	lgr := testlogger.NewWithConsoleLogger()
	log := LogWriter(lgr.Timestamp(), afero.NewMemMapFs())
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

	g.Expect(lgr.Errors.Len()).To(gomega.Equal(1))
	//t.Log(lgr.Errors.First())
	g.Expect(lgr.Errors.First().Key).To(gomega.Equal("error"))
	g.Expect(lgr.Errors.First().FindByKey("error").Val).To(gomega.Equal([]byte("Bang!")))
	g.Expect(lgr.Errors.First().FindByKey("t").Val.(time.Time).Equal(t0)).To(gomega.BeTrue())
	g.Expect(lgr.Errors.First().FindByKey("method").Val).To(gomega.Equal("PUT"))
	g.Expect(lgr.Errors.First().FindByKey("url").Val).To(gomega.Equal(u))
	g.Expect(lgr.Errors.First().FindByKey("duration").Val).To(gomega.Equal(123 * time.Microsecond))
}

func TestLogWriter_typical_PUT_short_content(t *testing.T) {
	g := gomega.NewWithT(t)

	u, _ := url.Parse("http://somewhere.com/a/b/c")
	reqHeader := make(http.Header)
	reqHeader.Set("Content-Type", "application/json; charset=UTF-8")
	reqHeader.Set("Content-Length", "18")

	lgr := testlogger.NewWithConsoleLogger()
	log := LogWriter(lgr.Timestamp(), afero.NewMemMapFs())
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

	g.Expect(lgr.Infos.Len()).To(gomega.Equal(1))
	//t.Log(lgr.Infos.First())
	g.Expect(lgr.Infos.First().Key).To(gomega.Equal("t"))
	g.Expect(lgr.Infos.First().FindByKey("t").Val.(time.Time).Equal(t0)).To(gomega.BeTrue())
	g.Expect(lgr.Infos.First().FindByKey("method").Val).To(gomega.Equal("PUT"))
	g.Expect(lgr.Infos.First().FindByKey("url").Val).To(gomega.Equal(u))
	g.Expect(lgr.Infos.First().FindByKey("duration").Val).To(gomega.Equal(time.Millisecond))
}

func TestLogWriter_typical_PUT_long_content(t *testing.T) {
	g := gomega.NewWithT(t)

	u, _ := url.Parse("http://somewhere.com/a/b/c")
	reqHeader := make(http.Header)
	reqHeader.Set("Content-Type", "application/json; charset=UTF-8")
	reqHeader.Set("Content-Length", "18")

	lgr := testlogger.NewWithConsoleLogger()
	log := LogWriter(lgr.Timestamp(), afero.NewMemMapFs())
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

	g.Expect(lgr.Infos.Len()).To(gomega.Equal(1))
	//t.Log(lgr.Infos.First())
	g.Expect(lgr.Infos.First().Key).To(gomega.Equal("t"))
	g.Expect(lgr.Infos.First().FindByKey("t").Val.(time.Time).Equal(t0)).To(gomega.BeTrue())
	g.Expect(lgr.Infos.First().FindByKey("method").Val).To(gomega.Equal("PUT"))
	g.Expect(lgr.Infos.First().FindByKey("url").Val).To(gomega.Equal(u))
	g.Expect(lgr.Infos.First().FindByKey("duration").Val).To(gomega.Equal(time.Millisecond))
}

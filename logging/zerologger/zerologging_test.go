package zerologger

import (
	"fmt"
	"github.com/onsi/gomega"
	"github.com/rickb777/httpclient/body"
	"github.com/rickb777/httpclient/logging"
	"github.com/rs/zerolog"
	"github.com/spf13/afero"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"
)

func init() {
	zerolog.TimeFieldFormat = time.RFC3339
}

const longJSON = `{"alpha":"some text","beta":"some more text","gamma":"this might drag on","delta":"and on past the 80 char threshold"}` + "\n"

var t0 = time.Date(2021, 04, 01, 10, 11, 12, 1000000, time.UTC)

func TestLogWriter_typical_GET_terse(t *testing.T) {
	g := gomega.NewWithT(t)

	u, _ := url.Parse("http://somewhere.com/a/b/c")
	reqHeader := make(http.Header)
	reqHeader.Set("Host", "somewhere.com")
	reqHeader.Set("Accept", "application/json")
	reqHeader.Set("Cookie", "a=123")
	reqHeader.Add("Cookie", "b=4556")

	resHeader := make(http.Header)
	resHeader.Set("Content-Type", "application/json; charset=UTF-8")
	resHeader.Set("Content-Length", "18")

	lgrBuf := &strings.Builder{}
	lgr := zerolog.New(lgrBuf)
	log := LogWriter(lgr, afero.NewMemMapFs())
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

	msg := lgrBuf.String()
	g.Expect(msg).To(gomega.ContainSubstring(`"level":"info"`))
	g.Expect(msg).To(gomega.ContainSubstring(`"at":"20`))
	g.Expect(msg).To(gomega.ContainSubstring(`"status":200`))
	g.Expect(msg).To(gomega.ContainSubstring(`"method":"GET"`))
	g.Expect(msg).To(gomega.ContainSubstring(`"url":"http://somewhere.com/a/b/c"`))
	g.Expect(msg).To(gomega.ContainSubstring(`"duration":1`))
}

func TestLogWriter_typical_GET_JSON_short_content(t *testing.T) {
	g := gomega.NewWithT(t)

	u, _ := url.Parse("http://somewhere.com/a/b/c?foo=1")
	reqHeader := make(http.Header)
	reqHeader.Set("Host", "somewhere.com")
	reqHeader.Set("Accept", "application/json")
	reqHeader.Set("Cookie", "a=123")
	reqHeader.Add("Cookie", "b=4556")

	resHeader := make(http.Header)
	resHeader.Set("Content-Type", "application/json; charset=UTF-8")
	resHeader.Set("Content-Length", "18")

	lgrBuf := &strings.Builder{}
	lgr := zerolog.New(lgrBuf)
	log := LogWriter(lgr, afero.NewMemMapFs())
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

	msg := lgrBuf.String()
	g.Expect(msg).To(gomega.ContainSubstring(`"level":"info"`))
	g.Expect(msg).To(gomega.ContainSubstring(`"at":"20`))
	g.Expect(msg).To(gomega.ContainSubstring(`"status":200`))
	g.Expect(msg).To(gomega.ContainSubstring(`"method":"GET"`))
	g.Expect(msg).To(gomega.ContainSubstring(`"url":"http://somewhere.com/a/b/c?foo=1"`))
	g.Expect(msg).To(gomega.ContainSubstring(`"duration":1`))
	g.Expect(msg).NotTo(gomega.ContainSubstring(`"req_body":`))
	g.Expect(msg).To(gomega.ContainSubstring(`"resp_body":"{\"A\":\"foo\",\"B\":7}"`))
}

func TestLogWriter_typical_GET_JSON_long_content(t *testing.T) {
	g := gomega.NewWithT(t)

	u, _ := url.Parse("http://somewhere.com/a/b/c?foo=1")
	reqHeader := make(http.Header)
	reqHeader.Set("Host", "somewhere.com")
	reqHeader.Set("Accept", "application/json")
	reqHeader.Set("Cookie", "a=123")
	reqHeader.Add("Cookie", "b=4556")

	resHeader := make(http.Header)
	resHeader.Set("Content-Type", "application/json; charset=UTF-8")
	resHeader.Set("Content-Length", "18")

	lgrBuf := &strings.Builder{}
	lgr := zerolog.New(lgrBuf)
	log := LogWriter(lgr, afero.NewMemMapFs())
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

	msg := lgrBuf.String()
	g.Expect(msg).To(gomega.ContainSubstring(`"level":"info"`))
	g.Expect(msg).To(gomega.ContainSubstring(`"at":"20`))
	g.Expect(msg).To(gomega.ContainSubstring(`"status":200`))
	g.Expect(msg).To(gomega.ContainSubstring(`"method":"GET"`))
	g.Expect(msg).To(gomega.ContainSubstring(`"url":"http://somewhere.com/a/b/c?foo=1"`))
	g.Expect(msg).To(gomega.ContainSubstring(`"duration":1`))
	g.Expect(msg).NotTo(gomega.ContainSubstring(`"req_body":`))
	g.Expect(msg).NotTo(gomega.ContainSubstring(`"resp_body":`))
	g.Expect(msg).To(gomega.ContainSubstring(`"resp_file":"2021-04-01_10-11-12-001_GET_somewhere.com_a_b_c_resp.json"`))
	g.Expect(msg).To(gomega.ContainSubstring(`"resp_body_len":119`))
}

func TestLogWriter_typical_GET_text_long_content(t *testing.T) {
	g := gomega.NewWithT(t)

	u, _ := url.Parse("http://somewhere.com/a/b/c?foo=1")
	reqHeader := make(http.Header)
	reqHeader.Set("Host", "somewhere.com")
	reqHeader.Set("Accept", "text/*")
	reqHeader.Set("Cookie", "a=123")
	reqHeader.Add("Cookie", "b=4556")

	resHeader := make(http.Header)
	resHeader.Set("Content-Type", "text/plain; charset=UTF-8")
	resHeader.Set("Content-Length", "18")

	lgrBuf := &strings.Builder{}
	lgr := zerolog.New(lgrBuf)
	log := LogWriter(lgr, afero.NewMemMapFs())
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

	msg := lgrBuf.String()
	g.Expect(msg).To(gomega.ContainSubstring(`"level":"info"`))
	g.Expect(msg).To(gomega.ContainSubstring(`"at":"20`))
	g.Expect(msg).To(gomega.ContainSubstring(`"status":200`))
	g.Expect(msg).To(gomega.ContainSubstring(`"method":"GET"`))
	g.Expect(msg).To(gomega.ContainSubstring(`"url":"http://somewhere.com/a/b/c?foo=1"`))
	g.Expect(msg).To(gomega.ContainSubstring(`"duration":1`))
	g.Expect(msg).NotTo(gomega.ContainSubstring(`"resp_body":`))
	g.Expect(msg).To(gomega.ContainSubstring(`"resp_file":"2021-04-01_10-11-12-001_GET_somewhere.com_a_b_c_resp.txt"`))
	g.Expect(msg).To(gomega.ContainSubstring(`"resp_body_len":167`))
}

func TestLogWriter_typical_GET_XML_long_content(t *testing.T) {
	g := gomega.NewWithT(t)

	u, _ := url.Parse("http://somewhere.com/a/b/c?foo=1")
	reqHeader := make(http.Header)
	reqHeader.Set("Host", "somewhere.com")
	reqHeader.Set("Accept", "application/xml")
	reqHeader.Set("Cookie", "a=123")
	reqHeader.Add("Cookie", "b=4556")

	resHeader := make(http.Header)
	resHeader.Set("Content-Type", "application/xml; charset=UTF-8")
	resHeader.Set("Content-Length", "18")

	lgrBuf := &strings.Builder{}
	lgr := zerolog.New(lgrBuf)
	log := LogWriter(lgr, afero.NewMemMapFs())
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

	msg := lgrBuf.String()
	g.Expect(msg).To(gomega.ContainSubstring(`"level":"info"`))
	g.Expect(msg).To(gomega.ContainSubstring(`"at":"20`))
	g.Expect(msg).To(gomega.ContainSubstring(`"status":200`))
	g.Expect(msg).To(gomega.ContainSubstring(`"method":"GET"`))
	g.Expect(msg).To(gomega.ContainSubstring(`"url":"http://somewhere.com/a/b/c?foo=1"`))
	g.Expect(msg).To(gomega.ContainSubstring(`"duration":1`))
	g.Expect(msg).NotTo(gomega.ContainSubstring(`"resp_body":`))
	g.Expect(msg).To(gomega.ContainSubstring(`"resp_file":"2021-04-01_10-11-12-001_GET_somewhere.com_a_b_c_resp.xml"`))
	g.Expect(msg).To(gomega.ContainSubstring(`"resp_body_len":127`))
}

func TestLogWriter_typical_GET_binary(t *testing.T) {
	g := gomega.NewWithT(t)

	u, _ := url.Parse("http://somewhere.com/a/b/c")
	reqHeader := make(http.Header)
	reqHeader.Set("Host", "somewhere.com")
	reqHeader.Set("Accept", "application/*")

	resHeader := make(http.Header)
	resHeader.Set("Content-Type", "application/octet-stream")
	resHeader.Set("Content-Length", "3")

	lgrBuf := &strings.Builder{}
	lgr := zerolog.New(lgrBuf)
	log := LogWriter(lgr, afero.NewMemMapFs())
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

	msg := lgrBuf.String()
	g.Expect(msg).To(gomega.ContainSubstring(`"level":"info"`))
	g.Expect(msg).To(gomega.ContainSubstring(`"at":"20`))
	g.Expect(msg).To(gomega.ContainSubstring(`"status":200`))
	g.Expect(msg).To(gomega.ContainSubstring(`"method":"GET"`))
	g.Expect(msg).To(gomega.ContainSubstring(`"url":"http://somewhere.com/a/b/c"`))
	g.Expect(msg).To(gomega.ContainSubstring(`"duration":1`))
	g.Expect(msg).NotTo(gomega.ContainSubstring(`"resp_body":`))
	g.Expect(msg).To(gomega.ContainSubstring(`"resp_body_len":3`))
}

func TestLogWriter_typical_PUT_headers_only_with_error(t *testing.T) {
	g := gomega.NewWithT(t)

	u, _ := url.Parse("http://somewhere.com/a/b/c")
	reqHeader := make(http.Header)
	reqHeader.Set("Host", "somewhere.com")
	reqHeader.Set("Content-Type", "application/json; charset=UTF-8")
	reqHeader.Set("Content-Length", "18")

	lgrBuf := &strings.Builder{}
	lgr := zerolog.New(lgrBuf)
	log := LogWriter(lgr, afero.NewMemMapFs())
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

	msg := lgrBuf.String()
	g.Expect(msg).To(gomega.ContainSubstring(`"level":"error"`))
	g.Expect(msg).To(gomega.ContainSubstring(`"error":"Bang!"`))
	g.Expect(msg).To(gomega.ContainSubstring(`"at":"20`))
	g.Expect(msg).To(gomega.ContainSubstring(`"status":0`))
	g.Expect(msg).To(gomega.ContainSubstring(`"method":"PUT"`))
	g.Expect(msg).To(gomega.ContainSubstring(`"url":"http://somewhere.com/a/b/c"`))
	g.Expect(msg).To(gomega.ContainSubstring(`"duration":0.123`))
}

func TestLogWriter_typical_PUT_short_content(t *testing.T) {
	g := gomega.NewWithT(t)

	u, _ := url.Parse("http://somewhere.com/a/b/c")
	reqHeader := make(http.Header)
	reqHeader.Set("Host", "somewhere.com")
	reqHeader.Set("Content-Type", "application/json; charset=UTF-8")
	reqHeader.Set("Content-Length", "18")

	lgrBuf := &strings.Builder{}
	lgr := zerolog.New(lgrBuf)
	log := LogWriter(lgr, afero.NewMemMapFs())
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

	msg := lgrBuf.String()
	g.Expect(msg).To(gomega.ContainSubstring(`"level":"info"`))
	g.Expect(msg).To(gomega.ContainSubstring(`"at":"20`))
	g.Expect(msg).To(gomega.ContainSubstring(`"status":204`))
	g.Expect(msg).To(gomega.ContainSubstring(`"method":"PUT"`))
	g.Expect(msg).To(gomega.ContainSubstring(`"url":"http://somewhere.com/a/b/c"`))
	g.Expect(msg).To(gomega.ContainSubstring(`"duration":1`))
	g.Expect(msg).To(gomega.ContainSubstring(`"req_body":"{\"A\":\"foo\",\"B\":7}"`))
	g.Expect(msg).NotTo(gomega.ContainSubstring(`"resp_body":`))
}

func TestLogWriter_typical_PUT_long_content(t *testing.T) {
	g := gomega.NewWithT(t)

	u, _ := url.Parse("http://somewhere.com/a/b/c")
	reqHeader := make(http.Header)
	reqHeader.Set("Host", "somewhere.com")
	reqHeader.Set("Content-Type", "application/json; charset=UTF-8")
	reqHeader.Set("Content-Length", "18")

	lgrBuf := &strings.Builder{}
	lgr := zerolog.New(lgrBuf)
	log := LogWriter(lgr, afero.NewMemMapFs())
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

	msg := lgrBuf.String()
	g.Expect(msg).To(gomega.ContainSubstring(`"level":"info"`))
	g.Expect(msg).To(gomega.ContainSubstring(`"at":"20`))
	g.Expect(msg).To(gomega.ContainSubstring(`"status":204`))
	g.Expect(msg).To(gomega.ContainSubstring(`"method":"PUT"`))
	g.Expect(msg).To(gomega.ContainSubstring(`"url":"http://somewhere.com/a/b/c"`))
	g.Expect(msg).To(gomega.ContainSubstring(`"duration":1`))
}

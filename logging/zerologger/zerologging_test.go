package zerologger

import (
	"fmt"
	"github.com/rickb777/expect"
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
	expect.String(msg).ToContain(t, `"level":"info"`)
	expect.String(msg).ToContain(t, `"at":"20`)
	expect.String(msg).ToContain(t, `"status":200`)
	expect.String(msg).ToContain(t, `"method":"GET"`)
	expect.String(msg).ToContain(t, `"url":"http://somewhere.com/a/b/c"`)
	expect.String(msg).ToContain(t, `"duration":1`)
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
	expect.String(msg).ToContain(t, `"level":"info"`)
	expect.String(msg).ToContain(t, `"at":"20`)
	expect.String(msg).ToContain(t, `"status":200`)
	expect.String(msg).ToContain(t, `"method":"GET"`)
	expect.String(msg).ToContain(t, `"url":"http://somewhere.com/a/b/c?foo=1"`)
	expect.String(msg).ToContain(t, `"duration":1`)
	expect.String(msg).Not().ToContain(t, `"req_body":`)
	expect.String(msg).ToContain(t, `"resp_body":"{\"A\":\"foo\",\"B\":7}"`)
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
	expect.String(msg).ToContain(t, `"level":"info"`)
	expect.String(msg).ToContain(t, `"at":"20`)
	expect.String(msg).ToContain(t, `"status":200`)
	expect.String(msg).ToContain(t, `"method":"GET"`)
	expect.String(msg).ToContain(t, `"url":"http://somewhere.com/a/b/c?foo=1"`)
	expect.String(msg).ToContain(t, `"duration":1`)
	expect.String(msg).Not().ToContain(t, `"req_body":`)
	expect.String(msg).Not().ToContain(t, `"resp_body":`)
	expect.String(msg).ToContain(t, `"resp_file":"2021-04-01_10-11-12-001_GET_somewhere.com_a_b_c_resp.json"`)
	expect.String(msg).ToContain(t, `"resp_body_len":119`)
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
	expect.String(msg).ToContain(t, `"level":"info"`)
	expect.String(msg).ToContain(t, `"at":"20`)
	expect.String(msg).ToContain(t, `"status":200`)
	expect.String(msg).ToContain(t, `"method":"GET"`)
	expect.String(msg).ToContain(t, `"url":"http://somewhere.com/a/b/c?foo=1"`)
	expect.String(msg).ToContain(t, `"duration":1`)
	expect.String(msg).Not().ToContain(t, `"resp_body":`)
	expect.String(msg).ToContain(t, `"resp_file":"2021-04-01_10-11-12-001_GET_somewhere.com_a_b_c_resp.txt"`)
	expect.String(msg).ToContain(t, `"resp_body_len":167`)
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
	expect.String(msg).ToContain(t, `"level":"info"`)
	expect.String(msg).ToContain(t, `"at":"20`)
	expect.String(msg).ToContain(t, `"status":200`)
	expect.String(msg).ToContain(t, `"method":"GET"`)
	expect.String(msg).ToContain(t, `"url":"http://somewhere.com/a/b/c?foo=1"`)
	expect.String(msg).ToContain(t, `"duration":1`)
	expect.String(msg).Not().ToContain(t, `"resp_body":`)
	expect.String(msg).ToContain(t, `"resp_file":"2021-04-01_10-11-12-001_GET_somewhere.com_a_b_c_resp.xml"`)
	expect.String(msg).ToContain(t, `"resp_body_len":127`)
}

func TestLogWriter_typical_GET_binary(t *testing.T) {
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
	expect.String(msg).ToContain(t, `"level":"info"`)
	expect.String(msg).ToContain(t, `"at":"20`)
	expect.String(msg).ToContain(t, `"status":200`)
	expect.String(msg).ToContain(t, `"method":"GET"`)
	expect.String(msg).ToContain(t, `"url":"http://somewhere.com/a/b/c"`)
	expect.String(msg).ToContain(t, `"duration":1`)
	expect.String(msg).Not().ToContain(t, `"resp_body":`)
	expect.String(msg).ToContain(t, `"resp_body_len":3`)
}

func TestLogWriter_typical_PUT_headers_only_with_error(t *testing.T) {
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
	expect.String(msg).ToContain(t, `"level":"error"`)
	expect.String(msg).ToContain(t, `"error":"Bang!"`)
	expect.String(msg).ToContain(t, `"at":"20`)
	expect.String(msg).ToContain(t, `"status":0`)
	expect.String(msg).ToContain(t, `"method":"PUT"`)
	expect.String(msg).ToContain(t, `"url":"http://somewhere.com/a/b/c"`)
	expect.String(msg).ToContain(t, `"duration":0.123`)
}

func TestLogWriter_typical_PUT_short_content(t *testing.T) {
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
	expect.String(msg).ToContain(t, `"level":"info"`)
	expect.String(msg).ToContain(t, `"at":"20`)
	expect.String(msg).ToContain(t, `"status":204`)
	expect.String(msg).ToContain(t, `"method":"PUT"`)
	expect.String(msg).ToContain(t, `"url":"http://somewhere.com/a/b/c"`)
	expect.String(msg).ToContain(t, `"duration":1`)
	expect.String(msg).ToContain(t, `"req_body":"{\"A\":\"foo\",\"B\":7}"`)
	expect.String(msg).Not().ToContain(t, `"resp_body":`)
}

func TestLogWriter_typical_PUT_long_content(t *testing.T) {
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
	expect.String(msg).ToContain(t, `"level":"info"`)
	expect.String(msg).ToContain(t, `"at":"20`)
	expect.String(msg).ToContain(t, `"status":204`)
	expect.String(msg).ToContain(t, `"method":"PUT"`)
	expect.String(msg).ToContain(t, `"url":"http://somewhere.com/a/b/c"`)
	expect.String(msg).ToContain(t, `"duration":1`)
}

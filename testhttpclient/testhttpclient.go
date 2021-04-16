package testhttpclient

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/onsi/gomega"
)

// TestingT is a simple *testing.T interface wrapper
type TestingT interface {
	Fatalf(format string, args ...interface{})
	Helper()
}

// MockResponse builds a http.Response.
func MockResponse(code int, body []byte, contentType string) *http.Response {
	res := &http.Response{
		StatusCode: code,
		Header:     make(http.Header),
		Body:       io.NopCloser(bytes.NewBuffer(body)),
	}

	res.Header.Set("Content-Length", strconv.Itoa(len(body)))
	res.Header.Set("Content-Type", contentType)
	return res
}

// Outcome defines a matching rule for an expected HTTP request outcome.
type Outcome struct {
	Response *http.Response
	Err      error
}

// MockHttpClient is a HttpClient that holds some stubbed outcomes.
type MockHttpClient struct {
	t                TestingT
	CapturedRequests []*http.Request
	capturedBodies   []string
	outcomes         map[string][]Outcome
}

func New(t TestingT) *MockHttpClient {
	return &MockHttpClient{t: t, outcomes: make(map[string][]Outcome)}
}

func (m *MockHttpClient) Reset() {
	m.CapturedRequests = nil
	m.outcomes = make(map[string][]Outcome)
}

func (m *MockHttpClient) CapturedBody(i int) string {
	if len(m.capturedBodies) == 0 {
		for _, req := range m.CapturedRequests {
			m.capturedBodies = append(m.capturedBodies, ReadString(req.Body))
		}
	}
	return m.capturedBodies[i]
}

func (m *MockHttpClient) RemainingOutcomes() (n int) {
	for _, o := range m.outcomes {
		n += len(o)
	}
	return n
}

// AddLiteralResponse adds an expected outcome that has a literal HTTP response as provided.
// An example might be
//
//    HTTP/1.1 200 OK
//    Content-Type: application/json; charset=UTF-8
//    Content-Length: 18
//
//    {"A":"foo","B":7}
//
func (m *MockHttpClient) AddLiteralResponse(method, url string, wholeResponse string) *MockHttpClient {
	return m.AddLiteralByteResponse(method, url, []byte(wholeResponse))
}

func (m *MockHttpClient) AddLiteralByteResponse(method, url string, wholeResponse []byte) *MockHttpClient {
	g := gomega.NewWithT(m.t)
	rdr := bufio.NewReader(bytes.NewBuffer(wholeResponse))
	res, err := http.ReadResponse(rdr, nil)
	g.Expect(err).NotTo(gomega.HaveOccurred())

	return m.AddResponse(method, url, res)
}

// AddResponse adds an expected outcome with a response or an error. If the error is not nil, the response will
// be ignored (so it should be nil).
func (m *MockHttpClient) AddResponse(method, url string, response *http.Response) *MockHttpClient {
	return m.AddOutcome(method, url, Outcome{Response: response})
}

func (m *MockHttpClient) AddError(method, url string, err error) *MockHttpClient {
	return m.AddOutcome(method, url, Outcome{Err: err})
}

// AddOutcome adds an outcome directly.
func (m *MockHttpClient) AddOutcome(method, url string, outcome Outcome) *MockHttpClient {
	match := fmt.Sprintf("%s %s", method, url)
	m.outcomes[match] = append(m.outcomes[match], outcome)
	return m
}

func (m *MockHttpClient) Do(req *http.Request) (*http.Response, error) {
	return m.RoundTrip(req)
}

func (m *MockHttpClient) RoundTrip(req *http.Request) (*http.Response, error) {
	m.CapturedRequests = append(m.CapturedRequests, req.Clone(context.Background()))

	match := fmt.Sprintf("%s %s", req.Method, req.URL)

	outcomes := m.outcomes[match]
	if len(outcomes) == 0 {
		var keys []string
		for k, v := range m.outcomes {
			if len(v) > 0 {
				keys = append(keys, fmt.Sprintf("%2d %s", len(v), k))
			}
		}
		m.t.Fatalf("Missing outcome for %s\nRemaining:\n%s", match, strings.Join(keys, "\n"))
	}

	o := outcomes[0]
	m.outcomes[match] = m.outcomes[match][1:]

	if o.Err != nil {
		return nil, o.Err
	}

	o.Response.Request = req
	return o.Response, nil
}

// TODO possibly remove this - see rest.ReadString

func ReadString(r io.Reader) string {
	buf := &bytes.Buffer{}
	buf.ReadFrom(r)
	return buf.String()
}

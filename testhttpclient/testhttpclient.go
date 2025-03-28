// Package testhttpclient provides a tool for testing code that uses HTTP client(s).
package testhttpclient

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/rickb777/expect"
	bodypkg "github.com/rickb777/httpclient/body"
)

// TestingT is a simple *testing.T interface wrapper
type TestingT interface {
	Fatalf(format string, args ...interface{})
	Helper()
}

const (
	ContentTypeApplicationJSON = "application/json; charset=UTF-8"
	ContentTypeApplicationXML  = "application/xml; charset=UTF-8"
)

// MockJSONResponse builds a JSON http.Response using a literal JSON string or a data struct.
func MockJSONResponse(code int, body interface{}) *http.Response {
	if b, ok := body.(string); ok {
		return MockResponse(code, []byte(b), ContentTypeApplicationJSON)
	}

	s := &bytes.Buffer{}
	must(json.NewEncoder(s).Encode(body))
	return MockResponse(code, s.Bytes(), ContentTypeApplicationJSON)
}

// MockXMLResponse builds an XML http.Response using a literal XML string or a data struct.
func MockXMLResponse(code int, body interface{}) *http.Response {
	if b, ok := body.(string); ok {
		return MockResponse(code, []byte(b), ContentTypeApplicationXML)
	}

	s := &bytes.Buffer{}
	must(xml.NewEncoder(s).Encode(body))
	return MockResponse(code, s.Bytes(), ContentTypeApplicationXML)
}

// MockResponse builds a http.Response. If contentType blank it is ignored.
func MockResponse(code int, body []byte, contentType string) *http.Response {
	body = withTrailingNewline(body)
	res := &http.Response{
		StatusCode: code,
		Header:     make(http.Header),
		Body:       bodypkg.NewBody(body),
	}

	res.Header.Set("Content-Length", strconv.Itoa(len(body)))
	if contentType != "" {
		res.Header.Set("Content-Type", contentType)
	}
	return res
}

//-------------------------------------------------------------------------------------------------

// Outcome defines a matching rule for an expected HTTP request outcome.
type Outcome struct {
	Response *http.Response
	Err      error
}

//-------------------------------------------------------------------------------------------------

// MockHttpClient is a HttpClient that holds some stubbed outcomes.
type MockHttpClient struct {
	t                expect.Tester
	CapturedRequests []*http.Request
	capturedBodies   []*bodypkg.Body
	outcomes         map[string][]Outcome
}

func New(t expect.Tester) *MockHttpClient {
	return &MockHttpClient{t: t, outcomes: make(map[string][]Outcome)}
}

// Reset deletes all outcomes and captured responses.
func (m *MockHttpClient) Reset() {
	m.CapturedRequests = nil
	m.capturedBodies = nil
	m.outcomes = make(map[string][]Outcome)
}

// CapturedBody gets the request body from the i'th request.
func (m *MockHttpClient) CapturedBody(i int) *bodypkg.Body {
	n := len(m.capturedBodies)
	if n < len(m.CapturedRequests) {
		for _, req := range m.CapturedRequests[n:] {
			body, err := bodypkg.Copy(req.Body)
			must(err)
			m.capturedBodies = append(m.capturedBodies, body)
		}
	}
	return m.capturedBodies[i]
}

// RemainingOutcomes describes the remaining outcomes. Typically, this should be empty
// at the end of a test (otherwise there might be a setup error).
func (m *MockHttpClient) RemainingOutcomes() []string {
	if len(m.outcomes) == 0 {
		return nil
	}

	info := make([]string, 0, len(m.outcomes))
	for u, o := range m.outcomes {
		if len(o) > 0 {
			info = append(info, fmt.Sprintf("%2d: %s", len(o), u))
		}
	}
	return info
}

// AddLiteralResponse adds an expected outcome that has a literal HTTP response as provided.
// An example might be
//
//	HTTP/1.1 200 OK
//	Content-Type: application/json; charset=UTF-8
//	Content-Length: 18
//
//	{"A":"foo","B":7}
func (m *MockHttpClient) AddLiteralResponse(method, url string, wholeResponse string) *MockHttpClient {
	return m.AddLiteralByteResponse(method, url, []byte(wholeResponse))
}

func (m *MockHttpClient) AddLiteralByteResponse(method, url string, wholeResponse []byte) *MockHttpClient {
	rdr := bufio.NewReader(bytes.NewBuffer(withTrailingNewline(wholeResponse)))
	res, err := http.ReadResponse(rdr, nil)
	expect.Error(err).Not().ToHaveOccurred(m.t)

	return m.AddResponse(method, url, res)
}

// AddResponse adds an expected outcome that returns a response.
func (m *MockHttpClient) AddResponse(method, url string, response *http.Response) *MockHttpClient {
	return m.AddOutcome(method, url, Outcome{Response: response})
}

// AddError adds an expected outcome that returns an error instead of a response.
func (m *MockHttpClient) AddError(method, url string, err error) *MockHttpClient {
	return m.AddOutcome(method, url, Outcome{Err: err})
}

// AddOutcome adds an outcome directly.
func (m *MockHttpClient) AddOutcome(method, url string, outcome Outcome) *MockHttpClient {
	match := fmt.Sprintf("%s %s", method, url)
	m.outcomes[match] = append(m.outcomes[match], outcome)
	return m
}

//-------------------------------------------------------------------------------------------------

// Do is a pluggable method that implements standard library behaviour using stubbed behaviours.
// See httpclient.HttpClient. This uses RoundTrip.
func (m *MockHttpClient) Do(req *http.Request) (*http.Response, error) {
	return m.RoundTrip(req)
}

// RoundTrip is a pluggable method that implements standard library http.RoundTripper behaviour
// using stubbed behaviours.
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

func withTrailingNewline(body []byte) []byte {
	if len(body) > 0 && body[len(body)-1] != '\n' {
		body = append(body, '\n')
	}
	return body
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}

package rest

import (
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/rickb777/acceptable/contenttype"
	"github.com/rickb777/expect"
	"github.com/rickb777/httpclient/internal/mytesting"
)

type data struct {
	A string
	B int
}

func Test_all_methods_various_inputs_without_response_204(t *testing.T) {
	helloWorld := "hello world"
	uv := url.Values{"a": []string{"1", "2"}}

	cases := map[string]any{
		"struct 21 application/json":               &data{A: "hello", B: 10},
		"string 11 application/json":               helloWorld,
		"*string 11 application/json":              &helloWorld,
		"form 7 application/x-www-form-urlencoded": uv,
		"bytes 11 ": []byte(helloWorld),
		"reader  ":  strings.NewReader(helloWorld),
	}

	for _, method := range []string{"GET", "HEAD", "PUT", "POST", "PUT", "PATCH", "DELETE"} {
		for tag, input := range cases {
			testClient := mytesting.StubHttpWithBody("HTTP/1.1 204 No Content\n\n")

			b, h1 := processRequestEntity(input)
			req := RESTRequest{req: httptest.NewRequest(method, "http://example.test/foo/bar", b)}
			mergeHeaders(req.req.Header, h1)
			res := req.HTTPRoundTrip(method, testClient)
			h2, code, err := res.Status()

			p := strings.Split(tag, " ")
			expect.Error(err).Not().ToHaveOccurred(t)
			expect.Any(code).ToBe(t, 204)
			expect.Map(h2).ToHaveLength(t, 0)
			expect.Any(testClient.Captured.Method).ToBe(t, method)
			expect.Any(testClient.Captured.Header.Get("Content-Length")).I(tag).ToBe(t, p[1])
			expect.Any(testClient.Captured.Header.Get("Content-Type")).I(tag).ToBe(t, p[2])
			expect.Any(testClient.Captured.Header.Get("Accept")).I(tag).ToBe(t, contenttype.ApplicationJSON)
		}
	}
}

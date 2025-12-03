package rest

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/rickb777/acceptable/contenttype"
	"github.com/rickb777/acceptable/header"
	"github.com/rickb777/acceptable/headername"
	"github.com/rickb777/expect"
	bodypkg "github.com/rickb777/httpclient/body"
	"github.com/rickb777/httpclient/internal/mytesting"
)

type data struct {
	A string
	B int
}

func Test_all_methods_various_inputs_without_response_204(t *testing.T) {
	helloWorld := "hello world"

	cases := map[string]func() any{
		"struct 21 application/json":               func() any { return &data{A: "hello", B: 10} },
		"string 11 application/json":               func() any { return helloWorld },
		"*string 11 application/json":              func() any { return &helloWorld },
		"form 7 application/x-www-form-urlencoded": func() any { return url.Values{"a": []string{"1", "2"}} },
		"bytes 11 ":  func() any { return []byte(helloWorld) },
		"reader 11 ": func() any { return strings.NewReader(helloWorld) },
	}

	for _, method := range []string{"GET", "HEAD", "PUT", "POST", "PATCH", "DELETE"} {
		for tag, input := range cases {
			testClient := mytesting.StubHttpWithBody("HTTP/1.1 204 No Content\n\n")

			cl := NewClient("http://example.test/foo", SetHttpClient(testClient))
			res, err := cl.Request(context.Background(), method, "/bar", input())
			if res != nil {
				res.Body.Close()
			}

			p := strings.Split(tag, " ")
			expect.Error(err).I("%s %q", method, tag).Not().ToHaveOccurred(t)
			expect.Number(res.StatusCode).I("%s %q", method, tag).ToBe(t, 204)
			expect.Map(res.Header).I("%s %q", method, tag).ToHaveLength(t, 0)
			expect.String(testClient.Captured.Method).I("%s %q", method, tag).ToBe(t, method)
			expect.String(testClient.Captured.Header.Get("Content-Length")).I("%s %q", method, tag).ToBe(t, p[1])
			expect.String(testClient.Captured.Header.Get("Content-Type")).I("%s %q", method, tag).ToBe(t, p[2])
			expect.String(testClient.Captured.Header.Get("Accept")).I("%s %q", method, tag).ToBe(t, contenttype.ApplicationJSON)
		}
	}
}

func Test_3xx_4xx_5xx(t *testing.T) {
	cases := []struct {
		contentType string
		headers     map[string]string
		msg         string
		statusCode  int
		transient   bool
	}{
		{headers: map[string]string{"Location": "/other"}, msg: `301: GET http://x.te/ moved permanently /other`, statusCode: http.StatusMovedPermanently},
		{headers: map[string]string{"Location": "/other"}, msg: `302: GET http://x.te/ found /other`, statusCode: http.StatusFound},
		{headers: map[string]string{"Location": "/other"}, msg: `303: GET http://x.te/ see other /other`, statusCode: http.StatusSeeOther},
		{headers: map[string]string{}, msg: `400: GET http://x.te/`, statusCode: http.StatusBadRequest},
		{headers: map[string]string{"Content-Type": "image/png"}, msg: `404: GET http://x.te/ image/png`, statusCode: http.StatusNotFound},
		{headers: map[string]string{"Content-Type": "text/plain"}, msg: `404: GET http://x.te/ text/plain that was bad`, statusCode: http.StatusNotFound},
		{headers: map[string]string{"Content-Type": "text/plain"}, msg: `500: GET http://x.te/ text/plain that was bad`, statusCode: http.StatusInternalServerError, transient: true},
		{headers: map[string]string{"Content-Type": "text/plain"}, msg: `503: GET http://x.te/ text/plain that was bad`, statusCode: http.StatusServiceUnavailable, transient: true},
	}

	for _, c := range cases {
		resp := fmt.Sprintf("HTTP/1.1 %d %s\n", c.statusCode, http.StatusText(c.statusCode))
		for k, v := range c.headers {
			resp += fmt.Sprintf("%s: %s\n", k, v)
		}
		resp += `Content-Length: 14

that was bad
`
		testClient := mytesting.StubHttpWithBody(resp)

		cl := NewClient("http://x.te", SetHttpClient(testClient))
		res, err := cl.Get(context.Background(), "/")

		expect.Error(err).I(c.statusCode).ToContain(t, c.msg)
		expect.Bool(err.(*RestError).IsTransient()).I(c.statusCode).ToBe(t, c.transient)
		expect.Bool(err.(*RestError).IsPermanent()).I(c.statusCode).ToBe(t, !c.transient)
		expect.Number(res.StatusCode).I(c.statusCode).ToBe(t, c.statusCode)
		expect.Number(len(res.Header)).I(c.statusCode).ToBeGreaterThanOrEqual(t, 1)
		expect.String(testClient.Captured.Method).I(c.statusCode).ToBe(t, "GET")
		expect.String(res.Body.String()).I(c.statusCode).ToBe(t, `that was bad`+"\n")
	}
}

func TestHead(t *testing.T) {
	testClient := mytesting.StubHttpWithBody(`HTTP/1.1 200 OK
Content-Type: application/json
Content-Length: 21

`)
	cl := NewClient("http://example.test/foo", SetHttpClient(testClient))

	res, err := cl.Head(context.Background(), "/bar")

	expect.Error(err).Not().ToHaveOccurred(t)
	expect.Number(res.StatusCode).ToBe(t, 200)
	expect.Map(res.Header).ToHaveLength(t, 1)
	expect.String(res.Header.Get(headername.ContentLength)).ToBe(t, "21")
	expect.String(res.Type.String()).ToBe(t, contenttype.ApplicationJSON)
	expect.String(res.Body.String()).ToBe(t, "")
	expect.String(testClient.Captured.URL.String()).ToBe(t, "http://example.test/foo/bar")
	expect.String(testClient.Captured.Method).ToBe(t, http.MethodHead)
	expect.String(testClient.Captured.Header.Get("Accept")).ToBe(t, contenttype.ApplicationJSON)
}

func TestGet(t *testing.T) {
	testClient := mytesting.StubHttpWithBody(`HTTP/1.1 200 OK
Content-Type: application/json
Content-Length: 21

{"A":"hello","B":10}
`)
	cl := NewClient("http://example.test/foo", SetHttpClient(testClient))

	res, err := cl.Get(context.Background(), "/bar")

	expect.Error(err).Not().ToHaveOccurred(t)
	expect.Number(res.StatusCode).ToBe(t, 200)
	expect.Map(res.Header).ToHaveLength(t, 1)
	expect.String(res.Header.Get(headername.ContentLength)).ToBe(t, "21")
	expect.String(res.Type.String()).ToBe(t, contenttype.ApplicationJSON)
	expect.String(res.Body.String()).ToBe(t, "{\"A\":\"hello\",\"B\":10}\n")
	expect.String(testClient.Captured.URL.String()).ToBe(t, "http://example.test/foo/bar")
	expect.String(testClient.Captured.Method).ToBe(t, http.MethodGet)
	expect.String(testClient.Captured.Header.Get("Accept")).ToBe(t, contenttype.ApplicationJSON)
}

func TestPost(t *testing.T) {
	testClient := mytesting.StubHttpWithBody(`HTTP/1.1 200 OK
Content-Type: application/json
Content-Length: 21

{"A":"hello","B":10}
`)
	cl := NewClient("http://example.test/foo", SetHttpClient(testClient))
	body := bodypkg.NewBodyString("hello world")

	res, err := cl.Post(context.Background(), "/bar", body)

	expect.Error(err).Not().ToHaveOccurred(t)
	expect.Number(res.StatusCode).ToBe(t, 200)
	expect.Map(res.Header).ToHaveLength(t, 1)
	expect.String(res.Header.Get(headername.ContentLength)).ToBe(t, "21")
	expect.String(res.Type.String()).ToBe(t, contenttype.ApplicationJSON)
	expect.String(res.Body.String()).ToBe(t, "{\"A\":\"hello\",\"B\":10}\n")
	expect.String(testClient.Captured.URL.String()).ToBe(t, "http://example.test/foo/bar")
	expect.String(testClient.Captured.Method).ToBe(t, http.MethodPost)
	expect.String(testClient.Captured.Header.Get("Accept")).ToBe(t, contenttype.ApplicationJSON)
}

func TestPut(t *testing.T) {
	testClient := mytesting.StubHttpWithBody(`HTTP/1.1 200 OK
Content-Type: application/json
Content-Length: 0

`)
	cl := NewClient("http://example.test/foo", SetHttpClient(testClient))
	body := bodypkg.NewBodyString("hello world")

	res, err := cl.Put(context.Background(), "/bar", body)

	expect.Error(err).Not().ToHaveOccurred(t)
	expect.Number(res.StatusCode).ToBe(t, 200)
	expect.Map(res.Header).ToHaveLength(t, 1)
	expect.String(res.Header.Get(headername.ContentLength)).ToBe(t, "0")
	expect.String(res.Type.String()).ToBe(t, contenttype.ApplicationJSON)
	expect.String(res.Body.String()).ToBe(t, "")
	expect.String(testClient.Captured.URL.String()).ToBe(t, "http://example.test/foo/bar")
	expect.String(testClient.Captured.Method).ToBe(t, http.MethodPut)
	expect.String(testClient.Captured.Header.Get("Accept")).ToBe(t, contenttype.ApplicationJSON)
}

func TestRequestOpts(t *testing.T) {
	testClient := mytesting.StubHttpWithBody(`HTTP/1.1 200 OK
Content-Type: application/json
Content-Length: 21

{"A":"hello","B":10}
`)
	cl := NewClient("http://example.test/foo", SetHttpClient(testClient))

	o1 := Query("a", "orses", "b", "for mutton")
	o3 := Query("[c]", `/\/`)
	h1 := Headers("X-Extra", "Foo")
	c2 := IfModifiedSince(time.Date(2010, 10, 10, 10, 10, 10, 0, time.UTC))
	c1 := IfNoneMatch(header.ETag{Hash: "abc123"})
	c3 := IfMatch(header.ETag{Hash: "xyz"}) // in reality, would not be used for GET requests

	res, err := cl.Get(context.Background(), "/bar", h1, o1, o3, c1, c2, c3)

	expect.Error(err).Not().ToHaveOccurred(t)
	expect.Number(res.StatusCode).ToBe(t, 200)
	expect.Map(res.Header).ToHaveLength(t, 1)
	expect.String(res.Header.Get(headername.ContentLength)).ToBe(t, "21")
	expect.String(res.Type.String()).ToBe(t, contenttype.ApplicationJSON)
	expect.String(res.Body.String()).ToBe(t, "{\"A\":\"hello\",\"B\":10}\n")
	expect.String(testClient.Captured.URL.String()).ToBe(t, "http://example.test/foo/bar?a=orses&b=for+mutton&%5Bc%5D=%2F%5C%2F")
	expect.String(testClient.Captured.Method).ToBe(t, http.MethodGet)
	expect.String(testClient.Captured.Header.Get("Accept")).ToBe(t, contenttype.ApplicationJSON)
	expect.String(testClient.Captured.Header.Get("X-Extra")).ToBe(t, "Foo")
	expect.String(testClient.Captured.Header.Get("If-None-Match")).ToBe(t, `"abc123"`)
	expect.String(testClient.Captured.Header.Get("If-Match")).ToBe(t, `"xyz"`)
	expect.String(testClient.Captured.Header.Get("If-Modified-Since")).ToBe(t, `Sun, 10 Oct 2010 10:10:10 GMT`)
}

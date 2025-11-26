package rest

import (
	"context"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/rickb777/acceptable/contenttype"
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
			expect.Error(err).Not().ToHaveOccurred(t)

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

func TestHead(t *testing.T) {
	testClient := mytesting.StubHttpWithBody(`HTTP/1.1 200 OK
Content-Type: application/json
Content-Length: 21

`)
	cl := NewClient("http://example.test/foo", SetHttpClient(testClient))

	res, err := cl.Head(context.Background(), "/bar", Headers("X-Extra", "Foo"), QueryKV("a", "orses"))

	expect.Error(err).Not().ToHaveOccurred(t)
	expect.Number(res.StatusCode).ToBe(t, 200)
	expect.Map(res.Header).ToHaveLength(t, 2)
	expect.String(res.Header.Get("Content-Length")).ToBe(t, "21")
	expect.String(res.Header.Get("Content-Type")).ToBe(t, contenttype.ApplicationJSON)
	expect.String(res.Body.String()).ToBe(t, "")
	expect.String(testClient.Captured.URL.String()).ToBe(t, "http://example.test/foo/bar?a=orses")
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

	res, err := cl.Get(context.Background(), "/bar", Headers("X-Extra", "Foo"), QueryKV("a", "orses"))

	expect.Error(err).Not().ToHaveOccurred(t)
	expect.Number(res.StatusCode).ToBe(t, 200)
	expect.Map(res.Header).ToHaveLength(t, 2)
	expect.String(res.Header.Get("Content-Length")).ToBe(t, "21")
	expect.String(res.Header.Get("Content-Type")).ToBe(t, contenttype.ApplicationJSON)
	expect.String(res.Body.String()).ToBe(t, "{\"A\":\"hello\",\"B\":10}\n")
	expect.String(testClient.Captured.URL.String()).ToBe(t, "http://example.test/foo/bar?a=orses")
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

	res, err := cl.Post(context.Background(), "/bar", body, Headers("X-Extra", "Foo"), QueryKV("a", "orses"))

	expect.Error(err).Not().ToHaveOccurred(t)
	expect.Number(res.StatusCode).ToBe(t, 200)
	expect.Map(res.Header).ToHaveLength(t, 2)
	expect.String(res.Header.Get("Content-Length")).ToBe(t, "21")
	expect.String(res.Header.Get("Content-Type")).ToBe(t, contenttype.ApplicationJSON)
	expect.String(res.Body.String()).ToBe(t, "{\"A\":\"hello\",\"B\":10}\n")
	expect.String(testClient.Captured.URL.String()).ToBe(t, "http://example.test/foo/bar?a=orses")
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

	res, err := cl.Put(context.Background(), "/bar", body, Headers("X-Extra", "Foo"), QueryKV("a", "orses"))

	expect.Error(err).Not().ToHaveOccurred(t)
	expect.Number(res.StatusCode).ToBe(t, 200)
	expect.Map(res.Header).ToHaveLength(t, 2)
	expect.String(res.Header.Get("Content-Length")).ToBe(t, "0")
	expect.String(res.Header.Get("Content-Type")).ToBe(t, contenttype.ApplicationJSON)
	expect.String(res.Body.String()).ToBe(t, "")
	expect.String(testClient.Captured.URL.String()).ToBe(t, "http://example.test/foo/bar?a=orses")
	expect.String(testClient.Captured.Method).ToBe(t, http.MethodPut)
	expect.String(testClient.Captured.Header.Get("Accept")).ToBe(t, contenttype.ApplicationJSON)
}

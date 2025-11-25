package rest_test

import (
	"context"
	"encoding/json"
	"errors"
	"net"
	"net/url"
	"os"
	"syscall"
	"testing"

	. "github.com/rickb777/acceptable/contenttype"
	"github.com/rickb777/expect"
	"github.com/rickb777/httpclient/internal/mytesting"
	"github.com/rickb777/httpclient/rest"
	"github.com/rickb777/httpclient/rest/temperror"
)

type data struct {
	A string
	B int
}

func TestGet_200_JSON_with_binding(t *testing.T) {
	testClient := mytesting.StubHttpWithBody(
		`HTTP/1.1 200 OK
Content-Type: application/json
Content-Length: 21

{"A":"hello","B":10}
`)

	var d data
	m := rest.Headers("Accept", "foo/bar")
	rh, code, err := rest.Request(context.Background(), "http://example.test/foo/bar").With(m).Get(testClient).Unmarshal(&d)

	expect.Error(err).Not().ToHaveOccurred(t)
	expect.Any(code).ToBe(t, 200)
	expect.Any(d).ToBe(t, data{A: "hello", B: 10})
	expect.Any(rh.Get("Content-Type")).ToBe(t, ApplicationJSON)
	expect.Any(testClient.Captured.Method).ToBe(t, "GET")
	expect.Any(testClient.Captured.Header.Get("Accept")).ToBe(t, "foo/bar")
}

func TestGet_200_JSON_as_map(t *testing.T) {
	testClient := mytesting.StubHttpWithBody(
		`HTTP/1.1 200 OK
Content-Type: application/json
Content-Length: 21

{"A":"hello","B":10}
`)

	dm, code, err := rest.Request(context.Background(), "http://example.test/foo/bar").Get(testClient).ToMap()

	expect.Error(err).Not().ToHaveOccurred(t)
	expect.Any(code).ToBe(t, 200)
	expect.Any(dm["A"]).ToBe(t, "hello")
	expect.Any(dm["B"]).ToBe(t, json.Number("10"))
	expect.Any(testClient.Captured.Method).ToBe(t, "GET")
}

func TestGet_non_JSON_error(t *testing.T) {
	testClient := mytesting.StubHttpWithBody(
		`HTTP/1.1 200 OK
Content-Type: application/octet-stream
Content-Length: 21

abc123-dkvfj;ikjas0d
`)

	var d data
	rh, code, err := rest.Request(context.Background(), "http://example.test/foo/bar").Get(testClient).Unmarshal(&d)

	expect.Error(err).ToHaveOccurred(t)
	expect.Bool(temperror.IsTransient(err)).I("%v", err).ToBeFalse(t)
	expect.Any(code).ToBe(t, 406)
	expect.Any(d).ToBe(t, data{})
	expect.Any(rh.Get("Content-Type")).ToBe(t, "application/octet-stream")
	expect.Any(testClient.Captured.Method).ToBe(t, "GET")
	expect.Any(testClient.Captured.Header.Get("Accept")).ToBe(t, ApplicationJSON)
}

func TestGet_200_bytes(t *testing.T) {
	testClient := mytesting.StubHttpWithBody(
		`HTTP/1.1 200 OK
Content-Type: application/octet-stream
Content-Length: 21

abc123-dkvfj;ikjas0d
`)

	var d []byte
	m := rest.Headers("Accept", "*/*")
	rh, code, err := rest.Request(context.Background(), "http://example.test/foo/bar").With(m).Get(testClient).Unmarshal(&d)

	expect.Error(err).Not().ToHaveOccurred(t)
	expect.Any(code).ToBe(t, 200)
	expect.Any(string(d)).ToBe(t, "abc123-dkvfj;ikjas0d\n")
	expect.Any(rh.Get("Content-Type")).ToBe(t, "application/octet-stream")
	expect.Any(testClient.Captured.Method).ToBe(t, "GET")
	expect.Any(testClient.Captured.Header.Get("Accept")).ToBe(t, "*/*")
}

func TestGet_200_string(t *testing.T) {
	testClient := mytesting.StubHttpWithBody(
		`HTTP/1.1 200 OK
Content-Type: application/json
Content-Length: 21

{"A":"hello","B":10}
`)

	var d string
	m := rest.Headers("Accept", "foo/bar")
	rh, code, err := rest.Request(context.Background(), "http://example.test/foo/bar").With(m).Get(testClient).Unmarshal(&d)

	expect.Error(err).Not().ToHaveOccurred(t)
	expect.Any(code).ToBe(t, 200)
	expect.Any(d).ToBe(t, `{"A":"hello","B":10}`+"\n")
	expect.Any(rh.Get("Content-Type")).ToBe(t, ApplicationJSON)
	expect.Any(testClient.Captured.Method).ToBe(t, "GET")
	expect.Any(testClient.Captured.Header.Get("Accept")).ToBe(t, "foo/bar")
}

func TestHead_200(t *testing.T) {
	testClient := mytesting.StubHttpWithBody(
		`HTTP/1.1 200 OK
Content-Type: application/json

`)

	rh, code, err := rest.Request(context.Background(), "http://example.test/foo/bar").Head(testClient).Status()

	expect.Error(err).Not().ToHaveOccurred(t)
	expect.Any(code).ToBe(t, 200)
	expect.Any(rh.Get("Content-Type")).ToBe(t, ApplicationJSON)
	expect.Any(testClient.Captured.Method).ToBe(t, "HEAD")
}

func TestDelete_200_string(t *testing.T) {
	testClient := mytesting.StubHttpWithBody(
		`HTTP/1.1 200 OK
Content-Type: application/json
Content-Length: 21

{"A":"hello","B":10}
`)

	var d string
	m := rest.Headers("Accept", "foo/bar")
	_, code, err := rest.Request(context.Background(), "http://example.test/foo/bar").With(m).Delete(testClient).Unmarshal(&d)

	expect.Error(err).Not().ToHaveOccurred(t)
	expect.Any(code).ToBe(t, 200)
	expect.Any(d).ToBe(t, `{"A":"hello","B":10}`+"\n")
	expect.Any(testClient.Captured.Method).ToBe(t, "DELETE")
	expect.Any(testClient.Captured.Header.Get("Accept")).ToBe(t, "foo/bar")
}

func TestGet_500(t *testing.T) {
	testClient := mytesting.StubHttpWithBody(
		`HTTP/1.1 500 Internal Server Error
Content-Type: application/json
Content-Length: 22

{"error":"failure"}
`)

	dm, code, err := rest.Request(context.Background(), "http://example.test/foo/bar").Get(testClient).ToMap()

	expect.Error(err).ToHaveOccurred(t)
	expect.Bool(temperror.IsTransient(err)).ToBeTrue(t)
	expect.Any(code).ToBe(t, 500)
	expect.Any(dm["error"]).ToBe(t, "failure")
	expect.String(err.Error()).ToContain(t, `500: GET http://example.test/foo/bar application/json {"error":"failure"}`)
}

func TestPut_values_with_no_response_204(t *testing.T) {
	testClient := mytesting.StubHttpWithBody("HTTP/1.1 204 No Content\n\n")

	d := data{A: "hello", B: 10}
	_, code, err := rest.JSON(context.Background(), "http://example.test/foo/bar", d).Put(testClient).Status()

	expect.Error(err).Not().ToHaveOccurred(t)
	expect.Any(code).ToBe(t, 204)
	expect.Any(testClient.Captured.Method).ToBe(t, "PUT")
	expect.Any(testClient.Captured.Header.Get("Content-Type")).ToBe(t, ApplicationJSON)
	expect.Any(testClient.Captured.Header.Get("Content-Length")).ToBe(t, "21")
	expect.Any(testClient.Captured.Header.Get("Accept")).ToBe(t, ApplicationJSON)
}

func TestPatch_values_with_no_response_204(t *testing.T) {
	testClient := mytesting.StubHttpWithBody("HTTP/1.1 204 No Content\n\n")

	d := map[string]interface{}{"A": "hello", "B": 10}
	_, code, err := rest.JSON(context.Background(), "http://example.test/foo/bar", d).Patch(testClient).Status()

	expect.Error(err).Not().ToHaveOccurred(t)
	expect.Any(code).ToBe(t, 204)
	expect.Any(testClient.Captured.Method).ToBe(t, "PATCH")
	expect.Any(testClient.Captured.Header.Get("Content-Type")).ToBe(t, ApplicationJSON)
	expect.Any(testClient.Captured.Header.Get("Content-Length")).ToBe(t, "21")
	expect.Any(testClient.Captured.Header.Get("Accept")).ToBe(t, ApplicationJSON)
}

func TestPost_values_with_JSON_response_struct_200(t *testing.T) {
	testClient := mytesting.StubHttpWithBody(
		`HTTP/1.1 200 OK
Content-Type: application/json
Content-Length: 21

{"A":"foo","B":7}
`)

	v := make(url.Values)
	v.Set("foo", "1")
	v.Set("bar", "2")
	var o data
	_, code, err := rest.Entity(context.Background(), "http://example.test/foo/bar", v, ApplicationForm).Post(testClient).Unmarshal(&o)

	expect.Error(err).Not().ToHaveOccurred(t)
	expect.Any(code).ToBe(t, 200)
	expect.Any(o).ToBe(t, data{A: "foo", B: 7})
	expect.Any(testClient.Captured.Method).ToBe(t, "POST")
	expect.Any(testClient.Captured.Header.Get("Content-Type")).ToBe(t, ApplicationForm)
	expect.Any(testClient.Captured.Header.Get("Content-Length")).ToBe(t, "11")
	expect.Any(testClient.Captured.Header.Get("Accept")).ToBe(t, ApplicationJSON)
}

func TestPost_values_with_no_input_and_JSON_response_as_map_200(t *testing.T) {
	testClient := mytesting.StubHttpWithBody(
		`HTTP/1.1 200 OK
Content-Type: application/json
Content-Length: 21

{"A":"foo","B":7}
`)

	dm, code, err := rest.Request(context.Background(), "http://example.test/foo/bar").Post(testClient).ToMap()

	expect.Error(err).Not().ToHaveOccurred(t)
	expect.Any(code).ToBe(t, 200)
	expect.Any(dm["A"]).ToBe(t, "foo")
	expect.Any(dm["B"]).ToBe(t, json.Number("7"))
	expect.Any(testClient.Captured.Method).ToBe(t, "POST")
	expect.Any(testClient.Captured.Header.Get("Content-Type")).ToBe(t, "")   // undefined
	expect.Any(testClient.Captured.Header.Get("Content-Length")).ToBe(t, "") // undefined
	expect.Any(testClient.Captured.Header.Get("Accept")).ToBe(t, ApplicationJSON)
	expect.Any(testClient.ReqBody).ToBe(t, "") // undefined
}

func TestPost_string_with_string_response_200(t *testing.T) {
	var i = `{"A":"foo","B":7}`

	// input is test both as string and as *string
	cases := []interface{}{i, &i}

	for _, c := range cases {
		testClient := mytesting.StubHttpWithBody(
			`HTTP/1.1 200 OK
Content-Type: application/json
Content-Length: 21

{"A":"foo","B":7}
`)

		var o string
		_, code, err := rest.Entity(context.Background(), "http://example.test/foo/bar", c, "").Post(testClient).Unmarshal(&o)

		expect.Error(err).Not().ToHaveOccurred(t)
		expect.Any(code).ToBe(t, 200)
		expect.Any(o).ToBe(t, `{"A":"foo","B":7}`+"\n")
		expect.Any(testClient.Captured.Method).ToBe(t, "POST")
		expect.Any(testClient.Captured.Header.Get("Content-Type")).ToBe(t, ApplicationJSON)
		expect.Any(testClient.Captured.Header.Get("Content-Length")).ToBe(t, "17")
		expect.Any(testClient.Captured.Header.Get("Accept")).ToBe(t, ApplicationJSON)
	}
}

func TestPost_500(t *testing.T) {
	testClient := mytesting.StubHttpWithBody(
		`HTTP/1.1 500 Internal Server Error
Content-Type: application/json
Content-Length: 22

{"error":"failure"}
`)

	d := data{A: "hello", B: 10}
	_, code, err := rest.JSON(context.Background(), "http://example.test/foo/bar", &d).Post(testClient).Status()

	expect.Error(err).ToHaveOccurred(t)
	expect.Bool(temperror.IsTransient(err)).ToBeTrue(t)
	expect.Any(code).ToBe(t, 500)
	expect.String(err.Error()).ToContain(t, `500: POST http://example.test/foo/bar application/json {"error":"failure"}`)
}

func TestGet_permanent_error(t *testing.T) {
	src := errors.New("error 1")
	testClient := &mytesting.StubHttp{Err: src}

	_, code, err := rest.Request(context.Background(), "http://example.test/foo/bar").Get(testClient).Status()

	expect.Any(err).ToBe(t, src)
	expect.Bool(temperror.IsTransient(err)).ToBeFalse(t)
	expect.Any(code).ToBe(t, 0)
}

func TestGet_transient_error(t *testing.T) {
	e1 := temperror.Wrap(errors.New("error 1"))
	e2 := &net.OpError{
		Op:  "dial",
		Net: "tcp",
		//Addr:   nil,
		Err: &os.SyscallError{Syscall: "connect", Err: syscall.ECONNREFUSED},
	}
	cases := []error{e1, e2}

	for _, ce := range cases {
		testClient := &mytesting.StubHttp{Err: ce}

		_, code, err := rest.Request(context.Background(), "http://example.test/foo/bar").Get(testClient).Status()

		//expect.Any(err).ToBe(t, ce)
		expect.Bool(temperror.IsTransient(err)).ToBeTrue(t)
		expect.Any(code).ToBe(t, 0)
	}

	for _, ce := range cases {
		testClient := &mytesting.StubHttp{Err: ce}

		_, code, err := rest.Request(context.Background(), "http://example.test/foo/bar").Put(testClient).Status()

		//expect.Any(err).ToBe(t, ce)
		expect.Bool(temperror.IsTransient(err)).ToBeTrue(t)
		expect.Any(code).ToBe(t, 0)
	}
}

func TestPost_permanent_error(t *testing.T) {
	src := errors.New("error 1")
	testClient := &mytesting.StubHttp{Err: src}

	_, code, err := rest.Request(context.Background(), "http://example.test/foo/bar").Post(testClient).Status()

	expect.Any(err).ToBe(t, src)
	expect.Bool(temperror.IsTransient(err)).ToBeFalse(t)
	expect.Any(code).ToBe(t, 0)
}

func TestPost_transient_error(t *testing.T) {
	src := temperror.Wrap(errors.New("error 1"))
	testClient := &mytesting.StubHttp{Err: src}

	_, code, err := rest.Request(context.Background(), "http://example.test/foo/bar").Post(testClient).Status()

	expect.Any(err).ToBe(t, src)
	expect.Bool(temperror.IsTransient(err)).ToBeTrue(t)
	expect.Any(code).ToBe(t, 0)
}

//-------------------------------------------------------------------------------------------------

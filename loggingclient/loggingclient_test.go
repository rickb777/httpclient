package loggingclient

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/rickb777/expect"
	"github.com/rickb777/httpclient/logging"
	"github.com/rickb777/httpclient/testhttpclient"
)

func TestLoggingClient_200_OK_Off(t *testing.T) {
	input := "Sunny day outside"
	target := "http://somewhere.com/a/b/c"

	lvl := logging.Off
	info := lvl.String()
	logging.Now = stubbedTime()
	req := httptest.NewRequest("POST", target, strings.NewReader(input))
	testClient := testhttpclient.New(t).AddLiteralResponse("POST", target,
		`HTTP/1.1 200 OK
Content-Type: application/json; charset=UTF-8
Content-Length: 18

{"A":"foo","B":7}
`)
	logged := false
	logger := func(item *logging.LogItem) {
		logged = true
	}

	client := New(testClient, logger, lvl)
	res, err := client.Do(req)

	expect.Error(err).Info(info).Not().Info(info).ToHaveOccurred(t)
	expect.Number(res.StatusCode).Info(info).ToBe(t, http.StatusOK)
	expect.Bool(logged).Info(info).ToBeFalse(t)
	buf := &bytes.Buffer{}
	buf.ReadFrom(req.Body)
	expect.String(buf.String()).Info(info).ToBe(t, input)
	buf.Reset()
	buf.ReadFrom(res.Body)
	expect.String(buf.String()).Info(info).ToBe(t, `{"A":"foo","B":7}`+"\n")
}

func TestLoggingClient_200_OK_WithHeadersAndBodies(t *testing.T) {
	input := "Sunny day outside"
	target := "http://somewhere.com/a/b/c"

	var lvl logging.Level
	for lvl = logging.Discrete; lvl <= logging.WithHeadersAndBodies; lvl++ {
		info := lvl.String()
		logging.Now = stubbedTime()
		req := httptest.NewRequest("POST", target, strings.NewReader(input))
		testClient := testhttpclient.New(t).AddLiteralResponse("POST", target,
			`HTTP/1.1 200 OK
Content-Type: application/json; charset=UTF-8
Content-Length: 18

{"A":"foo","B":7}
`)
		logged := false
		logger := func(item *logging.LogItem) {
			logged = true
			expect.String(item.Method).Info(info).ToBe(t, req.Method)
			expect.Any(item.URL).Info(info).ToBe(t, req.URL)
			if lvl == logging.WithHeadersAndBodies {
				expect.String(item.Request.Body.String()).Info(info).ToBe(t, input)
				expect.String(item.Response.Body.String()).Info(info).ToBe(t, `{"A":"foo","B":7}`+"\n")
			}
			expect.Number(item.StatusCode).Info(info).ToBe(t, http.StatusOK)
			expect.Error(item.Err).Info(info).Not().Info(info).ToHaveOccurred(t)
			expect.Any(item.Start).Info(info).ToBe(t, t0.Add(time.Second))
			expect.Number(item.Duration).Info(info).ToBe(t, time.Second)
			expect.Number(item.Level).Info(info).ToBe(t, lvl)
		}

		client := New(testClient, logger, lvl)
		res, err := client.Do(req)

		expect.Error(err).Info(info).Not().Info(info).ToHaveOccurred(t)
		expect.Number(res.StatusCode).Info(info).ToBe(t, http.StatusOK)
		expect.Bool(logged).Info(info).ToBeTrue(t)
		buf := &bytes.Buffer{}
		buf.ReadFrom(req.Body)
		expect.String(buf.String()).Info(info).ToBe(t, input)
		buf.Reset()
		buf.ReadFrom(res.Body)
		expect.String(buf.String()).Info(info).ToBe(t, `{"A":"foo","B":7}`+"\n")
	}
}

func TestLoggingClient_200_OK_Levels(t *testing.T) {
	target := "http://somewhere.com/a/b/c?foo=1&bar=2"
	req := httptest.NewRequest("GET", target, nil)

	cases := map[logging.Level]string{
		logging.Discrete:             "http://somewhere.com/a/b/c",
		logging.Summary:              "http://somewhere.com/a/b/c?foo=1&bar=2",
		logging.WithHeaders:          "http://somewhere.com/a/b/c?foo=1&bar=2",
		logging.WithHeadersAndBodies: "http://somewhere.com/a/b/c?foo=1&bar=2",
	}

	for lvl, expected := range cases {
		testClient := testhttpclient.New(t).AddLiteralResponse("GET", target,
			`HTTP/1.1 200 OK
Content-Type: application/json; charset=UTF-8
Content-Length: 18

{"A":"foo","B":7}
`)

		logged := false
		logger := func(item *logging.LogItem) {
			logged = true
			u, _ := url.Parse(expected)
			expect.Any(item.URL).Info(lvl).ToBe(t, u)
		}

		client := New(testClient, logger, lvl)
		_, _ = client.Do(req)
		expect.Bool(logged).Info(lvl).ToBeTrue(t)
	}
}

func TestLoggingClient_200_OK_variable(t *testing.T) {
	target := "http://somewhere.com/a/b/c"
	req := httptest.NewRequest("GET", target, nil)

	cases := map[logging.Level]int{}

	response := `HTTP/1.1 200 OK
Content-Type: application/json; charset=UTF-8
Content-Length: 18

{"A":"foo","B":7}
`
	testClient := testhttpclient.New(t).
		AddLiteralResponse("GET", target, response).
		AddLiteralResponse("GET", target, response).
		AddLiteralResponse("GET", target, response)

	logger := func(item *logging.LogItem) {
		cases[item.Level]++
	}

	vf := logging.NewVariableFilter(logging.Off)
	client := NewWithFilter(testClient, logger, vf)

	_, _ = client.Do(req)

	vf.SetLevel(logging.FixedLevel(logging.Discrete))

	_, _ = client.Do(req)

	vf.SetLevel(logging.FixedLevel(logging.Summary))

	_, _ = client.Do(req)

	expect.Number(cases[logging.Off]).ToBe(t, 0)
	expect.Number(cases[logging.Discrete]).ToBe(t, 1)
	expect.Number(cases[logging.Summary]).ToBe(t, 1)
}

func TestLoggingClient_error(t *testing.T) {
	target := "http://somewhere.com/a/b/c"
	req := httptest.NewRequest("GET", target, nil)
	theError := errors.New("Kaboom!")
	testClient := testhttpclient.New(t).AddError("GET", target, theError)

	logged := false
	logger := func(item *logging.LogItem) {
		logged = true
		expect.String(item.Method).ToBe(t, req.Method)
		expect.Any(item.URL).ToBe(t, req.URL)
		expect.Slice(item.Request.Body.Bytes()).ToHaveLength(t, 0)
		expect.Slice(item.Response.Body.Bytes()).ToHaveLength(t, 0)
		expect.Error(item.Err).ToContain(t, "Kaboom!")
		expect.Number(item.Duration).ToBeGreaterThan(t, 0)
	}

	client := New(testClient, logger, logging.Summary)
	_, err := client.Do(req)

	expect.Error(err).ToHaveOccurred(t)
	expect.Error(err).ToContain(t, "Kaboom!")
	expect.Bool(logged).ToBeTrue(t)
}

var t0 = time.Date(2021, 04, 01, 10, 0, 0, 0, time.UTC)

func stubbedTime() func() time.Time {
	t := t0
	return func() time.Time {
		t = t.Add(time.Second)
		return t
	}
}

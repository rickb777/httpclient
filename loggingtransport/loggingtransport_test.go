package loggingtransport

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/onsi/gomega"
	"github.com/rickb777/httpclient/logging"
	"github.com/rickb777/httpclient/testhttpclient"
)

func TestLoggingClient_200_OK_WithHeadersAndBodies(t *testing.T) {
	g := gomega.NewWithT(t)

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
			g.Expect(item.Method).To(gomega.Equal(req.Method), info)
			g.Expect(item.URL).To(gomega.Equal(req.URL), info)
			if lvl == logging.WithHeadersAndBodies {
				g.Expect(string(item.Request.Body)).To(gomega.Equal(input), info)
				g.Expect(string(item.Response.Body)).To(gomega.Equal(`{"A":"foo","B":7}`+"\n"), info)
			}
			g.Expect(item.StatusCode).To(gomega.Equal(http.StatusOK), info)
			g.Expect(item.Err).NotTo(gomega.HaveOccurred(), info)
			g.Expect(item.Start).To(gomega.Equal(t0.Add(time.Second)), info)
			g.Expect(item.Duration).To(gomega.Equal(time.Second), info)
			g.Expect(item.Level).To(gomega.Equal(lvl), info)
		}

		client := New(testClient, logger, lvl)
		res, err := client.RoundTrip(req)

		g.Expect(err).NotTo(gomega.HaveOccurred(), info)
		g.Expect(res.StatusCode).To(gomega.Equal(http.StatusOK), info)
		g.Expect(logged).To(gomega.BeTrue(), info)
		buf := &bytes.Buffer{}
		buf.ReadFrom(req.Body)
		g.Expect(buf.String()).To(gomega.Equal(input), info)
		buf.Reset()
		buf.ReadFrom(res.Body)
		g.Expect(buf.String()).To(gomega.Equal(`{"A":"foo","B":7}`+"\n"), info)
	}
}

func TestLoggingClient_200_OK_Levels(t *testing.T) {
	g := gomega.NewWithT(t)

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
			g.Expect(item.URL).To(gomega.Equal(u))
		}

		client := New(testClient, logger, lvl)
		_, _ = client.RoundTrip(req)
		g.Expect(logged).To(gomega.BeTrue())
	}
}

func TestLoggingClient_error(t *testing.T) {
	g := gomega.NewWithT(t)

	target := "http://somewhere.com/a/b/c"
	req := httptest.NewRequest("GET", target, nil)
	theError := errors.New("Kaboom!")
	testClient := testhttpclient.New(t).AddError("GET", target, theError)

	logged := false
	logger := func(item *logging.LogItem) {
		logged = true
		g.Expect(item.Method).To(gomega.Equal(req.Method))
		g.Expect(item.URL).To(gomega.Equal(req.URL))
		g.Expect(item.Request.Body).To(gomega.HaveLen(0))
		g.Expect(item.Response.Body).To(gomega.HaveLen(0))
		g.Expect(item.Err).To(gomega.HaveOccurred())
		g.Expect(item.Err.Error()).To(gomega.Equal("Kaboom!"))
		g.Expect(item.Duration).To(gomega.BeNumerically(">", 0))
	}

	client := New(testClient, logger, logging.Summary)
	_, err := client.RoundTrip(req)

	g.Expect(err).To(gomega.HaveOccurred())
	g.Expect(err.Error()).To(gomega.Equal("Kaboom!"))
	g.Expect(logged).To(gomega.BeTrue())
}

var t0 = time.Date(2021, 04, 01, 10, 0, 0, 0, time.UTC)

func stubbedTime() func() time.Time {
	t := t0
	return func() time.Time {
		t = t.Add(time.Second)
		return t
	}
}

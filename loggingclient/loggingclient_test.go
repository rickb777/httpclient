package loggingclient

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

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
		req := httptest.NewRequest("POST", target, strings.NewReader(input))
		testClient := testhttpclient.New(t).AddLiteralResponse("POST", target,
			`HTTP/1.1 200 OK
Content-Type: application/json; charset=UTF-8
Content-Length: 18

{"A":"foo","B":7}
`)

		logger := func(item *logging.LogItem) {
			g.Expect(item.Method).To(gomega.Equal(req.Method), info)
			g.Expect(item.URL).To(gomega.Equal(req.URL), info)
			if lvl == logging.WithHeadersAndBodies {
				g.Expect(string(item.Request.Body)).To(gomega.Equal(input), info)
				g.Expect(string(item.Response.Body)).To(gomega.Equal(`{"A":"foo","B":7}`+"\n"), info)
			}
			g.Expect(item.StatusCode).To(gomega.Equal(http.StatusOK), info)
			g.Expect(item.Err).NotTo(gomega.HaveOccurred(), info)
			g.Expect(item.Duration).To(gomega.BeNumerically(">", 0), info)
			g.Expect(item.Level).To(gomega.Equal(lvl), info)
		}

		client := New(testClient, logger, lvl)
		res, err := client.Do(req)

		g.Expect(err).NotTo(gomega.HaveOccurred(), info)
		g.Expect(res.StatusCode).To(gomega.Equal(http.StatusOK), info)
		buf := &bytes.Buffer{}
		io.Copy(buf, req.Body)
		g.Expect(buf.String()).To(gomega.Equal(input), info)
		buf.Reset()
		io.Copy(buf, res.Body)
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

		logger := func(item *logging.LogItem) {
			u, _ := url.Parse(expected)
			g.Expect(item.URL).To(gomega.Equal(u))
		}

		client := New(testClient, logger, lvl)
		_, _ = client.Do(req)
	}
}

func TestLoggingClient_error(t *testing.T) {
	g := gomega.NewWithT(t)

	target := "http://somewhere.com/a/b/c"
	req := httptest.NewRequest("GET", target, nil)
	theError := errors.New("Kaboom!")
	testClient := testhttpclient.New(t).AddError("GET", target, theError)

	logger := func(item *logging.LogItem) {
		g.Expect(item.Method).To(gomega.Equal(req.Method))
		g.Expect(item.URL).To(gomega.Equal(req.URL))
		g.Expect(item.Request.Body).To(gomega.HaveLen(0))
		g.Expect(item.Response.Body).To(gomega.HaveLen(0))
		g.Expect(item.Err).To(gomega.HaveOccurred())
		g.Expect(item.Err.Error()).To(gomega.Equal("Kaboom!"))
		g.Expect(item.Duration).To(gomega.BeNumerically(">", 0))
	}

	client := New(testClient, logger, 0)
	_, err := client.Do(req)

	g.Expect(err).To(gomega.HaveOccurred())
	g.Expect(err.Error()).To(gomega.Equal("Kaboom!"))
}

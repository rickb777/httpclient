package httpclient

import (
	"errors"
	"github.com/onsi/gomega"
	"github.com/rickb777/httpclient/testhttpclient"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLoggingClient_200_OK_WithHeadersAndBodies(t *testing.T) {
	g := gomega.NewWithT(t)

	target := "http://somewhere.com/a/b/c"
	req := httptest.NewRequest("GET", target, nil)

	var lvl Level
	for lvl = Discrete; lvl <= WithHeadersAndBodies; lvl++ {
		testClient := testhttpclient.New(t).AddLiteralResponse("GET", target,
			`HTTP/1.1 200 OK
Content-Type: application/json; charset=UTF-8
Content-Length: 18

{"A":"foo","B":7}
`)

		logger := func(item *LogItem) {
			g.Expect(item.Method).To(gomega.Equal(req.Method))
			g.Expect(item.URL).To(gomega.Equal(req.URL.String()))
			g.Expect(item.Request.Body).To(gomega.HaveLen(0))
			if lvl == WithHeadersAndBodies {
				g.Expect(string(item.Response.Body)).To(gomega.Equal(`{"A":"foo","B":7}` + "\n"))
			}
			g.Expect(item.StatusCode).To(gomega.Equal(http.StatusOK))
			g.Expect(item.Err).NotTo(gomega.HaveOccurred())
			g.Expect(item.Duration).To(gomega.BeNumerically(">", 0))
			g.Expect(item.Level).To(gomega.Equal(lvl))
		}

		client := NewLoggingClient(testClient, logger, lvl)
		res, err := client.Do(req)

		g.Expect(err).NotTo(gomega.HaveOccurred())
		g.Expect(res.StatusCode).To(gomega.Equal(http.StatusOK))
	}
}

func TestLoggingClient_200_OK_Levels(t *testing.T) {
	g := gomega.NewWithT(t)

	target := "http://somewhere.com/a/b/c?foo=1&bar=2"
	req := httptest.NewRequest("GET", target, nil)

	cases := map[Level]string{
		Discrete:             "http://somewhere.com/a/b/c",
		Summary:              "http://somewhere.com/a/b/c?foo=1&bar=2",
		WithHeaders:          "http://somewhere.com/a/b/c?foo=1&bar=2",
		WithHeadersAndBodies: "http://somewhere.com/a/b/c?foo=1&bar=2",
	}

	for lvl, expected := range cases {
		testClient := testhttpclient.New(t).AddLiteralResponse("GET", target,
			`HTTP/1.1 200 OK
Content-Type: application/json; charset=UTF-8
Content-Length: 18

{"A":"foo","B":7}
`)

		logger := func(item *LogItem) {
			g.Expect(item.URL).To(gomega.Equal(expected))
		}

		client := NewLoggingClient(testClient, logger, lvl)
		_, _ = client.Do(req)
	}
}

func TestLoggingClient_error(t *testing.T) {
	g := gomega.NewWithT(t)

	target := "http://somewhere.com/a/b/c"
	req := httptest.NewRequest("GET", target, nil)
	theError := errors.New("Kaboom!")
	testClient := testhttpclient.New(t).AddError("GET", target, theError)

	logger := func(item *LogItem) {
		g.Expect(item.Method).To(gomega.Equal(req.Method))
		g.Expect(item.URL).To(gomega.Equal(req.URL))
		g.Expect(item.Request.Body).To(gomega.HaveLen(0))
		g.Expect(item.Response.Body).To(gomega.HaveLen(0))
		g.Expect(item.Err).To(gomega.HaveOccurred())
		g.Expect(item.Err.Error()).To(gomega.Equal("Kaboom!"))
		g.Expect(item.Duration).To(gomega.BeNumerically(">", 0))
	}

	client := NewLoggingClient(testClient, logger, 0)
	_, err := client.Do(req)

	g.Expect(err).To(gomega.HaveOccurred())
	g.Expect(err.Error()).To(gomega.Equal("Kaboom!"))
}

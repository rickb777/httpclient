package hostheader

import (
	"github.com/onsi/gomega"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHostHeader_without_pre_existing_header(t *testing.T) {
	g := gomega.NewWithT(t)

	testcases := map[string]string{
		"https://www.example.com/":      "www.example.com",
		"https://www.example.com:3456/": "www.example.com:3456",
		"http://localhost/":             "localhost",
		"http://localhost:8080/":        "localhost:8080",
		"http://127.0.0.1:8080/":        "127.0.0.1:8080",
	}

	for u, expected := range testcases {
		hh := Wrap(tester(func(req *http.Request) {
			g.Expect(req.Header.Get("Host")).To(gomega.Equal(expected))
			g.Expect(req.URL.Host).NotTo(gomega.BeEmpty())
			g.Expect(req.URL.Scheme).NotTo(gomega.BeEmpty())
		}))

		req := httptest.NewRequest("GET", u, nil)
		_, err := hh.Do(req)
		g.Expect(err).NotTo(gomega.HaveOccurred())
	}
}

func TestHostHeader_with_pre_existing_header(t *testing.T) {
	g := gomega.NewWithT(t)

	testcases := map[string]string{
		"https://www.example.com/": "target.example.com",
		"http://127.0.0.1:8080/":   "target.example.com",
	}

	for u, expected := range testcases {
		hh := Wrap(tester(func(req *http.Request) {
			g.Expect(req.Header.Get("Host")).To(gomega.Equal(expected))
			g.Expect(req.URL.Host).NotTo(gomega.BeEmpty())
			g.Expect(req.URL.Scheme).NotTo(gomega.BeEmpty())
		}))

		req := httptest.NewRequest("GET", u, nil)
		req.Header.Set("Host", "target.example.com")
		_, err := hh.Do(req)
		g.Expect(err).NotTo(gomega.HaveOccurred())
	}
}

func TestHostHeader_inserts_other_headers(t *testing.T) {
	g := gomega.NewWithT(t)

	hh := Wrap(tester(func(req *http.Request) {
		g.Expect(req.Header.Get("X-Alpha")).To(gomega.Equal("foo"))
		g.Expect(req.Header.Get("X-Beta")).To(gomega.Equal("bar"))
	}), "X-Alpha", "foo", "X-Beta", "bar")

	req := httptest.NewRequest("GET", "https://www.example.com/", nil)
	req.Header.Set("Host", "target.example.com")
	_, err := hh.Do(req)
	g.Expect(err).NotTo(gomega.HaveOccurred())
}

type tester func(req *http.Request)

func (t tester) Do(req *http.Request) (*http.Response, error) {
	t(req)
	return nil, nil
}

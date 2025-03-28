package hostheader

import (
	"github.com/rickb777/expect"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHostHeader_without_pre_existing_header(t *testing.T) {
	testcases := map[string]string{
		"https://www.example.com/":      "www.example.com",
		"https://www.example.com:3456/": "www.example.com:3456",
		"http://localhost/":             "localhost",
		"http://localhost:8080/":        "localhost:8080",
		"http://127.0.0.1:8080/":        "127.0.0.1:8080",
	}

	for u, expected := range testcases {
		hh := Wrap(tester(func(req *http.Request) {
			expect.String(req.Header.Get("Host")).ToBe(t, expected)
			expect.String(req.URL.Host).Not().ToBeEmpty(t)
			expect.String(req.URL.Scheme).Not().ToBeEmpty(t)
		}))

		req := httptest.NewRequest("GET", u, nil)
		_, err := hh.Do(req)
		expect.Error(err).Not().ToHaveOccurred(t)
	}
}

func TestHostHeader_with_pre_existing_header(t *testing.T) {
	testcases := map[string]string{
		"https://www.example.com/": "target.example.com",
		"http://127.0.0.1:8080/":   "target.example.com",
	}

	for u, expected := range testcases {
		hh := Wrap(tester(func(req *http.Request) {
			expect.String(req.Header.Get("Host")).ToBe(t, expected)
			expect.String(req.URL.Host).Not().ToBeEmpty(t)
			expect.String(req.URL.Scheme).Not().ToBeEmpty(t)
		}))

		req := httptest.NewRequest("GET", u, nil)
		req.Header.Set("Host", "target.example.com")
		_, err := hh.Do(req)
		expect.Error(err).Not().ToHaveOccurred(t)
	}
}

func TestHostHeader_inserts_other_headers(t *testing.T) {
	hh := Wrap(tester(func(req *http.Request) {
		expect.String(req.Header.Get("X-Alpha")).ToBe(t, "foo")
		expect.String(req.Header.Get("X-Beta")).ToBe(t, "bar")
	}), "X-Alpha", "foo", "X-Beta", "bar")

	req := httptest.NewRequest("GET", "https://www.example.com/", nil)
	req.Header.Set("Host", "target.example.com")
	_, err := hh.Do(req)
	expect.Error(err).Not().ToHaveOccurred(t)
}

type tester func(req *http.Request)

func (t tester) Do(req *http.Request) (*http.Response, error) {
	t(req)
	return nil, nil
}

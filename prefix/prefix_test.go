package prefix

import (
	"github.com/rickb777/expect"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestPrefix_Wrap(t *testing.T) {
	testcases := map[string]interface{}{
		"https://www.example1.com/a/b?q=1#x":          "https://www.example1.com",
		"https://www.example1.com:3456/zzz/a/b?q=1#x": "https://www.example1.com:3456/zzz",
		"https://www.example2.com/a/b?q=1#x":          u("https://www.example2.com"),
	}

	for expected, input := range testcases {
		hh := Wrap(tester(func(req *http.Request) {
			expect.String(req.URL.String()).ToBe(t, expected)
		}), input)
		req := httptest.NewRequest("GET", "/a/b?q=1#x", nil)
		_, err := hh.Do(req)
		expect.Error(err).Not().ToHaveOccurred(t)
	}
}

func TestPrefix_WrapWithHost(t *testing.T) {
	testcases := map[string]interface{}{
		"https://www.example1.com/a/b?q=1#x":          "https://www.example1.com",
		"https://www.example1.com:3456/zzz/a/b?q=1#x": "https://www.example1.com:3456/zzz",
		"https://www.example2.com/a/b?q=1#x":          u("https://www.example2.com"),
	}

	for expected, input := range testcases {
		hh := WrapWithHost(tester(func(req *http.Request) {
			expect.String(req.URL.String()).ToBe(t, expected)
			expect.String(req.Header.Get("Host")).ToBe(t, u(expected).Host)
		}), input)
		req := httptest.NewRequest("GET", "/a/b?q=1#x", nil)
		_, err := hh.Do(req)
		expect.Error(err).Not().ToHaveOccurred(t)
	}
}

func u(s string) *url.URL {
	u, err := url.Parse(s)
	if err != nil {
		panic(err)
	}
	return u
}

type tester func(req *http.Request)

func (t tester) Do(req *http.Request) (*http.Response, error) {
	t(req)
	return nil, nil
}

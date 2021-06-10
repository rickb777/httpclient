// Package prefix provides fixed prefixing for HTTP client requests.
package prefix

import (
	"github.com/rickb777/httpclient"
	"github.com/rickb777/httpclient/hostheader"
	"net/http"
	"net/url"
)

type prefix struct {
	scheme, host, path string
	inner              httpclient.HttpClient
}

// WrapWithHost wraps a HttpClient so that any request URL with undefined scheme/host will
// use the prefix supplied here. The prefix pfx should be a URL: either *net.URL or string
// (otherwise it panics).
//
// The Host header is also defined on all requests. See package hostheader.
func WrapWithHost(inner httpclient.HttpClient, pfx interface{}) httpclient.HttpClient {
	return Wrap(hostheader.Wrap(inner), pfx)
}

// Wrap wraps a HttpClient so that any request URL with undefined scheme/host will use
// the prefix supplied here. The prefix pfx should be a URL: either *net.URL or string
// (otherwise it panics).
func Wrap(inner httpclient.HttpClient, pfx interface{}) httpclient.HttpClient {
	var u *url.URL
	var err error

	switch p := pfx.(type) {
	case string:
		u, err = url.Parse(p)
		if err != nil {
			panic(err) // coding or configuration error
		}
	case *url.URL:
		u = p
	}

	return prefix{
		scheme: u.Scheme,
		host:   u.Host,
		path:   u.Path,
		inner:  inner,
	}
}

func (p prefix) Do(req *http.Request) (*http.Response, error) {
	if req.URL.Scheme == "" {
		req.URL.Scheme = p.scheme
	}
	if req.URL.Host == "" {
		req.URL.Host = p.host
	}
	if p.path != "" {
		req.URL.Path = p.path + req.URL.Path
	}
	return p.inner.Do(req)
}

// Package hostheader provides a HttpClient wrapper that automatically inserts
// the Host header into the requests it makes. Host headers are a mandatory
// requirement (https://datatracker.ietf.org/doc/html/rfc7230#section-5.4).
//
// See also package prefix that extends this by inserting a prefix on all URLs.
package hostheader

import (
	"github.com/rickb777/httpclient"
	"net/http"
)

type hh struct {
	inner httpclient.HttpClient
}

// Wrap creates an automatic Host header inserter that wraps the next client.
func Wrap(next httpclient.HttpClient) httpclient.HttpClient {
	return &hh{inner: next}
}

func (hh *hh) Do(req *http.Request) (*http.Response, error) {
	if req.Header.Get("Host") == "" {
		req.Header.Set("Host", req.URL.Host)
	}
	return hh.inner.Do(req)
}

// Package hostheader provides a HttpClient wrapper that automatically inserts
// the Host header into the requests it makes. Host headers are a mandatory
// requirement (https://datatracker.ietf.org/doc/html/rfc7230#section-5.4).
//
// See also package prefix that extends this by inserting a prefix on all URLs.
package hostheader

import (
	"net/http"

	"github.com/rickb777/httpclient"
)

type hh struct {
	inner    httpclient.HttpClient
	metadata map[string]string
}

// Wrap creates an automatic Host header inserter that wraps the next client.
// It also inserts other headers as specified in the list of key/value pairs.
func Wrap(next httpclient.HttpClient, headerKeyVals ...string) httpclient.HttpClient {
	hc := &hh{inner: next}
	if len(headerKeyVals) > 1 {
		hc.metadata = make(map[string]string)
		for i := 1; i < len(headerKeyVals); i += 2 {
			hc.metadata[headerKeyVals[i-1]] = headerKeyVals[i]
		}
	}
	return hc
}

func (hh *hh) Do(req *http.Request) (*http.Response, error) {
	if req.Header.Get("Host") == "" {
		req.Header.Set("Host", req.URL.Host)
	}
	for h, v := range hh.metadata {
		req.Header.Set(h, v)
	}
	return hh.inner.Do(req)
}

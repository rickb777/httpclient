package rest

import (
	"crypto/tls"
	"net/http"
	"time"

	"github.com/rickb777/acceptable/contenttype"
	"github.com/rickb777/acceptable/headername"
	"github.com/rickb777/httpclient"
	"github.com/rickb777/httpclient/hostheader"
)

type HttpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

//-------------------------------------------------------------------------------------------------

// Headers builds a http.Header using the pairs of key-value strings.
func Headers(headerKeyVals ...string) http.Header {
	h := make(http.Header)
	for i := 1; i < len(headerKeyVals); i += 2 {
		h.Add(headerKeyVals[i-1], headerKeyVals[i])
	}
	return h
}

//-------------------------------------------------------------------------------------------------

var DefaultTransport http.RoundTripper

// DefaultClient gets a http.Client with the DefaultTransport as its round-tripper
// and a specified timeout. A timeout of zero means no timeout.
func DefaultClient(timeout time.Duration) httpclient.HttpClient {
	if timeout < 0 {
		panic(timeout)
	}
	return hostheader.Wrap(&http.Client{
		Transport: DefaultTransport,
		Timeout:   timeout,
	})
}

// RESTClient returns a client that sets request headers on every request.
//
//   - Host: (from URL)
//   - Accept: application/json
func RESTClient(timeout time.Duration) httpclient.HttpClient {
	return hostheader.Wrap(DefaultClient(timeout),
		headername.Accept, contenttype.ApplicationJSON)
}

// SetDefaultTLSConfig sets the TLS configuration used by the default transport.
func SetDefaultTLSConfig(cfg *tls.Config) {
	DefaultTransport.(*http.Transport).TLSClientConfig = cfg
}

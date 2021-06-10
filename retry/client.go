package retry

import (
	"github.com/rickb777/httpclient"
	"github.com/rs/zerolog"
	"net/http"
)

// retry is an HTTP client decorator that retries each request on network error.
type retry struct {
	inner httpclient.HttpClient
	cfg   RetryConfig
	lgr   zerolog.Logger
}

// New creates an HTTP client decorator that retries each request on network error.
func New(inner httpclient.HttpClient, cfg RetryConfig, lgr zerolog.Logger) httpclient.HttpClient {
	return &retry{
		inner: inner,
		cfg:   cfg,
		lgr:   lgr,
	}
}

func (r *retry) Do(req *http.Request) (response *http.Response, err error) {
	err = NewExponentialBackOff(r.cfg, req.URL.String(), r.lgr,
		func() error {
			response, err = r.inner.Do(req)
			return err
		})
	return response, err
}

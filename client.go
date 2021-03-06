package httpclient

import "net/http"

// HttpClient indicates the core function in http.Client, allowing features
// to be nested easily.
type HttpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type ControlledRedirectClient interface {
	HttpClient
	SetCheckRedirect(func(req *http.Request, via []*http.Request) error)
}

// DefaultClient is the http.Client zero-valued default.
var DefaultClient HttpClient = &http.Client{}

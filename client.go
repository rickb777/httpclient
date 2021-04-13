package httpclient

import "net/http"

// HttpClient indicates the core function in http.Client, allowing features
// to be nested easily.
type HttpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

package rest

import (
	"context"
	"net/http"
	"net/http/cookiejar"
	"os"
	"sync"

	"github.com/rickb777/httpclient"
	authpkg "github.com/rickb777/httpclient/auth"
	bodypkg "github.com/rickb777/httpclient/body"
	"golang.org/x/net/publicsuffix"
)

type RestClient interface {
	Request(ctx context.Context, method, path string, reqBody any, opts ...ReqOpt) (*http.Response, error)
	Head(ctx context.Context, path string, opts ...ReqOpt) (*Response, error)
	Get(ctx context.Context, path string, opts ...ReqOpt) (*Response, error)
	Put(ctx context.Context, path string, reqBody any, opts ...ReqOpt) (*Response, error)
	Post(ctx context.Context, path string, reqBody any, opts ...ReqOpt) (*Response, error)
	Delete(ctx context.Context, path string, reqBody any, opts ...ReqOpt) (*Response, error)
	ClearCookies()
}

// Response holds an HTTP response with the entity in a buffer.
type Response struct {
	// StatusCode the HTTP status code
	StatusCode int
	// Header the response headers
	Header http.Header
	// Body the buffered response entity
	Body *bodypkg.Body
	// Request the original request
	Request *http.Request
}

// client defines our structure
type client struct {
	root      string
	headers   http.Header
	hc        httpclient.HttpClient
	authMutex sync.Mutex
	auth      authpkg.Authenticator
	cookies   *cookiejar.Jar
}

//-------------------------------------------------------------------------------------------------

// NewClient creates a new Client. By default, this uses the default HTTP client.
func NewClient(uri string, opts ...ClientOpt) RestClient {
	cl := &client{
		root:    withoutTrailingSlash(uri),
		headers: make(http.Header),
		hc:      http.DefaultClient,
		auth:    authpkg.Anonymous,
	}
	cl.ClearCookies()

	for _, opt := range opts {
		opt(cl)
	}
	return cl
}

//-------------------------------------------------------------------------------------------------

type ClientOpt func(RestClient)

// AddHeader sets a request header that will be applied to all subsequent requests.
func AddHeader(key, value string) ClientOpt {
	return func(c RestClient) {
		c.(*client).headers.Add(key, value)
	}
}

// SetAuthentication sets the authentication credentials and method.
// Leave the authenticator method blank to allow HTTP challenges to
// select an appropriate method. Otherwise it should be "basic".
func SetAuthentication(authenticator authpkg.Authenticator) ClientOpt {
	return func(c RestClient) {
		c.(*client).auth = authenticator
	}
}

// SetHttpClient changes the http.Client. This allows control over
// the http.Transport, timeouts etc.
func SetHttpClient(httpClient httpclient.HttpClient) ClientOpt {
	return func(c RestClient) {
		c.(*client).hc = httpClient
	}
}

//-------------------------------------------------------------------------------------------------

func (c *client) ClearCookies() {
	jar, err := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
	if err != nil {
		log(err)
		os.Exit(1)
	}
	c.cookies = jar
}

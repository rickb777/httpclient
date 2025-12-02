package rest

import (
	"context"
	"net/http"
	"net/http/cookiejar"
	"os"
	"sync"

	"github.com/rickb777/acceptable/header"
	"github.com/rickb777/httpclient"
	authpkg "github.com/rickb777/httpclient/auth"
	bodypkg "github.com/rickb777/httpclient/body"
	"golang.org/x/net/publicsuffix"
)

// RestClient implements HTTP client requests as typically used for REST APIs etc.
//
// The Request method returns a [http.Response] that may contain a body as an [io.ReadCloser], which can
// be handled appropriately by the caller. The caller must close this body (if the response is not nil).
//
// The Head, Get, Put, Post, and Delete methods return a [Response] containing a buffered body that is
// simpler to use but potentially less performant for large bodies.
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
	// Type the content type of the response entity
	Type header.ContentType
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

// ClientOpt functions configure the client.
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

// ClearCookies drops any existing cookies.
func (c *client) ClearCookies() {
	jar, err := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
	if err != nil {
		log(err)
		os.Exit(1)
	}
	c.cookies = jar
}

package rest

import (
	"context"
	"fmt"
	"io"
	"net/http"
	urlpkg "net/url"
	"strconv"
	"strings"

	. "github.com/rickb777/acceptable/contenttype"
	"github.com/rickb777/acceptable/header"
	. "github.com/rickb777/acceptable/headername"
	authpkg "github.com/rickb777/httpclient/auth"
	bodypkg "github.com/rickb777/httpclient/body"
	"github.com/rickb777/httpclient/rest/temperror"
)

// ReqOpt optionally amends or enhances a request before it is sent.
type ReqOpt func(*http.Request)

// Headers adds header values to the request.
// kv is a list of key & value pairs.
func Headers(kv ...string) ReqOpt {
	return func(req *http.Request) {
		for i := 1; i < len(kv); i += 2 {
			req.Header.Add(kv[i-1], kv[i])
		}
	}
}

// QueryKV adds query values to the request.
// kv is a list of key & value pairs.
func QueryKV(kv ...string) ReqOpt {
	v := make(urlpkg.Values)
	for i := 1; i < len(kv); i += 2 {
		v.Add(kv[i-1], kv[i])
	}
	return Query(v)
}

// Query sets the query parameters on the request. Existing query parameters are replaced.
func Query(query urlpkg.Values) ReqOpt {
	return func(req *http.Request) {
		req.URL.RawQuery = query.Encode()
	}
}

// Request performs one or more round-trip HTTP requests, attempting to satisfy the authentication challenge
// if one is received. The bare *http.Response is returned; this contains the response entity as io.ReadCloser,
// which is useful if this is to be read as a stream. If res is not nil, the response body must be closed by
// the caller.
func (c *client) Request(ctx context.Context, method, path string, reqBody any, opts ...ReqOpt) (res *http.Response, err error) {
	return c.request(ctx, 1, method, path, reqBody, opts...)
}

// Request performs one or more round-trip HTTP requests, attempting to satisfy the authentication challenge
// if one is received. The bare *http.Response is returned; this contains the response entity as io.ReadCloser,
// which is useful if this is to be read as a stream. If res is not nil, the response body must be closed by
// the caller.
func (c *client) request(ctx context.Context, depth int, method, path string, reqBody any, opts ...ReqOpt) (res *http.Response, err error) {
	var req *http.Request
	// Buffer the body because, if authorization fails, we will need to read from it again.
	bodyBuf, hdrs, err := processRequestEntity(reqBody)
	if err != nil {
		return nil, err
	}

	u := c.root + withLeadingSlash(path) // TODO pathEscape removed here
	req, err = http.NewRequestWithContext(ctx, method, u, bodyBuf)

	if err != nil {
		return nil, err
	}

	cs := c.cookies.Cookies(req.URL)
	for _, c := range cs {
		req.Header.Add("Cookie", c.String())
	}

	// client-scoped headers
	for k, vals := range c.headers {
		for _, v := range vals {
			req.Header.Add(k, v)
		}
	}

	// request-scoped headers
	for _, opt := range opts {
		opt(req)
	}

	// headers determined by the request entity
	for k, vs := range hdrs {
		req.Header[k] = vs
	}

	// Make sure we read 'c.auth' only once because it may be substituted below,
	// which is unsafe to do when multiple goroutines are running at the same time.
	c.authMutex.Lock()
	auth := c.auth // make a duplicate
	c.authMutex.Unlock()

	// authentication headers
	auth.Authorize(req)

	res, err = c.hc.Do(req)
	if err != nil {
		return nil, err
	}

	if res.StatusCode == http.StatusUnauthorized && auth.Type() == "noAuth" {
		if depth > 3 {
			return nil, &RestError{Code: http.StatusUnauthorized, Request: req, Cause: fmt.Errorf("too many authentication retries")}
		}
		return c.repeat(ctx, depth, res, method, path, bodyBuf, opts...)
	} else if res.StatusCode == http.StatusUnauthorized {
		return res, newPathError("Authorize", c.root, res.StatusCode)
	}

	c.cookies.SetCookies(req.URL, res.Cookies())
	// note that res.Body is not yet closed

	return res, err
}

//-------------------------------------------------------------------------------------------------

func (c *client) repeat(ctx context.Context, depth int, res *http.Response, method, path string, body *bodypkg.Body, opts ...ReqOpt) (req *http.Response, err error) {
	wwwAuthenticateHeader := res.Header.Get("Www-Authenticate")
	wwwAuthenticateHeaderLC := strings.ToLower(wwwAuthenticateHeader)

	c.authMutex.Lock()
	auth := c.auth
	c.authMutex.Unlock()

	if strings.Contains(wwwAuthenticateHeaderLC, "digest") {
		c.authMutex.Lock()
		c.auth = authpkg.Digest(auth.User(), auth.Password()).DigestParts(wwwAuthenticateHeader)
		c.authMutex.Unlock()
	} else if strings.Contains(wwwAuthenticateHeaderLC, "basic") {
		c.authMutex.Lock()
		c.auth = authpkg.Basic(auth.User(), auth.Password())
		c.authMutex.Unlock()
	} else {
		return res, newPathError("Authorize", c.root, res.StatusCode)
	}

	_ = res.Body.Close()

	body.Rewind()
	return c.request(ctx, depth+1, method, path, body, opts...)
}

//-------------------------------------------------------------------------------------------------

func processRequestEntity(input any) (requestBody *bodypkg.Body, hdr http.Header, err error) {
	m := make(http.Header)
	m.Set(Accept, ApplicationJSON)
	m.Set(AcceptEncoding, "identity")

	switch data := input.(type) {
	case nil:
	case urlpkg.Values:
		m.Set(ContentType, ApplicationForm)
		requestBody = bodypkg.NewBodyString(data.Encode())
	case string:
		m.Set(ContentType, ApplicationJSON)
		requestBody = bodypkg.NewBodyString(data)
	case *string:
		m.Set(ContentType, ApplicationJSON)
		requestBody = bodypkg.NewBodyString(*data)
	case []byte:
		// must set earlier: m.Set(headername.ContentType, ...)
		m.Set(ContentLength, strconv.Itoa(len(data)))
		requestBody = bodypkg.NewBody(data)
	case io.Reader:
		// must set earlier: m.Set(headername.ContentType, ...)
		requestBody, err = bodypkg.Copy(data)
	case *bodypkg.Body:
		requestBody = data
	case ReqOpt:
		panic("ReqOpt passed instead of a body - did you mean this?")
	default:
		rb, err := bodypkg.JsonMarshalToString(data)
		if err != nil {
			panic(err)
		}
		rb += "\n" // required for Posix compliance
		requestBody = bodypkg.NewBodyString(rb)
		m.Set(ContentType, ApplicationJSON)
	}

	if requestBody != nil {
		m.Set(ContentLength, strconv.Itoa(len(requestBody.Bytes())))
	}

	return requestBody, m, err
}

//-------------------------------------------------------------------------------------------------

// Head performs a HEAD request. The response body is always empty.
func (c *client) Head(ctx context.Context, path string, opts ...ReqOpt) (response *Response, err error) {
	return responseOf(c.Request(ctx, http.MethodHead, path, nil, opts...))
}

//-------------------------------------------------------------------------------------------------

func (c *client) Get(ctx context.Context, path string, opts ...ReqOpt) (response *Response, err error) {
	return responseOf(c.Request(ctx, http.MethodGet, path, nil, opts...))
}

//-------------------------------------------------------------------------------------------------

func (c *client) Post(ctx context.Context, path string, reqBody any, opts ...ReqOpt) (response *Response, err error) {
	return responseOf(c.Request(ctx, http.MethodPost, path, reqBody, opts...))
}

//-------------------------------------------------------------------------------------------------

func (c *client) Put(ctx context.Context, path string, reqBody any, opts ...ReqOpt) (response *Response, err error) {
	return responseOf(c.Request(ctx, http.MethodPut, path, reqBody, opts...))
}

//-------------------------------------------------------------------------------------------------

func (c *client) Delete(ctx context.Context, path string, reqBody any, opts ...ReqOpt) (response *Response, err error) {
	return responseOf(c.Request(ctx, http.MethodDelete, path, reqBody, opts...))
}

//-------------------------------------------------------------------------------------------------

func responseOf(res *http.Response, err error) (*Response, error) {
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()
	body, err := bodypkg.Copy(res.Body)

	ct := header.ParseContentType(res.Header.Get(ContentType))
	delete(res.Header, ContentType)

	r := &Response{
		StatusCode: res.StatusCode,
		Request:    res.Request,
		Header:     res.Header,
		Type:       ct,
		Body:       body,
	}

	if res.StatusCode >= 400 {
		err = &RestError{
			Code:         res.StatusCode,
			Request:      res.Request,
			ResponseType: ct,
			Response:     body,
		}
	}

	if res.StatusCode >= 500 {
		err = temperror.Wrap(err)
	}

	return r, err
}

package rest

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	urlpkg "net/url"
	"strconv"
	"strings"

	. "github.com/rickb777/acceptable/contenttype"
	"github.com/rickb777/acceptable/header"
	"github.com/rickb777/acceptable/headername"
	authpkg "github.com/rickb777/httpclient/auth"
	bodypkg "github.com/rickb777/httpclient/body"
	"github.com/rickb777/httpclient/rest/temperror"
)

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

	// set the authentication headers
	auth.Authenticate(req)

	res, err = c.hc.Do(req)
	if err != nil {
		return nil, err
	}

	if res.StatusCode == http.StatusUnauthorized && auth.Type() == "noAuth" {
		if depth > 3 {
			r2, e2 := copyResponse(res, nil)
			return nil, &RestError{
				Response: r2,
				Cause:    errors.Join(e2, fmt.Errorf("too many authentication retries")),
			}
		}
		return c.repeat(ctx, depth, res, method, path, bodyBuf, opts...)
	} else if res.StatusCode == http.StatusUnauthorized {
		return res, newPathError("Authorize", req.URL.Path, res.StatusCode)
	}

	c.cookies.SetCookies(req.URL, res.Cookies())
	// note that res.Body is not yet closed

	return res, nil
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
	m.Set(headername.Accept, ApplicationJSON)
	m.Set(headername.AcceptEncoding, "identity")

	switch data := input.(type) {
	case nil:
	case urlpkg.Values:
		m.Set(headername.ContentType, ApplicationForm)
		requestBody = bodypkg.NewBodyString(data.Encode())
	case string:
		m.Set(headername.ContentType, ApplicationJSON)
		requestBody = bodypkg.NewBodyString(data)
	case *string:
		m.Set(headername.ContentType, ApplicationJSON)
		requestBody = bodypkg.NewBodyString(*data)
	case []byte:
		// must set earlier: m.Set(headername.ContentType, ...)
		m.Set(headername.ContentLength, strconv.Itoa(len(data)))
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
		m.Set(headername.ContentType, ApplicationJSON)
	}

	if requestBody != nil {
		m.Set(headername.ContentLength, strconv.Itoa(len(requestBody.Bytes())))
	}

	return requestBody, m, err
}

//-------------------------------------------------------------------------------------------------

// Head performs a HEAD request. The response body is always empty.
func (c *client) Head(ctx context.Context, path string, opts ...ReqOpt) (response *Response, err error) {
	return responseOf(c.Request(ctx, http.MethodHead, path, nil, opts...))
}

//-------------------------------------------------------------------------------------------------

// Get performs a GET request.
func (c *client) Get(ctx context.Context, path string, opts ...ReqOpt) (response *Response, err error) {
	return responseOf(c.Request(ctx, http.MethodGet, path, nil, opts...))
}

//-------------------------------------------------------------------------------------------------

// Post performs a POST request using the request body supplied, which can be nil.
func (c *client) Post(ctx context.Context, path string, reqBody any, opts ...ReqOpt) (response *Response, err error) {
	return responseOf(c.Request(ctx, http.MethodPost, path, reqBody, opts...))
}

//-------------------------------------------------------------------------------------------------

// Put performs a PUT request using the request body supplied, which can be nil.
func (c *client) Put(ctx context.Context, path string, reqBody any, opts ...ReqOpt) (response *Response, err error) {
	return responseOf(c.Request(ctx, http.MethodPut, path, reqBody, opts...))
}

//-------------------------------------------------------------------------------------------------

// Delete performs a DELETE request. The request body can be supplied but should normally be nil (see RFC-9110).
func (c *client) Delete(ctx context.Context, path string, reqBody any, opts ...ReqOpt) (response *Response, err error) {
	return responseOf(c.Request(ctx, http.MethodDelete, path, reqBody, opts...))
}

//-------------------------------------------------------------------------------------------------

func copyResponse(res *http.Response, err error) (Response, error) {
	var r Response

	// err might be nil, but we defer handling it so that anything available in the
	// response will also be available in the RESTError.

	if res != nil {
		defer res.Body.Close()
		body, e2 := bodypkg.Copy(res.Body)
		if e2 != nil {
			return Response{}, &RestError{
				Cause: errors.Join(err, e2),
			}
		}

		ct := header.ParseContentType(res.Header.Get(headername.ContentType))
		delete(res.Header, headername.ContentType)

		r = Response{
			StatusCode: res.StatusCode,
			Request:    res.Request,
			Header:     res.Header,
			Type:       ct,
			Body:       body,
		}
	}

	return r, err
}

//-------------------------------------------------------------------------------------------------

func responseOf(res *http.Response, err error) (*Response, error) {
	r, err2 := copyResponse(res, err)

	// err might be nil, but we defer handling it so that anything available in the
	// response will also be available in the RESTError.

	var httpStatusCodeError error

	if r.StatusCode >= 400 {
		httpStatusCodeError = &RestError{
			Response: r,
		}
	}

	if r.StatusCode >= 500 {
		httpStatusCodeError = temperror.Wrap(httpStatusCodeError)
	}

	return &r, errors.Join(err2, httpStatusCodeError)
}

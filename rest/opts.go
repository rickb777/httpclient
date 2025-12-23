package rest

import (
	"net/http"
	urlpkg "net/url"
	"strings"
	"time"

	"github.com/rickb777/acceptable/header"
	hdr "github.com/rickb777/acceptable/headername"
)

// ReqOpt optionally amends or enhances a request before it is sent.
type ReqOpt func(*http.Request)

// Query adds query values to the request.
// kv is a list of key & value pairs.
func Query(kv ...string) ReqOpt {
	return func(req *http.Request) {
		buf := &strings.Builder{}
		buf.WriteString(req.URL.RawQuery)
		sep := ""
		if buf.Len() > 0 {
			sep = "&"
		}
		for i := 1; i < len(kv); i += 2 {
			buf.WriteString(sep)
			buf.WriteString(urlpkg.QueryEscape(kv[i-1]))
			buf.WriteString("=")
			buf.WriteString(urlpkg.QueryEscape(kv[i]))
			sep = "&"
		}
		req.URL.RawQuery = buf.String()
	}
}

// Headers sets more header values on the request. This overwrites any
// pre-existing headers where they have the same names.
func Headers(more http.Header) ReqOpt {
	return func(req *http.Request) {
		for k, vs := range more {
			req.Header[k] = vs
		}
	}
}

// HeadersKV adds header values to the request.
// kv is a list of key & value pairs.
// Pre-existing headers are added to.
func HeadersKV(kv ...string) ReqOpt {
	return func(req *http.Request) {
		for i := 1; i < len(kv); i += 2 {
			req.Header.Add(kv[i-1], kv[i])
		}
	}
}

// IfModifiedSince makes a request conditional upon change history and is typically used for GET requests.
func IfModifiedSince(t time.Time) ReqOpt {
	return func(req *http.Request) {
		req.Header.Add(hdr.IfModifiedSince, header.FormatHTTPDateTime(t))
	}
}

// IfNoneMatch makes a request conditional upon ETags and is typically used for GET requests.
func IfNoneMatch(etag ...header.ETag) ReqOpt {
	return func(req *http.Request) {
		req.Header.Add(hdr.IfNoneMatch, header.ETags(etag).String())
	}
}

// IfMatch makes a request conditional upon ETags and is typically used for PUT, POST and DELETE requests.
// There is also an If-Unmodified-Since header, but If-Match takes precedence (see RFC-9110).
func IfMatch(etag ...header.ETag) ReqOpt {
	return func(req *http.Request) {
		req.Header.Add(hdr.IfMatch, header.ETags(etag).String())
	}
}

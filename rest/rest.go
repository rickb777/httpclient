package rest

import (
	"context"
	"errors"
	"io"
	"mime"
	"net/http"
	urlpkg "net/url"
	"strconv"

	. "github.com/rickb777/acceptable/contenttype"
	"github.com/rickb777/acceptable/header"
	. "github.com/rickb777/acceptable/headername"
	"github.com/rickb777/httpclient"
	bodypkg "github.com/rickb777/httpclient/body"
	"github.com/rickb777/httpclient/rest/temperror"
)

// Request constructs a request using a URL. The request will not have an entity.
func Request(ctx context.Context, url string) *RESTRequest {
	return Entity(ctx, url, nil, "")
}

// JSON constructs a request using a URL. The request will have a JSON entity.
func JSON(ctx context.Context, url string, input any) *RESTRequest {
	return Entity(ctx, url, input, ApplicationJSON)
}

// Entity constructs a request using a URL. The request will have an arbitrary entity.
func Entity(ctx context.Context, url string, input any, contentType string) *RESTRequest {
	rbr, m2 := processRequestEntity(input)

	req, err := http.NewRequestWithContext(ctx, "", url, rbr)
	if err != nil {
		return &RESTRequest{err: err}
	}

	mergeHeaders(req.Header, m2)

	if contentType != "" {
		req.Header.Set(ContentType, contentType)
	}

	return &RESTRequest{req, nil}
}

// Query adds query parameters to the request, as if a "?key=value" list had been
// attached to the URL.
func (rr *RESTRequest) Query(query urlpkg.Values) *RESTRequest {
	rr.req.URL.RawQuery = query.Encode()
	return rr
}

// With adds headers to the request.
func (rr *RESTRequest) With(m http.Header) *RESTRequest {
	mergeHeaders(rr.req.Header, m)
	return rr
}

// WithKV adds key-value pairs as headers to the request.
func (rr *RESTRequest) WithKV(headerKeyVals ...string) *RESTRequest {
	return rr.With(Headers(headerKeyVals...))
}

// Head performs an HTTP HEAD request.
// The response headers and status code are returned.
// In order to close resources correctly, it is essential to call one of
// [RESTResponse.Status], [RESTResponse.Unmarshal] or [RESTResponse.ToMap]
func (rr *RESTRequest) Head(httpClient httpclient.HttpClient) *RESTResponse {
	return rr.HTTPRoundTrip(http.MethodHead, httpClient)
}

// Get performs an HTTP GET request.
// The response headers and status code are returned.
// In order to close resources correctly, it is essential to call one of
// [RESTResponse.Status], [RESTResponse.Unmarshal] or [RESTResponse.ToMap]
func (rr *RESTRequest) Get(httpClient httpclient.HttpClient) *RESTResponse {
	return rr.HTTPRoundTrip(http.MethodGet, httpClient)
}

// Post performs an HTTP POST request.
// The response headers and status code are returned.
// In order to close resources correctly, it is essential to call one of
// [RESTResponse.Status], [RESTResponse.Unmarshal] or [RESTResponse.ToMap]
func (rr *RESTRequest) Post(httpClient httpclient.HttpClient) *RESTResponse {
	return rr.HTTPRoundTrip(http.MethodPost, httpClient)
}

// Put performs an HTTP PUT request.
// The response headers and status code are returned.
// In order to close resources correctly, it is essential to call one of
// [RESTResponse.Status], [RESTResponse.Unmarshal] or [RESTResponse.ToMap]
func (rr *RESTRequest) Put(httpClient httpclient.HttpClient) *RESTResponse {
	return rr.HTTPRoundTrip(http.MethodPut, httpClient)
}

// Patch performs an HTTP PATCH request.
// The response headers and status code are returned.
// In order to close resources correctly, it is essential to call one of
// [RESTResponse.Status], [RESTResponse.Unmarshal] or [RESTResponse.ToMap]
func (rr *RESTRequest) Patch(httpClient httpclient.HttpClient) *RESTResponse {
	return rr.HTTPRoundTrip(http.MethodPatch, httpClient)
}

// Delete performs an HTTP DELETE request.
// The response headers and status code are returned.
// In order to close resources correctly, it is essential to call one of
// [RESTResponse.Status], [RESTResponse.Unmarshal] or [RESTResponse.ToMap]
func (rr *RESTRequest) Delete(httpClient httpclient.HttpClient) *RESTResponse {
	return rr.HTTPRoundTrip(http.MethodDelete, httpClient)
}

//-------------------------------------------------------------------------------------------------

func processRequestEntity(input any) (io.Reader, http.Header) {
	m := make(http.Header)
	m.Set(Accept, ApplicationJSON)
	m.Set(AcceptEncoding, "identity")

	var requestBody *bodypkg.Body

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
		return bodypkg.NewBody(data), m
	case io.Reader:
		// must set earlier: m.Set(headername.ContentType, ...)
		return data, m
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

	return requestBody, m
}

//-------------------------------------------------------------------------------------------------

type RESTRequest struct {
	req *http.Request
	err error
}

// HTTPRoundTrip performs an arbitrary HTTP request.
// The response headers and status code are returned.
// In order to close resources correctly, it is essential to call one of
// [RESTResponse.Status], [RESTResponse.Unmarshal] or [RESTResponse.ToMap]
func (rr *RESTRequest) HTTPRoundTrip(method string, httpClient httpclient.HttpClient) *RESTResponse {
	if rr.err != nil {
		return &RESTResponse{Err: rr.err}
	}

	rr.req.Method = method

	res, err := httpClient.Do(rr.req)
	if err != nil {
		return &RESTResponse{Err: checkDoError(err)}
	}

	body := res.Body

	if res.StatusCode >= 300 {
		defer res.Body.Close()
		body, err = bodypkg.Copy(res.Body)
		re := &RESTError{
			cause:        err,
			Code:         res.StatusCode,
			Request:      rr.req,
			ResponseType: header.ParseContentTypeFromHeaders(res.Header),
			Response:     body.(*bodypkg.Body),
		}
		if res.StatusCode >= 500 {
			err = temperror.Wrap(re)
		} else {
			err = re
		}
	}

	return &RESTResponse{Req: rr.req, Res: res, Body: body, Err: err}
}

func checkDoError(err error) error {
	if temperror.NetworkConnectionError(err) {
		return temperror.Wrap(err)
	}
	return err
}

//-------------------------------------------------------------------------------------------------

type RESTResponse struct {
	Req  *http.Request
	Res  *http.Response
	Body io.ReadCloser
	Err  error
}

// ToMap extracts a map containing a tree of data from the JSON response.
// If the response does not contain JSON, the map will be nil.
func (rr *RESTResponse) ToMap() (data map[string]any, statusCode int, err error) {
	var output map[string]any
	_, code, err := rr.Unmarshal(&output)
	return output, code, err
}

// Status gets the status of the response.
func (rr *RESTResponse) Status() (respHeader http.Header, statusCode int, err error) {
	return rr.Unmarshal(nil)
}

// Unmarshal extracts data from the response. Typically, when the response contains JSON,
// output will be a pointer to a struct that matches the expected content. However,
// it may also be *string or *[]byte. See [RESTResponse.ToMap].
func (rr *RESTResponse) Unmarshal(output any) (respHeader http.Header, statusCode int, err error) {
	if rr.Res == nil {
		return nil, 0, rr.Err
	}

	if _, cachedAlready := rr.Body.(*bodypkg.Body); !cachedAlready {
		defer rr.Res.Body.Close()
	}

	if output == nil || rr.Res.StatusCode == http.StatusNoContent {
		return rr.Res.Header, rr.Res.StatusCode, rr.Err
	}

	switch data := output.(type) {
	case *string:
		entity, err := bodypkg.Copy(rr.Body)
		if err == nil {
			*data = entity.String()
		}
		return rr.Res.Header, rr.Res.StatusCode, errors.Join(rr.Err, err)

	case *[]byte:
		entity, err := bodypkg.Copy(rr.Body)
		if err == nil {
			*data = entity.Bytes()
		}
		return rr.Res.Header, rr.Res.StatusCode, errors.Join(rr.Err, err)

	default:
		if isContentType(rr.Res, ApplicationJSON) {
			err := bodypkg.JsonUnmarshal(rr.Body, output)
			return rr.Res.Header, rr.Res.StatusCode, errors.Join(rr.Err, err)
		}
	}

	return rr.Res.Header, http.StatusNotAcceptable, errors.Join(rr.Err, rr.notAcceptable())
}

func (rr *RESTResponse) notAcceptable() error {
	return &RESTError{
		Code:         http.StatusNotAcceptable,
		Request:      rr.Req,
		ResponseType: header.ParseContentTypeFromHeaders(rr.Res.Header),
	}
}

func isContentType(res *http.Response, expected string) bool {
	return bareContentType(res.Header) == expected
}

func bareContentType(hdrs http.Header) string {
	ct := hdrs.Get(ContentType)
	if ct == "" {
		return ""
	}
	lct, _, _ := mime.ParseMediaType(ct)
	return lct
}

//-------------------------------------------------------------------------------------------------

func mergeHeaders(reqHeader http.Header, m http.Header) {
	for k, v := range m {
		reqHeader[k] = v
	}
}

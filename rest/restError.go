package rest

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/rickb777/httpclient/rest/temperror"
)

type Body interface {
	Bytes() []byte
	fmt.Stringer
	io.Reader
}

type RestError struct {
	Response
	Cause error
}

// IsPermanent returns the opposite of [RestError.IsTransient]; the request should not be retried.
func (re *RestError) IsPermanent() bool {
	return !re.IsTransient()
}

// IsTransient returns true for server/network errors that can be retried.
// (Note that 401 authentication challenges should be responded to normally).
func (re *RestError) IsTransient() bool {
	switch re.Response.StatusCode {
	case http.StatusInternalServerError, http.StatusBadGateway, http.StatusServiceUnavailable, http.StatusGatewayTimeout:
		return true
	}
	if re.Cause != nil && temperror.NetworkConnectionError(re.Cause) {
		return true
	}
	return false
}

// Error makes it compatible with `error` interface.
func (re *RestError) Error() string {
	if 300 <= re.Response.StatusCode && re.Response.StatusCode < 400 {
		return fmt.Sprintf(`%d: %s %s %s %s`, re.StatusCode, re.Request.Method, re.Request.URL,
			strings.ToLower(http.StatusText(re.Response.StatusCode)), re.Header.Get("Location"))
	}
	if re.Type.MediaType == "" {
		return fmt.Sprintf(`%d: %s %s`, re.StatusCode, re.Request.Method, re.Request.URL)
	}
	if re.Type.IsTextual() {
		b := strings.TrimSpace(re.Response.Body.String())
		if len(b) > RESTErrorStringLimit {
			b = b[:RESTErrorStringLimit] + "..."
		}
		return fmt.Sprintf(`%d: %s %s %s %s`, re.StatusCode, re.Request.Method, re.Request.URL, re.Type, b)
	}
	return fmt.Sprintf(`%d: %s %s %s`, re.StatusCode, re.Request.Method, re.Request.URL, re.Type)
}

func (re *RestError) Unwrap() error {
	return re.Cause
}

//func (re *RestError) UnmarshalJSONResponse(value any) error {
//	if re.Response.Body == nil {
//		return nil
//	}
//	return body.JsonUnmarshal(re.Response.Body, value)
//}

var RESTErrorStringLimit = 100

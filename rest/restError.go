package rest

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/rickb777/acceptable/header"
	"github.com/rickb777/httpclient/body"
)

type Body interface {
	Bytes() []byte
	fmt.Stringer
	io.Reader
}

type RestError struct {
	Cause        error
	Code         int
	Request      *http.Request
	ResponseType header.ContentType
	Response     Body
}

// Error makes it compatible with `error` interface.
func (re *RestError) Error() string {
	if re.ResponseType.MediaType == "" {
		return fmt.Sprintf(`%d: %s %s`, re.Code, re.Request.Method, re.Request.URL)
	}
	if re.ResponseType.IsTextual() {
		b := strings.TrimSpace(re.Response.String())
		if len(b) > RESTErrorStringLimit {
			b = b[:RESTErrorStringLimit] + "..."
		}
		return fmt.Sprintf(`%d: %s %s %s %s`, re.Code, re.Request.Method, re.Request.URL, re.ResponseType, b)
	}
	return fmt.Sprintf(`%d: %s %s %s`, re.Code, re.Request.Method, re.Request.URL, re.ResponseType)
}

func (re *RestError) Unwrap() error {
	return re.Cause
}

func (re *RestError) UnmarshalJSONResponse(value any) error {
	return body.JsonUnmarshal(re.Response, value)
}

var RESTErrorStringLimit = 250

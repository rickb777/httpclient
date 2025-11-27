package rest

import (
	"fmt"
	"io"
	"strings"

	"github.com/rickb777/httpclient/body"
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

// Error makes it compatible with `error` interface.
func (re *RestError) Error() string {
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

func (re *RestError) UnmarshalJSONResponse(value any) error {
	if re.Response.Body == nil {
		return nil
	}
	return body.JsonUnmarshal(re.Response.Body, value)
}

var RESTErrorStringLimit = 250

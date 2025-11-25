package rest

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/rickb777/acceptable/header"
)

type Body interface {
	Bytes() []byte
	fmt.Stringer
	io.Reader
}

type RESTError struct {
	cause        error
	Code         int
	Request      *http.Request
	ResponseType header.ContentType
	Response     Body
}

// Error makes it compatible with `error` interface.
func (re *RESTError) Error() string {
	if re.ResponseType.Type == "" {
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

func (re *RESTError) Unwrap() error {
	return re.cause
}

func (re *RESTError) UnmarshalJSONResponse(value any) error {
	return JsonUnmarshal(re.Response, value)
}

var RESTErrorStringLimit = 250

package rest

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rickb777/acceptable/header"
	"github.com/rickb777/expect"
	"github.com/rickb777/httpclient/body"
)

func TestRESTError_Error(t *testing.T) {
	cases := map[string]*RestError{
		"404: GET http://localhost/foo": {
			Code:     404,
			Request:  httptest.NewRequest(http.MethodGet, "http://localhost/foo", nil),
			Response: body.NewBodyString("foo"),
		},
		"404: GET http://localhost/foo text/plain foo": {
			Code:         404,
			Request:      httptest.NewRequest(http.MethodGet, "http://localhost/foo", nil),
			ResponseType: header.ContentType{MediaType: "text/plain"},
			Response:     body.NewBodyString("foo"),
		},
		"404: GET http://localhost/foo image/png": {
			Code:         404,
			Request:      httptest.NewRequest(http.MethodGet, "http://localhost/foo", nil),
			ResponseType: header.ContentType{MediaType: "image/png"},
			Response:     body.NewBodyString("foo"),
		},
	}

	for expected, input := range cases {
		expect.String(input.Error()).ToBe(t, expected)
	}
}

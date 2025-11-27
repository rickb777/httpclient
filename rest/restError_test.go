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
			Response: Response{
				StatusCode: 404,
				Request:    httptest.NewRequest(http.MethodGet, "http://localhost/foo", nil),
				Body:       body.NewBodyString("foo"),
			},
		},
		"404: GET http://localhost/foo text/plain foo": {
			Response: Response{
				StatusCode: 404,
				Request:    httptest.NewRequest(http.MethodGet, "http://localhost/foo", nil),
				Type:       header.ContentType{MediaType: "text/plain"},
				Body:       body.NewBodyString("foo"),
			},
		},
		"404: GET http://localhost/foo image/png": {
			Response: Response{
				StatusCode: 404,
				Request:    httptest.NewRequest(http.MethodGet, "http://localhost/foo", nil),
				Type:       header.ContentType{MediaType: "image/png"},
				Body:       body.NewBodyString("foo"),
			},
		},
	}

	for expected, input := range cases {
		expect.String(input.Error()).ToBe(t, expected)
	}
}

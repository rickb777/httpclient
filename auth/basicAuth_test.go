package auth

import (
	"net/http/httptest"
	"testing"

	"github.com/rickb777/expect"
)

func TestBasic_Authorize(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)

	Basic("user", "password").Challenge([]string{`Basic realm="WallyWorld"`}).Authorize(req)

	expect.String(req.Header.Get("Authorization")).ToBe(t, "Basic dXNlcjpwYXNzd29yZA==")
}

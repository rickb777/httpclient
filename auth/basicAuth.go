package auth

import (
	"encoding/base64"
	"net/http"
	"strings"
)

// Basic provides HTTP basic authentication.
// See https://tools.ietf.org/html/rfc7617
func Basic(user string, pw string) Authenticator {
	return &basicAuth{
		user: user,
		pw:   pw,
	}
}

type basicAuth struct {
	user string
	pw   string
}

func (b *basicAuth) Challenge(ss []string) Authenticator {
	for _, s := range ss {
		if !strings.HasPrefix(s, "Basic") {
			panic("incorrect auth challenge: only 'Basic' expected here")
		}
	}
	return b
}

// Type identifies the Basic authenticator.
func (b *basicAuth) Type() string {
	return "Basic"
}

// User holds the BasicAuth username.
func (b *basicAuth) User() string {
	return b.user
}

// Password holds the BasicAuth password.
func (b *basicAuth) Password() string {
	return b.pw
}

// Authenticate the current request.
func (b *basicAuth) Authenticate(req *http.Request) {
	a := b.user + ":" + b.pw
	authorization := "Basic " + base64.StdEncoding.EncodeToString([]byte(a))
	req.Header.Set("Authorization", authorization)
}

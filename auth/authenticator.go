// Package auth provides HTTP client authentication support.
package auth

import (
	"net/http"
	"strings"
)

// Authenticator stub
type Authenticator interface {
	Type() string
	User() string
	Password() string
	// Challenge is the values from "WWW-Authenticate" header
	Challenge([]string) Authenticator
	Authenticate(*http.Request)
}

var Anonymous Authenticator = &noAuth{}

func Deferred(user string, pw string) Authenticator {
	return &noAuth{
		user: user,
		pw:   pw,
	}
}

// noAuth structure holds our credentials but doesn't use them.
type noAuth struct {
	user string
	pw   string
}

const None = "None"

// Type identifies the authenticator.
func (n *noAuth) Type() string {
	return None
}

// User returns the current user.
func (n *noAuth) User() string {
	return n.user
}

// Password returns the current password.
func (n *noAuth) Password() string {
	return n.pw
}

func (n *noAuth) Challenge(ss []string) Authenticator {
	if n.user != "" {
		for _, s := range ss {
			if strings.HasPrefix(s, "Digest") {
				return Digest(n.user, n.pw).Challenge(ss)
			} else if strings.HasPrefix(s, "Basic") {
				return Basic(n.user, n.pw).Challenge(ss)
			}
		}
	}
	return n
}

// Authenticate the current request
func (n *noAuth) Authenticate(_ *http.Request) {}

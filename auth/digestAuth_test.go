package auth

import (
	"net/http/httptest"
	"testing"

	"github.com/rickb777/expect"
)

// see https://datatracker.ietf.org/doc/html/rfc7616#section-3.9

func TestDigest_Authorize(t *testing.T) {
	req := httptest.NewRequest("GET", "/dir/index.html", nil)
	getCnonce = func() string { return "f2/wE4q74E6zIJEtWaHKaf5wv/H5QzzpXusqGemxURZJ" }

	const c1 = `Digest
		realm="http-auth@example.org",
		qop="auth, auth-int",
		algorithm=SHA-256,
		nonce="7ypf/xlj9XXwfDPEoM4URrv/xwf94BcCAzFZH4GiTo0v",
		opaque="FQhe/qaU925kfnzjCev0ciny7QMkPqMAFRtzCUYo5tdS"`

	const c2 = `Digest
		realm="http-auth@example.org",
		qop="auth, auth-int",
		algorithm=MD5,
		nonce="7ypf/xlj9XXwfDPEoM4URrv/xwf94BcCAzFZH4GiTo0v",
		opaque="FQhe/qaU925kfnzjCev0ciny7QMkPqMAFRtzCUYo5tdS"`

	digest := Deferred("Mufasa", "Circle of Life")
	digest.Challenge([]string{c1, c2}).Authenticate(req)

	expect.String(req.Header.Get("Authorization")).ToBe(t,
		`Digest username="Mufasa", `+
			`realm="http-auth@example.org", `+
			`uri="/dir/index.html", `+
			`algorithm=MD5, `+
			`nonce="7ypf/xlj9XXwfDPEoM4URrv/xwf94BcCAzFZH4GiTo0v", `+
			`nc=00000001, `+
			`cnonce="f2/wE4q74E6zIJEtWaHKaf5wv/H5QzzpXusqGemxURZJ", `+
			`response="8ca523f5e9506fed4657c9700eebdbec", `+
			`qop=auth, `+
			`opaque="FQhe/qaU925kfnzjCev0ciny7QMkPqMAFRtzCUYo5tdS"`)
}

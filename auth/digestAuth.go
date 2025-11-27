package auth

import (
	md5pkg "crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"strings"
	"unicode"
)

var _ Authenticator = &DigestAuth{}

// Digest implements HTTP digest authentication.
// See https://tools.ietf.org/html/rfc7616
// TODO finish this: it only supports MD5 and the tests are inadequate
func Digest(user string, pw string) *DigestAuth {
	return &DigestAuth{
		user:        user,
		pw:          pw,
		digestParts: map[string]string{},
	}
}

// DigestAuth structure holds our credentials.
type DigestAuth struct {
	user        string
	pw          string
	digestParts map[string]string
}

// Type identifies the Digest authenticator.
func (d *DigestAuth) Type() string {
	return "Digest"
}

// User holds the DigestAuth username.
func (d *DigestAuth) User() string {
	return d.user
}

// Password holds the DigestAuth password.
func (d *DigestAuth) Password() string {
	return d.pw
}

// Authenticate the current request.
func (d *DigestAuth) Authenticate(req *http.Request) {
	d.digestParts["username"] = d.user
	d.digestParts["password"] = d.pw
	req.Header.Set("Authorization", getDigestAuthentication(req, d.digestParts))
}

func (d *DigestAuth) Challenge(ss []string) Authenticator {
	for _, s := range ss {
		if !strings.HasPrefix(s, "Digest") {
			panic("incorrect auth challenge: only 'Digest' expected here")
		}
	}
	// TODO insert SHA-256, SHA-512-256
	// we support MD5
	for _, s := range ss {
		if strings.Contains(s, "algorithm=MD5") {
			return d.DigestParts(strings.TrimSpace(s[6:]))
		}
	}
	// we also default to MD5
	for _, s := range ss {
		if !strings.Contains(s, "algorithm") {
			return d.DigestParts(strings.TrimSpace(s[6:]))
		}
	}
	panic("unsupported Digest algorithm")
}

func (d *DigestAuth) DigestParts(wwwAuthenticateHeader string) Authenticator {
	d.digestParts = map[string]string{"algorithm": "MD5"} // contains our default algorithm
	wwwAuthenticateHeader = strings.TrimRightFunc(wwwAuthenticateHeader, unicode.IsSpace)

	// unwanted headers: domain, stale, charset, userhash
	wantedHeaders := []string{"nonce", "realm", "qop", "opaque", "algorithm", "entityBody"}

	for len(wwwAuthenticateHeader) > 0 {
		// We have to step through the header string token-by-token because commas can appear
		// either inside our outside double-quotes. Those outside double-quotes are parameter
		// separators. Those inside are part of their value.
		parts := strings.SplitN(wwwAuthenticateHeader, "=", 2)
		if len(parts) < 2 {
			return d
		}

		var key, value string
		key, wwwAuthenticateHeader = strings.TrimSpace(parts[0]), strings.TrimLeftFunc(parts[1], unicode.IsSpace)
		if strings.HasPrefix(wwwAuthenticateHeader, `"`) {
			end := strings.IndexByte(wwwAuthenticateHeader[1:], '"') + 1
			value = wwwAuthenticateHeader[1:end]
			wwwAuthenticateHeader = wwwAuthenticateHeader[end+1:]
		} else {
			end := strings.IndexByte(wwwAuthenticateHeader, ',')
			if end > 1 {
				value = wwwAuthenticateHeader[0:end]
				wwwAuthenticateHeader = wwwAuthenticateHeader[end:]
			} else {
				value = wwwAuthenticateHeader
				wwwAuthenticateHeader = ""
			}
		}

		if strings.HasPrefix(wwwAuthenticateHeader, `,`) {
			wwwAuthenticateHeader = strings.TrimLeftFunc(wwwAuthenticateHeader[1:], unicode.IsSpace)
		}

		for _, w := range wantedHeaders {
			if key == w {
				d.digestParts[w] = value
			}
		}
	}
	return d
}

func md5(text string) string {
	hasher := md5pkg.New()
	hasher.Write([]byte(text))
	return hex.EncodeToString(hasher.Sum(nil))
}

var getCnonce = func() string {
	b := make([]byte, 8)
	io.ReadFull(rand.Reader, b)
	return fmt.Sprintf("%x", b)[:16]
}

func getDigestAuthentication(req *http.Request, d map[string]string) string {
	// These are the correct ha1 and ha2 for qop=auth. We should probably check for other types of qop.

	var (
		ha1        string
		ha2        string
		nonceCount = "00000001"
		cnonce     = getCnonce()
		response   string
	)

	// 'ha1' value depends on value of "algorithm" field
	switch d["algorithm"] {
	case "MD5", "":
		ha1 = md5(d["username"] + ":" + d["realm"] + ":" + d["password"])
	case "MD5-sess":
		ha1 = md5(fmt.Sprintf("%s:%v:%s",
			md5(d["username"]+":"+d["realm"]+":"+d["password"]),
			nonceCount,
			cnonce))
	}

	// 'ha2' value depends on value of "qop" field
	qops := strings.Split(d["qop"], ",")
	chosenQop := ""
	switch strings.TrimSpace(qops[0]) {
	case "auth", "":
		ha2 = md5(req.Method + ":" + req.URL.Path)
		chosenQop = "auth"
	case "auth-int":
		entityBody := d["entityBody"]
		ha2 = md5(req.Method + ":" + req.URL.Path + ":" + md5(entityBody))
		chosenQop = "auth-int"
	}

	// 'response' value depends on value of "qop" field
	switch chosenQop {
	case "":
		response = md5(
			fmt.Sprintf("%s:%s:%s",
				ha1,
				d["nonce"],
				ha2,
			),
		)
	case "auth", "auth-int":
		response = md5(
			fmt.Sprintf("%s:%s:%v:%s:%s:%s",
				ha1,
				d["nonce"],
				nonceCount,
				cnonce,
				chosenQop,
				ha2,
			),
		)
	}

	var authentication strings.Builder
	fmt.Fprintf(&authentication, `Digest username="%s", realm="%s", uri="%s", algorithm=%s`,
		d["username"],
		d["realm"],
		req.URL.Path,
		d["algorithm"])

	fmt.Fprintf(&authentication, `, nonce="%s", nc=%v, cnonce="%s", response="%s"`,
		d["nonce"],
		nonceCount,
		cnonce,
		response)

	if chosenQop != "" {
		fmt.Fprintf(&authentication, `, qop=%s`, chosenQop)
	}

	if d["opaque"] != "" {
		fmt.Fprintf(&authentication, `, opaque="%s"`, d["opaque"])
	}

	return authentication.String()
}

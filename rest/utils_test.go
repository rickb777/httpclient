package rest

import (
	"fmt"
	"net/url"
	"path"
	"testing"
)

func TestJoin(t *testing.T) {
	eq(t, "/", "/", "")
	eq(t, "/", "/", "/")
	eq(t, "/foo", "", "/foo")
	eq(t, "foo/foo", "foo/", "/foo")
	eq(t, "foo/foo", "foo/", "foo")
}

func eq(t *testing.T, expected string, s0 string, s1 string) {
	s := path.Join(s0, s1)
	if s != expected {
		t.Error("For", "'"+s0+"','"+s1+"'", "expected", "'"+expected+"'", "got", "'"+s+"'")
	}
}

func ExamplePathEscape() {
	fmt.Println(pathEscape(""))
	fmt.Println(pathEscape("/"))
	fmt.Println(pathEscape("/web"))
	fmt.Println(pathEscape("/web/"))
	fmt.Println(pathEscape("/w e b/d a v/s%u&c#k:s/"))

	// Output:
	//
	// /
	// /web
	// /web/
	// /w%20e%20b/d%20a%20v/s%25u&c%23k:s/
}

func TestEscapeURL(t *testing.T) {
	ex := "https://foo.com/w%20e%20b/d%20a%20v/s%25u&c%23k:s/"
	u, _ := url.Parse("https://foo.com" + pathEscape("/w e b/d a v/s%u&c#k:s/"))
	if ex != u.String() {
		t.Error("expected: " + ex + " got: " + u.String())
	}
}

func TestWithTrailingSlash(t *testing.T) {
	cases := map[string]string{
		"":       "/",
		"/":      "/",
		"/a/bc/": "/a/bc/",
		"/a/bc":  "/a/bc/",
	}

	for input, expected := range cases {
		got := withTrailingSlash(input)
		if got != expected {
			t.Errorf("expected: %q got %q", expected, got)
		}
	}
}

func TestWithoutTrailingSlash(t *testing.T) {
	cases := map[string]string{
		"":       "",
		"/":      "",
		"/a/bc/": "/a/bc",
		"/a/bc":  "/a/bc",
	}

	for input, expected := range cases {
		got := withoutTrailingSlash(input)
		if got != expected {
			t.Errorf("expected: %q got %q", expected, got)
		}
	}
}

func TestWithSurroundingSlashes(t *testing.T) {
	cases := map[string]string{
		"":       "/",
		"/":      "/",
		"/a/bc/": "/a/bc/",
		"/a/bc":  "/a/bc/",
		"a/bc/":  "/a/bc/",
		"a/bc":   "/a/bc/",
	}

	for input, expected := range cases {
		got := withSurroundingSlashes(input)
		if got != expected {
			t.Errorf("expected: %q got %q", expected, got)
		}
	}
}

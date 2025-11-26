package rest

import (
	"fmt"
	"net/url"
	"os"
	"strings"
)

func log(msg interface{}) {
	fmt.Println(msg)
}

func newPathError(op string, path string, statusCode int) error {
	return newPathErrorErr(op, path, fmt.Errorf("%d", statusCode))
}

func newPathErrorErr(op string, path string, err error) error {
	return &os.PathError{
		Op:   op,
		Path: path,
		Err:  err,
	}
}

// pathEscape escapes all segments of a given path
func pathEscape(path string) string {
	s := strings.Split(path, "/")
	for i, e := range s {
		s[i] = url.PathEscape(e)
	}
	return strings.Join(s, "/")
}

// withoutTrailingSlash removes any trailing / from a string
func withoutTrailingSlash(s string) string {
	if strings.HasSuffix(s, "/") {
		return s[:len(s)-1]
	}
	return s
}

// withTrailingSlash appends a trailing / to a string
func withTrailingSlash(s string) string {
	if strings.HasSuffix(s, "/") {
		return s
	}
	return s + "/"
}

// withLeadingSlash prepends a leading / to a string
func withLeadingSlash(s string) string {
	if strings.HasPrefix(s, "/") {
		return s
	}
	return "/" + s
}

// withSurroundingSlashes appends and prepends a / if they are missing
func withSurroundingSlashes(s string) string {
	if !strings.HasPrefix(s, "/") {
		s = "/" + s
	}
	return withTrailingSlash(s)
}

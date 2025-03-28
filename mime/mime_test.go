package mime

import (
	"github.com/rickb777/expect"
	"testing"
)

func TestFileExtension(t *testing.T) {
	cases := map[string]string{
		"text/plain":               ".txt",
		"text/html":                ".html",
		"application/octet-stream": ".bin",
		"application/pdf":          ".pdf",
		"application/xml":          ".xml",
		"image/gif":                ".gif",
		"unknown/thing":            "",
	}
	for ct, exp := range cases {
		e := FileExtension(ct)
		expect.String(e).Info(ct).ToBe(t, exp)
	}
}

func TestIsTextual(t *testing.T) {
	t.Run("true", func(t *testing.T) {
		cases := []string{
			"text/plain", "text/html",
			"application/json",
			"application/calendar+json",
			"application/xml",
			"application/atom+xml",
			"application/vcard+xml",
			"image/svg+xml",
		}
		for _, c := range cases {
			r := IsTextual(c)
			expect.Bool(r).ToBeTrue(t)
		}
	})

	t.Run("false", func(t *testing.T) {
		cases := []string{
			"image/png",
			"application/octet-stream",
			"application/pdf",
			"audio/mp4",
			"unknown/thing",
		}
		for _, c := range cases {
			r := IsTextual(c)
			expect.Bool(r).ToBeFalse(t)
		}
	})
}

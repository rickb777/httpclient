package mime

import (
	"github.com/rickb777/expect"
	"testing"
)

func TestFileExtension_good(t *testing.T) {
	goodCases := map[string]string{
		"text/plain":               ".txt",
		"text/html":                ".html",
		"application/octet-stream": ".bin",
		"application/pdf":          ".pdf",
		"application/xml":          ".xml",
		"image/gif":                ".gif",
	}
	for ct, exp := range goodCases {
		e0 := FileExtension(ct)
		expect.String(e0).Info(ct).ToBe(t, exp)

		es, err := ExtensionsByType(ct)
		expect.Slice(es, err).Info(ct).ToContain(t, exp)
	}
}

func TestFileExtension_bad(t *testing.T) {
	badCases := map[string]string{
		"unknown/thing": "",
	}
	for ct, exp := range badCases {
		e0 := FileExtension(ct)
		expect.String(e0).Info(ct).ToBe(t, exp)
	}
}

func TestIsTextual(t *testing.T) {
	t.Run("true", func(t *testing.T) {
		cases := []string{
			"text/plain",
			"text/html",
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

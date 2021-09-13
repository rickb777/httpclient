package mime

import (
	"github.com/onsi/gomega"
	"testing"
)

func TestFileExtension(t *testing.T) {
	g := gomega.NewWithT(t)

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
		g.Expect(e).To(gomega.Equal(exp), ct)
	}
}

func TestIsTextual(t *testing.T) {
	g := gomega.NewWithT(t)

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
			g.Expect(r).To(gomega.BeTrue())
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
			g.Expect(r).To(gomega.BeFalse())
		}
	})
}

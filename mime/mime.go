// Package mime implements parts of the MIME spec similarly
// to the standard "mime" package.
// This package adds two functions, FileExtension and IsTextual. The
// other functions call through to the standard "mime" package functions
// of the same name.
package mime

import (
	mimepkg "mime"
	"strings"
	"sync"
)

var once sync.Once

// FileExtension returns the extensions known to be associated with the MIME
// type mimeType. This is similar to ExtensionsByType, except that these are treated
// as special cases:
//
// * "text/plain": ".txt"
// * "application/octet-stream": ".bin"
// * "application/xml": ".xml"
func FileExtension(mimeType string) string {
	once.Do(func() {
		mimepkg.AddExtensionType(".xml", "application/xml")
		mimepkg.AddExtensionType(".bin", "application/octet-stream")
		mimepkg.AddExtensionType(".txt", "text/plain")
	})

	ctl := strings.ToLower(mimeType)

	// these special cases ensure consistency across platforms
	// because the ordering of MIME type mappings is not predictable
	switch ctl {
	case "text/plain":
		return ".txt"
	case "application/xml":
		return ".xml"
	case "application/octet-stream":
		return ".bin"
	}

	exts, _ := mimepkg.ExtensionsByType(ctl)
	if len(exts) > 0 {
		return exts[len(exts)-1]
	}

	return ""
}

// ExtensionsByType returns the extensions known to be associated with the MIME
// type typ. The returned extensions will each begin with a leading dot, as in
// ".html". When typ has no associated extensions, ExtensionsByType returns an
// nil slice.
func ExtensionsByType(typ string) ([]string, error) {
	return mimepkg.ExtensionsByType(typ)
}

// IsTextual tests a media type to determine whether it describes text or binary content.
// Textual types are
//
// * "text/*"
// * "application/json"
// * "application/xml"
// * "application/*+json"
// * "application/*+xml"
// * "image/*+xml"
//
// where "*" is a wildcard.
func IsTextual(contentType string) bool {
	ct, _, _ := mimepkg.ParseMediaType(contentType)
	ps := strings.SplitN(strings.TrimSpace(ct), "/", 2)
	if len(ps) != 2 {
		return false
	}

	mainType, subType := ps[0], ps[1]

	if mainType == "text" {
		return true
	}

	if mainType == "application" {
		return subType == "json" ||
			subType == "xml" ||
			strings.HasSuffix(subType, "+xml") ||
			strings.HasSuffix(subType, "+json")
	}

	if mainType == "image" {
		return strings.HasSuffix(subType, "+xml")
	}

	return false
}

// FormatMediaType serializes mediatype t and the parameters
// param as a media type conforming to RFC 2045 and RFC 2616.
// The type and parameter names are written in lower-case.
// When any of the arguments result in a standard violation then
// FormatMediaType returns the empty string.
func FormatMediaType(t string, param map[string]string) string {
	return mimepkg.FormatMediaType(t, param)
}

// ParseMediaType parses a media type value and any optional
// parameters, per RFC 1521.  Media types are the values in
// Content-Type and Content-Disposition headers (RFC 2183).
// On success, ParseMediaType returns the media type converted
// to lowercase and trimmed of white space and a non-nil map.
// If there is an error parsing the optional parameter,
// the media type will be returned along with the error
// ErrInvalidMediaParameter.
// The returned map, params, maps from the lowercase
// attribute to the attribute value with its case preserved.
func ParseMediaType(v string) (mediatype string, params map[string]string, err error) {
	return mimepkg.ParseMediaType(v)
}

// TypeByExtension returns the MIME type associated with the file extension ext.
// The extension ext should begin with a leading dot, as in ".html".
// When ext has no associated type, TypeByExtension returns "".
//
// Extensions are looked up first case-sensitively, then case-insensitively.
//
// The built-in table is small but on unix it is augmented by the local
// system's MIME-info database or mime.types file(s) if available under one or
// more of these names:
//
//	/usr/local/share/mime/globs2
//	/usr/share/mime/globs2
//	/etc/mime.types
//	/etc/apache2/mime.types
//	/etc/apache/mime.types
//
// On Windows, MIME types are extracted from the registry.
//
// Text types have the charset parameter set to "utf-8" by default.
func TypeByExtension(ext string) string {
	return mimepkg.TypeByExtension(ext)
}

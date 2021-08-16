package mime

import (
	mimepkg "mime"
	"strings"
)

func FileExtension(mimeType string) string {
	ctl := strings.ToLower(mimeType)

	// two special cases to ensure consistency across platforms
	// because the ordering of MIME type mappings is not predictable
	switch ctl {
	case "text/plain":
		return ".txt"
	case "application/octet-stream":
		return ".bin"
	}

	exts, _ := mimepkg.ExtensionsByType(ctl)
	if len(exts) > 0 {
		return exts[0]
	}

	return ""
}

// IsTextual tests a media type (a.k.a. content type) to determine whether it
// describes text or binary content.
func IsTextual(contentType string) bool {
	cts := strings.SplitN(contentType, ";", 2)
	ps := strings.SplitN(strings.TrimSpace(cts[0]), "/", 2)
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

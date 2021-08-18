package file

import (
	"net/http"
	"os"
	"strings"
	"time"
	"unicode"
)

const (
	windowsPunct    = "_."
	nonWindowsPunct = "_=?&;:.,+@%"
)

var (
	// AllowedPunctuationInFilenames lists the punctuation characters that are tolerated
	// when converting URLs to filenames. This is initialised to a string appropriate for
	// the current OS.
	AllowedPunctuationInFilenames = nonWindowsPunct

	// FilenameTimestampFormat is the time format used for filenames containing a timestamp.
	FilenameTimestampFormat = "2006-01-02_15-04-05.000"
)

func init() {
	reset()
}

// reset is a seam for testing
func reset() {
	AllowedPunctuationInFilenames = nonWindowsPunct
	if os.PathSeparator == '\\' {
		AllowedPunctuationInFilenames = windowsPunct
	}
}

func FilenameTimestamp(t time.Time) string {
	return strings.Replace(t.Format(FilenameTimestampFormat), ".", "-", 1)
}

func UrlToFilename(path string) string {
	if path == "" {
		return ""
	}
	if path[0] == '/' {
		path = path[1:]
	}

	buf := &strings.Builder{}
	dash := false
	for _, c := range path {
		switch c {
		case '/':
			buf.WriteRune('_')
			dash = false
		default:
			if strings.IndexRune(AllowedPunctuationInFilenames, c) >= 0 {
				buf.WriteRune(c)
			} else if unicode.IsLetter(c) || unicode.IsDigit(c) {
				buf.WriteRune(c)
				dash = false
			} else if !dash {
				buf.WriteByte('-')
				dash = true
			}
		}
	}
	return buf.String()
}

// Hostname gets the "Host" header and removes any disallowed punctuation characters.
func Hostname(hdrs http.Header) string {
	host := UrlToFilename(hdrs.Get("Host"))
	if host != "" {
		host += "_"
	}
	return host
}

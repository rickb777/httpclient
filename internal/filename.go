package internal

import (
	"os"
	"strings"
	"time"
	"unicode"
)

const (
	windowsPunct    = "_."
	nonWindowsPunct = "_=?&;:.,+@%"
)

var AllowedPunctuationInFilenames = nonWindowsPunct

func init() {
	reset()
}

func reset() {
	AllowedPunctuationInFilenames = nonWindowsPunct
	if os.PathSeparator == '\\' {
		AllowedPunctuationInFilenames = windowsPunct
	}
}

func FilenameTimestamp(t time.Time) string {
	return t.Format("2006-01-02_15-04-05")
}

func UrlToFilename(path string) string {
	if path == "" {
		return ""
	}

	buf := &strings.Builder{}
	dash := false
	for _, c := range path[1:] {
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

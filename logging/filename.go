package logging

import (
	"os"
	"strings"
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

func urlToFilename(path string) string {
	if path == "" {
		return ""
	}
	return removePunctuation(path[1:])
}

func removePunctuation(s string) string {
	buf := &strings.Builder{}
	dash := false
	for _, c := range s {
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

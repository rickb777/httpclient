package logging

import (
	"fmt"
	"github.com/rickb777/httpclient/body"
	"github.com/rickb777/httpclient/file"
	"github.com/rickb777/httpclient/mime"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type LogContent struct {
	Header http.Header
	Body   *body.Body
}

// LogItem records information about one HTTP round-trip.
type LogItem struct {
	Method     string
	URL        *url.URL
	StatusCode int
	Request    LogContent
	Response   LogContent
	Err        error
	Start      time.Time
	Duration   time.Duration
	Level      Level
}

// ContentType gets the "Content-Type" header and returns the first part, i.e.
// excluding all parameters.
func (lc LogContent) ContentType() string {
	contentType := lc.Header.Get("Content-Type")
	return strings.SplitN(contentType, ";", 2)[0]
}

// FileExtension gets normal file extension for the content type represented
// by this content.
func (lc LogContent) FileExtension() string {
	return mime.FileExtension(lc.ContentType())
}

// IsTextual determines whether this content is normally consiudered to be
// textual. If the content is binary, the result is false.
func (lc LogContent) IsTextual() bool {
	return mime.IsTextual(lc.ContentType())
}

// FileName builds a filename that distinctively represents the request that
// raised the LogItem. This uses ItemFileName.
func (item *LogItem) FileName() string {
	return ItemFileName(item)
}

// ItemFileName builds a filename that distinctively represents the request that
// raised the LogItem.
//
// This is pluggable with alternative implementations. The default implementation
// concatenates the timestamp, the method, the hostname, and the URL. No extension
// is added -
var ItemFileName = func(item *LogItem) string {
	return fmt.Sprintf("%s_%s_%s%s",
		file.FilenameTimestamp(item.Start),
		item.Method,
		file.Hostname(item.Request.Header),
		file.UrlToFilename(item.URL.Path))
}

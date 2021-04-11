package httpclient

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"
)

type LogContent struct {
	Header http.Header
	Body   []byte
}

type LogItem struct {
	Method     string
	URL        string
	StatusCode int
	Request    LogContent
	Response   LogContent
	Err        error
	Duration   time.Duration
	Level      Level
}

type Logger func(item *LogItem)

// LogWriter returns a new Logger.
func LogWriter(out io.Writer) Logger {
	return func(item *LogItem) {
		// basic info
		fmt.Fprintf(out, "%-8s %s %d %s", item.Method, item.URL, item.StatusCode, item.Duration.Round(100*time.Microsecond))
		if item.Err != nil {
			fmt.Fprintf(out, " %v", item.Err)
		}
		fmt.Fprintln(out)

		// verbose info
		if item.Level >= WithHeaders {
			printPart(out, item.Request.Header, "-->", item.Request.Body)
			fmt.Fprintln(out)
			printPart(out, item.Response.Header, "<--", item.Response.Body)
			fmt.Fprintln(out, "--------")
		}
	}
}

func printPart(out io.Writer, hdrs http.Header, prefix string, body []byte) {
	printHeaders(out, hdrs, prefix)
	if IsTextual(hdrs.Get("Content-Type")) {
		if len(body) > 0 {
			fmt.Fprintln(out)
			io.Copy(out, bytes.NewBuffer(body))
		}
	} else if len(body) > 0 {
		fmt.Fprintf(out, "%s binary content [%d]byte\n", prefix, len(body))
	}
}

func printHeaders(out io.Writer, hdrs http.Header, prefix string) {
	if len(hdrs) == 0 {
		fmt.Fprintf(out, "%s no headers\n", prefix)
		return
	}

	var keys []string
	for k, _ := range hdrs {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		vs := hdrs[k]
		k += ":"
		fmt.Fprintf(out, "%s %-15s %s\n", prefix, k, vs[0])
		for _, v := range vs[1:] {
			fmt.Fprintf(out, "%s                 %s\n", prefix, v)
		}
	}
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

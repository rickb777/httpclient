package logging

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/go-xmlfmt/xmlfmt"
	"github.com/rickb777/httpclient"
	"io"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"
	"time"
)

// LongBodyThreshold is the body length threshold beyond which the body
// will be written to a text file (when the content is text). Otherwise
// it is written inline in the log.
var LongBodyThreshold = 100

type LogContent struct {
	Header http.Header
	Body   []byte
}

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

type Logger func(item *LogItem)

// LogWriter returns a new Logger.
func LogWriter(out io.Writer, dir string) Logger {
	if dir != "" && !strings.HasSuffix(dir, "/") {
		dir = dir + "/"
	}
	return func(item *LogItem) {
		// basic info
		fmt.Fprintf(out, "%-8s %s %d %s", item.Method, item.URL, item.StatusCode, item.Duration.Round(100*time.Microsecond))
		if item.Err != nil {
			fmt.Fprintf(out, " %v", item.Err)
		}
		fmt.Fprintln(out)

		// verbose info
		switch item.Level {
		case WithHeaders:
			printPart(out, item.Request.Header, true, "", nil)
			fmt.Fprintln(out)
			printPart(out, item.Response.Header, false, "", nil)
			fmt.Fprintln(out, "\n---")

		case WithHeadersAndBodies:
			file := fmt.Sprintf("%s%s_%s_%s", dir, item.Start.Format("2006-01-02_15-04-05"),
				item.Method, urlToFilename(item.URL.Path))
			printPart(out, item.Request.Header, true, file, item.Request.Body)
			fmt.Fprintln(out)
			printPart(out, item.Response.Header, false, file, item.Response.Body)
			fmt.Fprintln(out, "\n---")
		}
	}
}

func urlToFilename(path string) string {
	p := path
	if p == "" {
		return ""
	}
	return removePunctuation(p[1:])
}

func removePunctuation(s string) string {
	buf := &strings.Builder{}
	dash := false
	for _, c := range s {
		if 'A' <= c && c <= 'Z' {
			buf.WriteRune(c)
			dash = false
		} else if 'a' <= c && c <= 'z' {
			buf.WriteRune(c)
			dash = false
		} else if '0' <= c && c <= '9' {
			buf.WriteRune(c)
			dash = false
		} else if c == '/' {
			buf.WriteByte('_')
			dash = false
		} else if !dash {
			buf.WriteByte('-')
			dash = true
		}
	}
	return buf.String()
}

func printPart(out io.Writer, hdrs http.Header, isRequest bool, file string, body []byte) {
	prefix := ternary(isRequest, "-->", "<--")
	printHeaders(out, hdrs, prefix)
	contentType := hdrs.Get("Content-Type")
	if len(body) > 0 {
		if IsTextual(contentType) {
			saveBodyToFile(out, isRequest, contentType, file, body)
		} else {
			fmt.Fprintf(out, "%s binary content [%d]byte\n", prefix, len(body))
		}
	}
}

func saveBodyToFile(out io.Writer, isRequest bool, contentType string, file string, body []byte) {
	suffix := ternary(isRequest, "req", "res")
	name := fmt.Sprintf("%s_%s", file, suffix)
	cts := strings.SplitN(contentType, ";", 2)
	if len(body) > LongBodyThreshold {
		switch strings.ToLower(cts[0]) {
		case "application/json":
			writeBodyToFile(out, name, ".json", body)
		case "application/xml":
			writeBodyToFile(out, name, ".xml", body)
		default:
			writeBodyToFile(out, name, ".txt", body)
		}
	} else {
		// write short body inline
		fmt.Fprintln(out)
		fn := &httpclient.WithFinalNewline{W: out}
		io.Copy(fn, bytes.NewBuffer(body))
		fn.EnsureFinalNewline()
	}
}

func writeBodyToFile(out io.Writer, name, extn string, body []byte) {
	f, err := os.Create(name + extn)
	if err != nil {
		fmt.Fprintf(out, "logger open file error: %s\n", err)
		return
	}

	err = prettyPrinterFactory(extn)(f, body)
	if err != nil {
		fmt.Fprintf(out, "logger transcode error: %s\n", err)
		return
	}

	err = f.Close()
	if err != nil {
		fmt.Fprintf(out, "logger close error: %s\n", err)
	}

	fmt.Fprintf(out, "see %s%s\n", name, extn)
}

func printHeaders(out io.Writer, hdrs http.Header, prefix string) {
	if len(hdrs) == 0 {
		fmt.Fprintf(out, "%s no headers\n", prefix)
		return
	}

	var keys []string
	for k := range hdrs {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		vs := hdrs[k]
		k += ":"
		fmt.Fprintf(out, "%s %-16s %s\n", prefix, k, vs[0])
		for _, v := range vs[1:] {
			fmt.Fprintf(out, "%s                  %s\n", prefix, v)
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

func ternary(predicate bool, yes, no string) string {
	if predicate {
		return yes
	}
	return no
}

//-------------------------------------------------------------------------------------------------
// pretty printing via transcoding: implemented for JSON and XML only

type transcoder func(out io.Writer, body []byte) error

func prettyPrinterFactory(extension string) transcoder {
	switch extension {
	case ".json":
		return jsonTranscoder
	case ".xml":
		return xmlTranscoder
	}

	return func(out io.Writer, body []byte) error {
		fmt.Fprintln(out)
		fn := &httpclient.WithFinalNewline{W: out}
		_, err := bytes.NewBuffer(body).WriteTo(out)
		fn.EnsureFinalNewline()
		return err
	}
}

//-------------------------------------------------------------------------------------------------

func jsonTranscoder(out io.Writer, body []byte) error {
	var data interface{}
	err := json.NewDecoder(bytes.NewReader(body)).Decode(&data)
	if err != nil {
		return err
	}

	enc := json.NewEncoder(out)
	enc.SetIndent("", "  ")
	return enc.Encode(data)
}

//-------------------------------------------------------------------------------------------------

func xmlTranscoder(out io.Writer, body []byte) error {
	xml := xmlfmt.FormatXML(string(body), "", "    ")
	if strings.HasPrefix(xml, xmlfmt.NL) {
		xml = xml[len(xmlfmt.NL):]
	}
	_, err := fmt.Fprintln(out, xml)
	return err
}

func init() {
	xmlfmt.NL = "\n"
}

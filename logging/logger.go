package logging

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/go-xmlfmt/xmlfmt"
	"github.com/rickb777/httpclient"
	"github.com/spf13/afero"
	"io"
	"mime"
	"net/http"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// LongBodyThreshold is the body length threshold beyond which the body
// will be written to a text file (when the content is text). Otherwise
// it is written inline in the log.
var LongBodyThreshold = 100

// Logger is a function that processes log items, usually by writing them
// to a log file.
type Logger func(item *LogItem)

// Fs provides the filesystem. It can be stubbed for testing.
var Fs = afero.NewOsFs()

// Now provides the current time. It can be stubbed for testing.
var Now = func() time.Time {
	return time.Now().UTC()
}

// FileLogger returns a new Logger writing to a file in dir. The name
// of the file is provided.
// The same directory specifies where request and response bodies will be
// written as files. The current directory is used if this is "." or blank.
func FileLogger(name string) (Logger, error) {
	f, err := Fs.Create(name)
	if err != nil {
		return nil, err
	}
	return LogWriter(f, filepath.Dir(name)), nil
}

// LogWriter returns a new Logger.
// The directory dir specifies where request and response bodies will be
// written as files. The current directory is used if dir is "." or blank.
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
			file := fmt.Sprintf("%s%s_%s_%s", dir, timestamp(item.Start), item.Method, urlToFilename(item.URL.Path))
			printPart(out, item.Request.Header, true, file, item.Request.Body.Bytes())
			fmt.Fprintln(out)
			printPart(out, item.Response.Header, false, file, item.Response.Body.Bytes())
			fmt.Fprintln(out, "\n---")
		}
	}
}

func timestamp(t time.Time) string {
	return t.Format("2006-01-02_15-04-05")
}

func printPart(out io.Writer, hdrs http.Header, isRequest bool, file string, body []byte) {
	prefix := ternary(isRequest, "-->", "<--")
	printHeaders(out, hdrs, prefix)
	contentType := hdrs.Get("Content-Type")
	if len(body) == 0 {
		return
	}

	suffix := ternary(isRequest, "req", "resp")
	name := fmt.Sprintf("%s_%s", file, suffix)
	justType := strings.SplitN(contentType, ";", 2)[0]
	if len(body) > LongBodyThreshold {
		extn := fileExtension(justType)
		if extn != "" {
			writeBodyToFile(out, name, extn, body)
		} else {
			fmt.Fprintf(out, "%s binary content [%d]byte\n", prefix, len(body))
		}

	} else if IsTextual(justType) {
		// write short body inline
		fmt.Fprintln(out)
		fn := &httpclient.WithFinalNewline{W: out}
		io.Copy(fn, bytes.NewBuffer(body))
		fn.EnsureFinalNewline()

	} else {
		fmt.Fprintf(out, "%s binary content [%d]byte\n", prefix, len(body))
	}
}

func fileExtension(mimeType string) string {
	ctl := strings.ToLower(mimeType)

	// two special cases to ensure consistency across platforms
	// because the ordering of MIME type mappings is not predictable
	switch ctl {
	case "text/plain":
		return ".txt"
	case "application/octet-stream":
		return ".bin"
	}

	exts, _ := mime.ExtensionsByType(ctl)
	if len(exts) > 0 {
		return exts[0]
	}

	return ""
}

func writeBodyToFile(out io.Writer, name, extn string, body []byte) {
	f, err := Fs.Create(name + extn)
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
	return writePlainText
}

func writePlainText(out io.Writer, body []byte) error {
	fn := &httpclient.WithFinalNewline{W: out}
	_, err := bytes.NewBuffer(body).WriteTo(fn)
	fn.EnsureFinalNewline()
	return err
}

//-------------------------------------------------------------------------------------------------

func jsonTranscoder(out io.Writer, body []byte) error {
	var data interface{}
	err := json.NewDecoder(bytes.NewReader(body)).Decode(&data)
	if err != nil {
		return writePlainText(out, body)
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

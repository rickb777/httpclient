package internal

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
	"sort"
	"strings"
)

func PrintPart(out io.Writer, fs afero.Fs, hdrs http.Header, isRequest bool, file string, body []byte, longBodyThreshold int) {
	prefix := ternary(isRequest, "-->", "<--")
	printHeaders(out, hdrs, prefix)
	contentType := hdrs.Get("Content-Type")
	if len(body) == 0 {
		return
	}

	suffix := ternary(isRequest, "req", "resp")
	name := fmt.Sprintf("%s_%s", file, suffix)
	justType := strings.SplitN(contentType, ";", 2)[0]
	if len(body) > longBodyThreshold {
		extn := FileExtension(justType)
		if extn != "" {
			WriteBodyToFile(out, fs, name, extn, body)
		} else {
			fmt.Fprintf(out, "%s binary content [%d]byte\n", prefix, len(body))
		}

	} else if IsTextual(justType) {
		// write short body inline
		fn := &httpclient.WithFinalNewline{W: out}
		io.Copy(fn, bytes.NewBuffer(body))
		fn.EnsureFinalNewline()

	} else {
		fmt.Fprintf(out, "%s binary content [%d]byte\n", prefix, len(body))
	}
}

func WriteBodyToFile(out io.Writer, fs afero.Fs, name, extn string, body []byte) {
	f, err := fs.Create(name + extn)
	if err != nil {
		fmt.Fprintf(out, "logger open file error: %s\n", err)
		return
	}

	err = PrettyPrinterFactory(extn)(f, body)
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

	exts, _ := mime.ExtensionsByType(ctl)
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

func ternary(predicate bool, yes, no string) string {
	if predicate {
		return yes
	}
	return no
}

//-------------------------------------------------------------------------------------------------
// pretty printing via transcoding: implemented for JSON and XML only

type transcoder func(out io.Writer, body []byte) error

func PrettyPrinterFactory(extension string) transcoder {
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

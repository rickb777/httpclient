package internal

import (
	"bytes"
	"fmt"
	"github.com/rickb777/httpclient"
	filepkg "github.com/rickb777/httpclient/file"
	"github.com/rickb777/httpclient/mime"
	"github.com/spf13/afero"
	"io"
	"net/http"
	"os"
	"sort"
	"strings"
)

// PrintPart prints the headers and entity (body) for either the request or the response.
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
		extn := mime.FileExtension(justType)
		if extn != "" {
			WriteBodyToFile(out, fs, name, extn, body)
		} else {
			fmt.Fprintf(out, "%s binary content [%d]byte\n", prefix, len(body))
		}

	} else if mime.IsTextual(justType) {
		// write short body inline
		fn := &httpclient.WithFinalNewline{W: out}
		io.Copy(fn, bytes.NewBuffer(body))
		fn.EnsureFinalNewline()

	} else {
		fmt.Fprintf(out, "%s binary content [%d]byte\n", prefix, len(body))
	}
}

// WriteBodyToFile writes one body (entity) to a file.
func WriteBodyToFile(out io.Writer, fs afero.Fs, name, extn string, body []byte) {
	f, err := fs.OpenFile(name+extn, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		fmt.Fprintf(out, "logger open file error: %s\n", err)
		return
	}

	err = filepkg.PrettyPrint(extn, f, body)
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

func ternary(predicate bool, yes, no string) string {
	if predicate {
		return yes
	}
	return no
}

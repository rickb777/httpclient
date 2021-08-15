package body

import (
	"bytes"
	"encoding/json"
	"github.com/go-xmlfmt/xmlfmt"
	"github.com/rickb777/httpclient"
	"io"
	"strings"
)

// PrettyIndent sets the indentation used to pretty-print JSON and XML
// body files, when these are written (see WithHeadersAndBodies and
// LongBodyThreshold). When non-blank, PrettyIndent
var PrettyIndent = ""

//-------------------------------------------------------------------------------------------------

// PrettyPrint writes a body to a writer (usually a file). Pretty printing is
// implemented via transcoding for JSON and XML only. All other file times
// are written verbatim.
func PrettyPrint(extension string, out io.Writer, body []byte) error {
	switch extension {
	case ".json":
		return WriteJSONFile(out, body)
	case ".xml":
		return WriteXMLFile(out, body)
	}
	return writePlainText(out, body)
}

func writePlainText(out io.Writer, body []byte) error {
	fn := &httpclient.WithFinalNewline{W: out}
	_, err := bytes.NewBuffer(body).WriteTo(fn)
	fn.EnsureFinalNewline()
	return err
}

//-------------------------------------------------------------------------------------------------

// WriteJSONFile is a function to write JSON files. If PrettyIndent is non-blank, the
// result is pretty-printed JSON; otherwise, it is verbatim.
//
// An alternative function may be substituted if required.
var WriteJSONFile = func(out io.Writer, body []byte) error {
	if len(PrettyIndent) == 0 {
		return writePlainText(out, body)
	}

	var data interface{}
	err := json.NewDecoder(bytes.NewReader(body)).Decode(&data)
	if err != nil {
		return writePlainText(out, body)
	}

	var enc = json.NewEncoder(out)
	enc.SetIndent("", PrettyIndent)
	return enc.Encode(data)
}

// WriteXMLFile is a function to write XML files. If PrettyIndent is non-blank, the
// result is pretty-printed XML; otherwise, it is verbatim.
//
// An alternative function may be substituted if required.
var WriteXMLFile = func(out io.Writer, body []byte) error {
	if len(PrettyIndent) == 0 {
		return writePlainText(out, body)
	}

	xml := xmlfmt.FormatXML(string(body), "", PrettyIndent)
	if strings.HasPrefix(xml, xmlfmt.NL) {
		xml = xml[len(xmlfmt.NL):]
	}
	_, err := io.WriteString(out, xml)
	out.Write([]byte{'\n'})
	return err
}

//-------------------------------------------------------------------------------------------------

func init() {
	xmlfmt.NL = "\n"
}

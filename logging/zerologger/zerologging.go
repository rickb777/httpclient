package zerologger

import (
	"encoding/json"
	"fmt"
	bodypkg "github.com/rickb777/httpclient/file"
	"github.com/rickb777/httpclient/internal/mime"
	"github.com/rickb777/httpclient/logging"
	"github.com/rickb777/httpclient/logging/logger"
	"github.com/rs/zerolog"
	"github.com/spf13/afero"
	"log"
	"net/http"
	"sort"
	"strings"
	"time"
)

var (
	// DurationAsString enables writing durations as e.g. "21.4ms" instead of "21.432".
	DurationAsString = false
)

// LogWriter returns a new Logger.
// The filesystem fs specifies where request and response bodies will be
// written as files, if enabled by the item's level.
// If fs is nil, the OS filesystem is used.
func LogWriter(lgr zerolog.Logger, fs afero.Fs) logger.Logger {
	//lgr.Timestamp()
	return func(item *logging.LogItem) {
		var ze *zerolog.Event

		if item.Err != nil {
			// Error may contain invalid JSON e.g. `"`
			b, _ := json.Marshal(item.Err.Error())
			b = b[1 : len(b)-1]
			ze = lgr.Error().Bytes("error", b)
		} else {
			ze = lgr.Info()
		}

		// basic info
		ze = ze.
			Str("at", item.Start.Format(zerolog.TimeFieldFormat)).
			Str("method", item.Method).
			Stringer("url", item.URL).
			Int("status", item.StatusCode)

		if DurationAsString {
			ze = ze.Stringer("duration", item.Duration.Round(time.Microsecond))
		} else {
			ze = ze.Dur("duration", item.Duration.Round(time.Microsecond))
		}

		// verbose info
		out := &strings.Builder{} // should this be pooled?
		switch item.Level {
		case logging.WithHeaders:
			ze = printPart(ze, fs, item.Request.Header, true, "", nil, logger.LongBodyThreshold)
			ze = printPart(ze, fs, item.Response.Header, false, "", nil, logger.LongBodyThreshold)

		case logging.WithHeadersAndBodies:
			file := item.FileName()
			ze = printPart(ze, fs, item.Request.Header, true, file, item.Request.Body.Bytes(), logger.LongBodyThreshold)
			ze = printPart(ze, fs, item.Response.Header, false, file, item.Response.Body.Bytes(), logger.LongBodyThreshold)
		}

		ze.Msg(out.String())
	}
}

func printPart(ze *zerolog.Event, fs afero.Fs, hdrs http.Header, isRequest bool, file string, body []byte, longBodyThreshold int) *zerolog.Event {

	prefix := ternary(isRequest, "req", "resp")
	dict := printHeaders(hdrs)
	if dict != nil {
		ze = ze.Dict(prefix+"_headers", dict)
	}
	contentType := hdrs.Get("Content-Type")
	if len(body) == 0 {
		return ze
	}

	suffix := ternary(isRequest, "req", "resp")
	name := fmt.Sprintf("%s_%s", file, suffix)
	justType := strings.SplitN(contentType, ";", 2)[0]
	if len(body) > longBodyThreshold {
		extn := mime.FileExtension(justType)
		if extn != "" {
			ze = writeBodyToFile(ze, fs, prefix, name, extn, body)
		}
		ze = ze.Int(prefix+"_body_len", len(body))

	} else if mime.IsTextual(justType) {
		// write short body inline
		ze = ze.Str(prefix+"_body", strings.Trim(string(body), "\n"))

	} else {
		ze = ze.Int(prefix+"_body_len", len(body))
	}
	return ze
}

func writeBodyToFile(ze *zerolog.Event, fs afero.Fs, prefix, name, extn string, body []byte) *zerolog.Event {
	f, err := fs.Create(name + extn)
	if err != nil {
		log.Printf("logger open file error: %s\n", err)
		return ze
	}

	err = bodypkg.PrettyPrint(extn, f, body)
	if err != nil {
		log.Printf("logger transcode error: %s\n", err)
		return ze
	}

	err = f.Close()
	if err != nil {
		log.Printf("logger close error: %s\n", err)
	}

	return ze.Str(prefix+"_file", name+extn)
}

func printHeaders(hdrs http.Header) *zerolog.Event {
	if len(hdrs) == 0 {
		return nil
	}

	dict := zerolog.Dict()

	var keys []string
	for k := range hdrs {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		vs := hdrs[k]
		if len(vs) == 1 {
			dict = dict.Str(k, vs[0])
		} else {
			dict = dict.Strs(k, vs)
		}
	}

	return dict
}

func ternary(predicate bool, yes, no string) string {
	if predicate {
		return yes
	}
	return no
}

package ech0logger

import (
	"encoding/json"
	"fmt"
	"github.com/rickb777/ech0/v3"
	. "github.com/rickb777/httpclient/internal"
	"github.com/rickb777/httpclient/logging"
	"github.com/rickb777/httpclient/logging/logger"
	"github.com/spf13/afero"
	"strings"
	"time"
)

var (
	// DurationAsString enables writing durations as e.g. "21.4ms" instead of "21.432".
	// The precision used for strings is to the nearest tenth of a millisecond
	// (otherwise the precision is to the nearest microsecond).
	DurationAsString = false
)

// LogWriter returns a new Logger.
// The directory dir specifies where request and response bodies will be
// written as files, if enabled. The current directory is used if dir is "." or blank.
func LogWriter(lgr ech0.Zero, fs afero.Fs) logger.Logger {
	lgr.Timestamp()
	return func(item *logging.LogItem) {
		var ze ech0.ZeroEvent

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
			Stringer("t", item.Start).
			Str("method", item.Method).
			Stringer("url", item.URL).
			Int("status", item.StatusCode)

		if DurationAsString {
			ze = ze.Stringer("duration", item.Duration.Round(100*time.Microsecond))
		} else {
			ze = ze.Dur("duration", item.Duration.Round(time.Microsecond))
		}

		// verbose info
		out := &strings.Builder{} // should this be pooled?
		switch item.Level {
		case logging.WithHeaders:
			//TODO convert to use zerolog Dict
			PrintPart(out, fs, item.Request.Header, true, "", nil, logger.LongBodyThreshold)
			PrintPart(out, fs, item.Response.Header, false, "", nil, logger.LongBodyThreshold)
			fmt.Fprintln(out, "---")

		case logging.WithHeadersAndBodies:
			file := fmt.Sprintf("%s_%s_%s", FilenameTimestamp(item.Start), item.Method, UrlToFilename(item.URL.Path))
			PrintPart(out, fs, item.Request.Header, true, file, item.Request.Body.Bytes(), logger.LongBodyThreshold)
			PrintPart(out, fs, item.Response.Header, false, file, item.Response.Body.Bytes(), logger.LongBodyThreshold)
			fmt.Fprintln(out, "---")
		}

		ze.Msg(out.String())
	}
}

package logger

import (
	"bufio"
	"fmt"
	. "github.com/rickb777/httpclient/internal"
	"github.com/rickb777/httpclient/logging"
	"github.com/spf13/afero"
	"io"
	"log"
	"strings"
	"sync"
	"time"
)

// LongBodyThreshold is the body length threshold beyond which the body
// will be written to a text file (when the content is text). Otherwise
// it is written inline in the log.
var LongBodyThreshold = 100

// TimeFormat sets the appearance of the timestamp that prefixes each logged
// message.
var TimeFormat = "15:04:05"

// Logger is a function that processes log items, usually by writing them
// to a log file.
type Logger func(item *logging.LogItem)

// FileLogger returns a new Logger writing to a file in dir. The name
// of the file is provided.
// The same directory specifies where request and response bodies will be
// written as files. The current directory is used if this is "." or blank.
func FileLogger(name string, fs afero.Fs) (Logger, error) {
	f, err := fs.Create(name)
	if err != nil {
		return nil, err
	}
	return LogWriter(f, fs), nil
}

// LogWriter returns a new Logger. Log items are written without overlapping
// to ensure the log is easily readable.
// The filesystem fs specifies where request and response bodies will be
// written as files, if enabled by the item's level.
// If fs is nil, the OS filesystem is used.
func LogWriter(out io.Writer, fs afero.Fs) Logger {
	if fs == nil {
		fs = afero.NewOsFs()
	}

	var sw io.StringWriter
	if sw1, ok := out.(io.StringWriter); ok {
		sw = sw1
	} else {
		sw = bufio.NewWriterSize(out, 1024)
	}

	return func(item *logging.LogItem) {
		b := &strings.Builder{}

		// basic info
		if item.Err != nil {
			fmt.Fprintf(b, "%s %-8s %s %d %s %v\n", item.Start.Format(TimeFormat), item.Method,
				item.URL, item.StatusCode, item.Duration.Round(time.Microsecond), item.Err)
		} else {
			fmt.Fprintf(b, "%s %-8s %s %d %s\n", item.Start.Format(TimeFormat), item.Method,
				item.URL, item.StatusCode, item.Duration.Round(time.Microsecond))
		}

		// verbose info
		switch item.Level {
		case logging.WithHeaders:
			PrintPart(b, fs, item.Request.Header, true, "", nil, LongBodyThreshold)
			PrintPart(b, fs, item.Response.Header, false, "", nil, LongBodyThreshold)
			fmt.Fprintln(b, "---")

		case logging.WithHeadersAndBodies:
			file := fmt.Sprintf("%s_%s%s_%s", FilenameTimestamp(item.Start), Hostname(item.Request.Header), item.Method, UrlToFilename(item.URL.Path))
			PrintPart(b, fs, item.Request.Header, true, file, item.Request.Body.Bytes(), LongBodyThreshold)
			PrintPart(b, fs, item.Response.Header, false, file, item.Response.Body.Bytes(), LongBodyThreshold)
			fmt.Fprintln(b, "---")
		}

		_, err := sw.WriteString(b.String())
		handleErr(err)
	}
}

func handleErr(err error) {
	if err != nil {
		log.Printf("Warning: LogWriter %s.", err.Error())
	}
}

//-------------------------------------------------------------------------------------------------

type syncWriter struct {
	mu sync.Mutex
	w  io.Writer
}

// SyncWriter wraps a io.Writer so that each call to Write is protected by a
// mutex. Use this to wrap the call to writer's Write method if it is not
// thread safe.
//
// Note that you do not need this wrapper for os.File Write operations on POSIX
// and Windows systems as they are already thread-safe.
func SyncWriter(w io.Writer) io.Writer {
	return &syncWriter{w: w}
}

// Write writes (up to) len(p) bytes from p to the underlying data stream.
// It implements the io.Writer interface.
func (s *syncWriter) Write(p []byte) (n int, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.w.Write(p)
}

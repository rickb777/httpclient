package httpclient

import "io"

// WithFinalNewline adds a final newline to a written stream of bytes. This is
// intended for textual content sent via HTTP, for which the trailing newline
// is an often-forgotten Posix requirement. Wrap an existing Writer, then call
// EnsureFinalNewline after all the content has been written through.
type WithFinalNewline struct {
	W          io.Writer
	hasNewline bool
}

func (d *WithFinalNewline) Write(p []byte) (n int, err error) {
	if len(p) > 0 {
		n, err = d.W.Write(p)
		d.hasNewline = p[len(p)-1] == '\n'
	}
	return n, err
}

func (d *WithFinalNewline) EnsureFinalNewline() error {
	if d.hasNewline {
		return nil
	}
	_, err := d.W.Write([]byte{'\n'})
	return err
}

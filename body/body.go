// Package body provides a buffering utility allowing HTTP request and response
// bodies to be buffered so they can be read multiple times. Normally, the standard
// library http requests and responses do as little buffering as possible. When logging
// or other such processing is needed, it makes sense to buffer the outbound/inbound
// bodies exactly once, if possible, in order to reduce copying to a minimum.
package body

import (
	"bytes"
	"encoding/json"
	"io"
)

// A Body implements the io.Reader, io.Closer amd fmt.Stringer interfaces
// by reading from a byte slice.
// The zero value for Body operates like an empty io.Reader.
// Unlike a bytes.Reader, a Body provides methods to access the byte slice,
// which should not be modified in place.
// In addition, Body also implements io.Closer as a no-op.
type Body struct {
	b []byte
	i int64 // current reading index
}

// NewBody returns a new Body reading from a byte slice.
// It is similar to bytes.NewBuffer.
func NewBody(b []byte) *Body { return &Body{b: b} }

// NewBodyString returns a new Body reading from a string.
// It is similar to bytes.NewBufferString.
func NewBodyString(s string) *Body { return NewBody([]byte(s)) }

// CopyBody consumes a reader and returns its contents.
// If the reader is a *Body or a *bytes.Buffer, no copying is needed.
// Deprecated: use Copy instead.
func CopyBody(rdr io.Reader) (*Body, error) {
	return Copy(rdr)
}

// MustCopy is the same as Copy except that it panics on error.
func MustCopy(rdr io.Reader) *Body {
	b, err := Copy(rdr)
	if err != nil {
		panic(err)
	}
	return b
}

// Copy consumes a reader and returns its contents.
// If the reader is a *Body or a *bytes.Buffer, no copying is needed.
func Copy(rdr io.Reader) (*Body, error) {
	if rdr == nil {
		return nil, nil
	}

	switch v := rdr.(type) {
	case *bytes.Buffer:
		if v == nil {
			return nil, nil
		}
		return NewBody(v.Bytes()), nil
	case *Body:
		if v == nil {
			return nil, nil
		}
		v.i = 0
		return v, nil
	}

	buf := &bytes.Buffer{}
	_, err := buf.ReadFrom(rdr)
	if err == io.EOF || err == io.ErrUnexpectedEOF {
		err = nil
	}
	return NewBody(buf.Bytes()), err
}

// JSON encodes some value as a Body containing JSON data.
// This is a convenience for creating request entities.
// Any JSON encoding errors are silently discarded (e.g. from attempting to encode a Go channel).
func JSON(value any) *Body {
	b := &bytes.Buffer{}
	_ = json.NewEncoder(b).Encode(value)
	return &Body{b: b.Bytes()}
}

//-------------------------------------------------------------------------------------------------

// String gets the byte slice as a string regardless of the current read position.
// b may be nil, which yields a blank string.
func (b *Body) String() string { return string(b.Bytes()) }

// Bytes gets the byte slice regardless of the current read position.
// b may be nil, in which case a nil slice is returned.
func (b *Body) Bytes() []byte {
	if b == nil {
		return nil
	}
	return b.b
}

// Buffer gets the data in a form that is well suited to http.Request.Body.
// b may be nil, in which case nil is returned.
func (b *Body) Buffer() *bytes.Buffer {
	if b == nil {
		return nil
	}
	return bytes.NewBuffer(b.b)
}

// Read reads up to len(p) bytes into p the buffer, stopping if the buffer
// is drained or p is full. The return value n is the number of bytes read.
// If the buffer has no data to return, err is io.EOF (unless len(p) is zero);
// otherwise it is nil.
func (b *Body) Read(p []byte) (n int, err error) {
	if b == nil || b.i >= int64(len(b.b)) {
		return 0, io.EOF
	}
	n = copy(p, b.b[b.i:])
	b.i += int64(n)
	return
}

// Rewind rewinds the read pointer in the Body to zero and returns
// the modified Body. See [Body.Read]. b may be nil.
func (b *Body) Rewind() *Body {
	if b != nil {
		b.i = 0
	}
	return b
}

// Close implements the io.Closer interface as a no-op.
func (r *Body) Close() error {
	return nil
}

// Getter returns a function that allows the body to be read multiple
// times as used by http.Request.GetBody. b may be nil.
func (b *Body) Getter() func() (io.ReadCloser, error) {
	return func() (io.ReadCloser, error) {
		b.Rewind()
		return b, nil
	}
}

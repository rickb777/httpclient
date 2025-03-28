package body

import (
	"bytes"
	"github.com/rickb777/expect"
	"io"
	"io/ioutil"
	"testing"
)

func TestCopy_and_accessors(t *testing.T) {
	cases := []struct {
		input    io.Reader
		expected string
		isNil    bool
	}{
		// with content
		{input: NewBodyString("test string 1"), expected: "test string 1", isNil: false},
		{input: bytes.NewBufferString("test string 2"), expected: "test string 2", isNil: false},
		{input: ioutil.NopCloser(bytes.NewBufferString("test string 3")), expected: "test string 3", isNil: false},
		// various nil values
		{input: nil, expected: "", isNil: true},
		{input: (*bytes.Buffer)(nil), expected: "", isNil: true},
		{input: (*Body)(nil), expected: "", isNil: true},
	}

	for _, c := range cases {
		rdr := MustCopy(c.input)
		expect.String(rdr.Bytes()).ToEqual(t, c.expected)
		expect.String(rdr.String()).ToBe(t, c.expected)
		if c.isNil {
			expect.Any(rdr.Buffer()).ToBeNil(t)
		} else {
			expect.String(rdr.Buffer().Bytes()).ToEqual(t, c.expected)
		}

		if c.input != nil {
			buf := bytes.Buffer{}
			_, err := buf.ReadFrom(rdr)
			expect.Error(err).Not().ToHaveOccurred(t)
			expect.String(buf.String()).ToBe(t, c.expected)
		}
	}
}

func TestRewind(t *testing.T) {
	body := NewBodyString("abcdefghijklmnopqrst")

	p := make([]byte, 4)
	i, err := body.Read(p)
	expect.Error(err).Not().ToHaveOccurred(t)
	expect.Number(i).ToBe(t, 4)
	expect.String(p).ToEqual(t, "abcd")

	i, err = body.Read(p)
	expect.Error(err).Not().ToHaveOccurred(t)
	expect.Number(i).ToBe(t, 4)
	expect.String(p).ToEqual(t, "efgh")

	body = body.Rewind()

	i, err = body.Read(p)
	expect.Error(err).Not().ToHaveOccurred(t)
	expect.Number(i).ToBe(t, 4)
	expect.String(p).ToEqual(t, "abcd")

	body = nil
	p = make([]byte, 4)
	i, err = body.Read(p)
	expect.Error(err).ToHaveOccurred(t)
	expect.Number(i).ToBe(t, 0)
	expect.Slice(p).ToBe(t, []byte{0, 0, 0, 0}...) // unchanged
}

func TestGetter(t *testing.T) {
	body := NewBodyString("abcdefghijklmnopqrst")

	getter := body.Getter()

	//----- 1st pass -----
	rdr, err := getter()
	expect.Error(err).Not().ToHaveOccurred(t)

	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, rdr)
	expect.Error(err).Not().ToHaveOccurred(t)
	expect.String(buf.String()).ToBe(t, "abcdefghijklmnopqrst")

	//----- 2nd pass -----
	rdr, err = getter()
	expect.Error(err).Not().ToHaveOccurred(t)

	buf = new(bytes.Buffer)
	_, err = io.Copy(buf, rdr)
	expect.Error(err).Not().ToHaveOccurred(t)
	expect.String(buf.String()).ToBe(t, "abcdefghijklmnopqrst")
}

func TestClose(t *testing.T) {
	re := NewBodyString("test string")

	// When...
	err := re.Close()

	// Then...
	expect.Error(err).Not().ToHaveOccurred(t)
}

package body

import (
	"bytes"
	"io"
	"io/ioutil"
	"testing"

	. "github.com/onsi/gomega"
)

func TestCopy_and_accessors(t *testing.T) {
	g := NewGomegaWithT(t)

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
		g.Expect(rdr.Bytes()).To(Equal([]byte(c.expected)))
		g.Expect(rdr.String()).To(Equal(c.expected))
		if c.isNil {
			g.Expect(rdr.Buffer()).To(BeNil())
		} else {
			g.Expect(rdr.Buffer().Bytes()).To(Equal([]byte(c.expected)))
		}

		if c.input != nil {
			buf := bytes.Buffer{}
			_, err := buf.ReadFrom(rdr)
			g.Expect(err).NotTo(HaveOccurred())
			g.Expect(buf.String()).To(Equal(c.expected))
		}
	}
}

func TestRewind(t *testing.T) {
	g := NewGomegaWithT(t)
	body := NewBodyString("abcdefghijklmnopqrst")

	p := make([]byte, 4)
	i, err := body.Read(p)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(i).To(Equal(4))
	g.Expect(p).To(Equal([]byte("abcd")))

	i, err = body.Read(p)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(i).To(Equal(4))
	g.Expect(p).To(Equal([]byte("efgh")))

	body = body.Rewind()

	i, err = body.Read(p)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(i).To(Equal(4))
	g.Expect(p).To(Equal([]byte("abcd")))

	body = nil
	p = make([]byte, 4)
	i, err = body.Read(p)
	g.Expect(err).To(HaveOccurred())
	g.Expect(i).To(Equal(0))
	g.Expect(p).To(Equal([]byte{0, 0, 0, 0})) // unchanged
}

func TestGetter(t *testing.T) {
	g := NewGomegaWithT(t)
	body := NewBodyString("abcdefghijklmnopqrst")

	getter := body.Getter()

	//----- 1st pass -----
	rdr, err := getter()
	g.Expect(err).NotTo(HaveOccurred())

	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, rdr)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(buf.String()).To(Equal("abcdefghijklmnopqrst"))

	//----- 2nd pass -----
	rdr, err = getter()
	g.Expect(err).NotTo(HaveOccurred())

	buf = new(bytes.Buffer)
	_, err = io.Copy(buf, rdr)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(buf.String()).To(Equal("abcdefghijklmnopqrst"))
}

func TestClose(t *testing.T) {
	g := NewGomegaWithT(t)
	// Given...
	re := NewBodyString("test string")

	// When...
	err := re.Close()

	// Then...
	g.Expect(err).NotTo(HaveOccurred())
}

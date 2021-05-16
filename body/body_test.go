package body

import (
	"bytes"
	"io"
	"io/ioutil"
	"testing"

	. "github.com/onsi/gomega"
)

func TestCopyBody(t *testing.T) {
	g := NewGomegaWithT(t)
	cases := []struct {
		input    io.ReadCloser
		expected string
	}{
		{NewBodyString("test string"), "test string"},
		{ioutil.NopCloser(bytes.NewBufferString("test string")), "test string"},
		{nil, ""},
	}

	for _, c := range cases {
		rdr, err := CopyBody(c.input)
		g.Expect(err).NotTo(HaveOccurred())
		g.Expect(rdr.Bytes()).To(Equal([]byte(c.expected)))
		g.Expect(rdr.String()).To(Equal(c.expected))

		if c.input != nil {
			buf := bytes.Buffer{}
			_, err = buf.ReadFrom(rdr)
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

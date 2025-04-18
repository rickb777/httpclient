package file

import (
	"github.com/rickb777/expect"
	"testing"
)

func TestUrlToFilename(t *testing.T) {
	defer reset()

	cases := map[string]string{
		"U ":             "",
		"U /":            "",
		"U /aaa/bbb/ccc": "aaa_bbb_ccc",
		`U /A!B"C#D$E%F&G'H(I)J*K+L,/a:b;c<d=e>f?g@h[i\j]k^l` + "`/A{B|C}D~.": "A-B-C-D-E%F&G-H-I-J-K+L,_a:b;c-d=e-f?g@h-i-j-k-l-_A-B-C-D-.",
		`W /A!B"C#D$E%F&G'H(I)J*K+L,/a:b;c<d=e>f?g@h[i\j]k^l` + "`/A{B|C}D~.": "A-B-C-D-E-F-G-H-I-J-K-L-_a-b-c-d-e-f-g-h-i-j-k-l-_A-B-C-D-.",
	}
	for in, exp := range cases {
		switch in[0] {
		case 'U':
			AllowedPunctuationInFilenames = nonWindowsPunct
		case 'W':
			AllowedPunctuationInFilenames = windowsPunct
		}
		act := UrlToFilename(in[2:])
		expect.String(act).Info(in).ToBe(t, exp)
	}
}

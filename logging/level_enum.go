// generated code - do not edit
// github.com/rickb777/enumeration/v2 v2.3.1

package logging

import (
	"fmt"
)

const levelEnumStrings = "OffDiscreteSummaryWithHeadersWithHeadersAndBodies"

var levelEnumIndex = [...]uint16{0, 3, 11, 18, 29, 49}

// AllLevels lists all 5 values in order.
var AllLevels = []Level{
	Off, Discrete, Summary, WithHeaders,
	WithHeadersAndBodies,
}

// String returns the literal string representation of a Level, which is
// the same as the const identifier.
func (i Level) String() string {
	o := i.Ordinal()
	if o < 0 || o >= len(AllLevels) {
		return fmt.Sprintf("Level(%d)", i)
	}
	return levelEnumStrings[levelEnumIndex[o]:levelEnumIndex[o+1]]
}

// Tag returns the string representation of a Level. This is an alias for String.
func (i Level) Tag() string {
	return i.String()
}

// Ordinal returns the ordinal number of a Level.
func (i Level) Ordinal() int {
	switch i {
	case Off:
		return 0
	case Discrete:
		return 1
	case Summary:
		return 2
	case WithHeaders:
		return 3
	case WithHeadersAndBodies:
		return 4
	}
	return -1
}

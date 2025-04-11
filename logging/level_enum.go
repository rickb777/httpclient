// generated code - do not edit
// github.com/rickb777/enumeration/v4 dev

package logging

import (
	"fmt"
)

// AllLevels lists all 5 values in order.
var AllLevels = []Level{
	Off, Discrete, Summary, WithHeaders,
	WithHeadersAndBodies,
}

const (
	levelEnumStrings = "OffDiscreteSummaryWithHeadersWithHeadersAndBodies"
)

var (
	levelEnumIndex = [...]uint16{0, 3, 11, 18, 29, 49}
)

// String returns the literal string representation of a Level, which is
// the same as the const identifier but without prefix or suffix.
func (v Level) String() string {
	o := v.Ordinal()
	return v.toString(o, levelEnumStrings, levelEnumIndex[:])
}

// Ordinal returns the ordinal number of a Level. This is an integer counting
// from zero. It is *not* the same as the const number assigned to the value.
func (v Level) Ordinal() int {
	switch v {
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

func (v Level) toString(o int, concats string, indexes []uint16) string {
	if o < 0 || o >= len(AllLevels) {
		return fmt.Sprintf("Level(%d)", v)
	}
	return concats[indexes[o]:indexes[o+1]]
}

// IsValid determines whether a Level is one of the defined constants.
func (v Level) IsValid() bool {
	return v.Ordinal() >= 0
}

// Int returns the int value, which is not necessarily the same as the ordinal.
// This facilitates polymorphism (see enum.IntEnum).
func (v Level) Int() int {
	return int(v)
}

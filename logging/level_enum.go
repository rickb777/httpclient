// generated code - do not edit
// github.com/rickb777/enumeration/v2 v2.5.0

package logging

import (
	"database/sql/driver"
	"errors"
	"fmt"
	"github.com/rickb777/enumeration/v2/enum"
	"strconv"
	"strings"
)

const (
	levelEnumStrings = "OffDiscreteSummaryWithHeadersWithHeadersAndBodies"
	levelEnumInputs  = "offdiscretesummarywithheaderswithheadersandbodies"
)

var levelEnumIndex = [...]uint16{0, 3, 11, 18, 29, 49}

// AllLevels lists all 5 values in order.
var AllLevels = []Level{
	Off, Discrete, Summary, WithHeaders,
	WithHeadersAndBodies,
}

// AllLevelEnums lists all 5 values in order.
var AllLevelEnums = enum.IntEnums{
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

// Int returns the int value, which is not necessarily the same as the ordinal.
// It serves to facilitate polymorphism (see enum.IntEnum).
func (i Level) Int() int {
	return int(i)
}

// LevelOf returns a Level based on an ordinal number. This is the inverse of Ordinal.
// If the ordinal is out of range, an invalid Level is returned.
func LevelOf(i int) Level {
	if 0 <= i && i < len(AllLevels) {
		return AllLevels[i]
	}
	// an invalid result
	return Off + Discrete + Summary + WithHeaders + WithHeadersAndBodies + 1
}

// IsValid determines whether a Level is one of the defined constants.
func (i Level) IsValid() bool {
	switch i {
	case Off, Discrete, Summary, WithHeaders,
		WithHeadersAndBodies:
		return true
	}
	return false
}

// Parse parses a string to find the corresponding Level, accepting one of the string values or
// a number. The input representation is determined by levelMarshalTextRep. It is used by AsLevel.
// The input case does not matter.
//
// Usage Example
//
//	v := new(Level)
//	err := v.Parse(s)
//	...  etc
func (v *Level) Parse(s string) error {
	return v.parse(s, levelMarshalTextRep)
}

func (v *Level) parse(in string, rep enum.Representation) error {
	if rep == enum.Ordinal {
		if v.parseOrdinal(in) {
			return nil
		}
	} else {
		if v.parseNumber(in) {
			return nil
		}
	}

	s := strings.ToLower(in)

	if v.parseIdentifier(s) {
		return nil
	}

	return errors.New(in + ": unrecognised level")
}

// parseNumber attempts to convert a decimal value
func (v *Level) parseNumber(s string) (ok bool) {
	num, err := strconv.ParseInt(s, 10, 64)
	if err == nil {
		*v = Level(num)
		return v.IsValid()
	}
	return false
}

// parseOrdinal attempts to convert an ordinal value
func (v *Level) parseOrdinal(s string) (ok bool) {
	ord, err := strconv.Atoi(s)
	if err == nil && 0 <= ord && ord < len(AllLevels) {
		*v = AllLevels[ord]
		return true
	}
	return false
}

// parseIdentifier attempts to match an identifier.
func (v *Level) parseIdentifier(s string) (ok bool) {
	var i0 uint16 = 0

	for j := 1; j < len(levelEnumIndex); j++ {
		i1 := levelEnumIndex[j]
		p := levelEnumInputs[i0:i1]
		if s == p {
			*v = AllLevels[j-1]
			return true
		}
		i0 = i1
	}
	return false
}

// AsLevel parses a string to find the corresponding Level, accepting either one of the string values or
// a number. The input representation is determined by levelMarshalTextRep. It wraps Parse.
// The input case does not matter.
func AsLevel(s string) (Level, error) {
	var i = new(Level)
	err := i.Parse(s)
	return *i, err
}

// levelMarshalTextRep controls representation used for XML and other text encodings.
// By default, it is enum.Identifier and quoted strings are used.
var levelMarshalTextRep = enum.Identifier

// MarshalText converts values to a form suitable for transmission via JSON, XML etc.
// The representation is chosen according to levelMarshalTextRep.
func (i Level) MarshalText() (text []byte, err error) {
	return i.marshalText(levelMarshalTextRep, false)
}

// MarshalJSON converts values to bytes suitable for transmission via JSON.
// The representation is chosen according to levelMarshalTextRep.
func (i Level) MarshalJSON() ([]byte, error) {
	return i.marshalText(levelMarshalTextRep, true)
}

func (i Level) marshalText(rep enum.Representation, quoted bool) (text []byte, err error) {
	var bs []byte
	switch rep {
	case enum.Number:
		bs = []byte(strconv.FormatInt(int64(i), 10))
	case enum.Ordinal:
		bs = []byte(strconv.Itoa(i.Ordinal()))
	case enum.Tag:
		if quoted {
			bs = i.quotedString(i.Tag())
		} else {
			bs = []byte(i.Tag())
		}
	default:
		if quoted {
			bs = []byte(i.quotedString(i.String()))
		} else {
			bs = []byte(i.String())
		}
	}
	return bs, nil
}

func (i Level) quotedString(s string) []byte {
	b := make([]byte, len(s)+2)
	b[0] = '"'
	copy(b[1:], s)
	b[len(s)+1] = '"'
	return b
}

// UnmarshalText converts transmitted values to ordinary values.
func (i *Level) UnmarshalText(text []byte) error {
	return i.Parse(string(text))
}

// UnmarshalJSON converts transmitted JSON values to ordinary values. It allows both
// ordinals and strings to represent the values.
func (i *Level) UnmarshalJSON(text []byte) error {
	s := string(text)
	if s == "null" {
		// Ignore null, like in the main JSON package.
		return nil
	}
	s = strings.Trim(s, "\"")
	return i.Parse(s)
}

// levelStoreRep controls database storage via the Scan and Value methods.
// By default, it is enum.Identifier and quoted strings are used.
var levelStoreRep = enum.Identifier

// Scan parses some value, which can be a number, a string or []byte.
// It implements sql.Scanner, https://golang.org/pkg/database/sql/#Scanner
func (i *Level) Scan(value interface{}) (err error) {
	if value == nil {
		return nil
	}

	err = nil
	switch v := value.(type) {
	case int64:
		if levelStoreRep == enum.Ordinal {
			*i = LevelOf(int(v))
		} else {
			*i = Level(v)
		}
	case float64:
		*i = Level(v)
	case []byte:
		err = i.parse(string(v), levelStoreRep)
	case string:
		err = i.parse(v, levelStoreRep)
	default:
		err = fmt.Errorf("%T %+v is not a meaningful level", value, value)
	}

	return err
}

// Value converts the Level to a string.
// It implements driver.Valuer, https://golang.org/pkg/database/sql/driver/#Valuer
func (i Level) Value() (driver.Value, error) {
	switch levelStoreRep {
	case enum.Number:
		return int64(i), nil
	case enum.Ordinal:
		return int64(i.Ordinal()), nil
	case enum.Tag:
		return i.Tag(), nil
	default:
		return i.String(), nil
	}
}

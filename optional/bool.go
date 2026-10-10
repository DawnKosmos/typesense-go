// Package optional provides values that distinguish omission from explicit values.
package optional

import (
	"bytes"
	"fmt"
)

// Bool represents an optional boolean. Its zero value is None.
type Bool uint8

const (
	None Bool = iota
	False
	True
)

// FromBool returns an explicitly set boolean.
func FromBool(value bool) Bool {
	if value {
		return True
	}
	return False
}

// Value returns the boolean and whether it is set to a valid value.
func (b Bool) Value() (value bool, set bool) {
	return b == True, b == False || b == True
}

// IsZero lets encoding/json omit None from fields tagged with omitzero.
func (b Bool) IsZero() bool {
	return b == None
}

// MarshalJSON emits JSON booleans, never the underlying integer.
// None encodes as null when it is not omitted by the containing field's tags.
func (b Bool) MarshalJSON() ([]byte, error) {
	switch b {
	case None:
		return []byte("null"), nil
	case False:
		return []byte("false"), nil
	case True:
		return []byte("true"), nil
	default:
		return nil, fmt.Errorf("optional.Bool: invalid state %d", b)
	}
}

// UnmarshalJSON accepts true, false, and null; null becomes None.
func (b *Bool) UnmarshalJSON(data []byte) error {
	switch {
	case bytes.Equal(bytes.TrimSpace(data), []byte("null")):
		*b = None
	case bytes.Equal(bytes.TrimSpace(data), []byte("false")):
		*b = False
	case bytes.Equal(bytes.TrimSpace(data), []byte("true")):
		*b = True
	default:
		return fmt.Errorf("optional.Bool: expected true, false, or null")
	}
	return nil
}

package parsetypes

import (
	"bytes"
	"encoding/json"
	"fmt"
	"unsafe"
)

// IsScalar returns whether passed type is primitive.
func IsScalar(v any) bool {
	switch v.(type) {
	case string,
		bool,
		uint8,
		uint16,
		uint32,
		uint64,
		int,
		int8,
		int16,
		int32,
		int64,
		float32,
		float64:
		return true
	default:
		return false
	}
}

var _ json.Marshaler = (*TypeStringifyer)(nil)

type TypeStringifyer struct {
	v any
}

func (t *TypeStringifyer) MarshalJSON() ([]byte, error) {
	b := &bytes.Buffer{}
	err := json.NewEncoder(b).Encode(t.v)
	if err != nil {
		return nil, err
	}

	r := bytes.TrimSpace(b.Bytes())
	return r, nil
}

func (t *TypeStringifyer) String() string {
	b := &bytes.Buffer{}
	err := json.NewEncoder(b).Encode(t.v)
	if err != nil {
		b.Reset()
		fmt.Fprintf(b, "%#v", t.v)
	}

	r := bytes.TrimSpace(b.Bytes())
	return BytesAsString(r)
}

// Spew returns a lazy-evaluated stringer that formats a value in human-readable JSON way.
//
// Method used to represent unknown values to user without leaking internal Go type representation.
func Spew(v any) fmt.Stringer {
	return &TypeStringifyer{
		v: v,
	}
}

// BytesAsString casts byte slice into a string without copy.
func BytesAsString(b []byte) string {
	return unsafe.String(unsafe.SliceData(b), len(b))
}

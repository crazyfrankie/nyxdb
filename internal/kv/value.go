package kv

import "encoding/binary"

// Value represents value information that can be associated with a key
// and also contains internal Meta information.
type Value struct {
	Meta      byte
	UserMeta  byte
	ExpiresAt uint64
	Value     []byte

	Version uint64 // This field is not serialized. Only for internal usage.
}

func sizeVarint(target uint64) int {
	var n int
	for {
		n++
		target >>= 7
		if target == 0 {
			break
		}
	}

	return n
}

// EncodedSize gets the length of the encoding required for the corresponding Value.
func (v *Value) EncodedSize() uint32 {
	size := 2 + len(v.Value) // meta + userMeta
	ex := sizeVarint(v.ExpiresAt)

	return uint32(size + ex)
}

// Decode converts the given fragment into a Value structure for use.
// It uses the length of the fragment to infer the length of the Value field.
func (v *Value) Decode(b []byte) {
	v.Meta = b[0]
	v.UserMeta = b[1]
	var size int
	v.ExpiresAt, size = binary.Uvarint(b[2:])
	v.Value = b[2+size:]
}

// Encode encodes the value in the given Value into a fragment
// that is at least v.EncodedSize()
func (v *Value) Encode(b []byte) uint32 {
	b[0] = v.Meta
	b[1] = v.UserMeta
	size := binary.PutUvarint(b[2:], v.ExpiresAt)
	n := copy(b[2+size:], v.Value)

	return uint32(2 + size + n)
}

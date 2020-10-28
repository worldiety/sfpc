package sfpc

import (
	"encoding/binary"
	"github.com/worldiety/byteorder"
	"math"
)

const (
	embeddMinVal = -128
	embeddMaxVal = 123
	tvarpvint    = iota + embeddMaxVal // positive varuint
	tvarnvint                          // negative varuint
	tfloat64
	tfloat32
)

const (
	// MaxLen is the maximum encodable length
	MaxLen              = 9
	maxUvarintThreshold = byteorder.MaxInt56
)

//
// PutFloat is a naive float64 compression algorithm, which requires at worst 9 byte instead of 8.
// At best, it can encode a few low integers with 1 byte, everything else needs always one byte more.
//  * integer like values between −128 and 123 are encoded in a single byte
//  * larger numbers are encoded into either a positive or negative unsigned variable integer encoding.
//  * float values which can be expressed using float32, are encoded like that.
//  * if variable integer encoding would be too large or the precision is required, the float64 will be stored
//    as is, just with the prefix flag.
func PutFloat(dst []byte, v float64) int {
	const epsilon = 1e-9

	// rounding error is small enough, so it can be represented as an integer
	if _, frac := math.Modf(math.Abs(v)); frac < epsilon || frac > 1.0-epsilon {
		vi := int64(v)
		if vi >= int64(embeddMinVal) && vi <= int64(embeddMaxVal) {
			dst[0] = byte(int8(v))
			return 1
		}

		if vi > 0 && vi <= maxUvarintThreshold {
			dst[0] = tvarpvint
			return binary.PutUvarint(dst[1:], uint64(v)) + 1
		}

		if v < 0 && vi >= -(maxUvarintThreshold) {
			dst[0] = tvarnvint
			return binary.PutUvarint(dst[1:], uint64(-v)) + 1
		}

		// worst case, number too large for a variable integer encoding
		dst[0] = tfloat64
		byteorder.LE(dst[1:]).WriteFloat64(v)
		return 9
	}

	// rounding error is small enough for a float32
	tmp := v * 1000
	if _, frac := math.Modf(math.Abs(tmp)); frac < epsilon || frac > 1.0-epsilon && v <= 16777215 {
		dst[0] = tfloat32
		byteorder.LE(dst[1:]).WriteFloat32(float32(v))
		return 5
	}

	// worst case, to complex to compress
	dst[0] = tfloat64
	byteorder.LE(dst[1:]).WriteFloat64(v)
	return 9
}

// Float reads the variable float representation from the buffer and returns the amount of bytes read.
// See PutFloat for details.
func Float(src []byte) (float64, int) {
	ttype := src[0]
	if int8(ttype) <= embeddMaxVal {
		return float64(int8(ttype)), 1
	}

	if ttype == tvarpvint {
		v, l := binary.Uvarint(src[1:])
		return float64(v), l + 1
	}

	if ttype == tvarnvint {
		v, l := binary.Uvarint(src[1:])
		return -float64(v), l + 1
	}

	if ttype == tfloat32 {
		v := byteorder.LE(src[1:]).ReadFloat32()
		return float64(v), 5
	}

	// we avoid throwing a panic here, to allow inline
	v := byteorder.LE(src[1:]).ReadFloat64()
	return v, 9
}

package sfpc

import (
	"encoding/binary"
	"github.com/worldiety/byteorder"
	"github.com/x448/float16"
	"math"
)

const (
	embeddMinVal = -128
	embeddMaxVal = 115
	tscale0      = iota + embeddMaxVal // varint unscaled
	tscale1                            // varint scaled by 10
	tscale2                            // varint scaled by 100
	tscale3                            // varint scaled by 1.000
	tscale4                            // varint scaled by 10.000
	tpinf                              // plus infinity
	tninf                              // negative infinity
	tnan                               // not a number
	tfloat32                           // as a float32
	tfloat64                           // as a float64
)

const (
	// MaxLen is the maximum length, after encoding
	MaxLen              = 9
	maxUvarintThreshold = byteorder.MaxInt56

	// epsilon is used to check if a float fraction is smaller or larger than the given value.
	epsilon = 1e-9
)

// isInt returns for fractions smaller than 10^-8 true and that the value can be expressed as an int64, otherwise false.
func isInt(v float64) bool {
	if v > math.MaxInt64 || v < math.MinInt64 {
		return false
	}

	if _, frac := math.Modf(math.Abs(v)); frac < epsilon || frac > 1.0-epsilon {
		return true
	}

	return false
}

// float64equals returns true, if both floats have a difference which is smaller than 10^-8.
func float64equals(a, b float64) bool {
	if math.Abs(a) <math.Abs(b){
		return math.Abs(a-b)<=(math.Abs(b)*epsilon)
	}else{
		return math.Abs(a-b)<=(math.Abs(a)*epsilon)
	}
	return math.Abs(a-b) < epsilon
}

// asFloat16 returns the float16 value and true, if the given float64 can be expressed.
func asFloat16(v float64) (float16.Float16, bool) {
	if !float64equals(v, float64(float32(v))) {
		return 0, false
	}

	if float16.PrecisionFromfloat32(float32(v)) == float16.PrecisionExact {
		//	return float16.Fromfloat32(float32(v)), true
	}

	return float16.Fromfloat32(float32(v)), true
}

//
// PutFloat is a naive float64 compression algorithm, which requires at worst 9 byte instead of 8.
// At best, it can encode a few low integers with 1 byte, everything else needs always one byte more.
//  * integer like values between −128 and 123 are encoded in a single byte
//  * larger numbers are encoded into either a positive or negative unsigned variable integer encoding.
//  * float values which can be expressed using float32, are encoded like that.
//  * if variable integer encoding would be too large or the precision is required, the float64 will be stored
//    as is, just with the prefix flag.
func PutFloat(dst []byte, v float64) int {

	// check if it is worth inspecting for int representation
	if v < math.MaxInt64 && v > math.MinInt64 {

		// rounding error is small enough, so it can be embedded as an integer
		if _, frac := math.Modf(math.Abs(v)); frac < epsilon || frac > 1.0-epsilon {
			vi := int64(v)
			if vi >= int64(embeddMinVal) && vi <= int64(embeddMaxVal) {
				dst[0] = byte(int8(v))
				return 1
			}
		}

		// scale representation
		scaled := v
		for scaleStep := 0; scaleStep < 3; scaleStep++ {
			// rounding error is small enough, so it can be represented as an integer
			if _, frac := math.Modf(math.Abs(v)); frac < epsilon || frac > 1.0-epsilon {
				var tmp [11]byte
				n := binary.PutVarint(tmp[:], int64(v))
				if n < 8 {
					dst[0] = byte(tscale0 + scaleStep)
					copy(dst[1:], tmp[:n])
					return n + 1
				}
			}

			scaled *= 10
		}
	}

	if math.IsInf(v, 1) {
		dst[0] = tpinf
		return 1
	}

	if math.IsInf(v, -1) {
		dst[0] = tninf
		return 1
	}

	if math.IsNaN(v) {
		dst[0] = tnan
		return 1
	}

	// can be represented as float32
	if math.Abs(v-float64(float32(v))) < epsilon {
		dst[0] = tfloat32
		byteorder.LE(dst[1:]).WriteFloat32(float32(v))
		return 5
	}

	// any other case
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

	if ttype == tnan {
		return math.NaN(), 1
	}

	if ttype == tpinf {
		return math.Inf(1), 1
	}

	if ttype == tninf {
		return math.Inf(-1), 1
	}

	if ttype == tscale0 {
		v, l := binary.Varint(src[1:])
		return -float64(v), l + 1
	}

	if ttype == tscale1 {
		v, l := binary.Varint(src[1:])
		return -float64(v) / 10, l + 1
	}

	if ttype == tscale2 {
		v, l := binary.Varint(src[1:])
		return -float64(v) / 100, l + 1
	}

	if ttype == tscale3 {
		v, l := binary.Varint(src[1:])
		return -float64(v) / 1000, l + 1
	}

	if ttype == tscale4 {
		v, l := binary.Varint(src[1:])
		return -float64(v) / 10000, l + 1
	}

	if ttype == tfloat32 {
		v := byteorder.LE(src[1:]).ReadFloat32()
		return float64(v), 5
	}

	// we avoid throwing a panic here, to allow inline
	v := byteorder.LE(src[1:]).ReadFloat64()
	return v, 9
}

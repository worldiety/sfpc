/*
 * Copyright 2020 Torben Schinke
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package sfpc

import (
	"encoding/binary"
	"math"
)

const (
	tpscale0 = 128 + iota // positive varuint unscaled, 128 is -128 as int8
	tpscale1              // positive varuint scaled by 10
	tpscale2              // positive varuint scaled by 100
	tpscale3              // positive varuint scaled by 1.000
	tpscale4              // positive varuint scaled by 10.000

	tnscale0
	tnscale1 // negative varuint scaled by 10
	tnscale2 // negative varuint scaled by 100
	tnscale3 // negative varuint scaled by 1.000
	tnscale4 // negative varuint scaled by 10.000

	tpinf    // plus infinity
	tninf    // negative infinity
	tnan     // not a number
	tfloat32 // as a float32
	tfloat64 // as a float64, 142 is -114 as int8
)

const (
	embeddedMinSigned = 127 - embeddMaxVal
	embeddedMaxSigned = 127
	embeddMaxVal      = 240 // this is the unsigned representation, the int8 range is -113 to 128

	// MaxLen is the maximum length, after encoding.
	MaxLen = 9

	// epsilon is used to check if a float fraction is smaller or larger than the given value.
	epsilon = 1e-9

	maxInt56        = 1<<55 - 1
	minInt56        = -1 << 55
	varuintlenlimit = 8
	               //-1359034801756235
	maxIntInFloat64 = 9007199254740993 // 2^53−1
)

//
// PutFloat performs a variable length float64 compression algorithm, which requires at worst 9 byte instead of 8.
// It has a lossy accuracy for the fraction part of less than 10^-9. It works best for small decimal floats
// or natural numbers.
func PutFloat(dst []byte, v float64) int { //nolint:funlen,gocognit
	// check if it is worth inspecting for int representation
	//nolint:nestif
	if v < maxInt56 && v > minInt56 {
		// rounding error is small enough, so it can be encoded as an integer. We know, that it cannot be larger
		// than int64 range.
		if _, frac := math.Modf(math.Abs(v)); frac < epsilon || frac > 1.0-epsilon {
			vi := int64(math.Round(v))

			// encode in prefix, if small enough, this also includes signed values until -114
			if vi <= embeddedMaxSigned && vi >= embeddedMinSigned {
				dst[0] = byte(v)

				return 1
			}

			if v < 0 {
				n := binary.PutUvarint(dst[1:], uint64(-v))
				// think about checking for a shorter float32 representation
				if n <= varuintlenlimit {
					dst[0] = tnscale0

					return n + 1
				}
			} else {
				n := binary.PutUvarint(dst[1:], uint64(v))
				// think about checking for a shorter float32 representation
				if n <= varuintlenlimit {
					dst[0] = tpscale0

					return n + 1
				}
			} // jump out and try to encode as float32 or be lossless
		} else if v*10_000 < maxIntInFloat64 && v*10_000 > -maxIntInFloat64 {
			// we must not exceed the lossless representation of natural number in float64, otherwise
			// our scaling will be incorrect
			// we have a fraction, but check if we can scale and encode it efficiently using var uints
			scaler := float64(10) //nolint:gomnd
			for scaleStep := 1; scaleStep <= 4; scaleStep++ {
				scaled := v * scaler
				// rounding error is small enough, so it can be represented as an integer
				if _, frac := math.Modf(math.Abs(scaled)); frac < epsilon || frac > 1.0-epsilon {
					if v < 0 {
						n := binary.PutUvarint(dst[1:], uint64(-math.Round(scaled)))
						// a float32 cannot encode better
						if n <= varuintlenlimit {
							dst[0] = byte(tnscale0 + scaleStep)

							return n + 1
						}

						// need to many bytes, try to encode as float32 or lossless
						break
					} else {
						n := binary.PutUvarint(dst[1:], uint64(math.Round(scaled)))

						// a float32 cannot encode better
						if n <= varuintlenlimit {
							dst[0] = byte(tpscale0 + scaleStep)

							return n + 1
						}

						// need to many bytes, try to encode as float32 or lossless
						break
					}
				}
				scaler *= 10
			}
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

	// check, if a float32 has enough precision.
	// the following line looks like non-sense, also for an optimizing compiler. But gc does not
	// optimize it away and we get proper truncation today. Our unit-tests will proof that.
	// If one day this stops working, we need to use a method with go:noinline compiler directive.
	f32 := float64(float32(v))
	if i, frac := math.Modf(math.Abs(v - f32)); i == 0 && (frac < epsilon || frac > 1.0-epsilon) {
		dst[0] = tfloat32
		binary.LittleEndian.PutUint32(dst[1:], math.Float32bits(float32(v)))

		return 5 //nolint:gomnd
	}

	// any other case
	dst[0] = tfloat64
	binary.LittleEndian.PutUint64(dst[1:], math.Float64bits(v))

	return 9 //nolint:gomnd
}

// Float reads the variable float representation from the buffer and returns the amount of bytes read.
// See PutFloat for details.
func Float(src []byte) (float64, int) { //nolint:gocognit
	prefix := src[0]

	// return embedded value
	if int8(prefix) >= embeddedMinSigned {
		return float64(int8(prefix)), 1
	}

	// uvarint encoding
	//nolint:nestif
	if prefix >= tpscale0 && prefix <= tnscale4 {
		uv, l := binary.Uvarint(src[1:])
		v := float64(uv)

		if prefix >= tnscale0 {
			v *= -1
		}

		if prefix == tnscale0 || prefix == tpscale0 {
			return v, l + 1
		}

		if prefix == tpscale1 || prefix == tnscale1 {
			return v / 10, l + 1
		}

		if prefix == tpscale2 || prefix == tnscale2 {
			return v / 100, l + 1
		}

		if prefix == tpscale3 || prefix == tnscale3 {
			return v / 1000, l + 1
		}

		if prefix == tpscale4 || prefix == tnscale4 {
			return v / 10000, l + 1
		}
	}

	if prefix == tfloat32 {
		v := math.Float32frombits(binary.LittleEndian.Uint32(src[1:]))

		return float64(v), 5 //nolint: gomnd
	}

	if prefix == tnan {
		return math.NaN(), 1
	}

	if prefix == tpinf {
		return math.Inf(1), 1
	}

	if prefix == tninf {
		return math.Inf(-1), 1
	}

	// we avoid throwing a panic here, to allow theoretical inline, but method is probably to complex anyway
	v := math.Float64frombits(binary.LittleEndian.Uint64(src[1:]))

	return v, 9 //nolint:gomnd
}

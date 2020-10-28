# sfpc
sfpc is a simple floating point compressor for single numbers.
Actually it can not really compress floats in general, but it does well for
small integers:
* if the float fits into -128 and 120, it is just encoded as a single byte
* if the float has only a small rounding error, it is represented as varint or varuint. If
the varint would exceed 8 byte, the original float64 is stored, which caps at 9 byte, instead
of 10 (worst case varint) + 1 (our prefix).
* if the float64 can be represented as a float32 within some rounding errors, the float32 is accepted (5 byte).
* in any other case, the float64 format is kept (9 byte, worst case)

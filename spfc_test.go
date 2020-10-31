package sfpc_test

import (
	. "github.com/worldiety/sfpc"

	"encoding/binary"
	"math"
	"math/rand"
	"testing"
)

func BenchmarkEmptyCall(b *testing.B) {
	for i := 0; i < b.N; i++ {
		emptyCall()
	}
}

//go:noinline
func emptyCall() {

}

func BenchmarkPutFloat1(b *testing.B) {
	tmp := make([]byte, MaxLen)

	for i := 0; i < b.N; i++ {
		PutFloat(tmp, 25.25)
	}
}

func BenchmarkPutFloat2(b *testing.B) {
	tmp := make([]byte, MaxLen)

	for i := 0; i < b.N; i++ {
		PutFloat(tmp, 2525.25252525)
	}
}

func BenchmarkPutFloat3(b *testing.B) {
	tmp := make([]byte, MaxLen)

	for i := 0; i < b.N; i++ {
		PutFloat(tmp, math.MaxFloat64)
	}
}

func BenchmarkReadFloat1(b *testing.B) {
	tmp := make([]byte, MaxLen)
	PutFloat(tmp, 25.25)

	for i := 0; i < b.N; i++ {
		Float(tmp)
	}
}

func BenchmarkReadFloat2(b *testing.B) {
	tmp := make([]byte, MaxLen)
	PutFloat(tmp, 2525.25252525)

	for i := 0; i < b.N; i++ {
		Float(tmp)
	}
}

func BenchmarkReadFloat3(b *testing.B) {
	tmp := make([]byte, MaxLen)
	PutFloat(tmp, math.MaxFloat64)

	for i := 0; i < b.N; i++ {
		Float(tmp)
	}
}

func TestPutFloatLong(t *testing.T) {
	const rangeBound = 1_000

	tmp := make([]byte, MaxLen)

	for i := float64(-rangeBound) * 10; i <= rangeBound*10; i++ {
		for f := float64(-rangeBound); f <= rangeBound; f++ {
			v := i + (f * 1 / rangeBound)
			n := PutFloat(tmp, v)

			rv, rn := Float(tmp)
			if rn != n {
				t.Fatalf("%v: expected written length %v equal to read length %v", v, n, rn)
			}

			assertEquals(t, v, rv)
		}
	}
}

func TestPutFloatRandF64(t *testing.T) {
	const rangeBound = 10_000_000

	rd := rand.New(rand.NewSource(1234)) //nolint:gosec
	tmp := make([]byte, MaxLen)

	for i := 0; i < rangeBound; i++ {
		if _, err := rd.Read(tmp); err != nil {
			t.Fatal(err)
		}

		v := math.Float64frombits(binary.LittleEndian.Uint64(tmp))
		n := PutFloat(tmp, v)

		rv, rn := Float(tmp)
		if rn != n {
			t.Fatalf("%v: expected written length %v equal to read length %v", v, n, rn)
		}

		assertEquals(t, v, rv)
	}
}

func TestPutFloatRandF32(t *testing.T) {
	const rangeBound = 10_000_000

	rd := rand.New(rand.NewSource(767)) //nolint:gosec
	tmp := make([]byte, MaxLen)

	for i := 0; i < rangeBound; i++ {
		if _, err := rd.Read(tmp); err != nil {
			t.Fatal(err)
		}

		v := float64(math.Float32frombits(binary.LittleEndian.Uint32(tmp)))
		n := PutFloat(tmp, v)

		rv, rn := Float(tmp)
		if rn != n {
			t.Fatalf("%v: expected written length %v equal to read length %v", v, n, rn)
		}

		assertEquals(t, v, rv)
	}
}

func TestPutFloatRandI64(t *testing.T) {
	const rangeBound = 10_000_000

	rd := rand.New(rand.NewSource(3234)) //nolint:gosec
	tmp := make([]byte, MaxLen)

	for i := 0; i < rangeBound; i++ {
		v := float64(int64(rd.Uint64()))
		n := PutFloat(tmp, v)

		rv, rn := Float(tmp)
		if rn != n {
			t.Fatalf("%v: expected written length %v equal to read length %v", v, n, rn)
		}

		assertEquals(t, v, rv)
	}
}

func TestPutFloatRandI32(t *testing.T) {
	const rangeBound = 10_000_000

	rd := rand.New(rand.NewSource(345)) //nolint:gosec
	tmp := make([]byte, MaxLen)

	for i := 0; i < rangeBound; i++ {
		v := float64(int32(rd.Uint64()))
		n := PutFloat(tmp, v)

		rv, rn := Float(tmp)
		if rn != n {
			t.Fatalf("%v: expected written length %v equal to read length %v", v, n, rn)
		}

		assertEquals(t, v, rv)
	}
}

func TestPutFloatRandI16(t *testing.T) {
	const rangeBound = 10_000_000

	rd := rand.New(rand.NewSource(492)) //nolint:gosec
	tmp := make([]byte, MaxLen)

	for i := 0; i < rangeBound; i++ {
		v := float64(int16(rd.Uint64()))
		n := PutFloat(tmp, v)

		rv, rn := Float(tmp)
		if rn != n {
			t.Fatalf("%v: expected written length %v equal to read length %v", v, n, rn)
		}

		assertEquals(t, v, rv)
	}
}

func TestPutFloat(t *testing.T) {
	tmp := make([]byte, MaxLen)

	testTable := []struct {
		val    float64
		length int
	}{

		{1389062588068379648, 9},
		{-15, 1},

		{0, 1},
		{1, 1},
		{127, 1},
		{128, 3},
		{-113, 1},
		{-114, 2},
		{-115, 2},
		{-127, 2},
		{-128, 3},

		{255, 3},
		{2.55, 3},
		{25.5, 3},
		{256, 3},

		{1, 1},
		{0.1, 2},
		{0.01, 2},
		{0.001, 2},
		{0.0001, 2},
		{0.00001, 5}, // this is a float32

		{1, 1},
		{10, 1},
		{100, 1},
		{1000, 3},
		{10000, 3},
		{100000, 4},

		{1, 1},
		{1.1, 2},
		{1.01, 2},
		{1.001, 3},
		{1.0001, 3},
		{1.00001, 9}, // float32 is not exact enough

		{0, 1},
		{-129, 3},

		{float64(16383), 3},
		{float64(16384), 4},

		{float64(2097151), 4},
		{float64(2097152), 5},

		{float64(134217728), 5},
		{float64(268435455), 5},
		{float64(math.MaxInt32), 6},

		{float64(math.MaxInt64), 5},
		{float64(math.MinInt64), 5},
		{math.MaxFloat64, 9},
		{math.MaxFloat32, 5},

		{math.Inf(1), 1},
		{math.Inf(-1), 1},
		{math.NaN(), 1},


		{115, 1},
		{116, 1},
	}

	for _, s := range testTable {
		wN := PutFloat(tmp, s.val)

		v, rN := Float(tmp)
		if wN != rN {
			t.Fatalf("%v: written %v bytes but read %v bytes back", s.val, wN, rN)
		}

		if wN != s.length {
			t.Fatalf("length for %v should be %v but was %v", s.val, s.length, wN)
		}

		assertEquals(t, s.val, v)
	}

}

func assertEquals(t *testing.T, expected, b float64) {
	const epsilon = 1e-9

	t.Helper()

	if expected == b {
		return
	}

	if math.IsNaN(expected) && math.IsNaN(b) {
		return
	}

	if math.IsInf(expected, -1) && math.IsInf(b, -1) {
		return
	}

	if math.IsInf(expected, 1) && math.IsInf(b, 1) {
		return
	}

	if i, frac := math.Modf(math.Abs(expected - b)); i == 0 && (frac < epsilon || frac > 1.0-epsilon) {
		return
	}

	t.Fatalf("expected \n%.100f (%b) bot got \n%.100f (%b)", expected, math.Float64bits(expected), b, math.Float64bits(b))
}

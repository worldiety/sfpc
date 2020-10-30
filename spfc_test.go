package sfpc

import (
	"encoding/binary"
	"fmt"
	"github.com/worldiety/byteorder"
	"math"
	"math/rand"
	"strconv"
	"testing"
)



func TestBlb(t *testing.T) {
	fmt.Printf("%.100g\n", float64(0.2))

	a,err := strconv.ParseFloat("1.797693134862315708145274237317043567981e+308",64)
	if err !=nil{
		t.Fatal(err)
	}


	fmt.Println(math.Abs(a-(-a)))

	const epsilon = 1e-6

	table := []struct {
		val float64
	}{
		{5},
		{5.1},
		{5.01},
		{5.001},
		{0.2},
		{float64(float32(0.2))},
		{0.23},
		{0.245},
		{12.24},
		{99999999.99},
		{15.15},
		{4.60},
	}
	for _, s := range table {
		v := s.val
		for i := 0; i < 5; i++ {
			if _, frac := math.Modf(math.Abs(v)); frac < epsilon || frac > 1.0-epsilon {
				fmt.Println("is exact for ", v, i, frac)
				var tmp [11]byte
				n := binary.PutVarint(tmp[:], int64(v))
				fmt.Println("as varint requires: ", n, "bytes")
			} else {
				fmt.Println(v, "is not exact")
			}
			v *= 10
		}

	}

}

func TestIsInt(t *testing.T) {
	table := []struct {
		val   float64
		isInt bool
	}{
		{math.MaxFloat64, false},
		{-math.MaxFloat64, false},
		{math.MinInt64, true},
		{math.MaxInt64, true},

		// we cannot test min/max int properly, because float64 cannot express those values anyway
		{math.MinInt64 - 10000, false},
		{math.MaxInt64 + 10000, false},

		{5.0, true},
		{5.1, false},
		{5.01, false},
		{5.0000001, false},
		{5.00000001, true},
		{5.000000001, true},
		{5.0000000001, true},

		{-5.0, true},
		{-5.1, false},
		{-5.01, false},
		{-5.0000001, false},
		{-5.00000001, true},
		{-5.000000001, true},
		{-5.0000000001, true},
	}

	for _, s := range table {
		if isInt(s.val) != s.isInt {
			t.Fatalf("expected %v to be int == %v but is %v", s.val, s.isInt, !s.isInt)
		}
	}
}

func TestFloatIsEqual(t *testing.T) {
	table := []struct {
		a, b    float64
		isEqual bool
	}{
		{0.2, float64(float32(0.2)), true},
		{0.5, float64(float32(0.2) + float32(0.3)), true},
		{0.2, float64(float32(0.2) + float32(0.3) - 0.3), true},
		{math.Pi, float64(float32(math.Pi)), true},

		{5, 5, true},
		{5.1, 5.2, false},
		{5.01, 5.02, false},
		{5.001, 5.002, false},
		{5.0001, 5.0002, false},
		{5.00001, 5.00002, false},
		{5.000001, 5.000002, false},
		{5.0000001, 5.0000002, true},
		{5.00000001, 5.00000002, true},
		{5.000000001, 5.000000002, true},
	}

	for _, s := range table {
		if float64equals(s.a, s.b) != s.isEqual {
			if s.isEqual {
				t.Fatalf("expected %v and %v to be equal", s.a, s.b)
			} else {
				t.Fatalf("expected %v and %v to be unequal", s.a, s.b)
			}

		}
	}
}

func TestFloat16(t *testing.T) {
	table := []struct {
		v         float64
		isFloat16 bool
	}{
		{0.2, true},
		{0.5, true},
		{0.2, true},
		{math.Pi, true},

		{5, true},
		{5.1, false},
		{5.01, false},
		{5.001, false},
		{5.0001, false},
		{5.00001, false},
		{5.000001, false},
		{5.0000001, true},
		{5.00000001, true},
		{5.000000001, true},
	}

	for _, s := range table {
		v, ok := asFloat16(s.v)
		if ok != s.isFloat16 {
			if s.isFloat16 {
				t.Fatalf("expected %v to be representable as float16", s.v)
			} else {
				t.Fatalf("expected %v not to be representable as float16", s.v)
			}
		}

		if !float64equals(float64(v), s.v) {
			t.Fatalf("expected float64 %v but got %v", s.v, v)
		}
	}

}

func TestPutFloat(t *testing.T) {
	rd := rand.New(rand.NewSource(1234))
	tmp := make([]byte, MaxLen)

	testTable := []struct {
		val    float64
		length int
	}{
		{float64(byteorder.MaxInt64), 9},
		{-129, 3},
		{0, 1},
		{-128, 1},

		{123, 1},
		{124, 2},
		{math.NaN(), 9},
		{float64(byteorder.MaxInt56), 9},
	}

	for _, s := range testTable {
		wN := PutFloat(tmp, s.val)
		v, rN := Float(tmp)
		if wN != rN {
			t.Fatalf("written %v bytes but read %v bytes back", wN, rN)
		}

		if wN != s.length {
			t.Fatalf("length for %v should be %v but was %v", s.val, s.length, wN)
		}

		if v != s.val {
			if math.IsNaN(v) && math.IsNaN(s.val) {
				continue
			}

			t.Fatalf("expected \n%v (%b) bot got \n%v (%b)", s.val, math.Float64bits(s.val), v, math.Float64bits(v))
		}
	}

	for i := 0; i < 1_000_000_00; i++ {
		val := int32(rd.Uint32())
		wN := PutFloat(tmp, float64(val))
		v, rN := Float(tmp)
		if wN != rN {
			t.Fatalf("written %v bytes but read %v bytes back", wN, rN)
		}

		if int32(v) != val {
			t.Fatalf("expected %v bot got %v", val, v)
		}
	}

	for i := byteorder.MinInt24; i <= byteorder.MaxInt24; i++ {
		val := int32(rd.Uint32())
		wN := PutFloat(tmp, float64(val))
		v, rN := Float(tmp)
		if wN != rN {
			t.Fatalf("written %v bytes but read %v bytes back", wN, rN)
		}

		if int32(v) != val {
			t.Fatalf("expected %v bot got %v", val, v)
		}
	}

	for i := byteorder.MinInt8; i <= embeddMaxVal; i++ {
		if PutFloat(tmp, float64(i)) != 1 {
			t.Fatal("must be encoded as a single byte")
		}
	}

}

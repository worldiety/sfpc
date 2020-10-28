package sfpc

import (
	"github.com/worldiety/byteorder"
	"math"
	"math/rand"
	"testing"
)

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

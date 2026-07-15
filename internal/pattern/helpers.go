package pattern

import (
	"fmt"
	"math"
)

// sscanfImpl wraps fmt.Sscanf.
func sscanfImpl(s, format string, args ...interface{}) (int, error) {
	return fmt.Sscanf(s, format, args...)
}

// --- Continuous signals (ported from strudel sine/saw/tri/square/rand) ---

// Sine2 is a bipolar sine signal (-1..1).
var Sine2 = Signal(func(t float64) float64 {
	return math.Sin(2 * math.Pi * t)
})

// Sine is unipolar (0..1).
var Sine = Sine2.WithValue(func(v interface{}) interface{} {
	return (v.(float64) + 1) / 2
})

// Cosine2 is bipolar cosine.
var Cosine2 = Signal(func(t float64) float64 {
	return math.Cos(2 * math.Pi * t)
})

// Cosine is unipolar cosine.
var Cosine = Cosine2.WithValue(func(v interface{}) interface{} {
	return (v.(float64) + 1) / 2
})

// Saw is a unipolar sawtooth (0..1).
var Saw = Signal(func(t float64) float64 {
	return math.Mod(t, 1)
})

// Saw2 is a bipolar sawtooth (-1..1).
var Saw2 = Saw.WithValue(func(v interface{}) interface{} {
	return v.(float64)*2 - 1
})

// Isaw is an inverse sawtooth (1..0).
var Isaw = Signal(func(t float64) float64 {
	return 1 - math.Mod(t, 1)
})

// Isaw2 is bipolar inverse sawtooth.
var Isaw2 = Isaw.WithValue(func(v interface{}) interface{} {
	return v.(float64)*2 - 1
})

// Tri is a triangle wave (unipolar).
var Tri = fastcat(Isaw, Saw)

// Tri2 is bipolar triangle.
var Tri2 = fastcat(Isaw2, Saw2)

// Square is a unipolar square (0 or 1).
var Square = Signal(func(t float64) float64 {
	return math.Floor(math.Mod(t*2, 2))
})

// Square2 is bipolar square (-1 or 1).
var Square2 = Square.WithValue(func(v interface{}) interface{} {
	return v.(float64)*2 - 1
})

// RandN generates n random values per cycle.  Returns a function that creates
// the pattern to avoid global state issues.
func RandN(n int) Pattern {
	return Pure(0).Fast(FracInt(int64(n))).WithValue(func(_ interface{}) interface{} {
		return randFloat()
	})
}

// Rand is a continuous random signal (0..1).
var Rand = Signal(func(t float64) float64 {
	return math.Mod(t*0.5, 1.0)
})

// fastcat is the internal version that takes already-reified patterns.
func fastcat(pats ...Pattern) Pattern { return Fastcat(pats...) }

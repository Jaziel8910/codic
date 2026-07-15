package pattern

import (
	"fmt"
	"math"
)

// Fraction represents an exact rational number for Tidal's cyclic time system.
// All time positions and durations use fractions to avoid floating-point drift.
type Fraction struct {
	Num int64
	Den int64
}

// NewFrac creates a reduced fraction num/den.
func NewFrac(num, den int64) Fraction {
	if den == 0 {
		panic("pattern: fraction denominator is zero")
	}
	if den < 0 {
		num = -num
		den = -den
	}
	g := gcd(abs64(num), den)
	return Fraction{Num: num / g, Den: den / g}
}

// FracInt creates a fraction from an integer.
func FracInt(n int64) Fraction {
	return Fraction{Num: n, Den: 1}
}

// FracFloat converts a float64 to a fraction with reasonable precision.
func FracFloat(f float64) Fraction {
	if f == math.Trunc(f) && math.Abs(f) < 1e15 {
		return FracInt(int64(f))
	}
	// Use continued fraction approximation for musical values
	const maxDen int64 = 1000000
	n := int64(1)
	d := int64(1)
	bestN, bestD := n, d
	bestErr := math.Abs(f - float64(n)/float64(d))
	for d <= maxDen {
		if bestErr < 1e-12 {
			break
		}
		cf := int64(math.RoundToEven(f*float64(d)) / float64(1))
		n = cf * d
		err := math.Abs(f - float64(n)/float64(d))
		if err < bestErr {
			bestN, bestD, bestErr = n, d, err
		}
		d++
	}
	return NewFrac(bestN, bestD)
}

// --- Arithmetic ---

func (f Fraction) Add(o Fraction) Fraction {
	return NewFrac(f.Num*o.Den+o.Num*f.Den, f.Den*o.Den)
}

func (f Fraction) Sub(o Fraction) Fraction {
	return NewFrac(f.Num*o.Den-o.Num*f.Den, f.Den*o.Den)
}

func (f Fraction) Mul(o Fraction) Fraction {
	return NewFrac(f.Num*o.Num, f.Den*o.Den)
}

func (f Fraction) Div(o Fraction) Fraction {
	if o.Num == 0 {
		panic("pattern: division by zero fraction")
	}
	return NewFrac(f.Num*o.Den, f.Den*o.Num)
}

func (f Fraction) Neg() Fraction {
	return Fraction{Num: -f.Num, Den: f.Den}
}

// --- Comparisons ---

func (f Fraction) Equals(o Fraction) bool {
	return f.Num*o.Den == o.Num*f.Den
}

func (f Fraction) Cmp(o Fraction) int {
	l := f.Num * o.Den
	r := o.Num * f.Den
	switch {
	case l < r:
		return -1
	case l > r:
		return 1
	default:
		return 0
	}
}

func (f Fraction) Gt(o Fraction) bool  { return f.Cmp(o) > 0 }
func (f Fraction) Lt(o Fraction) bool  { return f.Cmp(o) < 0 }
func (f Fraction) Gte(o Fraction) bool { return f.Cmp(o) >= 0 }
func (f Fraction) Lte(o Fraction) bool { return f.Cmp(o) <= 0 }

func (f Fraction) Max(o Fraction) Fraction {
	if f.Gt(o) {
		return f
	}
	return o
}

func (f Fraction) Min(o Fraction) Fraction {
	if f.Lt(o) {
		return f
	}
	return o
}

// --- Cycle operations ---

// Floor returns the greatest integer <= f (toward negative infinity).
func (f Fraction) Floor() int64 {
	r := f.Num / f.Den
	if (f.Num%f.Den) != 0 && ((f.Num < 0) != (f.Den < 0)) {
		r--
	}
	return r
}

// Ceil returns the smallest integer >= f.
func (f Fraction) Ceil() int64 {
	r := f.Num / f.Den
	if (f.Num%f.Den) != 0 && !((f.Num < 0) != (f.Den < 0)) {
		r++
	}
	return r
}

// Sam returns the beginning of the cycle containing this fraction (floor to integer cycle).
func (f Fraction) Sam() Fraction {
	return FracInt(f.Floor())
}

// NextSam returns the beginning of the next cycle after this fraction.
func (f Fraction) NextSam() Fraction {
	return FracInt(f.Floor() + 1)
}

// WholeCycle returns the TimeSpan of the cycle containing this fraction.
func (f Fraction) WholeCycle() TimeSpan {
	sam := f.Sam()
	return TimeSpan{Begin: sam, End: sam.Add(FracInt(1))}
}

// --- Utilities ---

func (f Fraction) IsZero() bool { return f.Num == 0 }
func (f Fraction) IsInt() bool  { return f.Den == 1 }

// Float64 converts to a float64.
func (f Fraction) Float64() float64 {
	return float64(f.Num) / float64(f.Den)
}

// Show returns a human-readable string.
func (f Fraction) Show() string {
	if f.Den == 1 {
		return fmt.Sprintf("%d", f.Num)
	}
	return fmt.Sprintf("%d/%d", f.Num, f.Den)
}

func (f Fraction) String() string { return f.Show() }

// --- helpers ---

func gcd(a, b int64) int64 {
	for b != 0 {
		a, b = b, a%b
	}
	if a < 0 {
		a = -a
	}
	return a
}

func abs64(n int64) int64 {
	if n < 0 {
		return -n
	}
	return n
}

package pattern

import "math"

// --- Arithmetic on patterns (ported from strudel add/sub/mul/div/union) ---

// asNumberPattern converts the pattern's values to numeric (float64).
func (p Pattern) asNumber() Pattern {
	return p.WithEvent(func(h Hap) Hap {
		switch t := h.Value.(type) {
		case float64:
			return h
		case int:
			return h.WithValue(func(_ interface{}) interface{} { return float64(t) })
		case int64:
			return h.WithValue(func(_ interface{}) interface{} { return float64(t) })
		case Fraction:
			return h.WithValue(func(_ interface{}) interface{} { return t.Float64() })
		case string:
			if f, ok := tryParseNote(t); ok {
				return h.WithValue(func(_ interface{}) interface{} { return f })
			}
			if f, ok := tryParseFloat(t); ok {
				return h.WithValue(func(_ interface{}) interface{} { return f })
			}
		}
		return h.WithValue(func(_ interface{}) interface{} { return 0.0 })
	}).RemoveUndefineds()
}

// perLeft combines this (as numbers) with other using op, keeping this's structure.
func (p Pattern) operLeft(other Pattern, op func(float64, float64) float64) Pattern {
	return p.asNumber().WithValue(func(v interface{}) interface{} {
		return AppFunc(func(b interface{}) interface{} {
			bf := toNumber(b)
			return op(v.(float64), bf)
		})
	}).AppLeft(other.asNumber())
}

// Add adds other's numeric values to this pattern's.
func (p Pattern) Add(other Pattern) Pattern {
	return p.operLeft(other, func(a, b float64) float64 { return a + b })
}

// Sub subtracts other's numeric values from this pattern's.
func (p Pattern) Sub(other Pattern) Pattern {
	return p.operLeft(other, func(a, b float64) float64 { return a - b })
}

// Mul multiplies this pattern's numeric values by other's.
func (p Pattern) Mul(other Pattern) Pattern {
	return p.operLeft(other, func(a, b float64) float64 { return a * b })
}

// Div divides this pattern's numeric values by other's.
func (p Pattern) Div(other Pattern) Pattern {
	return p.operLeft(other, func(a, b float64) float64 { return a / b })
}

// Round rounds numeric values.
func (p Pattern) Round() Pattern {
	return p.asNumber().WithValue(func(v interface{}) interface{} { return math.Round(v.(float64)) })
}

// Floor rounds down.
func (p Pattern) Floor() Pattern {
	return p.asNumber().WithValue(func(v interface{}) interface{} { return math.Floor(v.(float64)) })
}

// Ceil rounds up.
func (p Pattern) Ceil() Pattern {
	return p.asNumber().WithValue(func(v interface{}) interface{} { return math.Ceil(v.(float64)) })
}

// Union merges the control maps of two patterns.
func (p Pattern) Union(other Pattern) Pattern {
	return p.WithValue(func(v interface{}) interface{} {
		return AppFunc(func(b interface{}) interface{} {
			return UnionControls(ToControlMap(v), ToControlMap(b))
		})
	}).AppLeft(other)
}

// Range maps a unipolar (0..1) pattern to the given range.
func (p Pattern) Range(lo, hi float64) Pattern {
	return p.WithValue(func(v interface{}) interface{} {
		f := toFloat(v)
		return lo + (hi-lo)*f
	})
}

// Range2 maps a bipolar (-1..1) pattern to the given range.
func (p Pattern) Range2(lo, hi float64) Pattern {
	return p.WithValue(func(v interface{}) interface{} {
		f := toFloat(v)
		return lo + (hi-lo)*((f+1)/2)
	})
}

// --- helpers ---

func toNumber(v interface{}) float64 {
	switch t := v.(type) {
	case float64:
		return t
	case int:
		return float64(t)
	case int64:
		return float64(t)
	case Fraction:
		return t.Float64()
	case string:
		if f, ok := tryParseNote(t); ok {
			return f
		}
		return 0
	default:
		return 0
	}
}

func toFloat(v interface{}) float64 {
	return toNumber(v)
}

// isTruthy interpreta un valor como booleano (estilo Strudel):
// true si es bool true, número no cero, o string no vacío/"0".
func isTruthy(v interface{}) bool {
	switch t := v.(type) {
	case bool:
		return t
	case float64:
		return t != 0
	case int:
		return t != 0
	case int64:
		return t != 0
	case Fraction:
		return !t.IsZero()
	case string:
		return t != "" && t != "0" && t != "false"
	case nil:
		return false
	default:
		return true
	}
}

func tryParseFloat(s string) (float64, bool) {
	var f float64
	if n, err := fmtSscanf(s, "%f", &f); err == nil && n == 1 {
		return f, true
	}
	return 0, false
}

func tryParseNote(s string) (float64, bool) {
	return noteToMidi(s)
}

// fmtSscanf wraps fmt.Sscanf to avoid importing fmt in this file.
func fmtSscanf(s, format string, args ...interface{}) (int, error) {
	return sscanfImpl(s, format, args...)
}

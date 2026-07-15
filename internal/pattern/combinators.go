package pattern

// --- Time transformations (ported from strudel _fast, _slow, _early, _late) ---

// Fast speeds up a pattern by the given factor (compresses time).
func (p Pattern) Fast(factor Fraction) Pattern {
	return p.WithQueryTime(func(t Fraction) Fraction { return t.Mul(factor) }).
		WithEventTime(func(t Fraction) Fraction { return t.Div(factor) })
}

// Slow is the inverse of Fast.
func (p Pattern) Slow(factor Fraction) Pattern {
	return p.Fast(FracInt(1).Div(factor))
}

// FastGap is like Fast but leaves gaps for factors > 1.
func (p Pattern) FastGap(factor Fraction) Pattern {
	old := p
	qf := func(span TimeSpan) TimeSpan {
		cycle := span.Begin.Sam()
		begin := cycle.Add(span.Begin.Sub(cycle).Mul(factor).Min(FracInt(1)))
		end := cycle.Add(span.End.Sub(cycle).Mul(factor).Min(FracInt(1)))
		return TimeSpan{Begin: begin, End: end}
	}
	ef := func(span TimeSpan) TimeSpan {
		cycle := span.Begin.Sam()
		begin := cycle.Add(span.Begin.Sub(cycle).Div(factor).Min(FracInt(1)))
		end := cycle.Add(span.End.Sub(cycle).Div(factor).Min(FracInt(1)))
		return TimeSpan{Begin: begin, End: end}
	}
	return old.WithQuerySpan(qf).WithEventSpan(ef).SplitQueries()
}

// CompressSpan compresses a pattern into a sub-span of the cycle.
func (p Pattern) CompressSpan(span TimeSpan) Pattern {
	b := span.Begin
	e := span.End
	if b.Gt(e) || b.Gt(FracInt(1)) || e.Gt(FracInt(1)) || b.Lt(FracInt(0)) || e.Lt(FracInt(0)) {
		return Silence()
	}
	return p.FastGap(FracInt(1).Div(e.Sub(b))).Early(b)
}

// Early shifts the pattern earlier in time by the offset.
func (p Pattern) Early(offset Fraction) Pattern {
	return p.WithQueryTime(func(t Fraction) Fraction { return t.Add(offset) }).
		WithEventTime(func(t Fraction) Fraction { return t.Sub(offset) })
}

// Late shifts the pattern later in time by the offset.
func (p Pattern) Late(offset Fraction) Pattern {
	return p.Early(FracInt(0).Sub(offset))
}

// Rev reverses the pattern within each cycle.
func (p Pattern) Rev() Pattern {
	old := p
	return Pattern{Query: func(s State) []Hap {
		span := s.Span
		cycle := span.Begin.Sam()
		nextCycle := span.Begin.NextSam()
		reflect := func(ts TimeSpan) TimeSpan {
			nb := cycle.Add(nextCycle.Sub(ts.End))
			ne := cycle.Add(nextCycle.Sub(ts.Begin))
			return TimeSpan{Begin: nb, End: ne}
		}
		reflectedSpan := reflect(span)
		haps := old.Query(s.SetSpan(reflectedSpan))
		out := make([]Hap, len(haps))
		for i, h := range haps {
			out[i] = h.WithSpan(reflect)
		}
		return out
	}}.SplitQueries()
}

// Struct re-structures the pattern according to a boolean pattern.
func (p Pattern) Struct(binary Pattern) Pattern {
	return binary.WithValue(func(b interface{}) interface{} {
		return AppFunc(func(val interface{}) interface{} {
			if isTruthy(b) {
				return val
			}
			return nil
		})
	}).AppLeft(p).RemoveUndefineds()
}

// Mask lets through only the parts of the pattern where the binary is true.
func (p Pattern) Mask(binary Pattern) Pattern {
	return binary.WithValue(func(b interface{}) interface{} {
		return AppFunc(func(val interface{}) interface{} {
			if isTruthy(b) {
				return val
			}
			return nil
		})
	}).AppRight(p).RemoveUndefineds()
}

// Segment samples the pattern at a fixed rate.
func (p Pattern) Segment(rate Fraction) Pattern {
	return p.Struct(Pure(true).Fast(rate))
}

// Invert swaps true/false in a binary pattern.
func (p Pattern) Invert() Pattern {
	return p.WithValue(func(x interface{}) interface{} { return !isTruthy(x) })
}

// Inv is alias for Invert.
func (p Pattern) Inv() Pattern { return p.Invert() }

// When applies a function conditionally based on a binary pattern.
func (p Pattern) When(binary Pattern, f func(Pattern) Pattern) Pattern {
	truePat := binary.FilterValues(func(v interface{}) bool { return isTruthy(v) })
	falsePat := binary.FilterValues(func(v interface{}) bool { return !isTruthy(v) })
	withPat := truePat.WithValue(func(_ interface{}) interface{} {
		return AppFunc(func(y interface{}) interface{} { return y })
	}).AppRight(f(p))
	withoutPat := falsePat.WithValue(func(_ interface{}) interface{} {
		return AppFunc(func(y interface{}) interface{} { return y })
	}).AppRight(p)
	return Stack(withPat, withoutPat)
}

// Off stacks the pattern with a time-shifted transformation.
func (p Pattern) Off(timePat Pattern, f func(Pattern) Pattern) Pattern {
	return Stack(p, f(p.Late(toFrac(timePat.FirstCycle()[0].Value))))
}

// Every applies a function every n cycles.
func (p Pattern) Every(n int, f func(Pattern) Pattern) Pattern {
	pats := make([]Pattern, n)
	for i := range pats {
		pats[i] = p
	}
	pats[0] = f(p)
	return SlowcatPrime(pats...)
}

// Iter rotates the pattern by 1/n each cycle, n cycles total.
func (p Pattern) Iter(n int) Pattern {
	pats := make([]Pattern, n)
	for i := 0; i < n; i++ {
		pats[i] = p.Early(NewFrac(int64(i), int64(n)))
	}
	return Slowcat(pats...)
}

// StutWith stacks n copies with time offset, each transformed by f.
func (p Pattern) StutWith(times int, time float64, f func(Pattern, int) Pattern) Pattern {
	pats := make([]Pattern, times)
	for i := 0; i < times; i++ {
		pats[i] = f(p.Early(FracFloat(time*float64(i))), i)
	}
	return Stack(pats...)
}

// Stut is a stutter effect with feedback.
func (p Pattern) Stut(times int, feedback, time float64) Pattern {
	return p.StutWith(times, time, func(pat Pattern, i int) Pattern {
		return pat.Velocity(mathPow(feedback, float64(i)))
	})
}

// Jux applies a function to the right channel, panning hard left/right.
func (p Pattern) Jux(f func(Pattern) Pattern) Pattern {
	by := 0.5
	left := p.WithValue(func(v interface{}) interface{} {
		return UnionControls(ToControlMap(v), ControlMap{"pan": 0.5 - by})
	})
	right := p.WithValue(func(v interface{}) interface{} {
		return UnionControls(ToControlMap(v), ControlMap{"pan": 0.5 + by})
	})
	return Stack(left, f(right))
}

// Append concatenates two patterns via fastcat.
func (p Pattern) Append(other Pattern) Pattern {
	return Fastcat(p, other)
}

// StacksWith stacks this pattern with others.
func (p Pattern) StacksWith(others ...Pattern) Pattern {
	return Stack(append([]Pattern{p}, others...)...)
}

func mathPow(base, exp float64) float64 {
	result := 1.0
	for i := 0; i < int(exp); i++ {
		result *= base
	}
	return result
}

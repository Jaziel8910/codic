package pattern

import "fmt"

// Silence is a pattern that produces no events.
var silence = Pattern{Query: func(State) []Hap { return nil }}

// Silence returns a pattern that produces no events.
func Silence() Pattern { return silence }

// Reify turns a value into a Pattern unless it already is one.
func Reify(v interface{}) Pattern {
	if p, ok := v.(Pattern); ok {
		return p
	}
	return Pure(v)
}

// Pure creates a discrete value that repeats once per cycle.
func Pure(value interface{}) Pattern {
	return Pattern{Query: func(s State) []Hap {
		var haps []Hap
		for _, sub := range s.Span.SpanCycles() {
			cycle := sub.Begin.WholeCycle()
			haps = append(haps, Hap{
				Whole: &cycle,
				Part:  sub,
				Value: value,
			})
		}
		return haps
	}}
}

// Steady creates a continuous value (no discrete onset structure).
func Steady(value interface{}) Pattern {
	return Pattern{Query: func(s State) []Hap {
		return []Hap{{Part: s.Span, Value: value}}
	}}
}

// Signal creates a continuous signal sampled at the midpoint of each query.
func Signal(f func(float64) float64) Pattern {
	return Pattern{Query: func(s State) []Hap {
		return []Hap{{
			Part:  s.Span,
			Value: f(s.Span.Midpoint().Float64()),
		}}
	}}
}

// Stacks multiple patterns so their events overlap in time.
func Stack(pats ...Pattern) Pattern {
	return Pattern{Query: func(s State) []Hap {
		var all []Hap
		for _, p := range pats {
			all = append(all, p.Query(s)...)
		}
		return all
	}}
}

// Slowcat concatenates patterns, one per cycle, without compressing.
func Slowcat(pats ...Pattern) Pattern {
	if len(pats) == 0 {
		return Silence()
	}
	return Pattern{Query: func(s State) []Hap {
		span := s.Span
		n := int64(len(pats))
		patIdx := span.Begin.Floor() % n
		if patIdx < 0 {
			patIdx += n
		}
		pat := pats[patIdx]
		// Offset so constituent patterns don't skip cycles
		offset := FracInt(span.Begin.Floor()).Sub(FracInt(span.Begin.Div(FracInt(n)).Floor()))
		return pat.WithEventTime(func(t Fraction) Fraction { return t.Add(offset) }).
			Query(s.SetSpan(span.WithTime(func(t Fraction) Fraction { return t.Sub(offset) })))
	}}.SplitQueries()
}

// SlowcatPrime is Slowcat but skips cycles (no offset math).
func SlowcatPrime(pats ...Pattern) Pattern {
	if len(pats) == 0 {
		return Silence()
	}
	return Pattern{Query: func(s State) []Hap {
		n := int64(len(pats))
		idx := s.Span.Begin.Floor() % n
		if idx < 0 {
			idx += n
		}
		return pats[idx].Query(s)
	}}.SplitQueries()
}

// Fastcat concatenates patterns, compressing each to fit within one cycle.
func Fastcat(pats ...Pattern) Pattern {
	return Slowcat(pats...).Fast(FracInt(int64(len(pats))))
}

// Cat is an alias for Fastcat.
func Cat(pats ...Pattern) Pattern { return Fastcat(pats...) }

// TimeCat concatenates patterns with temporal weights.
// Each arg is [Fraction{time}, Pattern].
func TimeCat(timePats ...[2]interface{}) Pattern {
	var total Fraction
	pats := make([]Pattern, 0, len(timePats))
	total = FracInt(0)
	for _, tp := range timePats {
		timeFrac := toFrac(tp[0])
		total = total.Add(timeFrac)
	}
	begin := FracInt(0)
	for _, tp := range timePats {
		timeFrac := toFrac(tp[0])
		end := begin.Add(timeFrac)
		// Compress span to [begin/total, end/total] and stack
		startFrac := begin.Div(total)
		endFrac := end.Div(total)
		pat := Reify(tp[1])
		pats = append(pats, compressSpan(pat, TimeSpan{Begin: startFrac, End: endFrac}))
		begin = end
	}
	return Stack(pats...)
}

func compressSpan(p Pattern, span TimeSpan) Pattern {
	b := span.Begin
	e := span.End
	if b.Gt(e) || b.Gt(FracInt(1)) || e.Gt(FracInt(1)) || b.Lt(FracInt(0)) || e.Lt(FracInt(0)) {
		return Silence()
	}
	factor := FracInt(1).Div(e.Sub(b))
	return p.FastGap(factor).Late(b)
}

// Sequence builds a pattern from args, auto-nesting slices.
// Sequence(a, []interface{}{b, c}, d) → fastcat(a, fastcat(b, c), d)
func Sequence(args ...interface{}) Pattern {
	return sequenceCount(args)
}

func sequenceCount(xs []interface{}) Pattern {
	if len(xs) == 0 {
		return Silence()
	}
	if len(xs) == 1 {
		return sequenceCountSingle(xs[0])
	}
	pats := make([]Pattern, len(xs))
	for i, x := range xs {
		pats[i] = sequenceCountSingle(x)
	}
	return Fastcat(pats...)
}

func sequenceCountSingle(x interface{}) Pattern {
	if s, ok := x.([]interface{}); ok {
		if len(s) == 0 {
			return Silence()
		}
		return sequenceCount(s)
	}
	return Reify(x)
}

// Polyrhythm stacks independently-timed sequences.
func Polyrhythm(xs ...interface{}) Pattern {
	if len(xs) == 0 {
		return Silence()
	}
	pats := make([]Pattern, 0, len(xs))
	for _, x := range xs {
		pats = append(pats, Sequence(x))
	}
	return Stack(pats...)
}

// Polymeter aligns patterns of different step counts to a common step.
func Polymeter(steps int, args ...interface{}) Pattern {
	if len(args) == 0 {
		return Silence()
	}
	pats := make([]Pattern, 0, len(args))
	for _, x := range args {
		seqPat, count := sequenceCountSingleN(x)
		if count == 0 {
			continue
		}
		if steps == count {
			pats = append(pats, seqPat)
		} else {
			pats = append(pats, seqPat.Fast(NewFrac(int64(steps), int64(count))))
		}
	}
	return Stack(pats...)
}

func sequenceCountSingleN(x interface{}) (Pattern, int) {
	if s, ok := x.([]interface{}); ok {
		if len(s) == 0 {
			return Silence(), 0
		}
		if len(s) == 1 {
			return sequenceCountSingleN(s[0])
		}
		pats := make([]Pattern, len(s))
		for i, v := range s {
			pats[i], _ = sequenceCountSingleN(v)
		}
		return Fastcat(pats...), len(s)
	}
	return Reify(x), 1
}

func toFrac(v interface{}) Fraction {
	switch t := v.(type) {
	case Fraction:
		return t
	case int:
		return FracInt(int64(t))
	case int64:
		return FracInt(t)
	case float64:
		return FracFloat(t)
	default:
		panic(fmt.Sprintf("pattern: cannot convert %v to Fraction", v))
	}
}

package pattern

// Pattern is the heart of the Tidal/Strudel system: a pure function from
// State (timespan) to a slice of Haps (events).  Patterns are immutable —
// every transformation returns a new Pattern that wraps the old one.
type Pattern struct {
	Query func(State) []Hap
}

// --- Core construction helpers (ported from strudel pattern.mjs) ---

// SplitQueries splits queries at cycle boundaries so each returned hap
// is constrained to a single cycle.  This simplifies many transformations.
func (p Pattern) SplitQueries() Pattern {
	old := p
	return Pattern{Query: func(state State) []Hap {
		var result []Hap
		for _, sub := range state.Span.SpanCycles() {
			result = append(result, old.Query(state.SetSpan(sub))...)
		}
		return result
	}}
}

// WithQuerySpan transforms the query timespan before querying.
func (p Pattern) WithQuerySpan(f func(TimeSpan) TimeSpan) Pattern {
	old := p
	return Pattern{Query: func(s State) []Hap {
		return old.Query(s.WithSpan(f))
	}}
}

// WithQueryTime transforms both begin and end of the query timespan.
func (p Pattern) WithQueryTime(f func(Fraction) Fraction) Pattern {
	return p.WithQuerySpan(func(ts TimeSpan) TimeSpan { return ts.WithTime(f) })
}

// WithEventSpan transforms the timespan of every returned hap.
func (p Pattern) WithEventSpan(f func(TimeSpan) TimeSpan) Pattern {
	old := p
	return Pattern{Query: func(s State) []Hap {
		haps := old.Query(s)
		out := make([]Hap, len(haps))
		for i, h := range haps {
			out[i] = h.WithSpan(f)
		}
		return out
	}}
}

// WithEventTime transforms begin and end of every hap's timespan.
func (p Pattern) WithEventTime(f func(Fraction) Fraction) Pattern {
	return p.WithEventSpan(func(ts TimeSpan) TimeSpan { return ts.WithTime(f) })
}

// WithEvents transforms the whole hap list.
func (p Pattern) WithEvents(f func([]Hap) []Hap) Pattern {
	old := p
	return Pattern{Query: func(s State) []Hap { return f(old.Query(s)) }}
}

// WithEvent transforms each hap individually.
func (p Pattern) WithEvent(f func(Hap) Hap) Pattern {
	return p.WithEvents(func(haps []Hap) []Hap {
		out := make([]Hap, len(haps))
		for i, h := range haps {
			out[i] = f(h)
		}
		return out
	})
}

// WithValue maps a function over each hap's value (Tidal's fmap).
func (p Pattern) WithValue(f func(interface{}) interface{}) Pattern {
	old := p
	return Pattern{Query: func(s State) []Hap {
		haps := old.Query(s)
		out := make([]Hap, len(haps))
		for i, h := range haps {
			out[i] = h.WithValue(f)
		}
		return out
	}}
}

// Fmap is an alias for WithValue.
func (p Pattern) Fmap(f func(interface{}) interface{}) Pattern { return p.WithValue(f) }

// FilterValues keeps only haps whose value passes the test.
func (p Pattern) FilterValues(test func(interface{}) bool) Pattern {
	old := p
	return Pattern{Query: func(s State) []Hap {
		haps := old.Query(s)
		out := haps[:0]
		for _, h := range haps {
			if test(h.Value) {
				out = append(out, h)
			}
		}
		return out
	}}
}

// RemoveUndefineds drops haps with nil values.
func (p Pattern) RemoveUndefineds() Pattern {
	return p.FilterValues(func(v interface{}) bool { return v != nil })
}

// OnsetsOnly keeps only haps that contain their onset.
func (p Pattern) OnsetsOnly() Pattern {
	return p.WithEvents(func(haps []Hap) []Hap {
		out := haps[:0]
		for _, h := range haps {
			if h.HasOnset() {
				out = append(out, h)
			}
		}
		return out
	})
}

// WithContext updates each hap's context.
func (p Pattern) WithContext(f func(HapContext) HapContext) Pattern {
	return p.WithEvent(func(h Hap) Hap { return h.SetContext(f(h.Context)) })
}

// WithLocation adds a source code location to each hap's context.
func (p Pattern) WithLocation(start, end [3]int) Pattern {
	loc := Location{
		Start: SourcePos{Line: start[0], Column: start[1], Offset: start[2]},
		End:   SourcePos{Line: end[0], Column: end[1], Offset: end[2]},
	}
	return p.WithContext(func(ctx HapContext) HapContext {
		locs := append([]Location{}, ctx.Locations...)
		locs = append(locs, loc)
		return HapContext{Locations: locs, Velocity: ctx.Velocity, HasVelocity: ctx.HasVelocity, Type: ctx.Type}
	})
}

// FirstCycle queries cycle 0..1, stripping context.  Useful for debugging.
func (p Pattern) FirstCycle() []Hap {
	return p.SplitQueries().Query(State{Span: TimeSpan{Begin: FracInt(0), End: FracInt(1)}})
}

// Apply applies a function to the pattern (pipeline helper).
func (p Pattern) Apply(f func(Pattern) Pattern) Pattern { return f(p) }

// Layer stacks multiple transformations of this pattern.
func (p Pattern) Layer(funcs ...func(Pattern) Pattern) Pattern {
	pats := make([]Pattern, 0, len(funcs))
	for _, f := range funcs {
		pats = append(pats, f(p))
	}
	return Stack(pats...)
}

// Edit stacks the pattern with the results of transformation functions.
func (p Pattern) Edit(funcs ...func(Pattern) Pattern) Pattern { return p.Layer(funcs...) }

// Pipe passes the pattern through a function.
func (p Pattern) Pipe(f func(Pattern) Pattern) Pattern { return f(p) }

// Superimpose stacks the original pattern with transformations of it.
func (p Pattern) Superimpose(funcs ...func(Pattern) Pattern) Pattern {
	pats := make([]Pattern, 0, len(funcs)+1)
	pats = append(pats, p)
	for _, f := range funcs {
		pats = append(pats, f(p))
	}
	return Stack(pats...)
}

// Duration sets the absolute duration of each event.
func (p Pattern) Duration(d Fraction) Pattern {
	return p.WithEventSpan(func(ts TimeSpan) TimeSpan {
		return TimeSpan{Begin: ts.Begin, End: ts.Begin.Add(d)}
	})
}

// Legato sets relative legato (multiplies each event's duration).
func (p Pattern) Legato(v Fraction) Pattern {
	return p.WithEventSpan(func(ts TimeSpan) TimeSpan {
		return TimeSpan{Begin: ts.Begin, End: ts.Begin.Add(ts.End.Sub(ts.Begin).Mul(v))}
	})
}

// Velocity multiplies the event's velocity in its context.
func (p Pattern) Velocity(v float64) Pattern {
	return p.WithContext(func(ctx HapContext) HapContext {
		if !ctx.HasVelocity {
			ctx.Velocity = 1.0
			ctx.HasVelocity = true
		}
		ctx.Velocity *= v
		return ctx
	})
}

// Bypass returns silence if on is truthy, else the pattern itself.
func (p Pattern) Bypass(on int) Pattern {
	if on != 0 {
		return Silence()
	}
	return p
}

// Hush returns silence.
func (p Pattern) Hush() Pattern { return Silence() }

// --- Applicative functor (ported from strudel appLeft/appRight/appBoth) ---

// AppLeft applies this pattern of AppFunc values to the other pattern of values,
// using the function pattern's whole/part spans. (Tidal's <*)
func (p Pattern) AppLeft(other Pattern) Pattern {
	funcPat := p
	valPat := other
	return Pattern{Query: func(state State) []Hap {
		var haps []Hap
		for _, fh := range funcPat.Query(state) {
			valHaps := valPat.Query(state.SetSpan(fh.Part))
			for _, vh := range valHaps {
				part, ok := intersect(fh.Part, vh.Part)
				if !ok {
					continue
				}
				fn, fok := fh.Value.(AppFunc)
				if !fok {
					continue
				}
				haps = append(haps, Hap{
					Whole:   fh.Whole,
					Part:    part,
					Value:   fn(vh.Value),
					Context: MergeContext(fh.Context, vh.Context),
				})
			}
		}
		return haps
	}}
}

// AppRight applies this pattern of AppFunc values, using the value pattern's
// whole/part spans. (Tidal's *>)
func (p Pattern) AppRight(other Pattern) Pattern {
	funcPat := p
	valPat := other
	return Pattern{Query: func(state State) []Hap {
		var haps []Hap
		for _, vh := range valPat.Query(state) {
			funcHaps := funcPat.Query(state.SetSpan(vh.Part))
			for _, fh := range funcHaps {
				part, ok := intersect(fh.Part, vh.Part)
				if !ok {
					continue
				}
				fn, fok := fh.Value.(AppFunc)
				if !fok {
					continue
				}
				haps = append(haps, Hap{
					Whole:   vh.Whole,
					Part:    part,
					Value:   fn(vh.Value),
					Context: MergeContext(fh.Context, vh.Context),
				})
			}
		}
		return haps
	}}
}

// AppBoth applies this pattern of AppFunc values to the other pattern,
// using the intersection of whole/part spans. (Tidal's <*>)
func (p Pattern) AppBoth(other Pattern) Pattern {
	funcPat := p
	valPat := other
	return Pattern{Query: func(state State) []Hap {
		var haps []Hap
		for _, fh := range funcPat.Query(state) {
			for _, vh := range valPat.Query(state) {
				part, ok := intersect(fh.Part, vh.Part)
				if !ok {
					continue
				}
				fn, fok := fh.Value.(AppFunc)
				if !fok {
					continue
				}
				var whole *TimeSpan
				if fh.Whole != nil && vh.Whole != nil {
					w, wok := intersect(*fh.Whole, *vh.Whole)
					if wok {
						whole = &w
					}
				}
				haps = append(haps, Hap{
					Whole:   whole,
					Part:    part,
					Value:   fn(vh.Value),
					Context: MergeContext(fh.Context, vh.Context),
				})
			}
		}
		return haps
	}}
}

// Join flattens a pattern of patterns into a single pattern.
func (p Pattern) Join() Pattern {
	old := p
	return Pattern{Query: func(state State) []Hap {
		var allHaps []Hap
		for _, h := range old.Query(state) {
			inner, ok := h.Value.(Pattern)
			if !ok {
				allHaps = append(allHaps, h)
				continue
			}
			for _, innerHap := range inner.Query(state.SetSpan(h.Part)) {
				var whole *TimeSpan
				if h.Whole != nil && innerHap.Whole != nil {
					w, ok := intersect(*h.Whole, *innerHap.Whole)
					if ok {
						whole = &w
					}
				}
				allHaps = append(allHaps, Hap{
					Whole:   whole,
					Part:    innerHap.Part,
					Value:   innerHap.Value,
					Context: MergeContext(h.Context, innerHap.Context),
				})
			}
		}
		return allHaps
	}}
}

// --- helpers ---

func intersect(a, b TimeSpan) (TimeSpan, bool) {
	r := a.Intersection(b)
	if r == nil {
		return TimeSpan{}, false
	}
	return *r, true
}

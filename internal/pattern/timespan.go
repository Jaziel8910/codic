package pattern

// TimeSpan represents a half-open interval [Begin, End) in Tidal's cyclic time.
type TimeSpan struct {
	Begin Fraction
	End   Fraction
}

// SpanCycles splits this timespan at cycle boundaries, returning sub-spans
// that each fit entirely within a single cycle. This is the core operation
// used by _splitQueries to simplify pattern queries.
func (t TimeSpan) SpanCycles() []TimeSpan {
	var spans []TimeSpan
	begin := t.Begin
	end := t.End
	endSam := end.Sam()
	for end.Gt(begin) {
		if begin.Sam().Equals(endSam) {
			spans = append(spans, TimeSpan{Begin: begin, End: t.End})
			break
		}
		nextBegin := begin.NextSam()
		spans = append(spans, TimeSpan{Begin: begin, End: nextBegin})
		begin = nextBegin
	}
	return spans
}

// WithTime applies a function to both Begin and End.
func (t TimeSpan) WithTime(f func(Fraction) Fraction) TimeSpan {
	return TimeSpan{Begin: f(t.Begin), End: f(t.End)}
}

// WithEnd applies a function to End only.
func (t TimeSpan) WithEnd(f func(Fraction) Fraction) TimeSpan {
	return TimeSpan{Begin: t.Begin, End: f(t.End)}
}

// Intersection returns the overlap of two timespans, or nil if they don't overlap.
func (t TimeSpan) Intersection(o TimeSpan) *TimeSpan {
	begin := t.Begin.Max(o.Begin)
	end := t.End.Min(o.End)
	if begin.Gt(end) {
		return nil
	}
	if begin.Equals(end) {
		// Zero-width intersection only valid at the very start
		if begin.Equals(t.End) && t.Begin.Lt(t.End) {
			return nil
		}
		if begin.Equals(o.End) && o.Begin.Lt(o.End) {
			return nil
		}
	}
	ts := TimeSpan{Begin: begin, End: end}
	return &ts
}

// Midpoint returns the center of the timespan.
func (t TimeSpan) Midpoint() Fraction {
	return t.Begin.Add(t.End.Sub(t.Begin).Div(FracInt(2)))
}

// Equals checks structural equality.
func (t TimeSpan) Equals(o TimeSpan) bool {
	return t.Begin.Equals(o.Begin) && t.End.Equals(o.End)
}

// Show returns a readable string.
func (t TimeSpan) Show() string {
	return t.Begin.Show() + " -> " + t.End.Show()
}

func (t TimeSpan) String() string { return t.Show() }

package pattern

import "fmt"

// AppFunc is a function value used by the applicative functor operations
// (appLeft, appRight, appBoth). When a Pattern carries AppFunc values,
// the applicative methods combine function-patterns with value-patterns.
type AppFunc func(interface{}) interface{}

// ControlMap is the set of control parameters for an event.
// Keys are parameter names like "note", "s", "gain", "cutoff"; values are
// strings, floats, or whatever the parameter carries.
type ControlMap map[string]interface{}

// UnionControls merges b into a (b wins on conflicts).
func UnionControls(a, b ControlMap) ControlMap {
	result := make(ControlMap, len(a)+len(b))
	for k, v := range a {
		result[k] = v
	}
	for k, v := range b {
		result[k] = v
	}
	return result
}

// ToControlMap converts any value into a ControlMap.  If the value is
// already a ControlMap it is returned as-is; otherwise it becomes the
// "s" key (the default sound selector, matching Tidal behaviour).
func ToControlMap(v interface{}) ControlMap {
	if cm, ok := v.(ControlMap); ok {
		return cm
	}
	return ControlMap{"s": v}
}

// HapContext stores metadata about an event: source locations, velocity, type.
type HapContext struct {
	Locations   []Location
	Velocity    float64
	Type        string // "midi", etc.
	HasVelocity bool
}

// Location records a source code position for real-time highlighting.
type Location struct {
	Start SourcePos
	End   SourcePos
}

type SourcePos struct {
	Line   int
	Column int
	Offset int
}

// MergeContext combines two contexts, concatenating locations and preferring
// velocity from b if present.
func MergeContext(a, b HapContext) HapContext {
	locs := make([]Location, 0, len(a.Locations)+len(b.Locations))
	locs = append(locs, a.Locations...)
	locs = append(locs, b.Locations...)
	c := HapContext{
		Locations:   locs,
		Type:        a.Type,
		Velocity:    a.Velocity,
		HasVelocity: a.HasVelocity,
	}
	if b.HasVelocity {
		c.HasVelocity = true
		c.Velocity = a.Velocity * b.Velocity
	}
	if b.Type != "" {
		c.Type = b.Type
	}
	return c
}

// Hap is a single event: a value active during the timespan Part.  Whole
// is the timespan of the event's "whole" extent (may be nil for continuous
// values).  This is called Hap (not Event) to match Strudel's terminology.
type Hap struct {
	Whole   *TimeSpan // nil for continuous signals
	Part    TimeSpan
	Value   interface{} // string, float64, ControlMap, or AppFunc
	Context HapContext
}

// WithSpan returns a new Hap with the function applied to both whole and part.
func (h Hap) WithSpan(f func(TimeSpan) TimeSpan) Hap {
	var whole *TimeSpan
	if h.Whole != nil {
		w := f(*h.Whole)
		whole = &w
	}
	return Hap{Whole: whole, Part: f(h.Part), Value: h.Value, Context: h.Context}
}

// WithValue returns a new Hap with the function applied to the value.
func (h Hap) WithValue(f func(interface{}) interface{}) Hap {
	return Hap{Whole: h.Whole, Part: h.Part, Value: f(h.Value), Context: h.Context}
}

// HasOnset returns true when the part begins at the same time as the whole.
func (h Hap) HasOnset() bool {
	return h.Whole != nil && h.Whole.Begin.Equals(h.Part.Begin)
}

// SetContext returns a new Hap with the given context.
func (h Hap) SetContext(ctx HapContext) Hap {
	return Hap{Whole: h.Whole, Part: h.Part, Value: h.Value, Context: ctx}
}

// Show returns a readable representation.
func (h Hap) Show() string {
	whole := "~"
	if h.Whole != nil {
		whole = h.Whole.Show()
	}
	return "(" + whole + ", " + h.Part.Show() + ", " + fmtVal(h.Value) + ")"
}

func (h Hap) String() string { return h.Show() }

func fmtVal(v interface{}) string {
	switch t := v.(type) {
	case string:
		return t
	case float64:
		return fmt.Sprintf("%g", t)
	case ControlMap:
		s := "{"
		first := true
		for k, val := range t {
			if !first {
				s += ", "
			}
			s += k + ": " + fmtVal(val)
			first = false
		}
		return s + "}"
	default:
		return fmt.Sprintf("%v", v)
	}
}

package pattern

// State is the input to a Pattern query: a timespan plus optional controls.
type State struct {
	Span     TimeSpan
	Controls map[string]interface{}
}

// SetSpan returns a new State with a different span.
func (s State) SetSpan(span TimeSpan) State {
	return State{Span: span, Controls: s.Controls}
}

// WithSpan applies a function to the span.
func (s State) WithSpan(f func(TimeSpan) TimeSpan) State {
	return s.SetSpan(f(s.Span))
}

// SetControls returns a new State with different controls.
func (s State) SetControls(c map[string]interface{}) State {
	return State{Span: s.Span, Controls: c}
}

package audio

// Envelope is a standard ADSR (Attack/Decay/Sustain/Release) amplitude envelope.
type Envelope struct {
	attack  float64 // seconds
	decay   float64 // seconds
	sustain float64 // 0..1
	release float64 // seconds
}

// NewEnvelope creates an ADSR envelope.
func NewEnvelope(attack, decay, sustain, release float64) *Envelope {
	return &Envelope{
		attack:  attack,
		decay:   decay,
		sustain: clamp(sustain, 0, 1),
		release: release,
	}
}

// Sample returns the amplitude at a given elapsed time (seconds).
func (e *Envelope) Sample(elapsed float64) float64 {
	a := e.attack
	d := e.decay
	s := e.sustain
	r := e.release

	switch {
	case elapsed < a:
		// Attack phase: 0 → 1
		if a <= 0 {
			return 1
		}
		return elapsed / a
	case elapsed < a+d:
		// Decay phase: 1 → sustain
		if d <= 0 {
			return s
		}
		return 1 - (1-s)*((elapsed-a)/d)
	case elapsed < a+d+r:
		// Release phase: sustain → 0 (approximate—stays in release)
		if r <= 0 {
			return 0
		}
		return s * (1 - (elapsed-a-d)/r)
	default:
		return 0
	}
}

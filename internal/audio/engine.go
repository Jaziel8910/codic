package audio

import (
	"math"
)

const (
	SampleRate     = 44100
	NumChannels    = 2
	BytesPerSample = 4 // 2 bytes * 2 channels (int16)
)

// Voice is a single sounding note/sample with an oscillator + envelope.
// It is used by both the offline WAV renderer and (historically) the live
// engine; it holds no audio-device state, so it is safe to build and sample
// without any realtime context.
type Voice struct {
	osc            Oscillator
	env            *Envelope
	pan            float64 // 0=left, 1=right, 0.5=center
	gain           float64
	duration       float64   // seconds remaining
	elapsed        float64   // seconds played
	sample         []float64 // pre-rendered sample (nil if using oscillator)
	samplePos      int
	sampleFloatPos float64
	sampleRate     int
	speed          float64 // playback rate for samples (1 = normal)
	looping        bool
	fx             []Processor // effects chain (filters, distortion, reverb, ...)
}

// NewVoice creates a voice from oscillator type, frequency, and duration.
func NewVoice(oscType string, freq float64, dur float64, gain float64, pan float64) *Voice {
	osc := newOscillator(oscType, freq)
	env := NewEnvelope(0.01, 0.1, 0.7, 0.2) // default ADSR
	return &Voice{
		osc:      osc,
		env:      env,
		gain:     gain,
		pan:      pan,
		duration: dur,
		elapsed:  0,
	}
}

// nextSample advances the voice by one sample and returns its value.
func (v *Voice) nextSample() float64 {
	var s float64

	if v.sample != nil {
		// Play pre-rendered sample with playback-rate support.
		sp := v.speed
		if sp == 0 {
			sp = 1
		}
		if v.sampleFloatPos >= float64(len(v.sample)) {
			if v.looping {
				v.sampleFloatPos = 0
			}
		}
		if v.sampleFloatPos < float64(len(v.sample)) {
			idx := int(v.sampleFloatPos)
			s = v.sample[idx]
			v.sampleFloatPos += sp
		} else {
			s = 0
		}
	} else if v.osc != nil {
		s = v.osc.Sample()
	}

	// Apply envelope
	s *= v.env.Sample(v.elapsed)

	// Apply effects chain
	for _, p := range v.fx {
		s = p.Process(s)
	}

	// Apply gain
	s *= v.gain

	v.elapsed += 1.0 / float64(SampleRate)
	return s
}

func (v *Voice) panMix(s float64) (float64, float64) {
	// Equal-power panning
	pan := clamp(v.pan, 0, 1)
	left := s * math.Cos(pan*math.Pi/2)
	right := s * math.Sin(pan*math.Pi/2)
	return left, right
}

// --- Helpers ---

func clamp(v, lo, hi float64) float64 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func midiToFreq(midi float64) float64 {
	return 440.0 * math.Pow(2, (midi-69)/12)
}

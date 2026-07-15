package audio

import "math"

// Oscillator generates audio waveforms.
type Oscillator interface {
	Sample() float64
	SetFreq(freq float64)
}

type sineOsc struct {
	phase      float64
	freq       float64
	sampleRate float64
}

type sawOsc struct {
	phase      float64
	freq       float64
	sampleRate float64
}

type squareOsc struct {
	phase      float64
	freq       float64
	sampleRate float64
}

type triOsc struct {
	phase      float64
	freq       float64
	sampleRate float64
}

type noiseOsc struct{}

func newOscillator(oscType string, freq float64) Oscillator {
	switch oscType {
	case "sine", "seno":
		return &sineOsc{freq: freq, sampleRate: SampleRate}
	case "sawtooth", "saw", "sierra", "isaw":
		return &sawOsc{freq: freq, sampleRate: SampleRate}
	case "square", "cuadrada":
		return &squareOsc{freq: freq, sampleRate: SampleRate}
	case "triangle", "tri", "triangulo":
		return &triOsc{freq: freq, sampleRate: SampleRate}
	case "noise", "ruido", "rand":
		return &noiseOsc{}
	case "cosine", "coseno":
		// cosine = sine shifted by a quarter period
		return &sineOsc{freq: freq, phase: 0.25, sampleRate: SampleRate}
	default:
		return &sineOsc{freq: freq, sampleRate: SampleRate}
	}
}

func (o *sineOsc) Sample() float64 {
	s := math.Sin(2 * math.Pi * o.phase)
	o.phase += o.freq / o.sampleRate
	if o.phase >= 1 {
		o.phase -= 1
	}
	return s
}

func (o *sineOsc) SetFreq(f float64) { o.freq = f }

func (o *sawOsc) Sample() float64 {
	s := 2*o.phase - 1
	o.phase += o.freq / o.sampleRate
	if o.phase >= 1 {
		o.phase -= 1
	}
	return s
}

func (o *sawOsc) SetFreq(f float64) { o.freq = f }

func (o *squareOsc) Sample() float64 {
	var s float64
	if o.phase < 0.5 {
		s = 1
	} else {
		s = -1
	}
	o.phase += o.freq / o.sampleRate
	if o.phase >= 1 {
		o.phase -= 1
	}
	return s
}

func (o *squareOsc) SetFreq(f float64) { o.freq = f }

func (o *triOsc) Sample() float64 {
	var s float64
	if o.phase < 0.5 {
		s = 4*o.phase - 1
	} else {
		s = 3 - 4*o.phase
	}
	o.phase += o.freq / o.sampleRate
	if o.phase >= 1 {
		o.phase -= 1
	}
	return s
}

func (o *triOsc) SetFreq(f float64) { o.freq = f }

func (o *noiseOsc) Sample() float64      { return randFloat()*2 - 1 }
func (o *noiseOsc) SetFreq(freq float64) {}

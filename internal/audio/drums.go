package audio

import "math"

// SynthesizeDrum generates a drum sample buffer for the given drum type.
// Returns a mono float64 buffer at SampleRate.
func SynthesizeDrum(name string) []float64 {
	switch name {
	case "bd", "kick", "bt":
		return synthKick()
	case "sd", "snare", "sn":
		return synthSnare()
	case "hh", "hihat", "ch":
		return synthHihat()
	case "cp", "clap":
		return synthClap()
	case "oh":
		return synthOpenHat()
	default:
		return synthKick()
	}
}

func synthKick() []float64 {
	dur := 0.3
	n := int(dur * SampleRate)
	buf := make([]float64, n)
	freq := 150.0
	for i := 0; i < n; i++ {
		t := float64(i) / SampleRate
		// Pitch envelope: 150Hz → 50Hz
		f := freq * math.Pow(0.1, t*8)
		// Amplitude envelope
		env := math.Exp(-t * 15)
		buf[i] = math.Sin(2*math.Pi*f*t) * env * 0.8
	}
	return buf
}

func synthSnare() []float64 {
	dur := 0.2
	n := int(dur * SampleRate)
	buf := make([]float64, n)
	for i := 0; i < n; i++ {
		t := float64(i) / SampleRate
		// Tone component
		tone := math.Sin(2*math.Pi*180*t) * 0.3
		// Noise component
		noise := (randFloat()*2 - 1) * 0.5
		// Amplitude envelope
		env := math.Exp(-t * 20)
		buf[i] = (tone + noise) * env * 0.7
	}
	return buf
}

func synthHihat() []float64 {
	dur := 0.06
	n := int(dur * SampleRate)
	buf := make([]float64, n)
	for i := 0; i < n; i++ {
		t := float64(i) / SampleRate
		noise := (randFloat()*2 - 1)
		env := math.Exp(-t * 60)
		// Highpass-ish: subtract lowpassed version
		buf[i] = noise * env * 0.3
	}
	return buf
}

func synthOpenHat() []float64 {
	dur := 0.3
	n := int(dur * SampleRate)
	buf := make([]float64, n)
	for i := 0; i < n; i++ {
		t := float64(i) / SampleRate
		noise := (randFloat()*2 - 1)
		env := math.Exp(-t * 12)
		buf[i] = noise * env * 0.25
	}
	return buf
}

func synthClap() []float64 {
	dur := 0.15
	n := int(dur * SampleRate)
	buf := make([]float64, n)
	for i := 0; i < n; i++ {
		t := float64(i) / SampleRate
		noise := (randFloat()*2 - 1)
		// Burst envelope (three taps)
		var env float64
		for _, offset := range []float64{0, 0.01, 0.02, 0.03} {
			if t >= offset {
				env += math.Exp(-(t - offset) * 40)
			}
		}
		buf[i] = noise * env * 0.35
	}
	return buf
}

// DrumSampleCache caches pre-synthesized drum samples.
var drumCache = map[string][]float64{}

// GetDrumSample returns a cached drum sample, synthesizing if needed.
func GetDrumSample(name string) []float64 {
	if s, ok := drumCache[name]; ok {
		return s
	}
	s := SynthesizeDrum(name)
	drumCache[name] = s
	return s
}

// IsValidDrum returns true if the name is a known drum sample.
func IsValidDrum(name string) bool {
	switch name {
	case "bd", "kick", "bt", "sd", "snare", "sn", "hh", "hihat", "ch",
		"cp", "clap", "oh":
		return true
	}
	return false
}

// IsOscillatorType returns true if the name is a known oscillator type.
func IsOscillatorType(name string) bool {
	switch name {
	case "sine", "seno", "sawtooth", "saw", "sierra", "square", "cuadrada",
		"triangle", "tri", "triangulo", "noise", "ruido", "cosine", "coseno",
		"rand", "isaw":
		return true
	}
	return false
}

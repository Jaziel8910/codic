package audio

import "math"

// Processor is any single-sample effect in the voice's fx chain.
type Processor interface {
	Process(x float64) float64
}

// --- Biquad filter (lowpass / highpass / bandpass) ---

const (
	filterLowpass = iota
	filterHighpass
	filterBandpass
)

type Biquad struct {
	typ                    int
	freq                   float64
	q                      float64
	a0, a1, a2, b0, b1, b2 float64
	x1, x2, y1, y2         float64
}

func NewBiquad(typ int, freq, q float64) *Biquad {
	if q <= 0 {
		q = 0.7
	}
	b := &Biquad{typ: typ, freq: freq, q: q}
	b.recompute()
	return b
}

func (b *Biquad) recompute() {
	sr := float64(SampleRate)
	w0 := 2 * math.Pi * b.freq / sr
	cosw := math.Cos(w0)
	sinw := math.Sin(w0)
	alpha := sinw / (2 * b.q)

	var b0, b1, b2, a0, a1, a2 float64
	switch b.typ {
	case filterLowpass:
		b0 = (1 - cosw) / 2
		b1 = 1 - cosw
		b2 = (1 - cosw) / 2
		a0 = 1 + alpha
		a1 = -2 * cosw
		a2 = 1 - alpha
	case filterHighpass:
		b0 = (1 + cosw) / 2
		b1 = -(1 + cosw)
		b2 = (1 + cosw) / 2
		a0 = 1 + alpha
		a1 = -2 * cosw
		a2 = 1 - alpha
	case filterBandpass:
		b0 = alpha
		b1 = 0
		b2 = -alpha
		a0 = 1 + alpha
		a1 = -2 * cosw
		a2 = 1 - alpha
	}
	b.a0 = b0 / a0
	b.a1 = b1 / a0
	b.a2 = b2 / a0
	b.b1 = a1 / a0
	b.b2 = a2 / a0
}

func (b *Biquad) Process(x float64) float64 {
	y := b.a0*x + b.a1*b.x1 + b.a2*b.x2 - b.b1*b.y1 - b.b2*b.y2
	b.x2 = b.x1
	b.x1 = x
	b.y2 = b.y1
	b.y1 = y
	return y
}

// --- Distortion (waveshaper) ---

type DistortionProc struct {
	amount float64
}

func NewDistortion(amount float64) *DistortionProc { return &DistortionProc{amount: amount} }

func (d *DistortionProc) Process(x float64) float64 {
	if d.amount <= 0 {
		return x
	}
	return math.Tanh(x*d.amount) / math.Tanh(d.amount)
}

// --- Bitcrush (sample + bit reduction) ---

type Crush struct {
	bits     float64
	step     float64 // sample-rate reduction
	phase    float64
	held     float64
	srFactor float64
}

func NewCrush(bits, sampleReduce float64) *Crush {
	if bits < 1 {
		bits = 1
	}
	if sampleReduce < 1 {
		sampleReduce = 1
	}
	return &Crush{
		bits:     bits,
		step:     math.Pow(0.5, bits-1),
		srFactor: sampleReduce,
	}
}

func (c *Crush) Process(x float64) float64 {
	c.phase += c.srFactor
	if c.phase >= 1 {
		c.phase -= 1
		c.held = x
	}
	// bit-depth reduction
	reduced := math.Round(c.held/c.step) * c.step
	return clamp(reduced, -1, 1)
}

// --- Vowel / formant filter ---

var vowelFormants = map[string][2]float64{
	"a": {800, 1150},
	"e": {400, 2100},
	"i": {300, 2700},
	"o": {450, 800},
	"u": {325, 700},
}

type Formant struct {
	bands []*Biquad
	dry   float64
}

func NewFormant(vowel string, dry float64) *Formant {
	v, ok := vowelFormants[vowel]
	if !ok {
		v = vowelFormants["a"]
	}
	if dry < 0 || dry > 1 {
		dry = 0.4
	}
	return &Formant{
		bands: []*Biquad{
			NewBiquad(filterBandpass, v[0], 12),
			NewBiquad(filterBandpass, v[1], 12),
		},
		dry: dry,
	}
}

func (f *Formant) Process(x float64) float64 {
	var wet float64
	for _, b := range f.bands {
		wet += b.Process(x)
	}
	wet *= 0.5
	return x*f.dry + wet*(1-f.dry)
}

// --- Phaser (allpass stages + LFO) ---

type Phaser struct {
	stages   []*allpass1
	lfoRate  float64
	lfoPhase float64
	mix      float64
	base     float64
	depth    float64
}

type allpass1 struct {
	a  float64
	x1 float64
	y1 float64
}

func (a *allpass1) process(x float64) float64 {
	y := a.a*x + a.x1 - a.a*a.y1
	a.x1 = x
	a.y1 = y
	return y
}

func NewPhaser(rate, depth, mix float64) *Phaser {
	if depth <= 0 {
		depth = 0.5
	}
	stages := make([]*allpass1, 6)
	for i := range stages {
		stages[i] = &allpass1{a: 0.5}
	}
	return &Phaser{
		stages:  stages,
		lfoRate: rate,
		mix:     clamp(mix, 0, 1),
		base:    400,
		depth:   depth,
	}
}

func (p *Phaser) Process(x float64) float64 {
	// Advance LFO
	p.lfoPhase += p.lfoRate * 2 * math.Pi / float64(SampleRate)
	if p.lfoPhase > 2*math.Pi {
		p.lfoPhase -= 2 * math.Pi
	}
	lfo := 0.5 + 0.5*math.Sin(p.lfoPhase)
	center := p.base + lfo*p.depth*3000
	a := (1 - center/float64(SampleRate)) / (1 + center/float64(SampleRate))
	for _, s := range p.stages {
		s.a = a
	}
	wet := x
	for _, s := range p.stages {
		wet = s.process(wet)
	}
	return x*(1-p.mix) + wet*p.mix
}

// --- Chorus (multiple modulated delays) ---

type Chorus struct {
	voices   []chorusVoice
	mix      float64
	lfoPhase float64
	rate     float64
}

type chorusVoice struct {
	buf   []float64
	pos   int
	amp   float64
	rate  float64
	phase float64
}

func NewChorus(rate, depth, mix float64) *Chorus {
	if depth <= 0 {
		depth = 0.5
	}
	rates := []float64{0.7, 1.1, 1.4}
	amps := []float64{1, 0.8, 0.6}
	voices := make([]chorusVoice, 3)
	for i := range voices {
		voices[i] = chorusVoice{
			buf:   make([]float64, int(0.03*float64(SampleRate))),
			amp:   amps[i],
			rate:  rates[i],
			phase: float64(i) * 2 * math.Pi / 3,
		}
	}
	return &Chorus{voices: voices, mix: clamp(mix, 0, 1), rate: rate}
}

func (c *Chorus) Process(x float64) float64 {
	c.lfoPhase += c.rate * 2 * math.Pi / float64(SampleRate)
	if c.lfoPhase > 2*math.Pi {
		c.lfoPhase -= 2 * math.Pi
	}
	var wet float64
	for i := range c.voices {
		v := &c.voices[i]
		v.phase += c.rate * 2 * math.Pi / float64(SampleRate)
		if v.phase > 2*math.Pi {
			v.phase -= 2 * math.Pi
		}
		offset := 0.5 + 0.5*math.Sin(v.phase)
		read := v.pos - int(offset*0.02*float64(SampleRate))
		if read < 0 {
			read += len(v.buf)
		}
		delayed := v.buf[read%len(v.buf)]
		v.buf[v.pos] = x
		v.pos++
		if v.pos >= len(v.buf) {
			v.pos = 0
		}
		wet += delayed * v.amp
	}
	wet *= 0.5
	return x*(1-c.mix) + wet*c.mix
}

// --- Tremolo (amplitude LFO) ---

type Tremolo struct {
	rate  float64
	depth float64
	phase float64
}

func NewTremolo(rate, depth float64) *Tremolo {
	if depth <= 0 {
		depth = 0.5
	}
	if rate <= 0 {
		rate = 5
	}
	return &Tremolo{rate: rate, depth: depth}
}

func (t *Tremolo) Process(x float64) float64 {
	t.phase += t.rate * 2 * math.Pi / float64(SampleRate)
	if t.phase > 2*math.Pi {
		t.phase -= 2 * math.Pi
	}
	lfo := 1 - t.depth*0.5*(1+math.Sin(t.phase))
	return x * lfo
}

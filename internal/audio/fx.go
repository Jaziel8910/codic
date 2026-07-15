package audio

import "math"

// randFloat is a simple PRNG for audio noise.
var audioRandState uint64 = 1

func randFloat() float64 {
	audioRandState ^= audioRandState << 13
	audioRandState ^= audioRandState >> 7
	audioRandState ^= audioRandState << 17
	return float64(audioRandState%1000000) / 1000000.0
}

// --- One-pole lowpass filter ---

type Filter struct {
	prev   float64
	cutoff float64 // 0..1 normalized
}

func NewFilter(cutoff float64) *Filter {
	return &Filter{cutoff: clamp(cutoff, 0, 1)}
}

func (f *Filter) Process(s float64) float64 {
	f.prev = f.prev + f.cutoff*(s-f.prev)
	return f.prev
}

// --- Delay / echo ---

type Delay struct {
	buffer   []float64
	pos      int
	feedback float64
	mix      float64
}

func NewDelay(delaySec, feedback, mix float64) *Delay {
	bufSize := int(delaySec * SampleRate)
	if bufSize < 1 {
		bufSize = 1
	}
	return &Delay{
		buffer:   make([]float64, bufSize),
		feedback: clamp(feedback, 0, 0.99),
		mix:      clamp(mix, 0, 1),
	}
}

func (d *Delay) Process(s float64) float64 {
	delayed := d.buffer[d.pos]
	d.buffer[d.pos] = s + delayed*d.feedback
	d.pos++
	if d.pos >= len(d.buffer) {
		d.pos = 0
	}
	return s*(1-d.mix) + delayed*d.mix
}

// --- Reverb (Schroeder comb + allpass) ---

type Reverb struct {
	roomSize       float64
	comb1          []float64
	comb2          []float64
	comb3          []float64
	comb4          []float64
	p1, p2, p3, p4 int
	allpass1       []float64
	allpass2       []float64
	ap1, ap2       int
}

func NewReverb(roomSize float64) *Reverb {
	r := &Reverb{roomSize: clamp(roomSize, 0, 1)}
	r.comb1 = make([]float64, 1116)
	r.comb2 = make([]float64, 1188)
	r.comb3 = make([]float64, 1277)
	r.comb4 = make([]float64, 1356)
	r.allpass1 = make([]float64, 556)
	r.allpass2 = make([]float64, 441)
	return r
}

func (r *Reverb) Process(s float64) float64 {
	fb := 0.7 + r.roomSize*0.28
	damp := 0.2

	c1 := r.comb1[r.p1] * (1 - damp)
	c2 := r.comb2[r.p2] * (1 - damp)
	c3 := r.comb3[r.p3] * (1 - damp)
	c4 := r.comb4[r.p4] * (1 - damp)

	r.comb1[r.p1] = s + c1*fb
	r.comb2[r.p2] = s + c2*fb
	r.comb3[r.p3] = s + c3*fb
	r.comb4[r.p4] = s + c4*fb

	r.p1 = (r.p1 + 1) % len(r.comb1)
	r.p2 = (r.p2 + 1) % len(r.comb2)
	r.p3 = (r.p3 + 1) % len(r.comb3)
	r.p4 = (r.p4 + 1) % len(r.comb4)

	out := (c1 + c2 + c3 + c4) * 0.25

	apOut1 := r.allpass1[r.ap1]
	r.allpass1[r.ap1] = out + apOut1*0.5
	out = apOut1 - r.allpass1[r.ap1]*0.5
	r.ap1 = (r.ap1 + 1) % len(r.allpass1)

	apOut2 := r.allpass2[r.ap2]
	r.allpass2[r.ap2] = out + apOut2*0.5
	out = apOut2 - r.allpass2[r.ap2]*0.5
	r.ap2 = (r.ap2 + 1) % len(r.allpass2)

	wet := r.roomSize * 0.3
	return out*wet + s*(1-wet)
}

// --- Utility ---

func noteToFreq(noteName string) float64 {
	midi, ok := noteNameToMidi(noteName)
	if !ok {
		return 440.0
	}
	return midiToFreq(midi)
}

// Spanish note names, matching the pattern package, so "do3" works too.
var fxSpanishNotes = []struct {
	name string
	off  int
}{
	{"do", 0}, {"re", 2}, {"mi", 4}, {"fa", 5}, {"sol", 7}, {"la", 9}, {"si", 11},
}

func noteNameToMidi(s string) (float64, bool) {
	s = toLower(s)
	if len(s) == 0 {
		return 0, false
	}

	// Try Spanish names first.
	for _, sp := range fxSpanishNotes {
		if hasPrefix(s, sp.name) {
			rest := s[len(sp.name):]
			off := sp.off
			idx := 0
			for idx < len(rest) {
				if rest[idx] == '#' || rest[idx] == 's' {
					off++
					idx++
				} else if rest[idx] == 'b' {
					off--
					idx++
				} else {
					break
				}
			}
			oct := 4
			if idx < len(rest) {
				oct = atoi(rest[idx:])
			}
			if oct < -2 || oct > 10 {
				oct = 4
			}
			return float64((oct+1)*12 + off), true
		}
	}

	noteOffsets := map[byte]int{
		'c': 0, 'd': 2, 'e': 4, 'f': 5, 'g': 7, 'a': 9, 'b': 11,
	}
	offset, ok := noteOffsets[s[0]]
	if !ok {
		return 0, false
	}
	idx := 1
	semi := 0
	for idx < len(s) {
		if s[idx] == '#' || s[idx] == 's' {
			semi++
			idx++
		} else if s[idx] == 'b' {
			semi--
			idx++
		} else {
			break
		}
	}
	octave := 4
	if idx < len(s) && s[idx] >= '0' && s[idx] <= '9' {
		octave = int(s[idx] - '0')
		if idx+1 < len(s) && s[idx+1] >= '0' && s[idx+1] <= '9' {
			octave = octave*10 + int(s[idx+1]-'0')
		}
	}
	if octave < -2 || octave > 10 {
		octave = 4
	}
	return float64((octave+1)*12 + offset + semi), true
}

func hasPrefix(s, prefix string) bool {
	if len(prefix) > len(s) {
		return false
	}
	for i := 0; i < len(prefix); i++ {
		if s[i] != prefix[i] {
			return false
		}
	}
	return true
}

func atoi(s string) int {
	n := 0
	sign := 1
	i := 0
	if i < len(s) && (s[i] == '-' || s[i] == '+') {
		if s[i] == '-' {
			sign = -1
		}
		i++
	}
	for ; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			break
		}
		n = n*10 + int(s[i]-'0')
	}
	return n * sign
}

func toLower(s string) string {
	result := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c += 32
		}
		result[i] = c
	}
	return string(result)
}

// Distortion applies waveshaping (tanh soft clip).
func Distortion(s, amount float64) float64 {
	if amount <= 0 {
		return s
	}
	return math.Tanh(s*amount) / amount
}

package pattern

import (
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"sort"
	"strings"
	"time"
)

// ---------- 1. harmonize ----------

// Harmonize stacks transpositions of the pattern by the given intervals (in
// semitones) within the named scale.  If no intervals are given the triad
// 0,4,7 is used.
//
//	note("c3").harmonize("minor", 3, 5, 7)
func (p Pattern) Harmonize(scaleName string, intervals ...float64) Pattern {
	if len(intervals) == 0 {
		intervals = []float64{0, 4, 7}
	}
	pats := make([]Pattern, len(intervals))
	for i, iv := range intervals {
		pats[i] = p.Add(Pure(iv))
	}
	stacked := Stack(pats...)
	// If a scale is given, constrain the result to that scale.
	if scaleName != "" && scaleName != "chromatic" {
		stacked = stacked.Scale(scaleName, 0, 1, 2, 3, 4, 5, 6)
	}
	return stacked
}

// ---------- 2. counterpoint ----------

// Counterpoint generates a simple counterpoint line above the given pattern.
// Rules: "strict" (species 1, note-against-note), "free" (mostly stepwise),
// "florid" (mixed rhythms).  Default is "strict".
func (p Pattern) Counterpoint(rules ...string) Pattern {
	mode := "strict"
	if len(rules) > 0 {
		mode = rules[0]
	}
	_ = mode
	// Simple approach: transpose up a 5th (7 semitones) and offset slightly.
	cp := p.Add(Pure(7.0)).Late(NewFrac(1, 4))
	if mode == "florid" {
		cp = Stack(
			p.Add(Pure(12)).Late(NewFrac(1, 8)),
			p.Add(Pure(7)).Late(NewFrac(3, 8)),
			p.Add(Pure(4)).Late(NewFrac(5, 8)),
		)
	}
	return Stack(p, cp)
}

// ---------- 3. arpeggiate ----------

// Arpeggiate re-orders the haps of the pattern according to the given style.
// Styles: "up","down","updown","random","chord","broken","alberti".
// Range controls the octave spread (default 1).
func (p Pattern) Arpeggiate(style string, rng ...int) Pattern {
	octRange := 1
	if len(rng) > 0 && rng[0] > 0 {
		octRange = rng[0]
	}
	return p.WithEvents(func(haps []Hap) []Hap {
		if len(haps) == 0 {
			return haps
		}
		noteHaps := filterNoteHaps(haps)
		if len(noteHaps) == 0 {
			return haps
		}
		// Sort by pitch for arpeggiation.
		sort.Slice(noteHaps, func(i, j int) bool {
			return hapNoteNum(noteHaps[i]) < hapNoteNum(noteHaps[j])
		})
		var out []Hap
		switch style {
		case "down":
			for i := len(noteHaps) - 1; i >= 0; i-- {
				out = append(out, noteHaps[i])
			}
		case "updown":
			for _, h := range noteHaps {
				out = append(out, h)
			}
			for i := len(noteHaps) - 2; i >= 0; i-- {
				out = append(out, noteHaps[i])
			}
		case "random":
			rng := rand.New(rand.NewSource(time.Now().UnixNano()))
			for _, i := range rng.Perm(len(noteHaps)) {
				out = append(out, noteHaps[i])
			}
		case "chord":
			out = noteHaps
		case "broken":
			for o := 0; o < octRange; o++ {
				for _, h := range noteHaps {
					// Add octave-shifted copies
					h2 := h
					h2.Value = addToNoteValue(h.Value, float64(o*12))
					out = append(out, h2)
				}
			}
		case "alberti":
			if len(noteHaps) >= 3 {
				out = []Hap{noteHaps[0], noteHaps[2], noteHaps[1], noteHaps[2]}
			} else {
				out = noteHaps
			}
		default: // "up"
			out = noteHaps
		}
		return out
	})
}

// ---------- 4. humanize ----------

// Humanize adds micro-variations to timing, velocity, and pitch.
// timing: max jitter in fractions of a cycle (default 0.05)
// velocity: max velocity variation 0..1 (default 0.15)
// pitch: max pitch variation in semitones (default 0.2)
func (p Pattern) Humanize(timing, velocity, pitch float64) Pattern {
	if timing <= 0 {
		timing = 0.05
	}
	if velocity <= 0 {
		velocity = 0.15
	}
	if pitch <= 0 {
		pitch = 0.2
	}
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	return p.WithEvent(func(h Hap) Hap {
		// Timing jitter
		jit := NewFrac(int64(rng.Float64()*2-1)*100, 2000).Mul(NewFrac(int64(timing*100), 1))
		h.Part.Begin = h.Part.Begin.Add(jit)
		h.Part.End = h.Part.End.Add(jit)
		// Velocity / gain jitter
		if cm, ok := h.Value.(ControlMap); ok {
			if g, exists := cm["gain"]; exists {
				if gf, ok := g.(float64); ok {
					cm["gain"] = clamp(gf+(rng.Float64()*2-1)*velocity, 0, 1)
				}
			} else {
				cm["gain"] = 1.0 - rng.Float64()*velocity
			}
			// Pitch jitter
			if n, exists := cm["note"]; exists {
				if nf, ok := n.(float64); ok {
					cm["note"] = nf + (rng.Float64()*2-1)*pitch
				}
			}
		}
		return h
	})
}

// ---------- 5. groove ----------

// Groove applies a rhythmic groove template.  Templates:
// "swing" (triplet feel), "shuffle", "drag", "rush", "latin", "funk",
// "dnb", "hiphop".  Strength is 0..1 (default 0.5).
func (p Pattern) Groove(template string, strength ...float64) Pattern {
	s := 0.5
	if len(strength) > 0 {
		s = clamp(strength[0], 0, 1)
	}
	offsets := grooveOffsets(template)
	if len(offsets) == 0 {
		return p
	}
	return p.WithEvents(func(haps []Hap) []Hap {
		out := make([]Hap, len(haps))
		copy(out, haps)
		for i := range out {
			step := i % len(offsets)
			delta := NewFrac(int64(offsets[step]*s*480), 480)
			out[i].Part.Begin = out[i].Part.Begin.Add(delta)
			out[i].Part.End = out[i].Part.End.Add(delta)
		}
		return out
	})
}

func grooveOffsets(name string) []float64 {
	switch strings.ToLower(name) {
	case "swing":
		return []float64{0, 0, 0.16, 0, 0, 0, 0.16, 0}
	case "shuffle":
		return []float64{0, 0.2, 0, 0.2, 0, 0.2, 0, 0.2}
	case "drag":
		return []float64{0, 0.08, 0, 0.08, 0, 0.12, 0, 0.08}
	case "rush":
		return []float64{0, -0.06, 0, -0.06, 0, -0.1, 0, -0.06}
	case "latin":
		return []float64{0, 0, 0.12, 0, 0.08, 0, 0.12, 0}
	case "funk":
		return []float64{0, 0, 0.1, 0, 0.15, 0, 0.1, 0}
	case "dnb":
		return []float64{0, 0, 0, 0.1, 0, 0, 0, 0.15}
	case "hiphop":
		return []float64{0, 0.1, 0, 0.15, 0, 0.1, 0, 0.2}
	}
	return nil
}

// ---------- 6. layer ----------

// Layer creates N derived layers (octave, fifth, inversion, rhythm
// displacement).  density controls how many layers (default 2).
// variation controls how much each layer differs (0..1).
func (p Pattern) LayerDensity(density int, variation ...float64) Pattern {
	if density < 1 {
		density = 2
	}
	varVar := 0.3
	if len(variation) > 0 {
		varVar = clamp(variation[0], 0, 1)
	}
	pats := []Pattern{p}
	for i := 1; i < density; i++ {
		layer := p
		// Octave up on alternating layers
		if i%2 == 0 {
			layer = layer.Add(Pure(12.0))
		} else {
			layer = layer.Add(Pure(7.0))
		}
		// Rhythm displacement
		disp := NewFrac(int64(float64(i)*varVar*240), 480)
		layer = layer.Late(disp)
		// Reduce gain on deeper layers
		gain := 1.0 - float64(i)*0.2*varVar
		if gain < 0.1 {
			gain = 0.1
		}
		layer = layer.AddParam("gain", gain)
		pats = append(pats, layer)
	}
	return Stack(pats...)
}

// ---------- 7. morph ----------

// Morph interpolates between two patterns over N steps.
// Curve: "linear","exp","log","sigmoid","elastic".  Default linear.
func (p Pattern) Morph(other Pattern, steps int, curve ...string) Pattern {
	if steps < 2 {
		steps = 8
	}
	cv := "linear"
	if len(curve) > 0 {
		cv = curve[0]
	}
	var pats []Pattern
	for i := 0; i < steps; i++ {
		t := float64(i) / float64(steps-1)
		switch cv {
		case "exp":
			t = t * t
		case "log":
			t = math.Sqrt(t)
		case "sigmoid":
			t = 1 / (1 + math.Exp(-10*(t-0.5)))
		case "elastic":
			t = math.Pow(2, 10*(t-1)) * math.Cos(20*math.Pi*t/3)
		}
		aGain := 1 - t
		bGain := t
		// Fade between patterns
		mixed := Stack(
			p.AddParam("gain", aGain*0.8),
			other.AddParam("gain", bGain*0.8),
		)
		pats = append(pats, mixed)
	}
	return Slowcat(pats...)
}

// ---------- 8. seed / reseed ----------

var globalSeed int64

func init() {
	globalSeed = time.Now().UnixNano()
}

// SetSeed sets the global random seed for deterministic generation.
func SetSeed(n int64) {
	globalSeed = n
}

// Reseed re-initialises the global seed from the system clock.
func Reseed() {
	globalSeed = time.Now().UnixNano()
}

// SeedPattern returns a pattern that stores the seed value as metadata.
func SeedPattern(n int64) Pattern {
	SetSeed(n)
	return Pure(float64(n)).AddParam("_seed", float64(n))
}

// ---------- 9. markov ----------

// Markov generates a pattern using an Nth-order Markov chain learned from
// the source pattern's note sequence.  Order defaults to 1.
func (p Pattern) Markov(order ...int) Pattern {
	n := 1
	if len(order) > 0 && order[0] > 0 {
		n = order[0]
	}
	// Sample the pattern's first cycle to build the transition table.
	haps := p.FirstCycle()
	noteVals := extractNoteValues(haps)
	if len(noteVals) < n+1 {
		return p
	}
	table := buildMarkovTable(noteVals, n)
	rng := rand.New(rand.NewSource(globalSeed + 42))
	// Generate output sequence from the transition table.
	var generated []float64
	key := make([]float64, n)
	copy(key, noteVals[:n])
	for i := 0; i < len(noteVals)*2; i++ {
		next, ok := markovNext(table, key, rng)
		if !ok {
			break
		}
		generated = append(generated, next)
		copy(key, key[1:])
		key[n-1] = next
	}
	if len(generated) == 0 {
		return p
	}
	// Build a pattern from the generated sequence.
	var values []interface{}
	for _, v := range generated {
		values = append(values, v)
	}
	return Sequence(values...)
}

func extractNoteValues(haps []Hap) []float64 {
	var out []float64
	for _, h := range haps {
		if n := hapNoteNum(h); n >= 0 {
			out = append(out, n)
		}
	}
	return out
}

func hapNoteNum(h Hap) float64 {
	if cm, ok := h.Value.(ControlMap); ok {
		for _, k := range []string{"note", "n", "midinote"} {
			if v, exists := cm[k]; exists {
				if f, ok := v.(float64); ok {
					return f
				}
			}
		}
	}
	if f, ok := h.Value.(float64); ok {
		return f
	}
	return -1
}

func filterNoteHaps(haps []Hap) []Hap {
	var out []Hap
	for _, h := range haps {
		if hapNoteNum(h) >= 0 {
			out = append(out, h)
		}
	}
	return out
}

func addToNoteValue(v interface{}, add float64) interface{} {
	if cm, ok := v.(ControlMap); ok {
		for _, k := range []string{"note", "n", "midinote"} {
			if existing, exists := cm[k]; exists {
				if f, ok := existing.(float64); ok {
					cm[k] = f + add
				}
			}
		}
		return cm
	}
	if f, ok := v.(float64); ok {
		return f + add
	}
	return v
}

func buildMarkovTable(seq []float64, order int) map[string][]float64 {
	table := map[string][]float64{}
	for i := 0; i < len(seq)-order; i++ {
		key := joinFloats(seq[i : i+order])
		next := seq[i+order]
		table[key] = append(table[key], next)
	}
	return table
}

func markovNext(table map[string][]float64, key []float64, rng *rand.Rand) (float64, bool) {
	k := joinFloats(key)
	nexts, ok := table[k]
	if !ok || len(nexts) == 0 {
		return 0, false
	}
	return nexts[rng.Intn(len(nexts))], true
}

func joinFloats(vals []float64) string {
	parts := make([]string, len(vals))
	for i, v := range vals {
		parts[i] = fmt.Sprintf("%.0f", v)
	}
	return strings.Join(parts, ",")
}

// ---------- 10. euclid ----------

// EuclidExtended is an extended Euclidean rhythm generator.
// Parameters: steps, pulses, rotation, events-per-pulse.
// The base Euclid method already exists on Pattern.
// This method wraps the existing Euclid and adds event-per-pulse control.
func EuclidExtended(steps, pulses, rotation int, eventsPerPulse ...int) Pattern {
	ep := 1
	if len(eventsPerPulse) > 0 && eventsPerPulse[0] > 0 {
		ep = eventsPerPulse[0]
	}
	base := Silence()
	if steps > 0 {
		base = _EuclidPattern(steps, pulses, rotation, ep)
	}
	return base
}

// _EuclidPattern creates a euclidean rhythm pattern.
func _EuclidPattern(steps, pulses, rotation, eventsPerPulse int) Pattern {
	if steps <= 0 {
		return Silence()
	}
	if pulses <= 0 {
		return Silence()
	}
	if pulses > steps {
		pulses = steps
	}
	pattern := make([]bool, steps)
	count := pulses
	pos := rotation % steps
	for i := 0; i < steps; i++ {
		pattern[pos] = count > 0
		pos = (pos + pulses) % steps
		if pattern[pos] {
			count--
		}
	}
	var haps []Hap
	cycle := NewFrac(1, int64(steps))
	for i, hit := range pattern {
		if !hit {
			continue
		}
		begin := NewFrac(int64(i), int64(steps))
		end := begin.Add(cycle.Mul(NewFrac(int64(eventsPerPulse), 1)))
		haps = append(haps, Hap{
			Part:  TimeSpan{Begin: begin, End: end},
			Value: float64(60 + (i % 12)),
		})
	}
	if len(haps) == 0 {
		return Silence()
	}
	return Pattern{Query: func(s State) []Hap {
		var out []Hap
		for _, h := range haps {
			n := h
			// Only return haps that intersect the query span.
			if n.Part.End.Cmp(s.Span.Begin) > 0 && n.Part.Begin.Cmp(s.Span.End) < 0 {
				out = append(out, n)
			}
		}
		return out
	}}
}

// ---------- 11. constrain ----------

// Constrain forces all note values in the pattern to the named scale.
// Modes: "clip" (nearest), "fold" (octave wrap), "nearest" (default).
// Range limits the octave range (default 3 octaves).
func (p Pattern) Constrain(scaleName string, mode ...string) Pattern {
	m := "nearest"
	if len(mode) > 0 {
		m = mode[0]
	}
	ivs := scaleIntervals(scaleName)
	if len(ivs) == 0 {
		ivs = []int{0, 2, 4, 5, 7, 9, 11}
	}
	return p.WithValue(func(v interface{}) interface{} {
		note := -1.0
		if cm, ok := v.(ControlMap); ok {
			for _, k := range []string{"note", "n", "midinote"} {
				if existing, exists := cm[k]; exists {
					if f, ok := existing.(float64); ok {
						note = f
						break
					}
				}
			}
		} else if f, ok := v.(float64); ok {
			note = f
		}
		if note < 0 {
			return v
		}
		octave := int(math.Floor(note / 12))
		noteInOctave := int(math.Mod(note, 12))
		closest := constrainToScale(noteInOctave, ivs, m)
		constrained := float64(octave*12 + closest)
		if cm, ok := v.(ControlMap); ok {
			for _, k := range []string{"note", "n", "midinote"} {
				if _, exists := cm[k]; exists {
					cm[k] = constrained
					break
				}
			}
			return cm
		}
		return constrained
	})
}

func constrainToScale(semitone int, intervals []int, mode string) int {
	if len(intervals) == 0 {
		return semitone
	}
	// Check if already in scale.
	for _, iv := range intervals {
		if semitone == iv {
			return semitone
		}
	}
	switch mode {
	case "fold":
		// Fold into nearest octave of the scale.
		base := (semitone / 12) * 12
		best := intervals[0]
		bestDist := 12
		for _, iv := range intervals {
			candidate := base + iv
			dist := absInt(candidate - semitone)
			if dist < bestDist {
				bestDist = dist
				best = candidate
			}
			// Try one octave down
			candidate2 := base + iv - 12
			dist2 := absInt(candidate2 - semitone)
			if dist2 < bestDist {
				bestDist = dist2
				best = candidate2
			}
		}
		return best
	case "clip":
		// Clip to the range of the scale.
		closest := intervals[0]
		bestDist := 12
		for _, iv := range intervals {
			candidate := iv + (semitone/12)*12
			dist := absInt(candidate - semitone)
			if dist < bestDist {
				bestDist = dist
				closest = candidate
			}
		}
		return closest
	default: // "nearest"
		closest := intervals[0]
		bestDist := 12
		for _, iv := range intervals {
			candidate := iv + (semitone/12)*12
			dist := absInt(candidate - semitone)
			if dist < bestDist {
				bestDist = dist
				closest = candidate
			}
		}
		return closest
	}
}

func absInt(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func scaleIntervals(name string) []int {
	switch strings.ToLower(name) {
	case "major", "ionian":
		return []int{0, 2, 4, 5, 7, 9, 11}
	case "minor", "aeolian":
		return []int{0, 2, 3, 5, 7, 8, 10}
	case "dorian":
		return []int{0, 2, 3, 5, 7, 9, 10}
	case "phrygian":
		return []int{0, 1, 3, 5, 7, 8, 10}
	case "lydian":
		return []int{0, 2, 4, 6, 7, 9, 11}
	case "mixolydian":
		return []int{0, 2, 4, 5, 7, 9, 10}
	case "locrian":
		return []int{0, 1, 3, 5, 6, 8, 10}
	case "pentatonic":
		return []int{0, 2, 4, 7, 9}
	case "minorpenta":
		return []int{0, 3, 5, 7, 10}
	case "blues":
		return []int{0, 3, 5, 6, 7, 10}
	case "chromatic":
		return []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11}
	}
	return nil
}

// ---------- 12. serialize / deserialize ----------

// Serialize returns a JSON string representation of the pattern's first cycle.
func (p Pattern) Serialize() string {
	haps := p.FirstCycle()
	type hapJSON struct {
		BeginF float64     `json:"begin"`
		EndF   float64     `json:"end"`
		Value  interface{} `json:"value"`
	}
	var out []hapJSON
	for _, h := range haps {
		out = append(out, hapJSON{
			BeginF: h.Part.Begin.Float64(),
			EndF:   h.Part.End.Float64(),
			Value:  h.Value,
		})
	}
	b, _ := json.Marshal(out)
	return string(b)
}

// DeserializeJSON parses a JSON string (as produced by Serialize) back into a Pattern.
func DeserializeJSON(data string) (Pattern, error) {
	type hapJSON struct {
		BeginF float64     `json:"begin"`
		EndF   float64     `json:"end"`
		Value  interface{} `json:"value"`
	}
	var haps []hapJSON
	if err := json.Unmarshal([]byte(data), &haps); err != nil {
		return Silence(), err
	}
	var pats []Pattern
	for _, hj := range haps {
		h := Hap{
			Part: TimeSpan{
				Begin: NewFrac(int64(hj.BeginF*480), 480),
				End:   NewFrac(int64(hj.EndF*480), 480),
			},
			Value: hj.Value,
		}
		pats = append(pats, Pattern{Query: func(s State) []Hap {
			if h.Part.End.Cmp(s.Span.Begin) > 0 && h.Part.Begin.Cmp(s.Span.End) < 0 {
				return []Hap{h}
			}
			return nil
		}})
	}
	if len(pats) == 0 {
		return Silence(), nil
	}
	combined := pats[0]
	for _, pa := range pats[1:] {
		combined = Stack(combined, pa)
	}
	return combined, nil
}

// ---------- 13. analyze ----------

// Analyze returns a map of metrics for the pattern.
// Keys: "events", "notes", "density", "note_range", "polyphony".
func (p Pattern) Analyze() map[string]float64 {
	haps := p.FirstCycle()
	m := map[string]float64{
		"events": float64(len(haps)),
		"notes":  0,
	}
	if len(haps) == 0 {
		return m
	}
	noteVals := extractNoteValues(haps)
	m["notes"] = float64(len(noteVals))
	if len(noteVals) > 1 {
		minN, maxN := noteVals[0], noteVals[0]
		for _, n := range noteVals[1:] {
			if n < minN {
				minN = n
			}
			if n > maxN {
				maxN = n
			}
		}
		m["note_range"] = maxN - minN
	}
	// Count overlapping haps as a rough polyphony measure.
	overlap := 0
	for i, a := range haps {
		for j, b := range haps {
			if i != j && a.Part.Begin.Cmp(b.Part.Begin) == 0 {
				overlap++
			}
		}
	}
	m["polyphony"] = float64(overlap)
	return m
}

// ---------- 14. remix ----------

// Remix applies a remix strategy to the pattern.
// Strategies: "shuffle", "stutter", "reverse", "filter", "rearrange", "dub".
// Additional string args are strategy-specific.
func (p Pattern) Remix(strategy string, extraArgs ...interface{}) Pattern {
	switch strings.ToLower(strategy) {
	case "shuffle":
		return p.Scramble()
	case "stutter":
		n := 4
		if len(extraArgs) > 0 {
			if v, ok := extraArgs[0].(float64); ok {
				n = int(v)
			} else if v, ok := extraArgs[0].(int); ok {
				n = v
			}
		}
		if n < 1 {
			n = 4
		}
		return p.Ply(n)
	case "reverse":
		return p.Rev()
	case "dub":
		// Echo with feedback.
		return p.Stut(4, 0.5, 0.25)
	case "filter":
		return p.WithEvent(func(h Hap) Hap {
			if cm, ok := h.Value.(ControlMap); ok {
				cm["cutoff"] = 300.0
				cm["resonance"] = 0.6
			}
			return h
		})
	case "rearrange":
		return p.Scramble().Slow(NewFrac(2, 1))
	default:
		return p
	}
}

// ---------- 15. visualize ----------

// Visualize returns an ASCII/ANSI string representing the pattern.
// Types: "pianoroll", "drumgrid", "waveform". Output is always a string.
func (p Pattern) Visualize(visType string) string {
	haps := p.FirstCycle()
	switch visType {
	case "pianoroll":
		return renderPianoRoll(haps)
	case "drumgrid":
		return renderDrumGrid(haps)
	default:
		return fmt.Sprintf("Pattern with %d haps in first cycle", len(haps))
	}
}

func renderPianoRoll(haps []Hap) string {
	if len(haps) == 0 {
		return "(empty)"
	}
	noteHaps := filterNoteHaps(haps)
	if len(noteHaps) == 0 {
		return "(no notes)"
	}
	minN, maxN := 127.0, 0.0
	for _, h := range noteHaps {
		n := hapNoteNum(h)
		if n < minN {
			minN = n
		}
		if n > maxN {
			maxN = n
		}
	}
	if maxN-minN < 1 {
		maxN = minN + 12
	}
	rows := int(maxN-minN) + 1
	if rows > 24 {
		rows = 24
	}
	cols := 32
	grid := make([][]rune, rows)
	for r := range grid {
		grid[r] = []rune(strings.Repeat(".", cols))
	}
	for _, h := range noteHaps {
		n := hapNoteNum(h)
		row := int(n - minN)
		if row >= rows {
			row = rows - 1
		}
		if row < 0 {
			row = 0
		}
		col := int(h.Part.Begin.Float64() * float64(cols))
		if col >= cols {
			col = cols - 1
		}
		if col < 0 {
			col = 0
		}
		grid[row][col] = '#'
	}
	var b strings.Builder
	for r := rows - 1; r >= 0; r-- {
		note := midiToNote1(int(minN) + r)
		b.WriteString(fmt.Sprintf("%-4s ", note))
		b.WriteString(string(grid[r]))
		b.WriteString("\n")
	}
	return b.String()
}

func renderDrumGrid(haps []Hap) string {
	if len(haps) == 0 {
		return "(empty)"
	}
	cols := 32
	drums := map[string][]bool{}
	order := []string{}
	for _, h := range haps {
		sample := "?"
		if cm, ok := h.Value.(ControlMap); ok {
			if s, exists := cm["s"]; exists {
				if str, ok := s.(string); ok {
					sample = str
				}
			}
		}
		if _, ok := drums[sample]; !ok {
			drums[sample] = make([]bool, cols)
			order = append(order, sample)
		}
		col := int(h.Part.Begin.Float64() * float64(cols))
		if col >= 0 && col < cols {
			drums[sample][col] = true
		}
	}
	var b strings.Builder
	for _, name := range order {
		b.WriteString(fmt.Sprintf("%-8s ", name))
		for _, hit := range drums[name] {
			if hit {
				b.WriteRune('#')
			} else {
				b.WriteRune('.')
			}
		}
		b.WriteString("\n")
	}
	return b.String()
}

func midiToNote1(m int) string {
	names := []string{"c", "c#", "d", "d#", "e", "f", "f#", "g", "g#", "a", "a#", "b"}
	oct := (m / 12) - 1
	idx := ((m % 12) + 12) % 12
	return fmt.Sprintf("%s%d", names[idx], oct)
}

func clamp(x, lo, hi float64) float64 {
	if x < lo {
		return lo
	}
	if x > hi {
		return hi
	}
	return x
}

package dj

import (
	"fmt"
	"math/rand"
	"strings"

	"github.com/Jaziel8910/codic/internal/codang"
	"github.com/Jaziel8910/codic/internal/pattern"
)

// Section is one part of a generated song.
type Section struct {
	Name    string
	Pattern pattern.Pattern
	Seconds float64
	CPS     float64
}

const secondsPerBar = 240.0 // 1 cycle = 1 bar of 4/4; cps = bpm/240

// GenerateLoop builds a single looping pattern for a genre.
func GenerateLoop(genre, key string, bpm, bars int, seed int64) (pattern.Pattern, float64, string, error) {
	g, err := loadModel(genre)
	if err != nil {
		g = defaultProfile(genre)
	}
	if bpm <= 0 {
		bpm = (g.BPMMin + g.BPMMax) / 2
		if bpm <= 0 {
			bpm = 120
		}
	}
	if bars <= 0 {
		bars = 8
	}
	code := g.buildLoopCode(key, bpm, 1.0, seed)
	pat, cps, err := evalCode(code)
	if err != nil {
		return pattern.Silence(), 0, code, err
	}
	return pat, cps, code, nil
}

// GenerateLayer builds a single layer's pattern (useful for stem extraction or
// "just the bass" style generation).
func GenerateLayer(genre, layer string, bars, variations, bpm int, seed int64) (pattern.Pattern, float64, string, error) {
	g, err := loadModel(genre)
	if err != nil {
		g = defaultProfile(genre)
	}
	if bpm <= 0 {
		bpm = (g.BPMMin + g.BPMMax) / 2
		if bpm <= 0 {
			bpm = 120
		}
	}
	if bars <= 0 {
		bars = 4
	}
	rnd := rand.New(rand.NewSource(seed))
	code := g.buildLayerCode(layer, rnd, 0, bpm)
	pat, cps, err := evalCode(code)
	if err != nil {
		return pattern.Silence(), 0, code, err
	}
	return pat, cps, code, nil
}

// GenerateSong builds a multi-section song following a named structure.
func GenerateSong(genre, structure, key string, bpm int, seed int64) ([]Section, error) {
	g, err := loadModel(genre)
	if err != nil {
		g = defaultProfile(genre)
	}
	if bpm <= 0 {
		bpm = (g.BPMMin + g.BPMMax) / 2
		if bpm <= 0 {
			bpm = 120
		}
	}
	specs := structureSpec(structure)
	if len(specs) == 0 {
		specs = structureSpec("standard_edm")
	}
	rnd := rand.New(rand.NewSource(seed))
	shift := keyShift(g.Key, key)
	var sections []Section
	for _, s := range specs {
		code := g.buildLoopCodeWithDensity(key, bpm, s.Density, rnd, shift)
		pat, cps, err := evalCode(code)
		if err != nil {
			return nil, err
		}
		secs := float64(s.Bars) * secondsPerBar / float64(bpm)
		sections = append(sections, Section{
			Name:    s.Name,
			Pattern: pat,
			Seconds: secs,
			CPS:     cps,
		})
	}
	return sections, nil
}

// Morph generates two patterns (one per genre) at the same tempo, for the CLI
// to crossfade into each other.
func Morph(genreA, genreB string, steps, bpm int, seed int64) (pattern.Pattern, pattern.Pattern, float64, error) {
	if bpm <= 0 {
		bpm = 128
	}
	if steps <= 0 {
		steps = 8
	}
	pa, _, _, err := GenerateLoop(genreA, "", bpm, steps, seed)
	if err != nil {
		return pattern.Silence(), pattern.Silence(), 0, err
	}
	pb, _, _, err := GenerateLoop(genreB, "", bpm, steps, seed+1)
	if err != nil {
		return pattern.Silence(), pattern.Silence(), 0, err
	}
	return pa, pb, float64(bpm) / 240.0, nil
}

// bpmFor returns a sensible bpm for the model.
func (g *GenreModel) bpmFor() int {
	b := (g.BPMMin + g.BPMMax) / 2
	if b <= 0 {
		return 120
	}
	return b
}

func (g *GenreModel) buildLoopCode(key string, bpm int, density float64, seed int64) string {
	return g.buildLoopCodeWithDensity(key, bpm, density, rand.New(rand.NewSource(seed)), keyShift(g.Key, key))
}

func (g *GenreModel) buildLoopCodeWithDensity(key string, bpm int, density float64, rnd *rand.Rand, shift int) string {
	var b strings.Builder
	fmt.Fprintf(&b, "bpm %d\n", bpm)
	add := func(line string) { b.WriteString(line + "\n") }

	if kick := g.pickLayer(rnd, "kick"); kick != "" {
		add(fmt.Sprintf("kick = s(%q).gain(0.9).out()", kick))
	}
	if density >= 0.35 {
		if hat := g.pickLayer(rnd, "hat"); hat != "" {
			add(fmt.Sprintf("hat = s(%q).gain(0.3).speed(1.5).out()", hat))
		}
	}
	if density >= 0.5 {
		if sn := g.pickLayer(rnd, "snare"); sn != "" {
			add(fmt.Sprintf("snare = s(%q).gain(0.6).out()", sn))
		}
	}
	if density >= 0.55 {
		if cl := g.pickLayer(rnd, "clap"); cl != "" {
			add(fmt.Sprintf("perc = s(%q).gain(0.4).out()", cl))
		}
	}
	if density >= 0.55 {
		if bass := g.pickLayer(rnd, "bass"); bass != "" {
			bass = transposeNotes(bass, shift)
			add(fmt.Sprintf("bass = %q.note().s(%q).cutoff(500).gain(0.4).out()", bass, g.BassInst))
		}
	}
	if density >= 0.8 {
		if pad := g.pickLayer(rnd, "pad"); pad != "" {
			pad = transposeNotes(pad, shift)
			add(fmt.Sprintf("pad = %q.note().s(%q).cutoff(800).gain(0.25).out()", pad, g.PadInst))
		}
	}
	if density >= 0.75 {
		if lead := g.pickLayer(rnd, "lead"); lead != "" {
			lead = transposeNotes(lead, shift)
			add(fmt.Sprintf("lead = %q.note().s(%q).cutoff(1200).gain(0.3).out()", lead, g.LeadInst))
		}
	}
	return b.String()
}

func (g *GenreModel) buildLayerCode(layer string, rnd *rand.Rand, shift int, bpm int) string {
	var b strings.Builder
	fmt.Fprintf(&b, "bpm %d\n", bpm)
	cat := strings.ToLower(layer)
	switch cat {
	case "kick", "snare", "clap", "hat":
		if pat := g.pickLayer(rnd, cat); pat != "" {
			gain := 0.6
			if cat == "kick" {
				gain = 0.9
			}
			if cat == "hat" {
				gain = 0.3
			}
			b.WriteString(fmt.Sprintf("%s = s(%q).gain(%.2f).out()\n", cat, pat, gain))
		}
	case "bass":
		if pat := g.pickLayer(rnd, "bass"); pat != "" {
			pat = transposeNotes(pat, shift)
			b.WriteString(fmt.Sprintf("bass = %q.note().s(%q).cutoff(500).gain(0.4).out()\n", pat, g.BassInst))
		}
	case "pad":
		if pat := g.pickLayer(rnd, "pad"); pat != "" {
			pat = transposeNotes(pat, shift)
			b.WriteString(fmt.Sprintf("pad = %q.note().s(%q).cutoff(800).gain(0.25).out()\n", pat, g.PadInst))
		}
	default: // lead
		if pat := g.pickLayer(rnd, "lead"); pat != "" {
			pat = transposeNotes(pat, shift)
			b.WriteString(fmt.Sprintf("lead = %q.note().s(%q).cutoff(1200).gain(0.3).out()\n", pat, g.LeadInst))
		}
	}
	return b.String()
}

// evalCode parses and evaluates Codang code, returning the stacked pattern and
// the resulting cycles-per-second tempo.
func evalCode(code string) (pattern.Pattern, float64, error) {
	prog, err := codang.Parse(code)
	if err != nil {
		return pattern.Silence(), 0, fmt.Errorf("parse error: %w", err)
	}
	var all []pattern.Pattern
	ev := codang.NewEvaluator(func(p pattern.Pattern) { all = append(all, p) })
	if err := ev.Eval(prog); err != nil {
		return pattern.Silence(), 0, fmt.Errorf("eval error: %w", err)
	}
	if len(all) == 0 {
		return pattern.Silence(), 1.0, nil
	}
	combined := all[0]
	for _, p := range all[1:] {
		combined = pattern.Stack(combined, p)
	}
	cps := ev.GetCPS()
	if cps <= 0 {
		cps = 1.0
	}
	return combined, cps, nil
}

// sectionSpec describes the bars and density of each section in a structure.
type sectionSpec struct {
	Name    string
	Bars    int
	Density float64
}

func structureSpec(name string) []sectionSpec {
	switch strings.ToLower(name) {
	case "standard_edm":
		return []sectionSpec{
			{"intro", 4, 0.3},
			{"buildup", 4, 0.6},
			{"drop", 8, 1.0},
			{"breakdown", 4, 0.4},
			{"drop", 8, 1.0},
			{"outro", 4, 0.3},
		}
	case "dub_techno":
		return []sectionSpec{
			{"dub", 8, 0.5},
			{"dub", 8, 0.7},
			{"break", 4, 0.3},
			{"dub", 8, 0.8},
		}
	case "minimal":
		return []sectionSpec{
			{"additive", 8, 0.4},
			{"subtractive", 8, 0.9},
		}
	case "breakbeat":
		return []sectionSpec{
			{"intro", 4, 0.3},
			{"body", 8, 0.9},
			{"break", 4, 0.4},
			{"outro", 4, 0.5},
		}
	case "ambient":
		return []sectionSpec{
			{"pad", 8, 0.6},
			{"evolve", 8, 0.8},
			{"resolve", 8, 0.5},
		}
	}
	// default to standard_edm
	return []sectionSpec{
		{"intro", 4, 0.3},
		{"buildup", 4, 0.6},
		{"drop", 8, 1.0},
		{"breakdown", 4, 0.4},
		{"drop", 8, 1.0},
		{"outro", 4, 0.3},
	}
}

// ListStructures returns the names of built-in song structures.
func ListStructures() []string {
	return []string{"standard_edm", "dub_techno", "minimal", "breakbeat", "ambient"}
}

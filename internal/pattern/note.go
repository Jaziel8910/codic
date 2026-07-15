package pattern

import (
	"math"
	"strings"
)

// Note names to semitone offsets from C. English single letters.
var noteOffsets = map[string]int{
	"c": 0, "d": 2, "e": 4, "f": 5, "g": 7, "a": 9, "b": 11,
}

// Spanish note names (do, re, mi, fa, sol, la, si) so Spanish-speaking users
// can write notes in their own language.
var spanishNotes = []struct {
	name string
	off  int
}{
	{"do", 0}, {"re", 2}, {"mi", 4}, {"fa", 5}, {"sol", 7}, {"la", 9}, {"si", 11},
}

var accidentals = map[string]int{
	"#": 1, "s": 1, "♯": 1,
	"b": -1, "♭": -1, "f": -1,
}

// noteToMidi converts a note name like "c4", "bb2", "fs#3", "do3", "sol#5" to a
// MIDI number. Accepts both English (c,d,e,f,g,a,b) and Spanish
// (do,re,mi,fa,sol,la,si) names. Returns false if not a note name.
func noteToMidi(s string) (float64, bool) {
	s = strings.ToLower(strings.TrimSpace(s))
	if len(s) == 0 {
		return 0, false
	}

	idx := 0
	offset, idx, ok := parseNoteName(s, idx)
	if !ok {
		return 0, false
	}
	semi := 0

	// Parse accidentals (can be multiple: "bs#", "bb", etc.)
	for idx < len(s) {
		acc, ok := accidentals[string(s[idx])]
		if !ok {
			break
		}
		semi += acc
		idx++
	}

	// Parse octave (optional, default 4)
	octave := 4
	if idx < len(s) {
		oct := ""
		if s[idx] == '-' || s[idx] == '+' {
			oct += string(s[idx])
			idx++
		}
		for idx < len(s) && s[idx] >= '0' && s[idx] <= '9' {
			oct += string(s[idx])
			idx++
		}
		if oct != "" {
			o := parseIntStr(oct)
			if o >= -2 && o <= 10 {
				octave = o
			}
		}
	}

	midi := (octave+1)*12 + offset + semi
	return float64(midi), true
}

// parseNoteName reads a note name (Spanish first, then English) at position i
// and returns its semitone offset and the new index.
func parseNoteName(s string, i int) (int, int, bool) {
	for _, sp := range spanishNotes {
		if strings.HasPrefix(s[i:], sp.name) {
			return sp.off, i + len(sp.name), true
		}
	}
	if i < len(s) {
		if off, ok := noteOffsets[string(s[i])]; ok {
			return off, i + 1, true
		}
	}
	return 0, i, false
}

// MidiToFreq converts a MIDI note number to frequency in Hz.
func MidiToFreq(midi float64) float64 {
	return 440.0 * math.Pow(2, (midi-69)/12)
}

// --- generation helpers ---

var randState uint64 = 1

func randFloat() float64 {
	// Simple xorshift PRNG - good enough for music patterns
	randState ^= randState << 13
	randState ^= randState >> 7
	randState ^= randState << 17
	return float64(randState%1000000) / 1000000.0
}

func parseIntStr(s string) int {
	neg := false
	idx := 0
	if idx < len(s) && (s[idx] == '-' || s[idx] == '+') {
		if s[idx] == '-' {
			neg = true
		}
		idx++
	}
	n := 0
	for idx < len(s) && s[idx] >= '0' && s[idx] <= '9' {
		n = n*10 + int(s[idx]-'0')
		idx++
	}
	if neg {
		n = -n
	}
	return n
}

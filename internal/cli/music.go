package cli

import (
	"fmt"
	"strings"
)

// noteToMidi parses note names like "c3", "fs4", "a#2", "do3", "sol#5" into a
// MIDI note number. Returns (note, true) on success.
func noteToMidi(s string) (int, bool) {
	s = strings.TrimSpace(strings.ToLower(s))
	if s == "" {
		return 0, false
	}
	spanish := []struct {
		name string
		semi int
	}{
		{"do", 0}, {"re", 2}, {"mi", 4}, {"fa", 5}, {"sol", 7}, {"la", 9}, {"si", 11},
	}
	for _, sp := range spanish {
		if strings.HasPrefix(s, sp.name) {
			rest := s[len(sp.name):]
			semi := sp.semi
			oct := 4
			if len(rest) > 0 && (rest[0] == '#' || rest[0] == 's') {
				semi++
				rest = rest[1:]
			} else if len(rest) > 0 && rest[0] == 'b' {
				semi--
				rest = rest[1:]
			}
			if rest != "" {
				if o, err := parseIntSafe2(rest); err == nil {
					oct = o
				}
			}
			return semi + (oct+1)*12, true
		}
	}
	names := map[rune]int{'c': 0, 'd': 2, 'e': 4, 'f': 5, 'g': 7, 'a': 9, 'b': 11}
	if len(s) == 0 {
		return 0, false
	}
	base := rune(s[0])
	semi, ok := names[base]
	if !ok {
		return 0, false
	}
	i := 1
	if i < len(s) && (s[i] == 's' || s[i] == '#') {
		i++
		semi++
	} else if i < len(s) && s[i] == 'b' {
		i++
		semi--
	}
	octStr := s[i:]
	oct := 4
	if octStr != "" {
		if o, err := parseIntSafe2(octStr); err == nil {
			oct = o
		} else {
			return 0, false
		}
	}
	return semi + (oct+1)*12, true
}

// midiToNote renders a MIDI note number as an English note name (e.g. 60 -> c4).
func midiToNote(m int) string {
	names := []string{"c", "c#", "d", "d#", "e", "f", "f#", "g", "g#", "a", "a#", "b"}
	oct := (m / 12) - 1
	idx := ((m % 12) + 12) % 12
	return fmt.Sprintf("%s%d", names[idx], oct)
}

// scaleIntervals returns the semitone offsets for a named scale.
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
	case "harmonicminor":
		return []int{0, 2, 3, 5, 7, 8, 11}
	case "melodicminor":
		return []int{0, 2, 3, 5, 7, 9, 11}
	case "penta", "pentatonic":
		return []int{0, 2, 4, 7, 9}
	case "minorpenta":
		return []int{0, 3, 5, 7, 10}
	case "blues":
		return []int{0, 3, 5, 6, 7, 10}
	case "chromatic":
		return []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11}
	}
	return []int{0, 2, 4, 5, 7, 9, 11} // default major
}

// scaleNotes returns the note names of a scale rooted at `root`.
func scaleNotes(root, scaleName string) ([]string, error) {
	base, ok := noteToMidi(root)
	if !ok {
		return nil, fmt.Errorf("invalid root note: %q", root)
	}
	ivs := scaleIntervals(scaleName)
	out := make([]string, 0, len(ivs))
	for _, iv := range ivs {
		out = append(out, midiToNote(base+iv))
	}
	return out, nil
}

// chordNotes returns the note names of a chord rooted at `root`.
func chordNotes(root, kind string) ([]string, error) {
	base, ok := noteToMidi(root)
	if !ok {
		return nil, fmt.Errorf("invalid root note: %q", root)
	}
	ivs := []int{0, 4, 7} // default major triad
	switch strings.ToLower(strings.TrimSuffix(kind, "7")) {
	case "maj", "major", "":
		ivs = []int{0, 4, 7}
	case "min", "minor":
		ivs = []int{0, 3, 7}
	case "dim":
		ivs = []int{0, 3, 6}
	case "aug":
		ivs = []int{0, 4, 8}
	case "sus2":
		ivs = []int{0, 2, 7}
	case "sus4":
		ivs = []int{0, 5, 7}
	}
	if strings.HasSuffix(strings.ToLower(kind), "7") {
		ivs = append(ivs, 10)
	}
	out := make([]string, 0, len(ivs))
	for _, iv := range ivs {
		out = append(out, midiToNote(base+iv))
	}
	return out, nil
}

func parseIntSafe2(s string) (int, error) {
	n := 0
	neg := false
	i := 0
	if i < len(s) && (s[i] == '-' || s[i] == '+') {
		if s[i] == '-' {
			neg = true
		}
		i++
	}
	for i < len(s) && s[i] >= '0' && s[i] <= '9' {
		n = n*10 + int(s[i]-'0')
		i++
	}
	if neg {
		n = -n
	}
	if i == 0 || i < len(s) {
		return 0, fmt.Errorf("not a number: %q", s)
	}
	return n, nil
}

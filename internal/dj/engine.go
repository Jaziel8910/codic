package dj

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// LayerSet holds learned/characteristic mini-notation patterns per musical layer.
type LayerSet struct {
	Kick  []string `json:"kick"`
	Snare []string `json:"snare"`
	Clap  []string `json:"clap"`
	Hat   []string `json:"hat"`
	Bass  []string `json:"bass"`
	Lead  []string `json:"lead"`
	Pad   []string `json:"pad"`
	Fx    []string `json:"fx"`
}

// GenreModel is the persisted knowledge of a genre used for procedural generation.
type GenreModel struct {
	Name       string   `json:"name"`
	BPMMin     int      `json:"bpm_min"`
	BPMMax     int      `json:"bpm_max"`
	Key        string   `json:"key"`
	BassInst   string   `json:"bass_inst"`
	LeadInst   string   `json:"lead_inst"`
	PadInst    string   `json:"pad_inst"`
	Layers     LayerSet `json:"layers"`
	Structures []string `json:"structures"`
	Count      int      `json:"track_count"`
	CreatedAt  string   `json:"created_at"`
	UpdatedAt  string   `json:"updated_at"`
}

// ---- directory / persistence ----

func djDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		home = "."
	}
	return filepath.Join(home, ".codic", "dj")
}

func genrePath(genre string) string {
	return filepath.Join(djDir(), "genres", sanitizeName(genre)+".json")
}

func sanitizeName(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	var b strings.Builder
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9', r == '_', r == '-':
			b.WriteRune(r)
		default:
			b.WriteRune('_')
		}
	}
	return b.String()
}

func nowStr() string { return time.Now().Format("2006-01-02") }

func loadModel(genre string) (*GenreModel, error) {
	data, err := os.ReadFile(genrePath(genre))
	if err != nil {
		return nil, err
	}
	var g GenreModel
	if err := json.Unmarshal(data, &g); err != nil {
		return nil, err
	}
	return &g, nil
}

func saveModel(genre string, g *GenreModel) error {
	path := genrePath(genre)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(g, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

// ListGenres returns the names of all learned genres.
func ListGenres() ([]string, error) {
	dir := filepath.Join(djDir(), "genres")
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var names []string
	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".json" {
			continue
		}
		names = append(names, e.Name()[:len(e.Name())-len(".json")])
	}
	sort.Strings(names)
	return names, nil
}

// Forget deletes a learned genre model.
func Forget(genre string) error {
	err := os.Remove(genrePath(genre))
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

// Stats loads a model and returns it (for inspection by the CLI).
func Stats(genre string) (*GenreModel, error) {
	return loadModel(genre)
}

// ---- learning ----

// Learn trains (or extends) a genre model from a set of .cdc files. When tags
// are supplied, only layers whose category is in the tag set are retained.
func Learn(genre string, files, tags []string) error {
	g, err := loadModel(genre)
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}
		g = defaultProfile(genre)
		g.Name = genre
		g.CreatedAt = nowStr()
	}
	g.UpdatedAt = nowStr()

	tagSet := map[string]bool{}
	for _, t := range tags {
		tagSet[strings.ToLower(strings.TrimSpace(t))] = true
	}

	for _, f := range files {
		src, err := os.ReadFile(f)
		if err != nil {
			return fmt.Errorf("reading %s: %w", f, err)
		}
		ls := learnFromSource(string(src))
		g.Layers = mergeLayers(g.Layers, ls, tagSet)
		g.Count++
	}
	return saveModel(genre, g)
}

func mergeLayers(dst, src LayerSet, tagSet map[string]bool) LayerSet {
	add := func(cat string, arr *[]string, vals []string) {
		if len(tagSet) > 0 && !tagSet[cat] {
			return
		}
		for _, v := range vals {
			if !containsStr(*arr, v) {
				*arr = append(*arr, v)
			}
		}
	}
	add("kick", &dst.Kick, src.Kick)
	add("snare", &dst.Snare, src.Snare)
	add("clap", &dst.Clap, src.Clap)
	add("hat", &dst.Hat, src.Hat)
	add("bass", &dst.Bass, src.Bass)
	add("lead", &dst.Lead, src.Lead)
	add("pad", &dst.Pad, src.Pad)
	add("fx", &dst.Fx, src.Fx)
	return dst
}

func containsStr(arr []string, s string) bool {
	for _, a := range arr {
		if a == s {
			return true
		}
	}
	return false
}

func learnFromSource(src string) LayerSet {
	var ls LayerSet
	for _, raw := range strings.Split(src, "\n") {
		line := strings.TrimSpace(raw)
		if line == "" || strings.HasPrefix(line, "#") ||
			strings.HasPrefix(line, "//") || strings.HasPrefix(line, "@") {
			continue
		}
		// sample-based layers: s("...")
		if idx := strings.Index(line, "s("); idx >= 0 {
			if pat := extractStringArg(line[idx:], "s"); pat != "" {
				if cat := classifyByTokens(pat); cat != "" {
					appendUnique(&ls, cat, pat)
				}
			}
		}
		// note-based layers: note("...") or .note("...")
		if strings.Contains(line, "note(") {
			if pat := extractStringArg(line, "note"); pat != "" {
				if cat := classifyNotes(pat); cat != "" {
					appendUnique(&ls, cat, pat)
				}
			}
		}
	}
	return ls
}

func appendUnique(ls *LayerSet, cat, pat string) {
	arr := layerArrPtr(ls, cat)
	if arr == nil {
		return
	}
	if !containsStr(*arr, pat) {
		*arr = append(*arr, pat)
	}
}

func classifyByTokens(pat string) string {
	low := strings.ToLower(pat)
	priority := []struct{ tok, cat string }{
		{"bd", "kick"}, {"kick", "kick"},
		{"sd", "snare"}, {"snare", "snare"}, {"clap", "snare"}, {"rim", "snare"},
		{"hh", "hat"}, {"hat", "hat"}, {"ho", "hat"}, {"oh", "hat"},
		{"cp", "clap"}, {"cow", "clap"}, {"shaker", "clap"}, {"perc", "clap"},
		{"bass", "bass"}, {"sub", "bass"}, {"808", "bass"}, {"reese", "bass"},
	}
	for _, p := range priority {
		if strings.Contains(low, p.tok) {
			return p.cat
		}
	}
	return ""
}

func classifyNotes(pat string) string {
	var sum, n, lo, hi int
	lo = 127
	hi = 0
	for _, f := range strings.Fields(pat) {
		if m, ok := noteNameToMidi(f); ok {
			sum += m
			n++
			if m < lo {
				lo = m
			}
			if m > hi {
				hi = m
			}
		}
	}
	if n == 0 {
		return ""
	}
	avg := sum / n
	if avg < 48 { // below c3
		return "bass"
	}
	return "lead"
}

// extractStringArg finds funcName("...") and returns the inner string.
func extractStringArg(line, fn string) string {
	idx := strings.Index(line, fn+"(")
	if idx < 0 {
		return ""
	}
	rest := line[idx+len(fn)+1:]
	q1 := strings.Index(rest, "\"")
	if q1 < 0 {
		return ""
	}
	rest2 := rest[q1+1:]
	q2 := strings.Index(rest2, "\"")
	if q2 < 0 {
		return ""
	}
	return rest2[:q2]
}

// ---- default genre profiles ----

// DefaultProfileFor returns the built-in default model for a genre (used when a
// genre has not been learned yet, or for exporting a base profile). It returns
// a fresh copy so callers can mutate it without affecting the shared defaults.
func DefaultProfileFor(genre string) *GenreModel {
	m := defaultProfile(genre)
	cp := *m
	return &cp
}

func defaultProfile(genre string) *GenreModel {
	if p, ok := DefaultProfiles[strings.ToLower(genre)]; ok {
		g := p
		if g.BassInst == "" {
			g.BassInst = "sawtooth"
		}
		if g.LeadInst == "" {
			g.LeadInst = "square"
		}
		if g.PadInst == "" {
			g.PadInst = "triangle"
		}
		return &g
	}
	// generic fallback so generation always produces something
	return &GenreModel{
		Name:     genre,
		BPMMin:   120,
		BPMMax:   140,
		Key:      "c minor",
		BassInst: "sawtooth",
		LeadInst: "square",
		PadInst:  "triangle",
		Layers: LayerSet{
			Kick:  []string{"bd*4"},
			Hat:   []string{"hh*8"},
			Snare: []string{"~ sd ~ sd"},
			Bass:  []string{"c2 c2 c2 c2"},
			Lead:  []string{"c3 e3 g3 c4"},
		},
		Structures: []string{"standard_edm"},
	}
}

// DefaultProfiles are built-in characteristic patterns for common genres. They
// are used both as a starting point for `dj learn` and as fallback when a genre
// has not been trained yet.
var DefaultProfiles = map[string]GenreModel{
	"techno": {
		BPMMin: 126, BPMMax: 140, Key: "a minor",
		BassInst: "sawtooth", LeadInst: "square", PadInst: "triangle",
		Layers: LayerSet{
			Kick:  []string{"bd*4", "bd(16,4)"},
			Hat:   []string{"hh*8", "hh*16"},
			Snare: []string{"~ sd ~ sd"},
			Clap:  []string{"~ cp ~ cp"},
			Bass:  []string{"a1 a1 a1 a1", "a1 ~ a1 a1"},
			Lead:  []string{"a2 c3 e3", "a2 e3 g3"},
			Pad:   []string{"a2 c3 e3 a3"},
		},
		Structures: []string{"standard_edm", "minimal"},
	},
	"house": {
		BPMMin: 120, BPMMax: 128, Key: "f minor",
		BassInst: "sawtooth", LeadInst: "square", PadInst: "triangle",
		Layers: LayerSet{
			Kick:  []string{"bd*4"},
			Hat:   []string{"~ hh ~ hh", "hh*8"},
			Snare: []string{"~ sd ~ sd"},
			Clap:  []string{"~ cp ~ cp", "cp*4"},
			Bass:  []string{"a1 a1 a1 a1", "f1 f1 a1 a1"},
			Lead:  []string{"a2 c3 e3", "f2 a2 c3"},
			Pad:   []string{"a2 c3 e3 a3"},
		},
		Structures: []string{"standard_edm"},
	},
	"dnb": {
		BPMMin: 160, BPMMax: 180, Key: "d minor",
		BassInst: "sawtooth", LeadInst: "square", PadInst: "triangle",
		Layers: LayerSet{
			Kick:  []string{"bd ~ ~ ~ ~ ~ bd ~", "bd ~ ~ ~ bd ~ ~ ~"},
			Snare: []string{"~ ~ sd ~ ~ ~ sd ~", "sd ~ ~ ~ sd ~ ~ ~"},
			Hat:   []string{"hh*16"},
			Bass:  []string{"d1 ~ d1 d2", "d1 d1 ~ d2 d1"},
			Lead:  []string{"d2 f2 a2", "d2 a2 c3"},
			Pad:   []string{"d2 f2 a2 d3"},
		},
		Structures: []string{"standard_edm", "breakbeat"},
	},
	"ambient": {
		BPMMin: 60, BPMMax: 90, Key: "c major",
		BassInst: "triangle", LeadInst: "sine", PadInst: "triangle",
		Layers: LayerSet{
			Kick: []string{"bd ~ ~ ~"},
			Hat:  []string{"hh ~ ~ ~"},
			Bass: []string{"c2 ~ g1 ~", "c2 e2 g2 ~"},
			Lead: []string{"c3 e3 g3", "e3 g3 c4"},
			Pad:  []string{"c3 e3 g3 c4", "a2 c3 e3 a3"},
		},
		Structures: []string{"ambient"},
	},
	"trance": {
		BPMMin: 128, BPMMax: 140, Key: "e minor",
		BassInst: "sawtooth", LeadInst: "square", PadInst: "sawtooth",
		Layers: LayerSet{
			Kick:  []string{"bd*4"},
			Hat:   []string{"hh*8", "oh*4"},
			Snare: []string{"~ sd ~ sd"},
			Bass:  []string{"e1 e1 e1 e1", "e1 g1 e1 b1"},
			Lead:  []string{"e2 g2 b2 e3", "e2 b2 e3 g3"},
			Pad:   []string{"e2 g2 b2 e3"},
		},
		Structures: []string{"standard_edm"},
	},
	"dub": {
		BPMMin: 60, BPMMax: 90, Key: "c minor",
		BassInst: "sine", LeadInst: "sine", PadInst: "triangle",
		Layers: LayerSet{
			Kick: []string{"bd ~ ~ ~", "bd ~ bd ~"},
			Hat:  []string{"hh ~ ~ ~"},
			Bass: []string{"c2 ~ c2 ~", "c2 g1 ~ c2"},
			Lead: []string{"c3 ~ e3 ~", "g2 ~ c3 ~"},
			Pad:  []string{"c3 g3 c4", "c3 eb3 g3 c4"},
		},
		Structures: []string{"dub_techno", "ambient"},
	},
	"minimal": {
		BPMMin: 120, BPMMax: 130, Key: "c minor",
		BassInst: "sawtooth", LeadInst: "square", PadInst: "triangle",
		Layers: LayerSet{
			Kick:  []string{"bd*4", "bd ~ bd ~"},
			Hat:   []string{"hh*8", "hh*4"},
			Snare: []string{"~ sd ~ ~"},
			Bass:  []string{"c2 ~ ~ c2", "c2 e2 ~ c2"},
			Lead:  []string{"c3 ~ e3 ~"},
			Pad:   []string{"c3 eb3 g3"},
		},
		Structures: []string{"minimal"},
	},
	"industrial": {
		BPMMin: 130, BPMMax: 150, Key: "c minor",
		BassInst: "sawtooth", LeadInst: "square", PadInst: "sawtooth",
		Layers: LayerSet{
			Kick:  []string{"bd*4", "bd(16,6)"},
			Hat:   []string{"hh*16"},
			Snare: []string{"sd ~ sd ~", "sd sd ~ ~"},
			Bass:  []string{"c1 c1 c1 c1", "c1 ~ c1 c1"},
			Lead:  []string{"c2 eb2 g2", "c2 g2 bb2"},
			Pad:   []string{"c2 eb2 g2 c3"},
		},
		Structures: []string{"standard_edm", "minimal"},
	},
	"breakbeat": {
		BPMMin: 130, BPMMax: 150, Key: "e minor",
		BassInst: "sawtooth", LeadInst: "square", PadInst: "triangle",
		Layers: LayerSet{
			Kick:  []string{"bd ~ ~ ~ ~ bd ~ ~", "bd ~ ~ ~ bd ~ bd ~"},
			Snare: []string{"~ ~ sd ~ ~ ~ sd ~ sd ~"},
			Hat:   []string{"hh*8", "oh*4"},
			Bass:  []string{"e1 e1 ~ e2 e1", "e1 g1 e1 b1"},
			Lead:  []string{"e2 g2 b2", "e2 b2 e3"},
			Pad:   []string{"e2 g2 b2 e3"},
		},
		Structures: []string{"breakbeat", "standard_edm"},
	},
}

// ---- note helpers ----

var noteOffsets = map[string]int{
	"c": 0, "c#": 1, "db": 1, "d": 2, "d#": 3, "eb": 3, "e": 4, "f": 5,
	"f#": 6, "gb": 6, "g": 7, "g#": 8, "ab": 8, "a": 9, "a#": 10, "bb": 10, "b": 11,
}

func noteNameToMidi(name string) (int, bool) {
	name = strings.TrimSpace(strings.ToLower(name))
	if name == "" {
		return 0, false
	}
	i := 0
	letter := name[i]
	i++
	if letter < 'a' || letter > 'g' {
		return 0, false
	}
	acc := ""
	for i < len(name) && (name[i] == '#' || name[i] == 'b') {
		acc += string(name[i])
		i++
	}
	key := string(letter) + acc
	semi, ok := noteOffsets[key]
	if !ok {
		return 0, false
	}
	oct := 4
	if i < len(name) {
		fmt.Sscanf(name[i:], "%d", &oct)
	}
	return (oct+1)*12 + semi, true
}

func midiToName(m int) string {
	names := []string{"c", "c#", "d", "d#", "e", "f", "f#", "g", "g#", "a", "a#", "b"}
	oct := m/12 - 1
	semi := m % 12
	if semi < 0 {
		semi += 12
		oct--
	}
	return names[semi] + fmt.Sprintf("%d", oct)
}

// transposeNotes shifts every note name in a mini-notation string by semis.
func transposeNotes(pat string, semis int) string {
	if semis == 0 {
		return pat
	}
	fields := strings.Fields(pat)
	for i, f := range fields {
		if m, ok := noteNameToMidi(f); ok {
			fields[i] = midiToName(m + semis)
		}
	}
	return strings.Join(fields, " ")
}

func rootOf(k string) int {
	k = strings.TrimSpace(k)
	if k == "" {
		return 60 // c4
	}
	parts := strings.Fields(k)
	if m, ok := noteNameToMidi(parts[0]); ok {
		return m
	}
	return 60
}

// keyShift returns the semitone offset to move from the genre's default key to
// the requested key.
func keyShift(fromKey, toKey string) int {
	if strings.TrimSpace(toKey) == "" {
		return 0
	}
	return rootOf(toKey) - rootOf(fromKey)
}

func layerArr(ls LayerSet, cat string) []string {
	switch cat {
	case "kick":
		return ls.Kick
	case "snare":
		return ls.Snare
	case "clap":
		return ls.Clap
	case "hat":
		return ls.Hat
	case "bass":
		return ls.Bass
	case "lead":
		return ls.Lead
	case "pad":
		return ls.Pad
	case "fx":
		return ls.Fx
	}
	return nil
}

func layerArrPtr(ls *LayerSet, cat string) *[]string {
	switch cat {
	case "kick":
		return &ls.Kick
	case "snare":
		return &ls.Snare
	case "clap":
		return &ls.Clap
	case "hat":
		return &ls.Hat
	case "bass":
		return &ls.Bass
	case "lead":
		return &ls.Lead
	case "pad":
		return &ls.Pad
	case "fx":
		return &ls.Fx
	}
	return nil
}

func genericLayer(cat string) []string {
	switch cat {
	case "kick":
		return []string{"bd*4"}
	case "hat":
		return []string{"hh*8"}
	case "snare":
		return []string{"~ sd ~ sd"}
	case "clap":
		return []string{"~ cp ~ cp"}
	case "bass":
		return []string{"c2 c2 c2 c2"}
	case "lead":
		return []string{"c3 e3 g3 c4"}
	case "pad":
		return []string{"c3 e3 g3 c4"}
	}
	return nil
}

// pickLayer returns a random pattern for cat, preferring learned data, then the
// genre default, then a generic fallback.
func (g *GenreModel) pickLayer(rnd *rand.Rand, cat string) string {
	arr := layerArr(g.Layers, cat)
	if len(arr) == 0 {
		arr = genericLayer(cat)
	}
	if len(arr) == 0 {
		return ""
	}
	return arr[rnd.Intn(len(arr))]
}

// Package project implements high-level music project features for Codic:
// albums (collections of tracks), third-party audio stems, and export to the
// open DAWproject format so your music opens in any DAW (Bitwig, Reaper, etc.).
package project

import (
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/Jaziel8910/codic/internal/pattern"
)

// Track is one lane of music in an album: either an audio stem (a file you
// loaded) or a pattern you wrote in Codang.
type Track struct {
	Name         string
	Kind         TrackKind
	StemPath     string          // only for Kind == KindStem
	Pattern      pattern.Pattern // only for Kind == KindPattern
	Color        string
	Gain         float64
	Pan          float64
	StartCycle   float64
	LengthCycles float64
}

type TrackKind int

const (
	KindStem TrackKind = iota
	KindPattern
)

// Project is a whole album / song.
type Project struct {
	Name    string
	BPM     float64
	Tracks  []*Track
	nextRes int
}

// New creates an empty project with a default tempo.
func New(name string) *Project {
	return &Project{Name: name, BPM: 120}
}

// SetName changes the album / project title.
func (p *Project) SetName(name string) { p.Name = name }

// SetBPM changes the tempo used for export and timing.
func (p *Project) SetBPM(bpm float64) {
	if bpm > 0 {
		p.BPM = bpm
	}
}

// AddStem registers a third-party audio file as a track (drums you downloaded,
// a vocal take, a sample pack loop...).
func (p *Project) AddStem(name, path string, opts ...TrackOption) {
	t := &Track{Name: name, Kind: KindStem, StemPath: path, Gain: 1, Pan: 0,
		Color: "#6cc7ff", StartCycle: 0, LengthCycles: 8}
	for _, o := range opts {
		o(t)
	}
	p.Tracks = append(p.Tracks, t)
}

// AddPattern registers a Codang pattern as a track (it becomes MIDI notes on
// export).
func (p *Project) AddPattern(name string, pat pattern.Pattern, opts ...TrackOption) {
	t := &Track{Name: name, Kind: KindPattern, Pattern: pat, Gain: 1, Pan: 0,
		Color: "#ff7ad9", StartCycle: 0, LengthCycles: 4}
	for _, o := range opts {
		o(t)
	}
	p.Tracks = append(p.Tracks, t)
}

// TrackOption customises a track (gain, pan, colour, position, length).
type TrackOption func(*Track)

// WithGain sets the track volume (0..1).
func WithGain(g float64) TrackOption { return func(t *Track) { t.Gain = g } }

// WithPan sets stereo position (-1 left .. 1 right).
func WithPan(p float64) TrackOption { return func(t *Track) { t.Pan = p } }

// WithColor sets the track colour (hex like "#ff0000").
func WithColor(c string) TrackOption { return func(t *Track) { t.Color = c } }

// AtCycle sets the start position in cycles (1 cycle = 1 bar at 4/4).
func AtCycle(c float64) TrackOption { return func(t *Track) { t.StartCycle = c } }

// ForCycles sets how many cycles long the track is.
func ForCycles(c float64) TrackOption { return func(t *Track) { t.LengthCycles = c } }

// ExportDAWProject writes the project to a .dawproject file (open format).
func (p *Project) ExportDAWProject(path string) error {
	doc := p.buildDAWProject()
	out, err := xml.MarshalIndent(doc, "", "  ")
	if err != nil {
		return err
	}
	header := []byte(`<?xml version="1.0" encoding="UTF-8"?>` + "\n")
	full := append(header, out...)
	if dir := filepath.Dir(path); dir != "" {
		_ = os.MkdirAll(dir, 0o755)
	}
	return os.WriteFile(path, full, 0o644)
}

// ---- dawproject XML structures ----

type dawProject struct {
	XMLName     xml.Name       `xml:"Project"`
	XMLNS       string         `xml:"xmlns,attr"`
	Version     string         `xml:"version,attr"`
	Application dawApplication `xml:"Application"`
	Transport   dawTransport   `xml:"Transport"`
	Structure   dawStructure   `xml:"Structure"`
	Resources   dawResources   `xml:"Resources"`
	Decoders    dawDecoders    `xml:"Decoders"`
}

type dawApplication struct {
	Name    string `xml:"name,attr"`
	Version string `xml:"version,attr"`
}

type dawTransport struct {
	Tempo dawTempo `xml:"Tempo"`
}

type dawTempo struct {
	Start string `xml:"start,attr"`
}

type dawStructure struct {
	Tracks dawTracks `xml:"Tracks"`
}

type dawTracks struct {
	Track []dawTrack `xml:"Track"`
}

type dawTrack struct {
	Name  string   `xml:"name,attr"`
	Color string   `xml:"color,attr,omitempty"`
	Clips dawClips `xml:"Clips"`
}

type dawClips struct {
	Audio []dawAudioClip `xml:"AudioClip"`
	Midi  []dawMidiClip  `xml:"MidiClip"`
}

type dawAudioClip struct {
	Content string  `xml:"content,attr"`
	Name    string  `xml:"name,attr"`
	Length  string  `xml:"length,attr"`
	Time    dawTime `xml:"Time"`
}

type dawMidiClip struct {
	Content string   `xml:"content,attr"`
	Name    string   `xml:"name,attr"`
	Length  string   `xml:"length,attr"`
	Time    dawTime  `xml:"Time"`
	Notes   dawNotes `xml:"Notes"`
}

type dawTime struct {
	Time string `xml:"time,attr"`
}

type dawNotes struct {
	Note []dawNote `xml:"Note"`
}

type dawNote struct {
	Time     string `xml:"time,attr"`
	Pitch    string `xml:"pitch,attr"`
	Velocity string `xml:"velocity,attr"`
	Duration string `xml:"duration,attr"`
}

type dawResources struct {
	Audio []dawAudioRef `xml:"AudioFileReference"`
}

type dawAudioRef struct {
	ID   string `xml:"id,attr"`
	Path string `xml:"path,attr"`
}

type dawDecoders struct{}

func (p *Project) buildDAWProject() dawProject {
	cycleSec := 60.0 / p.BPM
	doc := dawProject{
		XMLNS:       "http://www.dawproject.org",
		Version:     "1.0.0",
		Application: dawApplication{Name: "Codic", Version: "0.1.0"},
		Transport:   dawTransport{Tempo: dawTempo{Start: fmt.Sprintf("%.4f", p.BPM)}},
		Resources:   dawResources{},
	}
	trackList := []dawTrack{}
	for _, t := range p.Tracks {
		dt := dawTrack{Name: t.Name, Color: t.Color}
		startSec := t.StartCycle * cycleSec
		lenSec := t.LengthCycles * cycleSec
		switch t.Kind {
		case KindStem:
			refID := fmt.Sprintf("res_%d", p.nextRes)
			p.nextRes++
			doc.Resources.Audio = append(doc.Resources.Audio,
				dawAudioRef{ID: refID, Path: t.StemPath})
			dt.Clips.Audio = append(dt.Clips.Audio, dawAudioClip{
				Content: refID, Name: t.Name,
				Length: fmt.Sprintf("%.4f", lenSec),
				Time:   dawTime{Time: fmt.Sprintf("%.4f", startSec)},
			})
		case KindPattern:
			notes := p.patternNotes(t, cycleSec)
			dt.Clips.Midi = append(dt.Clips.Midi, dawMidiClip{
				Content: fmt.Sprintf("res_%d", p.nextRes),
				Name:    t.Name,
				Length:  fmt.Sprintf("%.4f", lenSec),
				Time:    dawTime{Time: fmt.Sprintf("%.4f", startSec)},
				Notes:   dawNotes{Note: notes},
			})
			p.nextRes++
		}
		trackList = append(trackList, dt)
	}
	doc.Structure.Tracks = dawTracks{Track: trackList}
	return doc
}

// patternNotes turns a Codang pattern into DAWproject MIDI notes.
func (p *Project) patternNotes(t *Track, cycleSec float64) []dawNote {
	span := t.LengthCycles
	if span <= 0 {
		span = 4
	}
	haps := t.Pattern.Query(pattern.State{Span: pattern.TimeSpan{
		Begin: pattern.FracInt(0), End: pattern.FracInt(int64(span))}})
	if len(haps) == 0 {
		// fall back to first cycle if the pattern is sparse
		haps = t.Pattern.FirstCycle()
	}
	notes := make([]dawNote, 0, len(haps))
	const maxNotes = 8192
	for i, h := range haps {
		if i >= maxNotes {
			break
		}
		pitch, ok := hapPitch(h)
		if !ok {
			continue
		}
		start := h.Part.Begin.Float64() * cycleSec
		dur := h.Part.End.Sub(h.Part.Begin).Float64() * cycleSec
		if dur <= 0 {
			dur = 0.1 * cycleSec
		}
		gain := hapGain(h)
		vel := int(40 + gain*87)
		if vel > 127 {
			vel = 127
		}
		notes = append(notes, dawNote{
			Time:     strconv.FormatFloat(start, 'f', 4, 64),
			Pitch:    strconv.Itoa(pitch),
			Velocity: strconv.Itoa(vel),
			Duration: strconv.FormatFloat(dur, 'f', 4, 64),
		})
	}
	return notes
}

// drumMap maps classic Strudel/Tidal sample names to General MIDI notes so
// drum patterns still land on the right pads when opened in a DAW.
var drumMap = map[string]int{
	"bd": 36, "sd": 38, "sd!": 38, "rim": 37, "rim!": 37,
	"lt": 41, "mt": 43, "ht": 45, "hh": 42, "hh!": 42, "oh": 46,
	"cy": 49, "cym": 49, "cp": 39, "cl": 39, "claves": 39,
	"tom": 45, "cb": 36, "cowbell": 56, "shaker": 70, "perc": 47,
	"kick": 36, "sn": 38, "snr": 38, "hat": 42, "openhat": 46,
}

// hapPitch extracts a MIDI note number from a hap's value.
func hapPitch(h pattern.Hap) (int, bool) {
	val := h.Value
	switch v := val.(type) {
	case pattern.ControlMap:
		if m, ok := numField(v, "midinote"); ok {
			return int(m + 0.5), true
		}
		if m, ok := numField(v, "note"); ok {
			return int(m + 0.5), true
		}
		if m, ok := numField(v, "n"); ok {
			return int(m + 0.5), true
		}
		if s, ok := v["s"].(string); ok {
			if d, ok := drumMap[strings.TrimSpace(s)]; ok {
				return d, true
			}
		}
		return 0, false
	case string:
		if d, ok := drumMap[strings.TrimSpace(v)]; ok {
			return d, true
		}
		if f, ok := noteToMidi(v); ok {
			return f, true
		}
		return 0, false
	case float64:
		return int(v + 0.5), true
	}
	return 0, false
}

func numField(cm pattern.ControlMap, key string) (float64, bool) {
	if v, ok := cm[key]; ok {
		if f, ok := v.(float64); ok {
			return f, true
		}
	}
	return 0, false
}

func hapGain(h pattern.Hap) float64 {
	if cm, ok := h.Value.(pattern.ControlMap); ok {
		if g, ok := numField(cm, "gain"); ok {
			return clamp01(g)
		}
		if v, ok := numField(cm, "velocity"); ok {
			return clamp01(v)
		}
	}
	return 0.7
}

func clamp01(x float64) float64 {
	if x < 0 {
		return 0
	}
	if x > 1 {
		return 1
	}
	return x
}

// noteToMidi parses note names like "c3", "fs4", "a#2", "do3", "sol#5".
func noteToMidi(s string) (int, bool) {
	s = strings.TrimSpace(strings.ToLower(s))
	if s == "" {
		return 0, false
	}
	// Spanish names (do, re, mi, fa, sol, la, si)
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
			if len(rest) > 0 {
				if rest[0] == '#' || rest[0] == 's' {
					semi++
					rest = rest[1:]
				} else if rest[0] == 'b' {
					semi--
					rest = rest[1:]
				}
			}
			if rest != "" {
				if o, err := strconv.Atoi(rest); err == nil {
					oct = o
				}
			}
			return semi + (oct+1)*12, true
		}
	}
	// English single-letter names
	names := map[rune]int{'c': 0, 'd': 2, 'e': 4, 'f': 5, 'g': 7, 'a': 9, 'b': 11}
	base := rune(s[0])
	semi := 0
	if b, ok := names[base]; ok {
		semi = b
	} else {
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
		o, err := strconv.Atoi(octStr)
		if err != nil {
			return 0, false
		}
		oct = o
	}
	return semi + (oct+1)*12, true
}

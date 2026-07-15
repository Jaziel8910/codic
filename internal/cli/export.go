package cli

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/Jaziel8910/codic/internal/audio"
	"github.com/Jaziel8910/codic/internal/pattern"
)

func exportCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "export",
		Short: "Export audio / projects to many formats",
	}
	cmd.AddCommand(exportWavCmd())
	cmd.AddCommand(exportMp3Cmd())
	cmd.AddCommand(exportFlacCmd())
	cmd.AddCommand(exportOggCmd())
	cmd.AddCommand(exportVideoCmd())
	cmd.AddCommand(exportStemsCmd())
	cmd.AddCommand(exportMidiCmd())
	cmd.AddCommand(exportDawprojectCmd())
	return cmd
}

func exportRender(file string, secs float64) ([]float64, error) {
	combined, cps, err := collectPatternsFromFile(file)
	if err != nil {
		return nil, err
	}
	if secs <= 0 {
		secs = 8
	}
	buf, err := audio.RenderPattern(combined, cps, secs)
	if err != nil {
		return nil, err
	}
	return normalizePeak(buf), nil
}

func exportWavCmd() *cobra.Command {
	var secs float64
	cmd := &cobra.Command{
		Use:   "wav <file.cdc> [out.wav]",
		Short: "Render a pattern to WAV",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			buf, err := exportRender(args[0], secs)
			if err != nil {
				return err
			}
			out := strings.TrimSuffix(args[0], ".cdc") + ".wav"
			if len(args) > 1 {
				out = args[1]
			}
			if err := audio.WriteWAVFile(out, buf, audio.SampleRate); err != nil {
				return err
			}
			fmt.Printf("exported WAV -> %s\n", out)
			return nil
		},
	}
	cmd.Flags().Float64Var(&secs, "sec", 8, "length in seconds")
	return cmd
}

func ffmpegConvert(in, out string) error {
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		return fmt.Errorf("ffmpeg not found (install ffmpeg to export %s)", out)
	}
	c := exec.Command("ffmpeg", "-y", "-i", in, out)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	return c.Run()
}

func exportViaFfmpeg(ext string) *cobra.Command {
	return &cobra.Command{
		Use:   fmt.Sprintf("%s <file.cdc> [out.%s]", ext, ext),
		Short: fmt.Sprintf("Render a pattern and convert to %s via ffmpeg", ext),
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			buf, err := exportRender(args[0], 8)
			if err != nil {
				return err
			}
			tmp, err := tempWAV("export_preview.wav")
			if err != nil {
				return err
			}
			if err := audio.WriteWAVFile(tmp, buf, audio.SampleRate); err != nil {
				return err
			}
			out := strings.TrimSuffix(args[0], ".cdc") + "." + ext
			if len(args) > 1 {
				out = args[1]
			}
			if err := ffmpegConvert(tmp, out); err != nil {
				return err
			}
			fmt.Printf("exported %s -> %s\n", strings.ToUpper(ext), out)
			return nil
		},
	}
}

func exportMp3Cmd() *cobra.Command  { return exportViaFfmpeg("mp3") }
func exportFlacCmd() *cobra.Command { return exportViaFfmpeg("flac") }
func exportOggCmd() *cobra.Command  { return exportViaFfmpeg("ogg") }

func exportVideoCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "video <file.cdc> [out.mp4]",
		Short: "Render audio and mux into a video via ffmpeg",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			buf, err := exportRender(args[0], 8)
			if err != nil {
				return err
			}
			tmp, err := tempWAV("video_audio.wav")
			if err != nil {
				return err
			}
			if err := audio.WriteWAVFile(tmp, buf, audio.SampleRate); err != nil {
				return err
			}
			out := strings.TrimSuffix(args[0], ".cdc") + ".mp4"
			if len(args) > 1 {
				out = args[1]
			}
			if _, err := exec.LookPath("ffmpeg"); err != nil {
				return fmt.Errorf("ffmpeg not found (needed for video export)")
			}
			c := exec.Command("ffmpeg", "-y", "-i", tmp, "-f", "lavfi", "-i",
				"color=c=blue:s=640x360:d=8", "-shortest", out)
			c.Stdout = os.Stdout
			c.Stderr = os.Stderr
			if err := c.Run(); err != nil {
				return err
			}
			fmt.Printf("exported VIDEO -> %s\n", out)
			return nil
		},
	}
	return cmd
}

func exportStemsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "stems [dir]",
		Short: "Render each track of the project to a separate WAV",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			dir := "."
			if len(args) > 0 {
				dir = args[0]
			}
			pf, err := loadProjectFile(projectYAMLPath(dir))
			if err != nil {
				return err
			}
			if err := os.MkdirAll("stems", 0o755); err != nil {
				return err
			}
			for i, t := range pf.Tracks {
				code, e := os.ReadFile(t.File)
				if e != nil {
					return e
				}
				combined, cps, e := collectPatterns(string(code))
				if e != nil {
					return e
				}
				buf, e := audio.RenderPattern(combined, cps, trackSeconds(t, cps))
				if e != nil {
					return e
				}
				buf = normalizePeak(buf)
				out := filepath.Join("stems", fmt.Sprintf("%02d_%s.wav", i, sanitize(t.Name)))
				if e := audio.WriteWAVFile(out, buf, audio.SampleRate); e != nil {
					return e
				}
				fmt.Printf("stem %d -> %s\n", i, out)
			}
			return nil
		},
	}
}

func exportMidiCmd() *cobra.Command {
	var secs float64
	cmd := &cobra.Command{
		Use:   "midi <file.cdc> [out.mid]",
		Short: "Export a pattern to a Standard MIDI File",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			combined, cps, err := collectPatternsFromFile(args[0])
			if err != nil {
				return err
			}
			if secs <= 0 {
				secs = 8
			}
			notes := patternNotesMIDI(combined, cps, secs)
			out := strings.TrimSuffix(args[0], ".cdc") + ".mid"
			if len(args) > 1 {
				out = args[1]
			}
			if err := writeMIDIFile(out, notes); err != nil {
				return err
			}
			fmt.Printf("exported MIDI -> %s (%d notes)\n", out, len(notes))
			return nil
		},
	}
	cmd.Flags().Float64Var(&secs, "sec", 8, "length in seconds")
	return cmd
}

func exportDawprojectCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "dawproject [out.dawproject]",
		Short: "Export the current project to DAWproject format",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			pf, err := loadProjectFile(projectYAMLPath("."))
			if err != nil {
				return err
			}
			p, err := buildEngineProject(pf)
			if err != nil {
				return err
			}
			out := "album.dawproject"
			if len(args) > 0 {
				out = args[0]
			}
			if err := p.ExportDAWProject(out); err != nil {
				return err
			}
			fmt.Printf("exported DAWproject -> %s\n", out)
			return nil
		},
	}
}

// ---- minimal MIDI writer ----

type midiNote struct {
	start float64
	dur   float64
	pitch int
	vel   int
}

func patternNotesMIDI(pat pattern.Pattern, cps, secs float64) []midiNote {
	span := int64(secs * cps)
	if span < 1 {
		span = 1
	}
	haps := pat.Query(pattern.State{Span: pattern.TimeSpan{Begin: pattern.FracInt(0), End: pattern.FracInt(span)}})
	var notes []midiNote
	for _, h := range haps {
		pitch, ok := hapMidiPitch(h.Value)
		if !ok {
			continue
		}
		start := h.Part.Begin.Float64() / cps
		dur := h.Part.End.Sub(h.Part.Begin).Float64() / cps
		if dur <= 0 {
			dur = 0.2
		}
		vel := int(90)
		if cm, ok := h.Value.(pattern.ControlMap); ok {
			if g, ok := numFieldCM(cm, "gain"); ok {
				vel = int(40 + g*87)
				if vel > 127 {
					vel = 127
				}
			}
		}
		notes = append(notes, midiNote{start: start, dur: dur, pitch: pitch, vel: vel})
	}
	return notes
}

func hapMidiPitch(v interface{}) (int, bool) {
	drumMap := map[string]int{"bd": 36, "sd": 38, "hh": 42, "oh": 46, "cp": 39, "ht": 45, "lt": 41, "mt": 43, "cy": 49}
	switch val := v.(type) {
	case pattern.ControlMap:
		for _, k := range []string{"midinote", "note", "n"} {
			if m, ok := numFieldCM(val, k); ok {
				return int(m + 0.5), true
			}
		}
		if s, ok := val["s"].(string); ok {
			if d, ok := drumMap[strings.TrimSpace(s)]; ok {
				return d, true
			}
		}
	case string:
		if d, ok := drumMap[strings.TrimSpace(val)]; ok {
			return d, true
		}
		if m, ok := noteToMidi(val); ok {
			return m, true
		}
	case float64:
		return int(val + 0.5), true
	}
	return 0, false
}

func numFieldCM(cm pattern.ControlMap, key string) (float64, bool) {
	if v, ok := cm[key]; ok {
		if f, ok := v.(float64); ok {
			return f, true
		}
	}
	return 0, false
}

func writeMIDIFile(path string, notes []midiNote) error {
	var track bytes.Buffer
	// delta-time helper (variable length)
	vlq := func(n int) []byte {
		b := []byte{byte(n & 0x7f)}
		n >>= 7
		for n > 0 {
			b = append([]byte{byte(n&0x7f) | 0x80}, b...)
			n >>= 7
		}
		return b
	}
	// note on / off events
	type ev struct {
		tick int
		data []byte
	}
	var evs []ev
	ticksPerSec := 480.0 / 1.0 // 480 ticks per second assumption
	for _, n := range notes {
		on := int(n.start * ticksPerSec)
		off := int((n.start + n.dur) * ticksPerSec)
		evs = append(evs, ev{on, []byte{0x90, byte(n.pitch), byte(n.vel)}})
		evs = append(evs, ev{off, []byte{0x80, byte(n.pitch), 0}})
	}
	// sort by tick
	for i := 1; i < len(evs); i++ {
		for j := i; j > 0 && evs[j].tick < evs[j-1].tick; j-- {
			evs[j], evs[j-1] = evs[j-1], evs[j]
		}
	}
	last := 0
	for _, e := range evs {
		d := e.tick - last
		if d < 0 {
			d = 0
		}
		track.Write(vlq(d))
		track.Write(e.data)
		last = e.tick
	}
	track.Write(vlq(0))
	track.Write([]byte{0xff, 0x2f, 0x00}) // end of track

	var head bytes.Buffer
	head.Write([]byte("MThd"))
	head.Write([]byte{0, 0, 0, 6})
	head.Write([]byte{0, 0})       // format 0
	head.Write([]byte{0, 1})       // 1 track
	head.Write([]byte{0x01, 0xe0}) // 480 ticks per beat
	head.Write([]byte("MTrk"))
	th := make([]byte, 4)
	binary.BigEndian.PutUint32(th, uint32(track.Len()))
	head.Write(th)
	head.Write(track.Bytes())

	return os.WriteFile(path, head.Bytes(), 0o644)
}

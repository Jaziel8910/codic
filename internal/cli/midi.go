package cli

import (
	"encoding/binary"
	"fmt"
	"math"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/Jaziel8910/codic/internal/audio"
)

type midiEvent struct {
	tick int
	ch   int
	note int
	vel  int
	on   bool
}

// parseSMF parses a Standard MIDI File into note on/off events with absolute
// ticks. It supports format 0 and 1 and reads tempo to convert ticks to
// seconds.
func parseSMF(data []byte) ([]midiEvent, float64, error) {
	if len(data) < 14 || string(data[0:4]) != "MThd" {
		return nil, 0, fmt.Errorf("not a MIDI file")
	}
	// header length (should be 6)
	_ = data[4:8]
	// format := int(binary.BigEndian.Uint16(data[8:10]))
	ntracks := int(binary.BigEndian.Uint16(data[10:12]))
	ticksPerBeat := int(binary.BigEndian.Uint16(data[12:14]))
	if ntracks < 1 {
		return nil, 0, fmt.Errorf("no tracks")
	}
	pos := 14
	tempo := 500000.0 // us per quarter
	var events []midiEvent
	for tr := 0; tr < ntracks; tr++ {
		if pos+8 > len(data) || string(data[pos:pos+4]) != "MTrk" {
			break
		}
		lenb := int(binary.BigEndian.Uint32(data[pos+4 : pos+8]))
		pos += 8
		end := pos + lenb
		abs := 0
		running := byte(0)
		for pos < end {
			// delta time VLQ
			dt := 0
			for {
				if pos >= end {
					break
				}
				b := data[pos]
				pos++
				dt = (dt << 7) | int(b&0x7f)
				if b&0x80 == 0 {
					break
				}
			}
			abs += dt
			if pos >= end {
				break
			}
			status := data[pos]
			if status&0x80 == 0 {
				// running status
				status = running
			} else {
				pos++
				running = status
			}
			switch {
			case status == 0xff: // meta
				if pos >= end {
					break
				}
				mtype := data[pos]
				pos++
				// length VLQ
				l := 0
				for {
					if pos >= end {
						break
					}
					b := data[pos]
					pos++
					l = (l << 7) | int(b&0x7f)
					if b&0x80 == 0 {
						break
					}
				}
				if mtype == 0x51 && l >= 3 { // tempo
					tempo = float64(int(data[pos])<<16 | int(data[pos+1])<<8 | int(data[pos+2]))
				}
				pos += l
			case status == 0xf0 || status == 0xf7: // sysex
				l := 0
				for {
					if pos >= end {
						break
					}
					b := data[pos]
					pos++
					l = (l << 7) | int(b&0x7f)
					if b&0x80 == 0 {
						break
					}
				}
				pos += l
			case status&0xf0 == 0x90: // note on
				note := int(data[pos])
				vel := int(data[pos+1])
				pos += 2
				events = append(events, midiEvent{tick: abs, ch: int(status & 0x0f), note: note, vel: vel, on: vel > 0})
			case status&0xf0 == 0x80: // note off
				note := int(data[pos])
				pos += 2
				events = append(events, midiEvent{tick: abs, ch: int(status & 0x0f), note: note, vel: 0, on: false})
			case status&0xf0 == 0xa0 || status&0xf0 == 0xb0 || status&0xf0 == 0xe0:
				pos += 2
			case status&0xf0 == 0xc0 || status&0xf0 == 0xd0:
				pos += 1
			default:
				pos++
			}
		}
	}
	secPerTick := (tempo / 1e6) / float64(ticksPerBeat)
	return events, secPerTick, nil
}

func midiCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "midi",
		Short: "Import, export and map MIDI",
	}
	cmd.AddCommand(midiImportCmd())
	cmd.AddCommand(midiExportCmd())
	cmd.AddCommand(midiListCmd())
	cmd.AddCommand(midiPlayCmd())
	cmd.AddCommand(midiLearnCmd())
	cmd.AddCommand(midiMapCmd())
	cmd.AddCommand(midiDeviceCmd())
	cmd.AddCommand(midiClockCmd())
	return cmd
}

func midiImportCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "import <file.mid> [out.cdc]",
		Short: "Convert a MIDI file into a Codang pattern",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			data, err := os.ReadFile(args[0])
			if err != nil {
				return err
			}
			events, secPerTick, err := parseSMF(data)
			if err != nil {
				return err
			}
			out := strings.TrimSuffix(args[0], ".mid") + ".cdc"
			if len(args) > 1 {
				out = args[1]
			}
			var b strings.Builder
			b.WriteString("@name midi-import\n@bpm 120\n\n")
			for _, e := range events {
				if !e.on {
					continue
				}
				sec := float64(e.tick) * secPerTick
				start := sec / 2.0 // quarter = 2 cycles at 4/4-ish
				b.WriteString(fmt.Sprintf("note(%d).late(%s).out()\n", e.note, fmt.Sprintf("%.3f", start)))
			}
			if err := os.WriteFile(out, []byte(b.String()), 0o644); err != nil {
				return err
			}
			fmt.Printf("imported %d events -> %s\n", len(events), out)
			return nil
		},
	}
}

func midiExportCmd() *cobra.Command {
	var secs float64
	cmd := &cobra.Command{
		Use:   "export <file.cdc> [out.mid]",
		Short: "Export a pattern to MIDI",
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
			fmt.Printf("exported MIDI -> %s\n", out)
			return nil
		},
	}
	cmd.Flags().Float64Var(&secs, "sec", 8, "length in seconds")
	return cmd
}

func midiListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List .mid files in the current directory",
		RunE: func(cmd *cobra.Command, args []string) error {
			entries, err := os.ReadDir(".")
			if err != nil {
				return err
			}
			var files []string
			for _, e := range entries {
				if !e.IsDir() && strings.HasSuffix(strings.ToLower(e.Name()), ".mid") {
					files = append(files, e.Name())
				}
			}
			if len(files) == 0 {
				fmt.Println("(no .mid files)")
				return nil
			}
			sort.Strings(files)
			for _, f := range files {
				fmt.Printf("  %s\n", f)
			}
			return nil
		},
	}
}

func midiPlayCmd() *cobra.Command {
	var secs float64
	cmd := &cobra.Command{
		Use:   "play <file.mid> [out.wav]",
		Short: "Render a MIDI file to audio (via a simple synth) and play it",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			data, err := os.ReadFile(args[0])
			if err != nil {
				return err
			}
			events, secPerTick, err := parseSMF(data)
			if err != nil {
				return err
			}
			if secs <= 0 {
				secs = float64(events[len(events)-1].tick)*secPerTick + 1
			}
			n := int(secs*audio.SampleRate) * 2
			buf := make([]float64, n)
			sr := float64(audio.SampleRate)
			for _, e := range events {
				if !e.on || e.vel == 0 {
					continue
				}
				start := int(float64(e.tick) * secPerTick * sr)
				if start >= len(buf)/2 {
					continue
				}
				freq := 440 * math.Exp2((float64(e.note)-69.0)/12.0)
				dur := int(0.3 * sr)
				for i := 0; i < dur && start*2+2*i+1 < len(buf); i++ {
					s := math.Sin(2 * 3.141592653589793 * freq * float64(i) / sr)
					env := 1.0 - float64(i)/float64(dur)
					buf[start*2+2*i] = s * env * 0.3
					buf[start*2+2*i+1] = s * env * 0.3
				}
			}
			out := strings.TrimSuffix(args[0], ".mid") + ".wav"
			if len(args) > 1 {
				out = args[1]
			}
			if err := audio.WriteWAVFile(out, buf, audio.SampleRate); err != nil {
				return err
			}
			fmt.Printf("rendered MIDI -> %s\n", out)
			return openPlayer(out)
		},
	}
	cmd.Flags().Float64Var(&secs, "sec", 0, "length in seconds (0 = full)")
	return cmd
}

func midiLearnCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "learn <note> <sample>",
		Short: "Map a MIDI note number to a sample name",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			note, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("note must be a number")
			}
			mp := map[string]string{}
			_ = registryLoad("midi", "map", &mp)
			mp[strconv.Itoa(note)] = args[1]
			if err := registrySave("midi", "map", &mp); err != nil {
				return err
			}
			fmt.Printf("learned note %d -> %s\n", note, args[1])
			return nil
		},
	}
}

func midiMapCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "map",
		Short: "Show the current MIDI note -> sample mappings",
		RunE: func(cmd *cobra.Command, args []string) error {
			mp := map[string]string{}
			if err := registryLoad("midi", "map", &mp); err != nil {
				fmt.Println("(no mappings)")
				return nil
			}
			keys := make([]string, 0, len(mp))
			for k := range mp {
				keys = append(keys, k)
			}
			sort.Strings(keys)
			for _, k := range keys {
				fmt.Printf("  note %s -> %s\n", k, mp[k])
			}
			return nil
		},
	}
}

func midiDeviceCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "device",
		Short: "List available MIDI devices (offline info)",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("MIDI devices are accessed through your OS MIDI stack.")
			fmt.Println("Codic is in offline mode; use `codic midi import` to bring in .mid files.")
			return nil
		},
	}
}

func midiClockCmd() *cobra.Command {
	var bpm float64
	cmd := &cobra.Command{
		Use:   "clock",
		Short: "Print a MIDI-clock reference for a given tempo",
		RunE: func(cmd *cobra.Command, args []string) error {
			spb := 60.0 / bpm
			fmt.Printf("BPM      : %.0f\n", bpm)
			fmt.Printf("beat (1/4): %.3f s\n", spb)
			fmt.Printf("bar (4/4) : %.3f s\n", spb*4)
			fmt.Printf("24ppqn    : %.3f ms per tick\n", spb*1000/24)
			return nil
		},
	}
	cmd.Flags().Float64Var(&bpm, "bpm", 120, "tempo")
	return cmd
}

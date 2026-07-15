package cli

import (
	"fmt"
	"math"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

func extrasCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "extras",
		Short: "Practice tools: metronome, tuner, tap tempo, theory",
	}
	cmd.AddCommand(extrasMetronomeCmd())
	cmd.AddCommand(extrasTunerCmd())
	cmd.AddCommand(extrasTapCmd())
	cmd.AddCommand(extrasBpmCmd())
	cmd.AddCommand(extrasKeysigCmd())
	cmd.AddCommand(extrasChordidCmd())
	cmd.AddCommand(extrasScaCmd())
	cmd.AddCommand(extrasQuizCmd())
	return cmd
}

func extrasMetronomeCmd() *cobra.Command {
	var bpm, secs float64
	cmd := &cobra.Command{
		Use:   "metronome [out.wav]",
		Short: "Render a metronome click track",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			code := `s("cp").out()`
			out := "metronome.wav"
			if len(args) > 0 {
				out = args[0]
			}
			if err := renderToFileString(fmt.Sprintf("@bpm %.0f\n\n%s", bpm, code), out, secs); err != nil {
				return err
			}
			fmt.Printf("metronome @ %.0f bpm -> %s\n", bpm, out)
			return nil
		},
	}
	cmd.Flags().Float64Var(&bpm, "bpm", 120, "tempo")
	cmd.Flags().Float64Var(&secs, "sec", 8, "length in seconds")
	return cmd
}

func extrasTunerCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "tuner <freq-hz>",
		Short: "Identify the nearest note and cents offset for a frequency",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			f, err := parseFloat(args[0])
			if err != nil {
				return fmt.Errorf("frequency must be a number")
			}
			if f <= 0 {
				return fmt.Errorf("frequency must be positive")
			}
			midi := 69 + 12*math.Log2(f/440.0)
			nearest := math.Round(midi)
			cents := (midi - nearest) * 100
			note := midiToNote(int(nearest))
			fmt.Printf("frequency: %.2f Hz\n", f)
			fmt.Printf("nearest  : %s (MIDI %d)\n", note, int(nearest))
			fmt.Printf("offset   : %+.1f cents\n", cents)
			return nil
		},
	}
}

func extrasTapCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "tap <interval-seconds...>",
		Short: "Compute tempo from a list of tap intervals (seconds)",
		Args:  cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			var sum float64
			for _, a := range args {
				v, err := parseFloat(a)
				if err != nil || v <= 0 {
					return fmt.Errorf("intervals must be positive numbers")
				}
				sum += v
			}
			avg := sum / float64(len(args))
			bpm := 60.0 / avg
			fmt.Printf("average interval: %.3f s\n", avg)
			fmt.Printf("tempo          : %.1f BPM\n", bpm)
			return nil
		},
	}
}

func extrasBpmCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "bpm <file.cdc>",
		Short: "Show the tempo of a pattern (from @bpm)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			data, err := os.ReadFile(args[0])
			if err != nil {
				return err
			}
			if b, ok := scanMeta(string(data), "bpm"); ok {
				fmt.Printf("bpm: %s\n", b)
				return nil
			}
			fmt.Println("no @bpm found")
			return nil
		},
	}
}

// scanMeta finds a top-level "@key value" metadata directive in src.
func scanMeta(src, key string) (string, bool) {
	for _, line := range strings.Split(src, "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "@"+key) {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) >= 2 {
			return fields[1], true
		}
	}
	return "", false
}

func extrasKeysigCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "keysig <root> [major|minor]",
		Short: "Show the key signature for a root note",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			mode := "major"
			if len(args) > 1 {
				mode = args[1]
			}
			notes, err := scaleNotes(args[0], mode)
			if err != nil {
				return err
			}
			fmt.Printf("key: %s %s\n", strings.ToUpper(args[0]), mode)
			fmt.Printf("scale: %s\n", strings.Join(notes, " "))
			return nil
		},
	}
}

func extrasChordidCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "chordid <note...>",
		Short: "Identify a chord from a list of note names",
		Args:  cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			var midis []int
			for _, n := range args {
				m, ok := noteToMidi(n)
				if !ok {
					return fmt.Errorf("invalid note %q", n)
				}
				midis = append(midis, m%12)
			}
			name := identifyChord(midis)
			fmt.Printf("notes: %s\n", strings.Join(args, " "))
			fmt.Printf("chord : %s\n", name)
			return nil
		},
	}
}

func extrasScaCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "sca <root> <scale>",
		Short: "Print the notes of a scale",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			notes, err := scaleNotes(args[0], args[1])
			if err != nil {
				return err
			}
			fmt.Printf("%s %s: %s\n", args[0], args[1], strings.Join(notes, " "))
			return nil
		},
	}
}

func extrasQuizCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "quiz",
		Short: "Generate a random ear-training question",
		RunE: func(cmd *cobra.Command, args []string) error {
			r := rand.New(rand.NewSource(time.Now().UnixNano()))
			all := []string{"c", "c#", "d", "d#", "e", "f", "f#", "g", "g#", "a", "a#", "b"}
			root := all[r.Intn(len(all))]
			scales := []string{"major", "minor", "dorian", "pentatonic", "blues"}
			sc := scales[r.Intn(len(scales))]
			notes, _ := scaleNotes(root, sc)
			fmt.Printf("What scale is this?\n  root : %s\n  type : %s\n  notes: %s\n", root, sc, strings.Join(notes, " "))
			return nil
		},
	}
}

func identifyChord(pcs []int) string {
	set := map[int]bool{}
	for _, p := range pcs {
		set[p] = true
	}
	has := func(n int) bool { return set[n%12] }
	switch {
	case has(0) && has(4) && has(7):
		if has(10) {
			return "minor 7th"
		}
		if has(11) {
			return "dominant 7th"
		}
		if has(3) {
			return "major 7th"
		}
		return "major triad"
	case has(0) && has(3) && has(7):
		return "minor triad"
	case has(0) && has(3) && has(6):
		return "diminished triad"
	case has(0) && has(4) && has(8):
		return "augmented triad"
	default:
		return "unknown / clusters"
	}
}

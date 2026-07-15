package cli

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/Jaziel8910/codic/internal/audio"
	"github.com/Jaziel8910/codic/internal/codang"
	"github.com/Jaziel8910/codic/internal/pattern"
)

func patternCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pattern",
		Short: "Create and manipulate Codang patterns (.cdc)",
	}
	cmd.AddCommand(patternNewCmd())
	cmd.AddCommand(patternListCmd())
	cmd.AddCommand(patternShowCmd())
	cmd.AddCommand(patternRandomCmd())
	cmd.AddCommand(patternSliceCmd())
	cmd.AddCommand(patternCombineCmd())
	cmd.AddCommand(patternTransformCmd())
	cmd.AddCommand(patternArpCmd())
	cmd.AddCommand(patternExportCmd())
	cmd.AddCommand(patternValidateCmd())
	return cmd
}

func patternNewCmd() *cobra.Command {
	var bpm int
	cmd := &cobra.Command{
		Use:   "new <name.cdc>",
		Short: "Create a new empty pattern file",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			file := args[0]
			if !strings.HasSuffix(file, ".cdc") {
				file += ".cdc"
			}
			skel := fmt.Sprintf("@name %s\n@bpm %d\n\ns(\"bd\").out()\n", strings.TrimSuffix(filepath.Base(file), ".cdc"), bpm)
			if err := os.WriteFile(file, []byte(skel), 0o644); err != nil {
				return err
			}
			fmt.Printf("created pattern %s\n", file)
			return nil
		},
	}
	cmd.Flags().IntVar(&bpm, "bpm", 120, "tempo")
	return cmd
}

func patternListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List .cdc pattern files in the current directory",
		RunE: func(cmd *cobra.Command, args []string) error {
			entries, err := os.ReadDir(".")
			if err != nil {
				return err
			}
			var files []string
			for _, e := range entries {
				if !e.IsDir() && strings.HasSuffix(e.Name(), ".cdc") {
					files = append(files, e.Name())
				}
			}
			if len(files) == 0 {
				fmt.Println("(no .cdc files)")
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

func patternShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show <file.cdc>",
		Short: "Show a pattern's metadata and structure",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			code, err := os.ReadFile(args[0])
			if err != nil {
				return err
			}
			prog, err := codang.Parse(string(code))
			if err != nil {
				return fmt.Errorf("parse error: %w", err)
			}
			fmt.Printf("file: %s\n", args[0])
			fmt.Printf("bpm : %s\n", prog.Metadata["bpm"])
			fmt.Printf("key : %s\n", prog.Metadata["key"])
			fmt.Printf("name: %s\n", prog.Metadata["name"])
			combined, cps, err := collectPatterns(string(code))
			if err != nil {
				fmt.Printf("render: %v\n", err)
				return nil
			}
			fmt.Printf("cps : %.4f (%.1f bpm)\n", cps, cps*240)
			fmt.Printf("events (first cycle): %d\n", len(combined.FirstCycle()))
			return nil
		},
	}
}

func patternRandomCmd() *cobra.Command {
	var bpm int
	var bars int
	cmd := &cobra.Command{
		Use:   "random [out.cdc]",
		Short: "Generate a random drum pattern",
		Args:  cobra.RangeArgs(0, 1),
		RunE: func(cmd *cobra.Command, args []string) error {
			out := "random.cdc"
			if len(args) > 0 {
				out = args[0]
			}
			r := rand.New(rand.NewSource(time.Now().UnixNano()))
			drums := []string{"bd", "sd", "hh", "oh", "cp"}
			var b strings.Builder
			b.WriteString(fmt.Sprintf("@bpm %d\n\n", bpm))
			for bar := 0; bar < bars; bar++ {
				for step := 0; step < 16; step++ {
					if r.Float64() < 0.35 {
						d := drums[r.Intn(len(drums))]
						b.WriteString(fmt.Sprintf("s(\"%s\").every(%d, x=>x).out()\n", d, 4))
					}
				}
			}
			if err := os.WriteFile(out, []byte(b.String()), 0o644); err != nil {
				return err
			}
			fmt.Printf("generated random pattern -> %s\n", out)
			return nil
		},
	}
	cmd.Flags().IntVar(&bpm, "bpm", 120, "tempo")
	cmd.Flags().IntVar(&bars, "bars", 2, "number of bars")
	return cmd
}

func patternSliceCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "slice <file.cdc> <n>",
		Short: "Split a pattern into n separate bar files",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			n, err := strconv.Atoi(args[1])
			if err != nil || n < 1 {
				return fmt.Errorf("n must be a positive integer")
			}
			code, err := os.ReadFile(args[0])
			if err != nil {
				return err
			}
			base := strings.TrimSuffix(filepath.Base(args[0]), ".cdc")
			for i := 0; i < n; i++ {
				part := fmt.Sprintf("%s_part%d.cdc", base, i)
				if err := os.WriteFile(part, code, 0o644); err != nil {
					return err
				}
			}
			fmt.Printf("sliced %s into %d parts\n", args[0], n)
			return nil
		},
	}
}

func patternCombineCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "combine <a.cdc> <b.cdc> [out.cdc]",
		Short: "Stack two patterns into one",
		Args:  cobra.RangeArgs(2, 3),
		RunE: func(cmd *cobra.Command, args []string) error {
			a, err := os.ReadFile(args[0])
			if err != nil {
				return err
			}
			b, err := os.ReadFile(args[1])
			if err != nil {
				return err
			}
			out := "combined.cdc"
			if len(args) > 2 {
				out = args[2]
			}
			merged := fmt.Sprintf("// %s\n%s\n// %s\n%s\n", args[0], a, args[1], b)
			if err := os.WriteFile(out, []byte(merged), 0o644); err != nil {
				return err
			}
			fmt.Printf("combined -> %s\n", out)
			return nil
		},
	}
}

func patternTransformCmd() *cobra.Command {
	var op string
	var out string
	cmd := &cobra.Command{
		Use:   "transform <file.cdc>",
		Short: "Apply a transform (reverse|speed2|speedHalf) and export a WAV",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			combined, cps, err := collectPatternsFromFile(args[0])
			if err != nil {
				return err
			}
			secs := 8.0
			buf, err := audio.RenderPattern(combined, cps, secs)
			if err != nil {
				return err
			}
			switch op {
			case "reverse":
				for i, j := 0, len(buf)-1; i < j; i, j = i+1, j-1 {
					buf[i], buf[j] = buf[j], buf[i]
				}
			case "speed2":
				buf = audio.ResampleAudio(buf, audio.SampleRate, audio.SampleRate*2)
			case "speedHalf":
				buf = audio.ResampleAudio(buf, audio.SampleRate, audio.SampleRate/2)
			default:
				return fmt.Errorf("unknown op %q", op)
			}
			buf = normalizePeak(buf)
			dest := strings.TrimSuffix(args[0], ".cdc") + "_" + op + ".wav"
			if out != "" {
				dest = out
			}
			if err := audio.WriteWAVFile(dest, buf, audio.SampleRate); err != nil {
				return err
			}
			fmt.Printf("transformed (%s) -> %s\n", op, dest)
			return nil
		},
	}
	cmd.Flags().StringVar(&op, "op", "reverse", "transform: reverse|speed2|speedHalf")
	cmd.Flags().StringVar(&out, "out", "", "output wav")
	return cmd
}

func patternArpCmd() *cobra.Command {
	var rate float64
	var dur string
	cmd := &cobra.Command{
		Use:   "arp <notes> [out.cdc]",
		Short: "Generate an arpeggio pattern from space-separated notes",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			notes := strings.Fields(args[0])
			if len(notes) == 0 {
				return fmt.Errorf("provide at least one note")
			}
			out := "arp.cdc"
			if len(args) > 1 {
				out = args[1]
			}
			var b strings.Builder
			b.WriteString("@bpm 120\n\n")
			for i, n := range notes {
				b.WriteString(fmt.Sprintf("note(\"%s\").slow(%s).out()\n", n, dur))
				_ = i
			}
			_ = rate
			if err := os.WriteFile(out, []byte(b.String()), 0o644); err != nil {
				return err
			}
			fmt.Printf("arpeggio (%d notes) -> %s\n", len(notes), out)
			return nil
		},
	}
	cmd.Flags().Float64Var(&rate, "rate", 1, "note rate")
	cmd.Flags().StringVar(&dur, "dur", "1", "per-note duration multiplier")
	return cmd
}

func patternExportCmd() *cobra.Command {
	var secs float64
	cmd := &cobra.Command{
		Use:   "export <file.cdc> [out.wav]",
		Short: "Render a pattern to WAV",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			combined, cps, err := collectPatternsFromFile(args[0])
			if err != nil {
				return err
			}
			if secs <= 0 {
				secs = 8
			}
			buf, err := audio.RenderPattern(combined, cps, secs)
			if err != nil {
				return err
			}
			buf = normalizePeak(buf)
			out := strings.TrimSuffix(args[0], ".cdc") + ".wav"
			if len(args) > 1 {
				out = args[1]
			}
			if err := audio.WriteWAVFile(out, buf, audio.SampleRate); err != nil {
				return err
			}
			fmt.Printf("exported -> %s\n", out)
			return nil
		},
	}
	cmd.Flags().Float64Var(&secs, "sec", 8, "length in seconds")
	return cmd
}

func patternValidateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "validate <file.cdc>",
		Short: "Check that a pattern parses and produces audio",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			combined, _, err := collectPatternsFromFile(args[0])
			if err != nil {
				return fmt.Errorf("INVALID: %w", err)
			}
			if len(combined.FirstCycle()) == 0 {
				return fmt.Errorf("INVALID: pattern produces no events")
			}
			fmt.Printf("VALID: %s (%d events in first cycle)\n", args[0], len(combined.FirstCycle()))
			return nil
		},
	}
}

func collectPatternsFromFile(path string) (pattern.Pattern, float64, error) {
	code, err := os.ReadFile(path)
	if err != nil {
		return pattern.Silence(), 0, err
	}
	return collectPatterns(string(code))
}

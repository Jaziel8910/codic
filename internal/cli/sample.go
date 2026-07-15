package cli

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/Jaziel8910/codic/internal/audio"
)

func sampleCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sample",
		Short: "Inspect and manage the sample library",
	}
	cmd.AddCommand(sampleListCmd())
	cmd.AddCommand(sampleSearchCmd())
	cmd.AddCommand(sampleInfoCmd())
	cmd.AddCommand(sampleAddCmd())
	cmd.AddCommand(sampleRemoveCmd())
	cmd.AddCommand(sampleNormalizeCmd())
	cmd.AddCommand(sampleLoopCmd())
	cmd.AddCommand(samplePitchCmd())
	cmd.AddCommand(sampleReverseCmd())
	cmd.AddCommand(sampleExportCmd())
	return cmd
}

func ensureSamples() error {
	if audio.SampleRoot() == "" {
		_ = audio.InitSamples()
	}
	return nil
}

func sampleListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all samples",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := ensureSamples(); err != nil {
				return err
			}
			names := audio.SampleNames()
			if len(names) == 0 {
				fmt.Println("(no samples indexed - run `codic doctor` to locate the library)")
				return nil
			}
			sort.Strings(names)
			for _, n := range names {
				fmt.Printf("  %-14s x%d\n", n, audio.SampleCount(n))
			}
			return nil
		},
	}
}

func sampleSearchCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "search <term>",
		Short: "Search samples by name",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := ensureSamples(); err != nil {
				return err
			}
			term := strings.ToLower(args[0])
			names := audio.SampleNames()
			sort.Strings(names)
			found := 0
			for _, n := range names {
				if strings.Contains(n, term) {
					fmt.Printf("  %-14s x%d\n", n, audio.SampleCount(n))
					found++
				}
			}
			if found == 0 {
				fmt.Printf("no samples match %q\n", args[0])
			}
			return nil
		},
	}
}

func sampleInfoCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "info <name>",
		Short: "Show sample details (rate, duration, variants)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := ensureSamples(); err != nil {
				return err
			}
			name, n := audio.NormalizeSampleName(args[0])
			path, ok := audio.SamplePath(name, n)
			if !ok {
				return fmt.Errorf("sample %q not found", name)
			}
			samples, rate, err := audio.DecodeWAVFile(path)
			if err != nil {
				return err
			}
			dur := float64(len(samples)) / float64(rate)
			fmt.Printf("name    : %s\n", name)
			fmt.Printf("variant : %d\n", n)
			fmt.Printf("file    : %s\n", path)
			fmt.Printf("rate    : %d Hz\n", rate)
			fmt.Printf("frames  : %d\n", len(samples))
			fmt.Printf("duration: %.3f s\n", dur)
			fmt.Printf("variants: %d\n", audio.SampleCount(name))
			return nil
		},
	}
}

func sampleAddCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "add <file.wav> [name]",
		Short: "Add a WAV file to the sample library",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := ensureSamples(); err != nil {
				return err
			}
			root := audio.SampleRoot()
			if root == "" {
				return fmt.Errorf("no sample root configured")
			}
			name := strings.TrimSuffix(filepath.Base(args[0]), filepath.Ext(args[0]))
			if len(args) > 1 {
				name = args[1]
			}
			dest := filepath.Join(root, "user", name+filepath.Ext(args[0]))
			if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
				return err
			}
			data, err := os.ReadFile(args[0])
			if err != nil {
				return err
			}
			if err := os.WriteFile(dest, data, 0o644); err != nil {
				return err
			}
			fmt.Printf("added sample %q -> %s\n", name, dest)
			return nil
		},
	}
}

func sampleRemoveCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "remove <name>",
		Short: "Remove a sample from the library",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := ensureSamples(); err != nil {
				return err
			}
			name, n := audio.NormalizeSampleName(args[0])
			path, ok := audio.SamplePath(name, n)
			if !ok {
				return fmt.Errorf("sample %q not found", name)
			}
			if err := os.Remove(path); err != nil {
				return err
			}
			fmt.Printf("removed %s\n", path)
			return nil
		},
	}
}

func decodeSampleArg(arg string) ([]float64, int, string, error) {
	name, n := audio.NormalizeSampleName(arg)
	path, ok := audio.SamplePath(name, n)
	if !ok {
		if _, err := os.Stat(arg); err == nil {
			s, r, e := audio.DecodeWAVFile(arg)
			return s, r, arg, e
		}
		return nil, 0, "", fmt.Errorf("sample %q not found", name)
	}
	s, r, err := audio.DecodeWAVFile(path)
	return s, r, path, err
}

func sampleNormalizeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "normalize <name> [out.wav]",
		Short: "Peak-normalize a sample and export it",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			samples, rate, _, err := decodeSampleArg(args[0])
			if err != nil {
				return err
			}
			samples = normalizePeak(samples)
			out := args[0] + "_norm.wav"
			if len(args) > 1 {
				out = args[1]
			}
			if err := audio.WriteWAVFile(out, samples, rate); err != nil {
				return err
			}
			fmt.Printf("normalized -> %s\n", out)
			return nil
		},
	}
}

func sampleLoopCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "loop <name> [out.wav]",
		Short: "Export a sample ready for seamless looping",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			samples, rate, _, err := decodeSampleArg(args[0])
			if err != nil {
				return err
			}
			xf := rate / 20
			if xf > len(samples)/2 {
				xf = len(samples) / 2
			}
			for i := 0; i < xf; i++ {
				a := float64(i) / float64(xf)
				samples[i] = samples[i]*a + samples[len(samples)-xf+i]*(1-a)
			}
			samples = samples[:len(samples)-xf]
			out := args[0] + "_loop.wav"
			if len(args) > 1 {
				out = args[1]
			}
			if err := audio.WriteWAVFile(out, samples, rate); err != nil {
				return err
			}
			fmt.Printf("looped -> %s\n", out)
			return nil
		},
	}
}

func samplePitchCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "pitch <name> <semitones> [out.wav]",
		Short: "Transpose a sample by N semitones",
		Args:  cobra.RangeArgs(2, 3),
		RunE: func(cmd *cobra.Command, args []string) error {
			semi, err := parseFloat(args[1])
			if err != nil {
				return fmt.Errorf("semitones must be a number")
			}
			samples, rate, _, err := decodeSampleArg(args[0])
			if err != nil {
				return err
			}
			factor := 1.0
			if semi != 0 {
				factor = 1.0 / math.Exp2(semi/12.0)
			}
			newRate := int(float64(rate) * factor)
			out := args[0] + "_pitch.wav"
			if len(args) > 2 {
				out = args[2]
			}
			if err := audio.WriteWAVFile(out, samples, newRate); err != nil {
				return err
			}
			fmt.Printf("pitched %+.1f semitones -> %s\n", semi, out)
			return nil
		},
	}
}

func sampleReverseCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "reverse <name> [out.wav]",
		Short: "Reverse a sample",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			samples, rate, _, err := decodeSampleArg(args[0])
			if err != nil {
				return err
			}
			for i, j := 0, len(samples)-1; i < j; i, j = i+1, j-1 {
				samples[i], samples[j] = samples[j], samples[i]
			}
			out := args[0] + "_rev.wav"
			if len(args) > 1 {
				out = args[1]
			}
			if err := audio.WriteWAVFile(out, samples, rate); err != nil {
				return err
			}
			fmt.Printf("reversed -> %s\n", out)
			return nil
		},
	}
}

func sampleExportCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "export <name> [out.wav]",
		Short: "Export a sample to a standalone WAV",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			samples, rate, _, err := decodeSampleArg(args[0])
			if err != nil {
				return err
			}
			out := args[0] + ".wav"
			if len(args) > 1 {
				out = args[1]
			}
			if err := audio.WriteWAVFile(out, samples, rate); err != nil {
				return err
			}
			fmt.Printf("exported -> %s\n", out)
			return nil
		},
	}
}

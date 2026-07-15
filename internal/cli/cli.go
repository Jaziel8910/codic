package cli

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"runtime"

	"strings"

	"github.com/spf13/viper"

	"github.com/Jaziel8910/codic/internal/audio"
	"github.com/Jaziel8910/codic/internal/codang"
	"github.com/Jaziel8910/codic/internal/pattern"
)

// Version is set at build time via -ldflags.
var Version = "v2.0.0-dev"

// CodicDir returns the global CODIC workspace directory, created inside the
// user's home folder. This is where albums, projects, sounds, docs and
// backups live.
func CodicDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "CODIC"
	}
	return filepath.Join(home, "CODIC")
}

// ensureCodicDir creates ~/.codic if it does not exist.
func ensureCodicDir() error {
	return os.MkdirAll(CodicDir(), 0o755)
}

// collectPatterns parses and evaluates Codang code, returning the combined
// pattern (stacked from every .out() call) and the tempo in cps.
func collectPatterns(code string) (pattern.Pattern, float64, error) {
	prog, err := codang.Parse(code)
	if err != nil {
		return pattern.Silence(), 0, fmt.Errorf("parse error: %w", err)
	}
	var all []pattern.Pattern
	eval := codang.NewEvaluator(func(p pattern.Pattern) {
		all = append(all, p)
	})
	if err := eval.Eval(prog); err != nil {
		return pattern.Silence(), 0, fmt.Errorf("eval error: %w", err)
	}
	if len(all) == 0 {
		return pattern.Silence(), 0, fmt.Errorf("no pattern produced (did you call .out()?)")
	}
	combined := all[0]
	for _, p := range all[1:] {
		combined = pattern.Stack(combined, p)
	}
	cps := eval.GetCPS()
	if cps <= 0 {
		cps = 1.0
	}
	return combined, cps, nil
}

// renderToFile renders Codang code from a file to a WAV path.
func renderToFile(inPath, outPath string, seconds float64, normalize bool) error {
	code, err := os.ReadFile(inPath)
	if err != nil {
		return fmt.Errorf("reading %s: %w", inPath, err)
	}
	combined, cps, err := collectPatterns(string(code))
	// Validate the song: the declared @type must be met (a real, complete song).
	if prog, perr := codang.Parse(string(code)); perr == nil {
		errs, warns := codang.ValidateSong(prog)
		if len(errs) > 0 {
			return fmt.Errorf("cancion invalida (%s):\n - %s", filepath.Base(inPath), strings.Join(errs, "\n - "))
		}
		for _, w := range warns {
			fmt.Fprintf(os.Stderr, "aviso: %s\n", w)
		}
		// Derive length from @cycles + tempo when no explicit -d flag was given.
		if seconds <= 0 {
			if d := codang.SongDuration(prog, 0); d > 0 {
				seconds = d
			}
		}
	}
	if err != nil {
		return err
	}
	if seconds <= 0 {
		if d := viper.GetFloat64("default_duration"); d > 0 {
			seconds = d
		} else {
			seconds = 8
		}
	}
	buf, err := audio.RenderPattern(combined, cps, seconds)
	if err != nil {
		return fmt.Errorf("render error: %w", err)
	}
	if normalize {
		buf = normalizePeak(buf)
	}
	if err := audio.WriteWAVFile(outPath, buf, audio.SampleRate); err != nil {
		return fmt.Errorf("write error: %w", err)
	}
	return nil
}

// normalizePeak applies peak normalization so the loudest sample hits ~0.9.
func normalizePeak(samples []float64) []float64 {
	var peak float64
	for _, s := range samples {
		if a := math.Abs(s); a > peak {
			peak = a
		}
	}
	if peak < 1e-6 {
		return samples
	}
	gain := 0.9 / peak
	out := make([]float64, len(samples))
	for i, s := range samples {
		out[i] = s * gain
	}
	return out
}

// openPlayer opens a rendered audio file with the operating system's default
// player.
func openPlayer(path string) error {
	var cmd string
	var args []string
	switch runtime.GOOS {
	case "windows":
		cmd = "rundll32"
		args = []string{"url.dll,FileProtocolHandler", path}
	case "darwin":
		cmd = "open"
		args = []string{path}
	default:
		cmd = "xdg-open"
		args = []string{path}
	}
	return runCommand(cmd, args...)
}

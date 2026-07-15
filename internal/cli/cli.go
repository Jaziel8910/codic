package cli

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"runtime"

	"github.com/Jaziel8910/codic/internal/audio"
	"github.com/Jaziel8910/codic/internal/codang"
	"github.com/Jaziel8910/codic/internal/pattern"
)

// Version is set at build time via -ldflags.
var Version = "v2.0.0-dev"

// CodicDir returns the user-level config directory (~/.codic).
func CodicDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ".codic"
	}
	return filepath.Join(home, ".codic")
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
	if err != nil {
		return err
	}
	if seconds <= 0 {
		seconds = 8
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

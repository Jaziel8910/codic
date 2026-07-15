package cli

import (
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

// runCommand runs a command detached (so the parent CLI can exit).
func runCommand(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdin = nil
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Start()
}

// tempWAV returns a path inside ~/.codic/tmp for a transient render.
func tempWAV(name string) (string, error) {
	dir := filepath.Join(CodicDir(), "tmp")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	if name == "" {
		name = "preview.wav"
	}
	return filepath.Join(dir, name), nil
}

// newRand returns a deterministic-enough source for shuffle operations.
func newRand() *rand.Rand {
	return rand.New(rand.NewSource(time.Now().UnixNano()))
}

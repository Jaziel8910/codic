// Package dist embeds the small distribution assets (stdlib, templates,
// examples) and knows how to scaffold the user-level ~/.codic
// directory. The (large) sample bank is downloaded on demand by the
// CLI `codic install` command, never embedded.
package dist

import (
	"embed"
	"io/fs"
	"os"
	"path/filepath"
)

//go:embed all:data
var dataFS embed.FS

// Assets returns the embedded distribution filesystem.
func Assets() fs.FS { return dataFS }

// Install scaffolds the user directory ~/.codic with the embedded
// stdlib, templates and examples, and creates the runtime folders
// (instruments, dj, packages, samples, out). It does NOT fetch the
// sample bank — that is done by the CLI install command so it can show
// progress and tolerate being offline.
func Install(home string) error {
	if err := os.MkdirAll(home, 0o755); err != nil {
		return err
	}
	dirs := []struct{ from, to string }{
		{"data/stdlib", "stdlib"},
		{"data/templates", "templates"},
		{"data/examples", "examples"},
	}
	for _, d := range dirs {
		if err := copyTree(d.from, filepath.Join(home, d.to)); err != nil {
			return err
		}
	}
	for _, sub := range []string{"instruments", "dj", "packages", "samples", "out"} {
		if err := os.MkdirAll(filepath.Join(home, sub), 0o755); err != nil {
			return err
		}
	}
	return nil
}

// copyTree mirrors an embedded directory (rel, rooted at "data/...") into dst.
func copyTree(rel, dst string) error {
	return fs.WalkDir(dataFS, rel, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		relp, _ := filepath.Rel(rel, p)
		target := filepath.Join(dst, relp)
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		b, err := dataFS.ReadFile(p)
		if err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		return os.WriteFile(target, b, 0o644)
	})
}

// Manifest returns the embedded manifest.json bytes.
func Manifest() ([]byte, error) { return dataFS.ReadFile("data/manifest.json") }

package cli

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"

	"github.com/Jaziel8910/codic/internal/dist"
)

// type manifestJSON mirrors dist/data/manifest.json.
type manifestJSON struct {
	Version     string `json:"version"`
	SamplesURL string `json:"samples_url"`
}

// installCmd scaffolds the full user environment: config, stdlib,
// templates, example modules and (optionally) the sample bank.
func installCmd() *cobra.Command {
	var withSamples bool
	cmd := &cobra.Command{
		Use:   "install",
		Short: "Install Codic: config, stdlib, templates and sample bank",
		Long: `Scaffolds ~/.codic with the standard library, project templates,
example modules and (with --samples) the drum/synth sample bank.

Run this once after first download, or any time you want to
re-sync your local environment.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			home := CodicDir()
			fmt.Printf("installing Codic into %s\n", home)

			if err := dist.Install(home); err != nil {
				return fmt.Errorf("scaffold failed: %w", err)
			}
			fmt.Println("  - stdlib, templates, examples")

			p, cerr := ensureConfigFile()
			if cerr != nil {
				return cerr
			}
			if _, serr := os.Stat(p); serr != nil {
				defaults := `# Codic configuration
default_duration: 8
sample_rate: 44100
output_dir: ` + filepath.Join(home, "out") + `
samples_dir: ` + filepath.Join(home, "samples") + `
`
				if werr := os.WriteFile(p, []byte(defaults), 0o644); werr != nil {
					return werr
				}
			}
			fmt.Println("  - config.yaml")

			if withSamples {
				if err := installSamples(home); err != nil {
					fmt.Fprintf(os.Stderr, "  ! samples: %v\n", err)
					fmt.Fprintln(os.Stderr, "    you can run `codic install --samples` later (needs internet)")
				} else {
					fmt.Println("  - sample bank")
				}
			} else {
				fmt.Println("  . skip sample bank (use `codic install --samples`)")
			}

			fmt.Printf("\ndone. try:  codic play %s\n",
				filepath.Join(home, "examples", "basic.cdc"))
			return nil
		},
	}
	cmd.Flags().BoolVar(&withSamples, "samples", false, "also download the sample bank (needs internet)")
	return cmd
}

// installSamples downloads the sample bank from the release and unzips it
// into ~/.codic/samples. Tolerates being offline (returns an error
// that the caller prints without aborting the rest of the install).
func installSamples(home string) error {
	dest := filepath.Join(home, "samples")
	if _, err := os.Stat(filepath.Join(dest, "strudel.json")); err == nil {
		return nil // already present
	}

	url, err := sampleURL()
	if err != nil {
		return err
	}

	fmt.Printf("  > downloading sample bank from %s\n", url)
	tmp := filepath.Join(home, "samples.zip")
	if err := downloadFile(url, tmp); err != nil {
		return err
	}
	defer os.Remove(tmp)

	if err := unzip(tmp, dest); err != nil {
		return err
	}
	return nil
}

// sampleURL reads the embedded manifest to find the release asset URL.
func sampleURL() (string, error) {
	b, err := dist.Manifest()
	if err != nil {
		return "", err
	}
	var m manifestJSON
	if err := json.Unmarshal(b, &m); err != nil {
		return "", err
	}
	if m.SamplesURL == "" {
		return "", fmt.Errorf("no samples_url in manifest")
	}
	return m.SamplesURL, nil
}

// downloadFile streams url into dst, reporting rough progress.
func downloadFile(url, dst string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download returned %s", resp.Status)
	}

	f, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer f.Close()

	buf := make([]byte, 1<<20)
	var written int64
	last := time.Now()
	for {
		n, err := resp.Body.Read(buf)
		if n > 0 {
			if _, werr := f.Write(buf[:n]); werr != nil {
				return werr
			}
			written += int64(n)
			if time.Since(last) > 500*time.Millisecond {
				fmt.Printf("    %d MB\n", written>>20)
				last = time.Now()
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
	}
	fmt.Printf("    %d MB downloaded\n", written>>20)
	return nil
}

// unzip extracts every entry of zipPath into destDir.
func unzip(zipPath, destDir string) error {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer r.Close()

	if err := os.MkdirAll(destDir, 0o755); err != nil {
		return err
	}
	for _, zf := range r.File {
		target := filepath.Join(destDir, zf.Name)
		if !within(destDir, target) {
			continue
		}
		if zf.FileInfo().IsDir() {
			if err := os.MkdirAll(target, 0o755); err != nil {
				return err
			}
			continue
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		rc, err := zf.Open()
		if err != nil {
			return err
		}
		out, err := os.OpenFile(target, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o644)
		if err != nil {
			rc.Close()
			return err
		}
		if _, err := io.Copy(out, rc); err != nil {
			out.Close()
			rc.Close()
			return err
		}
		out.Close()
		rc.Close()
	}
	return nil
}

// within reports whether target is inside base (zip-slip protection).
func within(base, target string) bool {
	rel, err := filepath.Rel(base, target)
	if err != nil {
		return false
	}
	return rel != ".." && !filepath.HasPrefix(rel, ".."+string(filepath.Separator))
}

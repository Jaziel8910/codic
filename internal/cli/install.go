package cli

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/Jaziel8910/codic/internal/dist"
)

// type manifestJSON mirrors dist/data/manifest.json.
type manifestJSON struct {
	Version    string `json:"version"`
	SamplesURL string `json:"samples_url"`
	Home       string `json:"home"`
}

// installCmd scaffolds the full CODIC workspace: config, stdlib, templates,
// examples, documentation, the QUICKSTART.docx guide and (optionally) the
// sample bank. It also backs up any pre-existing workspace first.
func installCmd() *cobra.Command {
	var withSamples bool
	cmd := &cobra.Command{
		Use:   "install",
		Short: "Install Codic: workspace, config, stdlib, templates and sample bank",
		Long: `Scaffolds the CODIC global workspace with the standard library,
project templates, example modules, documentation and (with --samples) the
drum/synth sample bank.

Run this once after first download, or any time you want to re-sync your
local environment.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			home := CodicDir()
			fmt.Printf("installing Codic into %s\n", home)

			if _, err := os.Stat(home); err == nil {
				fmt.Println("  . existing workspace found — backing up first")
				if bp, berr := createBackup(home); berr == nil {
					fmt.Printf("    backup: %s\n", bp)
				} else {
					fmt.Fprintf(os.Stderr, "    ! backup failed: %v\n", berr)
				}
			}

			if err := dist.Install(home); err != nil {
				return fmt.Errorf("scaffold failed: %w", err)
			}
			fmt.Println("  - stdlib, templates, examples, docs")

			p, cerr := ensureConfigFile()
			if cerr != nil {
				return cerr
			}
			if _, serr := os.Stat(p); serr != nil {
				defaults := `# Codic configuration
default_duration: 8
sample_rate: 44100
output_dir: ` + filepath.Join(home, "out") + `
sounds_dir: ` + filepath.Join(home, "sounds") + `
`
				if werr := os.WriteFile(p, []byte(defaults), 0o644); werr != nil {
					return werr
				}
			}
			fmt.Println("  - config.yaml")

			if err := createQuickstartDocx(filepath.Join(home, "QUICKSTART.docx")); err != nil {
				fmt.Fprintf(os.Stderr, "  ! quickstart: %v\n", err)
			} else {
				fmt.Println("  - QUICKSTART.docx")
			}

			if err := installSelf(home); err != nil {
				fmt.Fprintf(os.Stderr, "  ! self-install: %v\n", err)
			} else {
				fmt.Println("  - codic command in CODIC/bin")
			}
			added, perr := addToPath(filepath.Join(home, "bin"))
			if perr != nil {
				fmt.Fprintf(os.Stderr, "  ! could not add to PATH automatically: %v\n", perr)
				fmt.Fprintln(os.Stderr, "    add CODIC\\bin to PATH to run `codic` from anywhere")
			} else if added {
				fmt.Println("  - added CODIC/bin to PATH (open a new terminal to use `codic`)")
			} else {
				fmt.Println("  . CODIC/bin already on PATH")
			}

			if err := createDesktopShortcut(home); err != nil {
				fmt.Fprintf(os.Stderr, "  ! shortcut: %v\n", err)
			} else {
				fmt.Println("  - acceso directo 'CODIC' en el Escritorio")
			}

			if withSamples {
				if err := installSamples(home); err != nil {
					fmt.Fprintf(os.Stderr, "  ! samples: %v\n", err)
					fmt.Fprintln(os.Stderr, "    you can run `codic install --samples` later (needs internet)")
				} else {
					fmt.Println("  - sample bank (sounds/)")
				}
			} else {
				fmt.Println("  . skip sample bank (use `codic install --samples`)")
			}

			fmt.Printf("\ndone. your workspace is at: %s\n", home)
			fmt.Printf("try:  codic play %s\n", filepath.Join(home, "examples", "basic.cdc"))
			return nil
		},
	}
	cmd.Flags().BoolVar(&withSamples, "samples", false, "also download the sample bank (needs internet)")
	return cmd
}

// backupCmd zips the CODIC workspace (minus backups/ and out/) into backups/.
func backupCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "backup",
		Short: "Back up the CODIC workspace to CODIC/backups",
		RunE: func(cmd *cobra.Command, args []string) error {
			home := CodicDir()
			if _, err := os.Stat(home); err != nil {
				return fmt.Errorf("CODIC workspace not found at %s (run `codic install`)", home)
			}
			path, err := createBackup(home)
			if err != nil {
				return err
			}
			fmt.Printf("backup created: %s\n", path)
			return nil
		},
	}
}

// createBackup writes a timestamped zip of home into home/backups, skipping
// the backups and out directories (to avoid recursion and re-saving renders).
func createBackup(home string) (string, error) {
	if err := os.MkdirAll(filepath.Join(home, "backups"), 0o755); err != nil {
		return "", err
	}
	stamp := time.Now().Format("20060102-150405")
	outPath := filepath.Join(home, "backups", "CODIC-"+stamp+".zip")
	f, err := os.Create(outPath)
	if err != nil {
		return "", err
	}
	defer f.Close()
	zw := zip.NewWriter(f)
	defer zw.Close()

	base := home
	err = filepath.Walk(base, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(base, p)
		if err != nil {
			return err
		}
		if rel == "." {
			return nil
		}
		if rel == "backups" && info.IsDir() {
			return filepath.SkipDir
		}
		if rel == "out" && info.IsDir() {
			return filepath.SkipDir
		}
		hdr, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}
		hdr.Name = filepath.ToSlash(rel)
		if info.IsDir() {
			hdr.Name += "/"
		}
		w, err := zw.Create(hdr.Name)
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		in, err := os.Open(p)
		if err != nil {
			return err
		}
		defer in.Close()
		_, err = io.Copy(w, in)
		return err
	})
	if err != nil {
		return "", err
	}
	return outPath, nil
}

// installSamples downloads the sample bank from the release and unzips it
// into CODIC/sounds. Tolerates being offline (returns an error that the
// caller prints without aborting the rest of the install).
func installSamples(home string) error {
	dest := filepath.Join(home, "sounds")
	if _, err := os.Stat(filepath.Join(dest, "strudel.json")); err == nil {
		return nil // already present
	}

	url, err := sampleURL()
	if err != nil {
		return err
	}

	fmt.Printf("  > downloading sample bank from %s\n", url)
	tmp := filepath.Join(home, "sounds.zip")
	if err := downloadFile(url, tmp); err != nil {
		return err
	}
	defer os.Remove(tmp)

	if err := unzip(tmp, dest); err != nil {
		return err
	}
	// The release zip may contain a top-level "samples/" folder; flatten it
	// so strudel.json ends up directly under dest.
	if err := flattenSamplesDir(dest); err != nil {
		return err
	}
	return nil
}

// flattenSamplesDir moves the contents of a nested "samples/" subfolder up
// into dest when the archive was published with a redundant top directory.
func flattenSamplesDir(dest string) error {
	nested := filepath.Join(dest, "samples")
	if _, err := os.Stat(filepath.Join(nested, "strudel.json")); err != nil {
		return nil // not nested — already in the right place
	}
	entries, err := os.ReadDir(nested)
	if err != nil {
		return err
	}
	for _, e := range entries {
		src := filepath.Join(nested, e.Name())
		dst := filepath.Join(dest, e.Name())
		if _, err := os.Stat(dst); err == nil {
			continue // keep existing file at dest
		}
		if err := os.Rename(src, dst); err != nil {
			return err
		}
	}
	return os.RemoveAll(nested)
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

// createQuickstartDocx writes a human-friendly .docx guide into the
// workspace (only if it does not already exist, so user edits are kept).
func createQuickstartDocx(path string) error {
	if _, err := os.Stat(path); err == nil {
		return nil
	}
	home := CodicDir()
	paras := []string{
		"Codic - Guia rapida para humanos",
		"",
		"Codic es un estudio de musica que controlas escribiendo texto (lenguaje Codang, archivos .cdc).",
		"Tu carpeta de trabajo global se llama CODIC y vive en tu carpeta de usuario.",
		"",
		"Como empezar:",
		"1) Abre la terminal (Boton Inicio -> escribe 'PowerShell' -> Enter).",
		"2) Escribe:  codic install --samples   (descarga los sonidos una vez, necesita internet).",
		"3) Prueba:   codic play \"" + home + "\\examples\\basic.cdc\"",
		"",
		"Comandos utiles:",
		"  codic new track <nombre>    crea una cancion nueva",
		"  codic play <archivo.cdc>    genera audio y lo reproduce",
		"  codic render <archivo.cdc>  guarda el audio en un archivo .wav",
		"  codic backup                hace una copia de seguridad de CODIC",
		"",
		"Para que una IA te ayude con tu musica: dile que lea AGENTS.md y COMMANDS.md de esta carpeta.",
		"La documentacion completa esta en la carpeta docs/.",
	}
	b, err := makeDocxZip(paras)
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0o644)
}

// makeDocxZip builds a minimal but valid .docx (Office Open XML) from a list
// of paragraphs and returns the file bytes.
func makeDocxZip(paras []string) ([]byte, error) {
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	write := func(name, content string) error {
		w, err := zw.Create(name)
		if err != nil {
			return err
		}
		_, err = io.WriteString(w, content)
		return err
	}

	contentTypes := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">
<Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/>
<Default Extension="xml" ContentType="application/xml"/>
<Override PartName="/word/document.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.document.main+xml"/>
</Types>`
	if err := write("[Content_Types].xml", contentTypes); err != nil {
		return nil, err
	}

	rels := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
<Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="word/document.xml"/>
</Relationships>`
	if err := write("_rels/.rels", rels); err != nil {
		return nil, err
	}

	var body strings.Builder
	body.WriteString(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>`)
	body.WriteString(`<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main"><w:body>`)
	for _, p := range paras {
		body.WriteString(`<w:p><w:r><w:t xml:space="preserve">`)
		body.WriteString(escapeXML(p))
		body.WriteString(`</w:t></w:r></w:p>`)
	}
	body.WriteString(`</w:body></w:document>`)
	if err := write("word/document.xml", body.String()); err != nil {
		return nil, err
	}

	if err := zw.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func escapeXML(s string) string {
	return strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;").Replace(s)
}

// installSelf copies the running executable into CODIC/bin so the user gets a
// stable `codic` command. It is idempotent (no-op if already self).
func installSelf(home string) error {
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	binDir := filepath.Join(home, "bin")
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		return err
	}
	name := "codic"
	if runtime.GOOS == "windows" {
		name = "codic.exe"
	}
	dst := filepath.Join(binDir, name)
	if filepath.Clean(exe) == filepath.Clean(dst) {
		return nil
	}
	return copyFile(exe, dst)
}

// copyFile copies src to dst, preserving the executable bit on Unix.
func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	if runtime.GOOS != "windows" {
		_ = out.Chmod(0o755)
	}
	return nil
}

// addToPath adds dir to the user's PATH (Windows) so `codic` is available
// from any terminal. It edits only the USER portion (via .NET, which does not
// truncate at 1024 chars like setx.exe does) and is idempotent.
// Returns whether it was added (false if already present or unsupported here).
func addToPath(dir string) (bool, error) {
	if runtime.GOOS != "windows" {
		return false, nil
	}
	if strings.Contains(os.Getenv("PATH"), dir) {
		return false, nil
	}
	ps := fmt.Sprintf(
		`$p=[Environment]::GetEnvironmentVariable('PATH','User'); if($p -notlike '*%s*'){ [Environment]::SetEnvironmentVariable('PATH', ($p.TrimEnd(';')+';%s'), 'User') }`,
		dir, dir)
	cmd := exec.Command("powershell", "-NoProfile", "-Command", ps)
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return false, err
	}
	// Refresh this process's PATH so later steps see it.
	os.Setenv("PATH", os.Getenv("PATH")+string(os.PathListSeparator)+dir)
	return true, nil
}

// createDesktopShortcut puts a "CODIC.lnk" on the user's Desktop pointing at
// the workspace, so it is easy to find. Best-effort; never fatal.
func createDesktopShortcut(home string) error {
	if runtime.GOOS != "windows" {
		return nil
	}
	profile := os.Getenv("USERPROFILE")
	if profile == "" {
		return nil
	}
	desktop := filepath.Join(profile, "Desktop")
	if _, err := os.Stat(desktop); err != nil {
		return nil
	}
	lnk := filepath.Join(desktop, "CODIC.lnk")
	if _, err := os.Stat(lnk); err == nil {
		return nil // already there
	}
	ps := fmt.Sprintf(
		`$ws=New-Object -ComObject WScript.Shell; $s=$ws.CreateShortcut('%s'); $s.TargetPath='%s'; $s.Description='Codic workspace'; $s.Save()`,
		lnk, home)
	cmd := exec.Command("powershell", "-NoProfile", "-Command", ps)
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

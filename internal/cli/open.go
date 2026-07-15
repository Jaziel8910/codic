package cli

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/spf13/cobra"
)

// openCmd opens the CODIC workspace (or a part of it) in the OS file manager.
func openCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "open [docs|sounds|finals|projects|backups|bin|quickstart|<ruta>]",
		Short: "Open the CODIC workspace in your file manager",
		Long: `Opens the CODIC global workspace in the operating system's file
manager so you can browse your music, docs and sounds with a double click.

With no argument it opens the CODIC folder. Named shortcuts:
  docs  sounds  finals  projects  backups  bin  quickstart`,
		RunE: func(cmd *cobra.Command, args []string) error {
			home := CodicDir()
			target := home
			if len(args) > 0 {
				switch args[0] {
				case "docs":
					target = filepath.Join(home, "docs")
				case "sounds", "samples":
					target = filepath.Join(home, "sounds")
				case "finals":
					target = filepath.Join(home, "finals")
				case "projects":
					target = filepath.Join(home, "projects")
				case "backups":
					target = filepath.Join(home, "backups")
				case "bin":
					target = filepath.Join(home, "bin")
				case "quickstart":
					target = filepath.Join(home, "QUICKSTART.docx")
				default:
					target = filepath.Join(home, args[0])
				}
			}
			if _, err := os.Stat(target); err != nil {
				return fmt.Errorf("no se encuentra: %s", target)
			}
			if err := openInFileManager(target); err != nil {
				return err
			}
			fmt.Printf("abierto: %s\n", target)
			return nil
		},
	}
	return cmd
}

// openInFileManager opens path with the platform's default file manager.
func openInFileManager(path string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("explorer", path)
	case "darwin":
		cmd = exec.Command("open", path)
	default:
		cmd = exec.Command("xdg-open", path)
	}
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	// Start (don't wait): explorer/xdg-open return immediately.
	return cmd.Start()
}

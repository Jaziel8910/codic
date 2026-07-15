package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/Jaziel8910/codic/internal/audio"
)

// Package is a manifest for a shareable Codic package.
type Package struct {
	Name         string            `json:"name"`
	Version      string            `json:"version"`
	Description  string            `json:"description"`
	Dependencies map[string]string `json:"dependencies"`
}

func pkgDir() string {
	root := CodicDir()
	d := filepath.Join(root, "packages")
	os.MkdirAll(d, 0o755)
	return d
}

func pkgManifestPath(name string) string {
	return filepath.Join(pkgDir(), name, "package.json")
}

func pkgLoad(name string) (*Package, error) {
	p := pkgManifestPath(name)
	data, err := os.ReadFile(p)
	if err != nil {
		return nil, errNotFound("package", name)
	}
	var pkg Package
	if err := json.Unmarshal(data, &pkg); err != nil {
		return nil, err
	}
	return &pkg, nil
}

func pkgSave(pkg *Package) error {
	dir := filepath.Join(pkgDir(), pkg.Name)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(pkg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, "package.json"), data, 0o644)
}

func pkgCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pkg",
		Short: "Package manager for Codic patterns and libraries",
	}
	cmd.AddCommand(pkgInitCmd())
	cmd.AddCommand(pkgAddCmd())
	cmd.AddCommand(pkgRemoveCmd())
	cmd.AddCommand(pkgListCmd())
	cmd.AddCommand(pkgBuildCmd())
	return cmd
}

func pkgInitCmd() *cobra.Command {
	var version, desc string
	cmd := &cobra.Command{
		Use:   "init <name>",
		Short: "Create a new package manifest",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			pkg := &Package{
				Name:         args[0],
				Version:      version,
				Description:  desc,
				Dependencies: map[string]string{},
			}
			if pkg.Version == "" {
				pkg.Version = "0.1.0"
			}
			if err := pkgSave(pkg); err != nil {
				return err
			}
			fmt.Printf("created package %q at %s\n", pkg.Name, filepath.Join(pkgDir(), pkg.Name))
			return nil
		},
	}
	cmd.Flags().StringVar(&version, "version", "0.1.0", "initial version")
	cmd.Flags().StringVar(&desc, "desc", "", "description")
	return cmd
}

func pkgAddCmd() *cobra.Command {
	var ver string
	cmd := &cobra.Command{
		Use:   "add <package> <source>",
		Short: "Add a dependency to a package",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			pkg, err := pkgLoad(args[0])
			if err != nil {
				return err
			}
			if ver == "" {
				ver = "latest"
			}
			pkg.Dependencies[args[1]] = ver
			if err := pkgSave(pkg); err != nil {
				return err
			}
			fmt.Printf("added dependency %q (%s) to %s\n", args[1], ver, pkg.Name)
			return nil
		},
	}
	cmd.Flags().StringVar(&ver, "version", "latest", "dependency version")
	return cmd
}

func pkgRemoveCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "remove <package> <dependency>",
		Short: "Remove a dependency",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			pkg, err := pkgLoad(args[0])
			if err != nil {
				return err
			}
			if _, ok := pkg.Dependencies[args[1]]; !ok {
				return fmt.Errorf("dependency %q not found in %s", args[1], pkg.Name)
			}
			delete(pkg.Dependencies, args[1])
			if err := pkgSave(pkg); err != nil {
				return err
			}
			fmt.Printf("removed %q from %s\n", args[1], pkg.Name)
			return nil
		},
	}
}

func pkgListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List installed packages",
		RunE: func(cmd *cobra.Command, args []string) error {
			d := pkgDir()
			ents, err := os.ReadDir(d)
			if err != nil {
				return err
			}
			if len(ents) == 0 {
				fmt.Println("(no packages)")
				return nil
			}
			for _, e := range ents {
				if !e.IsDir() {
					continue
				}
				p, err := pkgLoad(e.Name())
				if err != nil {
					fmt.Printf("  %s (unreadable)\n", e.Name())
					continue
				}
				fmt.Printf("  %s v%s — %s\n", p.Name, p.Version, p.Description)
				if len(p.Dependencies) > 0 {
					fmt.Printf("    deps: %d\n", len(p.Dependencies))
				}
			}
			return nil
		},
	}
}

func pkgBuildCmd() *cobra.Command {
	var out string
	cmd := &cobra.Command{
		Use:   "build <package>",
		Short: "Render all patterns in a package into one WAV",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			pkg, err := pkgLoad(args[0])
			if err != nil {
				return err
			}
			srcDir := filepath.Join(pkgDir(), pkg.Name)
			var all []float64
			count := 0
			err = filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}
				if info.IsDir() || filepath.Ext(path) != ".cdc" {
					return nil
				}
				buf, e := djRender(path, 8)
				if e != nil {
					fmt.Printf("  skip %s: %v\n", info.Name(), e)
					return nil
				}
				all = append(all, buf...)
				for i := 0; i < audio.SampleRate; i++ {
					all = append(all, 0)
				}
				count++
				return nil
			})
			if err != nil {
				return err
			}
			if count == 0 {
				return fmt.Errorf("no .cdc patterns found in package %s", pkg.Name)
			}
			if out == "" {
				out = pkg.Name + "-build.wav"
			}
			if err := audio.WriteWAVFile(out, all, audio.SampleRate); err != nil {
				return err
			}
			fmt.Printf("built %d patterns of %s -> %s\n", count, pkg.Name, out)
			return nil
		},
	}
	cmd.Flags().StringVar(&out, "out", "", "output wav file")
	return cmd
}

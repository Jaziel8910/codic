package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/Jaziel8910/codic/internal/audio"
)

func doctorCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "doctor",
		Short: "Diagnose the codic installation",
	}
	cmd.AddCommand(&cobra.Command{
		Use:   "check",
		Short: "Run a full diagnostic",
		RunE: func(cmd *cobra.Command, args []string) error {
			ok := true
			if err := ensureCodicDir(); err != nil {
				fmt.Printf("[FAIL] CODIC: %v\n", err)
				ok = false
			} else {
				fmt.Printf("[OK]   CODIC at %s\n", CodicDir())
			}
			if _, err := os.Stat(CodicDir() + string(os.PathSeparator) + "config.yaml"); err != nil {
				fmt.Println("[WARN] no config.yaml (run `codic config init`)")
			} else {
				fmt.Println("[OK]   config.yaml present")
			}
			if err := audio.InitSamples(); err != nil {
				fmt.Printf("[WARN] sample bank unavailable: %v\n", err)
			} else {
				fmt.Println("[OK]   sample bank loaded")
			}
			fmt.Printf("[INFO] engine: offline WAV render (SampleRate=%d)\n", audio.SampleRate)
			if !ok {
				return fmt.Errorf("doctor found problems")
			}
			fmt.Println("all checks passed")
			return nil
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:   "deps",
		Short: "Verify dependencies",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Go modules: OK (built with this toolchain)")
			fmt.Println("audio engine: offline render (no realtime deps)")
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:   "paths",
		Short: "Show directory structure",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("config dir : %s\n", CodicDir())
			fmt.Printf("config file: %s\n", configPathOr(CodicDir()+string(filepath.Separator)+"config.yaml"))
			fmt.Printf("tmp dir    : %s\n", CodicDir()+string(filepath.Separator)+"tmp")
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:   "samples",
		Short: "Check the sample bank",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := audio.InitSamples(); err != nil {
				return fmt.Errorf("samples: %v", err)
			}
			fmt.Println("samples: loaded")
			return nil
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:   "repair",
		Short: "Repair common issues",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := ensureCodicDir(); err != nil {
				return err
			}
			fmt.Println("repaired: ensured CODIC exists")
			return nil
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:   "env",
		Short: "Show relevant environment variables",
		Run: func(cmd *cobra.Command, args []string) {
			for _, k := range []string{"EDITOR", "CODIC_CONFIG", "HOME", "USERPROFILE"} {
				fmt.Printf("%s=%s\n", k, os.Getenv(k))
			}
		},
	})
	return cmd
}

func configPathOr(def string) string {
	if p := configPath(); p != "" {
		return p
	}
	return def
}

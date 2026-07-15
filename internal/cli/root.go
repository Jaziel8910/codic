package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/Jaziel8910/codic/internal/audio"
)

var cfgFile string

// NewRootCmd builds the root cobra command and all subcommands.
func NewRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "codic",
		Short: "Codic — CLI-first music production studio (Codang)",
		Long: `Codic is a CLI-first music production studio.
Write music in the Codang (.cdc) language and render offline to WAV,
or open it with your OS player. No TUI, no realtime engine.`,
		SilenceUsage:  true,
		SilenceErrors: true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// Load config (best-effort; ~/.codic/config.yaml is optional).
			if err := initConfig(); err != nil {
				return err
			}
			// Sample bank (Strudel/Dirt-Samples) is optional; if it
			// is missing we hint at `codic install --samples`.
			if err := audio.InitSamples(); err != nil {
				fmt.Fprintln(os.Stderr,
					"tip: no sample bank found — run `codic install --samples` to get sounds")
			}
			return nil
		},
	}

	root.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default CODIC/config.yaml)")

	root.AddCommand(newCmd())
	root.AddCommand(playCmd())
	root.AddCommand(renderCmd())
	root.AddCommand(evalCmd())
	root.AddCommand(runCmd())
	root.AddCommand(watchCmd())
	root.AddCommand(serveCmd())
	root.AddCommand(versionCmd())
	root.AddCommand(configCmd())
	root.AddCommand(doctorCmd())

	root.AddCommand(projectCmd())
	root.AddCommand(instrumentCmd())
	root.AddCommand(fxCmd())
	root.AddCommand(sampleCmd())
	root.AddCommand(trackCmd())
	root.AddCommand(patternCmd())
	root.AddCommand(exportCmd())
	root.AddCommand(midiCmd())
	root.AddCommand(langCmd())
	root.AddCommand(extrasCmd())
	root.AddCommand(djCmd())
	root.AddCommand(pkgCmd())
	root.AddCommand(installCmd())
	root.AddCommand(backupCmd())

	return root
}

// initConfig wires up Viper to read ~/.codic/config.yaml.
func initConfig() error {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		if err := ensureCodicDir(); err != nil {
			return err
		}
		viper.AddConfigPath(CodicDir())
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
	}
	viper.SetEnvPrefix("CODIC")
	viper.AutomaticEnv()

	// Defaults
	viper.SetDefault("default_duration", 8.0)
	viper.SetDefault("sample_rate", 44100)
	viper.SetDefault("output_dir", filepath.Join(CodicDir(), "out"))

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			// Config exists but is broken — warn, don't fail.
			fmt.Fprintf(os.Stderr, "warning: config error: %v\n", err)
		}
	}
	return nil
}

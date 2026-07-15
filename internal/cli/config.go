package cli

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func configCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage the CODIC/config.yaml configuration",
	}
	cmd.AddCommand(configInitCmd())
	cmd.AddCommand(configGetCmd())
	cmd.AddCommand(configSetCmd())
	cmd.AddCommand(configListCmd())
	cmd.AddCommand(configResetCmd())
	cmd.AddCommand(configEditCmd())
	cmd.AddCommand(configImportCmd())
	cmd.AddCommand(configExportCmd())
	return cmd
}

func configPath() string {
	return viper.ConfigFileUsed()
}

func ensureConfigFile() (string, error) {
	if p := configPath(); p != "" {
		return p, nil
	}
	if err := ensureCodicDir(); err != nil {
		return "", err
	}
	p := CodicDir() + string(os.PathSeparator) + "config.yaml"
	viper.SetConfigFile(p)
	return p, nil
}

func configInitCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Create a default config file",
		RunE: func(cmd *cobra.Command, args []string) error {
			p, err := ensureConfigFile()
			if err != nil {
				return err
			}
			if _, err := os.Stat(p); err == nil {
				fmt.Printf("config already exists: %s\n", p)
				return nil
			}
			viper.Set("default_duration", 8.0)
			viper.Set("sample_rate", 44100)
			viper.Set("output_dir", CodicDir()+string(os.PathSeparator)+"out")
			if err := viper.WriteConfig(); err != nil {
				return err
			}
			fmt.Printf("created config: %s\n", p)
			return nil
		},
	}
}

func configGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <key>",
		Short: "Print a config value",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !viper.IsSet(args[0]) {
				return fmt.Errorf("key %q not set", args[0])
			}
			fmt.Println(viper.Get(args[0]))
			return nil
		},
	}
}

func configSetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "set <key> <value>",
		Short: "Set a config value and persist it",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			p, err := ensureConfigFile()
			if err != nil {
				return err
			}
			viper.Set(args[0], coerce(args[1]))
			if err := viper.WriteConfig(); err != nil {
				return err
			}
			fmt.Printf("set %s in %s\n", args[0], p)
			return nil
		},
	}
}

func configListCmd() *cobra.Command {
	var asJSON bool
	cmd := &cobra.Command{
		Use:   "list",
		Short: "Print the full configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			if asJSON {
				b, err := marshalJSON(viper.AllSettings())
				if err != nil {
					return err
				}
				fmt.Println(b)
				return nil
			}
			for k, v := range viper.AllSettings() {
				fmt.Printf("%s = %v\n", k, v)
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&asJSON, "json", false, "output as JSON")
	return cmd
}

func configResetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "reset [key]",
		Short: "Reset a key (or the whole config) to defaults",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			p, err := ensureConfigFile()
			if err != nil {
				return err
			}
			if len(args) == 0 {
				viper.Reset()
				viper.Set("default_duration", 8.0)
				viper.Set("sample_rate", 44100)
			} else {
				viper.Set(args[0], nil)
			}
			if err := viper.WriteConfig(); err != nil {
				return err
			}
			fmt.Printf("reset %s\n", p)
			return nil
		},
	}
}

func configEditCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "edit",
		Short: "Open the config in $EDITOR",
		RunE: func(cmd *cobra.Command, args []string) error {
			p, err := ensureConfigFile()
			if err != nil {
				return err
			}
			if _, err := os.Stat(p); err != nil {
				viper.Set("default_duration", 8.0)
				viper.Set("sample_rate", 44100)
				if err := viper.WriteConfig(); err != nil {
					return err
				}
			}
			editor := os.Getenv("EDITOR")
			if editor == "" {
				editor = defaultEditor()
			}
			c := exec.Command(editor, p)
			c.Stdin = os.Stdin
			c.Stdout = os.Stdout
			c.Stderr = os.Stderr
			return c.Run()
		},
	}
}

func configImportCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "import <path>",
		Short: "Import a config file",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			p, err := ensureConfigFile()
			if err != nil {
				return err
			}
			data, err := os.ReadFile(args[0])
			if err != nil {
				return err
			}
			if err := os.WriteFile(p, data, 0o644); err != nil {
				return err
			}
			viper.SetConfigFile(p)
			fmt.Printf("imported config -> %s\n", p)
			return nil
		},
	}
}

func configExportCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "export <path>",
		Short: "Export the current config to a file",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := viper.WriteConfigAs(args[0]); err != nil {
				return err
			}
			fmt.Printf("exported config -> %s\n", args[0])
			return nil
		},
	}
}

func defaultEditor() string {
	if runtime.GOOS == "windows" {
		return "notepad"
	}
	return "vi"
}

func coerce(s string) interface{} {
	switch strings.ToLower(s) {
	case "true":
		return true
	case "false":
		return false
	}
	if f, err := parseFloat(s); err == nil {
		return f
	}
	return s
}

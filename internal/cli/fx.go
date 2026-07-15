package cli

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/spf13/cobra"
)

// Fx is a named effect chain preset stored under ~/.codic/fx/<name>.json.
type Fx struct {
	Name   string             `json:"name"`
	Type   string             `json:"type"` // reverb|delay|distortion|filter|chorus
	Desc   string             `json:"desc"`
	Params map[string]float64 `json:"params"`
	Chain  string             `json:"chain"` // codang post-processing expression
}

func fxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "fx",
		Short: "Manage effects / FX chains",
	}
	cmd.AddCommand(fxNewCmd())
	cmd.AddCommand(fxListCmd())
	cmd.AddCommand(fxShowCmd())
	cmd.AddCommand(fxEditCmd())
	cmd.AddCommand(fxAddCmd())
	cmd.AddCommand(fxRemoveCmd())
	cmd.AddCommand(fxPresetsCmd())
	cmd.AddCommand(fxTestCmd())
	return cmd
}

func fxNewCmd() *cobra.Command {
	var typ, desc, chain string
	cmd := &cobra.Command{
		Use:   "new <name>",
		Short: "Create a new FX chain",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			f := Fx{
				Name:   args[0],
				Type:   typ,
				Desc:   desc,
				Params: map[string]float64{},
				Chain:  chain,
			}
			if f.Type == "" {
				f.Type = "reverb"
			}
			if f.Chain == "" {
				f.Chain = `# no processing`
			}
			if err := registrySave("fx", f.Name, &f); err != nil {
				return err
			}
			fmt.Printf("created fx %q (%s)\n", f.Name, f.Type)
			return nil
		},
	}
	cmd.Flags().StringVar(&typ, "type", "", "type: reverb|delay|distortion|filter|chorus")
	cmd.Flags().StringVar(&desc, "desc", "", "description")
	cmd.Flags().StringVar(&chain, "chain", "", "codang post-processing expression")
	return cmd
}

func fxListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List saved FX chains",
		RunE: func(cmd *cobra.Command, args []string) error {
			names, err := registryList("fx")
			if err != nil {
				return err
			}
			if len(names) == 0 {
				fmt.Println("(no fx — create one with `codic fx new`)")
				return nil
			}
			sort.Strings(names)
			for _, n := range names {
				var f Fx
				_ = registryLoad("fx", n, &f)
				fmt.Printf("  %-18s %-12s %s\n", f.Name, f.Type, f.Desc)
			}
			return nil
		},
	}
}

func fxShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show <name>",
		Short: "Show an FX chain definition",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var f Fx
			if err := registryLoad("fx", args[0], &f); err != nil {
				return errNotFound("fx", args[0])
			}
			b, _ := marshalJSON(f)
			fmt.Println(b)
			return nil
		},
	}
}

func fxEditCmd() *cobra.Command {
	var typ, desc, chain string
	cmd := &cobra.Command{
		Use:   "edit <name>",
		Short: "Edit an FX chain",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var f Fx
			if err := registryLoad("fx", args[0], &f); err != nil {
				return errNotFound("fx", args[0])
			}
			if typ != "" {
				f.Type = typ
			}
			if desc != "" {
				f.Desc = desc
			}
			if chain != "" {
				f.Chain = chain
			}
			if err := registrySave("fx", f.Name, &f); err != nil {
				return err
			}
			fmt.Printf("updated fx %q\n", f.Name)
			return nil
		},
	}
	cmd.Flags().StringVar(&typ, "type", "", "type")
	cmd.Flags().StringVar(&desc, "desc", "", "description")
	cmd.Flags().StringVar(&chain, "chain", "", "codang post-processing expression")
	return cmd
}

func fxAddCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "add <name> <to-file.cdc> [out.cdc]",
		Short: "Apply an FX chain to a .cdc file",
		Args:  cobra.RangeArgs(2, 3),
		RunE: func(cmd *cobra.Command, args []string) error {
			var f Fx
			if err := registryLoad("fx", args[0], &f); err != nil {
				return errNotFound("fx", args[0])
			}
			code, err := os.ReadFile(args[1])
			if err != nil {
				return err
			}
			out := strings.TrimSuffix(args[1], ".cdc") + "_" + sanitize(f.Name) + ".cdc"
			if len(args) > 2 {
				out = args[2]
			}
			applied := fmt.Sprintf("%s\n// fx:%s\n%s\n.out()\n", code, f.Name, f.Chain)
			if err := os.WriteFile(out, []byte(applied), 0o644); err != nil {
				return err
			}
			fmt.Printf("applied fx %q -> %s\n", f.Name, out)
			return nil
		},
	}
}

func fxRemoveCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "remove <name>",
		Short: "Delete an FX chain",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := registryRemove("fx", args[0]); err != nil {
				return err
			}
			fmt.Printf("deleted fx %q\n", args[0])
			return nil
		},
	}
}

func fxPresetsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "presets",
		Short: "List built-in FX presets",
		RunE: func(cmd *cobra.Command, args []string) error {
			for _, p := range builtinFx() {
				fmt.Printf("  %-12s %-12s %s\n", p.Name, p.Type, p.Desc)
			}
			return nil
		},
	}
}

func fxTestCmd() *cobra.Command {
	var seconds float64
	cmd := &cobra.Command{
		Use:   "test <name> [out.wav]",
		Short: "Render a short demo through an FX chain",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			var f Fx
			if err := registryLoad("fx", args[0], &f); err != nil {
				for _, p := range builtinFx() {
					if p.Name == args[0] {
						f = p
					}
				}
				if f.Name == "" {
					return errNotFound("fx", args[0])
				}
			}
			code := `s("bd")|.every(4, x=>x.speed(1.5))` + "\n" + f.Chain + "\n.out()"
			out := "fx_test.wav"
			if len(args) > 1 {
				out = args[1]
			}
			return renderToFileString(code, out, seconds)
		},
	}
	cmd.Flags().Float64Var(&seconds, "sec", 3, "demo length in seconds")
	return cmd
}

func builtinFx() []Fx {
	return []Fx{
		{Name: "reverb", Type: "reverb", Desc: "hall reverb", Params: map[string]float64{"size": 0.6, "decay": 4}},
		{Name: "delay", Type: "delay", Desc: "ping-pong delay", Params: map[string]float64{"time": 0.3, "fb": 0.4}},
		{Name: "distortion", Type: "distortion", Desc: "soft clip", Params: map[string]float64{"drive": 0.5}},
		{Name: "filter", Type: "filter", Desc: "low-pass sweep", Params: map[string]float64{"cutoff": 1200}},
		{Name: "chorus", Type: "chorus", Desc: "wide chorus", Params: map[string]float64{"rate": 0.5, "depth": 0.3}},
	}
}

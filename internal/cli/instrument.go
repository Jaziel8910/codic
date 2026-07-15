package cli

import (
	"fmt"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/Jaziel8910/codic/internal/audio"
)

// Instrument is a named synth/sampler definition stored as JSON under
// ~/.codic/instruments/<name>.json.
type Instrument struct {
	Name     string             `json:"name"`
	Engine   string             `json:"engine"` // synth | sampler | fm | sub
	Desc     string             `json:"desc"`
	Params   map[string]float64 `json:"params"`
	Template string             `json:"template"` // codang expression using $note
}

func instrumentCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "instrument",
		Short: "Manage instruments (synths / samplers)",
	}
	cmd.AddCommand(instrumentNewCmd())
	cmd.AddCommand(instrumentListCmd())
	cmd.AddCommand(instrumentShowCmd())
	cmd.AddCommand(instrumentEditCmd())
	cmd.AddCommand(instrumentCloneCmd())
	cmd.AddCommand(instrumentDeleteCmd())
	cmd.AddCommand(instrumentPresetsCmd())
	cmd.AddCommand(instrumentTestCmd())
	return cmd
}

func instrumentNewCmd() *cobra.Command {
	var engine, desc, template string
	cmd := &cobra.Command{
		Use:   "new <name>",
		Short: "Create a new instrument",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			inst := Instrument{
				Name:     args[0],
				Engine:   engine,
				Desc:     desc,
				Params:   map[string]float64{"gain": 1, "pan": 0},
				Template: template,
			}
			if inst.Engine == "" {
				inst.Engine = "synth"
			}
			if inst.Template == "" {
				inst.Template = fmt.Sprintf("s(\"%s\")", "bd")
			}
			if err := registrySave("instruments", inst.Name, &inst); err != nil {
				return err
			}
			fmt.Printf("created instrument %q (%s)\n", inst.Name, inst.Engine)
			return nil
		},
	}
	cmd.Flags().StringVar(&engine, "engine", "", "engine: synth|sampler|fm|sub")
	cmd.Flags().StringVar(&desc, "desc", "", "description")
	cmd.Flags().StringVar(&template, "template", "", "codang template using $note")
	return cmd
}

func instrumentListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List saved instruments",
		RunE: func(cmd *cobra.Command, args []string) error {
			names, err := registryList("instruments")
			if err != nil {
				return err
			}
			if len(names) == 0 {
				fmt.Println("(no instruments — create one with `codic instrument new`)")
				return nil
			}
			sort.Strings(names)
			for _, n := range names {
				var inst Instrument
				_ = registryLoad("instruments", n, &inst)
				fmt.Printf("  %-18s %-8s %s\n", inst.Name, inst.Engine, inst.Desc)
			}
			return nil
		},
	}
}

func instrumentShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show <name>",
		Short: "Show an instrument definition",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var inst Instrument
			if err := registryLoad("instruments", args[0], &inst); err != nil {
				return errNotFound("instrument", args[0])
			}
			b, _ := marshalJSON(inst)
			fmt.Println(b)
			return nil
		},
	}
}

func instrumentEditCmd() *cobra.Command {
	var engine, desc, template string
	cmd := &cobra.Command{
		Use:   "edit <name>",
		Short: "Edit an instrument",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var inst Instrument
			if err := registryLoad("instruments", args[0], &inst); err != nil {
				return errNotFound("instrument", args[0])
			}
			if engine != "" {
				inst.Engine = engine
			}
			if desc != "" {
				inst.Desc = desc
			}
			if template != "" {
				inst.Template = template
			}
			if err := registrySave("instruments", inst.Name, &inst); err != nil {
				return err
			}
			fmt.Printf("updated instrument %q\n", inst.Name)
			return nil
		},
	}
	cmd.Flags().StringVar(&engine, "engine", "", "engine: synth|sampler|fm|sub")
	cmd.Flags().StringVar(&desc, "desc", "", "description")
	cmd.Flags().StringVar(&template, "template", "", "codang template using $note")
	return cmd
}

func instrumentCloneCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "clone <name> <newname>",
		Short: "Clone an instrument",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			var inst Instrument
			if err := registryLoad("instruments", args[0], &inst); err != nil {
				return errNotFound("instrument", args[0])
			}
			inst.Name = args[1]
			if err := registrySave("instruments", inst.Name, &inst); err != nil {
				return err
			}
			fmt.Printf("cloned %q -> %q\n", args[0], args[1])
			return nil
		},
	}
}

func instrumentDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete <name>",
		Short: "Delete an instrument",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := registryRemove("instruments", args[0]); err != nil {
				return err
			}
			fmt.Printf("deleted instrument %q\n", args[0])
			return nil
		},
	}
}

func instrumentPresetsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "presets",
		Short: "List built-in instrument presets",
		RunE: func(cmd *cobra.Command, args []string) error {
			for _, p := range builtinInstruments() {
				fmt.Printf("  %-12s %-8s %s\n", p.Name, p.Engine, p.Desc)
			}
			return nil
		},
	}
}

func instrumentTestCmd() *cobra.Command {
	var note string
	var seconds float64
	cmd := &cobra.Command{
		Use:   "test <name> [out.wav]",
		Short: "Render a short demo of an instrument",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			var inst Instrument
			if err := registryLoad("instruments", args[0], &inst); err != nil {
				// fall back to built-in preset
				for _, p := range builtinInstruments() {
					if p.Name == args[0] {
						inst = p
					}
				}
				if inst.Name == "" {
					return errNotFound("instrument", args[0])
				}
			}
			n := note
			if n == "" {
				n = "c3"
			}
			code := strings.ReplaceAll(inst.Template, "$note", n)
			out := "instrument_test.wav"
			if len(args) > 1 {
				out = args[1]
			}
			return renderToFileString(code, out, seconds)
		},
	}
	cmd.Flags().StringVar(&note, "note", "", "note to play (default c3)")
	cmd.Flags().Float64Var(&seconds, "sec", 2, "demo length in seconds")
	return cmd
}

func renderToFileString(code, out string, seconds float64) error {
	combined, cps, err := collectPatterns(code)
	if err != nil {
		return err
	}
	buf, err := audio.RenderPattern(combined, cps, seconds)
	if err != nil {
		return err
	}
	return audio.WriteWAVFile(out, buf, audio.SampleRate)
}

// builtinInstruments returns a small set of ready-made presets.
func builtinInstruments() []Instrument {
	return []Instrument{
		{Name: "kick", Engine: "synth", Desc: "punchy kick", Template: `s("bd")`, Params: map[string]float64{"gain": 1}},
		{Name: "snare", Engine: "synth", Desc: "snappy snare", Template: `s("sd")`, Params: map[string]float64{"gain": 0.9}},
		{Name: "bass", Engine: "synth", Desc: "sub bass", Template: `s("bass")`, Params: map[string]float64{"gain": 0.8}},
		{Name: "lead", Engine: "synth", Desc: "lead pluck", Template: `s("bleep")`, Params: map[string]float64{"gain": 0.7}},
		{Name: "pad", Engine: "synth", Desc: "soft pad", Template: `s("casio")`, Params: map[string]float64{"gain": 0.6}},
		{Name: "hat", Engine: "synth", Desc: "hi-hat", Template: `s("hh")`, Params: map[string]float64{"gain": 0.5}},
	}
}

package cli

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/Jaziel8910/codic/internal/audio"
	"github.com/Jaziel8910/codic/internal/codang"
)

func langCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "lang",
		Short: "Codang language tools (repl, format, lint, docs)",
	}
	cmd.AddCommand(langReplCmd())
	cmd.AddCommand(langFmtCmd())
	cmd.AddCommand(langLintCmd())
	cmd.AddCommand(langDocCmd())
	cmd.AddCommand(langParseCmd())
	cmd.AddCommand(langTokenizeCmd())
	cmd.AddCommand(langSpecCmd())
	cmd.AddCommand(langModuleCmd())
	cmd.AddCommand(langCompletionsCmd())
	cmd.AddCommand(langEvalCmd())
	cmd.AddCommand(langTypesCmd())
	return cmd
}

func langReplCmd() *cobra.Command {
	var secs float64
	cmd := &cobra.Command{
		Use:   "repl",
		Short: "Interactive Codang REPL (blank line renders a preview)",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("codic repl - type codang, blank line to preview, 'quit' to exit")
			sc := bufio.NewScanner(os.Stdin)
			var src strings.Builder
			for {
				fmt.Print("codic> ")
				if !sc.Scan() {
					break
				}
				line := sc.Text()
				if strings.TrimSpace(line) == "quit" {
					break
				}
				if strings.TrimSpace(line) == "" {
					if src.Len() > 0 {
						combined, cps, err := collectPatterns(src.String())
						if err != nil {
							fmt.Printf("error: %v\n", err)
						} else {
							buf, _ := audio.RenderPattern(combined, cps, secs)
							out := "repl_preview.wav"
							if err := audio.WriteWAVFile(out, buf, audio.SampleRate); err != nil {
								fmt.Printf("render error: %v\n", err)
							} else {
								fmt.Printf("preview -> %s (%d events/cycle)\n", out, len(combined.FirstCycle()))
							}
						}
						src.Reset()
					}
					continue
				}
				src.WriteString(line)
				src.WriteString("\n")
			}
			return nil
		},
	}
	cmd.Flags().Float64Var(&secs, "sec", 2, "preview length in seconds")
	return cmd
}

func langFmtCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "fmt <file.cdc>",
		Short: "Normalize whitespace in a Codang file",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			data, err := os.ReadFile(args[0])
			if err != nil {
				return err
			}
			lines := strings.Split(string(data), "\n")
			var out []string
			blank := 0
			for _, l := range lines {
				s := strings.TrimRight(l, " \t")
				if strings.TrimSpace(s) == "" {
					blank++
					if blank > 1 {
						continue
					}
				} else {
					blank = 0
				}
				out = append(out, s)
			}
			for len(out) > 0 && strings.TrimSpace(out[len(out)-1]) == "" {
				out = out[:len(out)-1]
			}
			if err := os.WriteFile(args[0], []byte(strings.Join(out, "\n")+"\n"), 0o644); err != nil {
				return err
			}
			fmt.Printf("formatted %s\n", args[0])
			return nil
		},
	}
}

func langLintCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "lint <file.cdc>",
		Short: "Check a Codang file for common problems",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			data, err := os.ReadFile(args[0])
			if err != nil {
				return err
			}
			src := string(data)
			problems := 0
			if _, e := codang.Parse(src); e != nil {
				fmt.Printf("ERROR: parse failed: %v\n", e)
				problems++
			}
			if !strings.Contains(src, ".out()") {
				fmt.Println("WARN: no .out() call - nothing will be produced")
				problems++
			}
			if strings.Contains(src, "speed(") && strings.Contains(src, "superavage") {
				fmt.Println("INFO: using superavage (good!)")
			}
			if prog, e := codang.Parse(src); e == nil {
				errs, warns := codang.ValidateSong(prog)
				if len(errs) > 0 {
					for _, er := range errs {
						fmt.Printf("ERROR: %s\n", er)
						problems++
					}
				}
				for _, w := range warns {
					fmt.Printf("WARN: %s\n", w)
				}
			}
			if problems == 0 {
				fmt.Printf("%s: OK\n", args[0])
			}
			return nil
		},
	}
}

func langDocCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "doc",
		Short: "Print a quick reference for the Codang language",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Print(langReference)
			return nil
		},
	}
}

func langParseCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "parse <file.cdc>",
		Short: "Parse a file and show its AST summary",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			data, err := os.ReadFile(args[0])
			if err != nil {
				return err
			}
			prog, err := codang.Parse(string(data))
			if err != nil {
				return fmt.Errorf("parse error: %w", err)
			}
			fmt.Printf("statements: %d\n", len(prog.Statements))
			fmt.Printf("metadata  : %v\n", prog.Metadata)
			return nil
		},
	}
}

func langTokenizeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "tokenize <file.cdc>",
		Short: "Tokenize a file and print the token stream",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			data, err := os.ReadFile(args[0])
			if err != nil {
				return err
			}
			lex := codang.NewLexer(string(data))
			toks, err := lex.Tokenize()
			if err != nil {
				return err
			}
			for _, t := range toks {
				fmt.Println(t.String())
			}
			return nil
		},
	}
}

func langSpecCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "spec",
		Short: "Print the EBNF-like grammar for Codang",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Print(codangGrammar)
		},
	}
}

func langModuleCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "module",
		Short: "Manage Codang modules (imports / stdlib)",
	}
	cmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List available modules in stdlib",
		RunE: func(cmd *cobra.Command, args []string) error {
			modDir := filepath.Join(CodicDir(), "stdlib")
			entries, err := os.ReadDir(modDir)
			if err != nil {
				if os.IsNotExist(err) {
					fmt.Println("(no modules installed — run `codic lang module init` to set up)")
					return nil
				}
				return err
			}
			found := 0
			for _, e := range entries {
				if !e.IsDir() && (strings.HasSuffix(e.Name(), ".cdc") || strings.HasSuffix(e.Name(), ".md")) {
					fmt.Printf("  %s\n", e.Name())
					found++
				}
			}
			if found == 0 {
				fmt.Println("(no .cdc modules found)")
			}
			return nil
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:   "info <module>",
		Short: "Show details of a module",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			modDir := filepath.Join(CodicDir(), "stdlib")
			path := filepath.Join(modDir, args[0])
			if !strings.HasSuffix(path, ".cdc") {
				path += ".cdc"
			}
			data, err := os.ReadFile(path)
			if err != nil {
				return fmt.Errorf("module %q not found", args[0])
			}
			prog, err := codang.Parse(string(data))
			if err != nil {
				fmt.Printf("file: %s\n(raw)\n%s\n", path, data)
				return nil
			}
			fmt.Printf("module  : %s\n", args[0])
			fmt.Printf("path    : %s\n", path)
			fmt.Printf("metadata: %v\n", prog.Metadata)
			fmt.Printf("stmts   : %d\n", len(prog.Statements))
			return nil
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:   "new <name>",
		Short: "Create a new module skeleton in stdlib",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			modDir := filepath.Join(CodicDir(), "stdlib")
			if err := os.MkdirAll(modDir, 0o755); err != nil {
				return err
			}
			file := filepath.Join(modDir, args[0])
			if !strings.HasSuffix(file, ".cdc") {
				file += ".cdc"
			}
			if _, err := os.Stat(file); err == nil {
				return fmt.Errorf("module %q already exists", file)
			}
			skeleton := fmt.Sprintf("# Module: %s\n# Import with: import %q\n\n", args[0], args[0])
			if err := os.WriteFile(file, []byte(skeleton), 0o644); err != nil {
				return err
			}
			fmt.Printf("created module -> %s\n", file)
			return nil
		},
	})
	return cmd
}

var codangGrammar = `Codang EBNF-like Grammar
========================

Program      = { Statement | Metadata } .
Metadata     = "@" key value .
Statement    = Assignment | FuncDef | CallExpr .

Assignment   = identifier "=" Expression .
FuncDef      = "func" identifier "(" [ Parameters ] ")" ":" StatementList "end".
Parameters   = identifier { "," identifier } .

Expression   = CallExpr | MethodChain | InfixExpr | Atom .
MethodChain  = Expression "." identifier "(" [ Arguments ] ")" .
InfixExpr    = Expression operator Expression .
Atom         = string | number | identifier | "(" Expression ")" .

CallExpr     = identifier "(" [ Arguments ] ")" .
Arguments    = Expression { "," Expression } .

Built-in functions:
  s(name)        → select sample
  note(notes)    → pitched notes
  gain(v), pan(v), cutoff(v), resonance(v)
  fast(n), slow(n), rev(), speed(n)
  stack(a, b), cat(a, b), slowcat(a, b)
  euclid(steps, pulses, rotation)

Operators:
  |  pipe (method chain)
  +  stack
  *  repeat

Mini-notation (inside strings):
  "bd sd hh"       sequence
  "[bd sd] hh"     grouping
  "bd*3"           repeat
  "<bd cp>"        alternate
  "(3,8)"          euclidean rhythm
  "bd?"            random 50%
  "bd~hh"          rest
`

func langCompletionsCmd() *cobra.Command {
	var shell string
	cmd := &cobra.Command{
		Use:   "completions [out-file]",
		Short: "Generate shell completions for codic",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			root := cmd.Root()
			var w *os.File = os.Stdout
			if len(args) > 0 {
				f, err := os.Create(args[0])
				if err != nil {
					return err
				}
				defer f.Close()
				w = f
			}
			switch shell {
			case "bash":
				return root.GenBashCompletion(w)
			case "zsh":
				return root.GenZshCompletion(w)
			case "fish":
				return root.GenFishCompletion(w, true)
			case "powershell":
				return root.GenPowerShellCompletionWithDesc(w)
			default:
				return fmt.Errorf("unknown shell %q (bash|zsh|fish|powershell)", shell)
			}
		},
	}
	cmd.Flags().StringVar(&shell, "shell", "bash", "shell: bash|zsh|fish|powershell")
	return cmd
}

func langEvalCmd() *cobra.Command {
	var secs float64
	cmd := &cobra.Command{
		Use:   "eval <expression>",
		Short: "Evaluate a Codang expression and show a preview",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			combined, cps, err := collectPatterns(args[0])
			if err != nil {
				return err
			}
			buf, _ := audio.RenderPattern(combined, cps, secs)
			out := "eval_preview.wav"
			if err := audio.WriteWAVFile(out, buf, audio.SampleRate); err != nil {
				return err
			}
			fmt.Printf("evaluated -> %s (cps=%.4f, events/cycle=%d)\n", out, cps, len(combined.FirstCycle()))
			return nil
		},
	}
	cmd.Flags().Float64Var(&secs, "sec", 2, "preview length in seconds")
	return cmd
}

func langTypesCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "types",
		Short: "List the 10 Codang song types (mini-loop to full-prod-song)",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Tipos de cancion en Codang (elige bien - el tipo exige la complejidad):")
			fmt.Println()
			for _, t := range codang.SongTypeList {
				fmt.Println(t.Summary())
				fmt.Println()
			}
			fmt.Printf("Usa @type <id> al inicio de tu .cdc. Validos: %s\n", codang.ValidTypeList())
		},
	}
}

var langReference = `Codang quick reference
====================
Samples : s("bd"), s("sd"), s("hh") ...            trigger a sample
Stack   : a.cat(b)                                 layer patterns
Repeat  : p.every(n, f)                            transform every n cycles
Speed   : p.speed(2) / p.slow(2)                  change playback rate
Notes   : note("c3").out()                         play a pitched note
Math    : + - * / %                                arithmetic
Chains  : p |> f |> g                              pipe transformations
Metadata: @bpm 120 / @key "c minor" / @name "x"   song header
Output  : .out()                                   emit the pattern
`

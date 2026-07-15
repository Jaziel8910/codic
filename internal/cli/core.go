package cli

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/Jaziel8910/codic/internal/audio"
	"github.com/Jaziel8910/codic/internal/codang"
	"github.com/Jaziel8910/codic/internal/pattern"
)

// ---------- new ----------

func newCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "new [type] [name]",
		Short: "Scaffold a track, project, dj set, template or instrument",
		Args:  cobra.MinimumNArgs(1),
		RunE:  runNew,
	}
	return cmd
}

func runNew(cmd *cobra.Command, args []string) error {
	typ := args[0]
	name := "untitled"
	if len(args) > 1 {
		name = args[1]
	}
	switch typ {
	case "track":
		return newTrack(name)
	case "project":
		return newProject(name)
	case "dj":
		return newDJ(name)
	case "template":
		return newTemplate(name)
	case "instrument":
		return newInstrument(name)
	default:
		return fmt.Errorf("unknown type %q (use: track|project|dj|template|instrument)", typ)
	}
}

func newTrack(name string) error {
	name = sanitize(name)
	path := name + ".cdc"
	if _, err := os.Stat(path); err == nil {
		return fmt.Errorf("%s already exists", path)
	}
	tpl := fmt.Sprintf(`# %s
# Codang track — render with: codic render %s out.wav
bpm 120

kick = s("bd").euclid(8, 2, 0).out()
hat  = s("hh").euclid(8, 3, 1).speed(2).gain(0.4).out()
bass = "c2 e2 g2 a2".note().s("sawtooth").cutoff(600).gain(0.3).out()
`, name, path)
	if err := os.WriteFile(path, []byte(tpl), 0o644); err != nil {
		return err
	}
	fmt.Printf("created track %s\n", path)
	return nil
}

func newProject(name string) error {
	name = sanitize(name)
	dir := name
	if err := os.MkdirAll(filepath.Join(dir, "tracks"), 0o755); err != nil {
		return err
	}
	yaml := fmt.Sprintf(`project: %q
type: album
bpm: 120
key: "c minor"
tracks: []
metadata:
  artist: ""
  label: ""
  release_date: ""
  cover: ""
`, name)
	if err := os.WriteFile(filepath.Join(dir, "project.yaml"), []byte(yaml), 0o644); err != nil {
		return err
	}
	fmt.Printf("created project %s/ (project.yaml)\n", dir)
	return nil
}

func newDJ(name string) error {
	name = sanitize(name)
	path := name + ".cdc"
	tpl := fmt.Sprintf(`# DJ set: %s
# Train with: codic dj learn --genre %s *.cdc
bpm 128
s("bd*4").euclid(16, 4, 0).out()
`, name, name)
	return os.WriteFile(path, []byte(tpl), 0o644)
}

func newTemplate(name string) error {
	path := sanitize(name) + ".cdc"
	tpl := "# Codang template\nbpm 120\n\n"
	return os.WriteFile(path, []byte(tpl), 0o644)
}

func newInstrument(name string) error {
	path := sanitize(name) + ".cdc"
	tpl := fmt.Sprintf("# Instrument: %s\n# wavetable / synth definition\n", name)
	return os.WriteFile(path, []byte(tpl), 0o644)
}

func sanitize(s string) string {
	s = strings.TrimSuffix(s, ".cdc")
	s = strings.TrimSuffix(s, ".yaml")
	return strings.ReplaceAll(s, string(filepath.Separator), "_")
}

// ---------- play ----------

func playCmd() *cobra.Command {
	var duration float64
	var noOpen, normalize bool
	cmd := &cobra.Command{
		Use:   "play [file.cdc]",
		Short: "Render a .cdc file and open it in the OS player",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			tmp, err := tempWAV("play_preview.wav")
			if err != nil {
				return err
			}
			fmt.Printf("rendering %s -> %s\n", args[0], tmp)
			if err := renderToFile(args[0], tmp, duration, normalize); err != nil {
				return err
			}
			if noOpen {
				fmt.Printf("rendered (--no-open): %s\n", tmp)
				return nil
			}
			fmt.Println("opening in OS player...")
			return openPlayer(tmp)
		},
	}
	cmd.Flags().Float64VarP(&duration, "duration", "d", 0, "render duration in seconds (default from config)")
	cmd.Flags().BoolVar(&noOpen, "no-open", false, "do not open the player, just render to temp")
	cmd.Flags().BoolVar(&normalize, "normalize", true, "peak-normalize the output")
	return cmd
}

// ---------- render ----------

func renderCmd() *cobra.Command {
	var duration float64
	var normalize, stems bool
	var format string
	cmd := &cobra.Command{
		Use:   "render [file.cdc] [out.wav]",
		Short: "Render a .cdc file to a permanent audio file",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			in := args[0]
			out := "out.wav"
			if len(args) > 1 {
				out = args[1]
			}
			if format != "" && format != "wav" {
				fmt.Fprintf(os.Stderr, "warning: --format %s not supported in phase 0, rendering WAV\n", format)
			}
			if stems {
				fmt.Fprintln(os.Stderr, "warning: --stems not yet implemented in phase 0")
			}
			fmt.Printf("rendering %s -> %s\n", in, out)
			if err := renderToFile(in, out, duration, normalize); err != nil {
				return err
			}
			fmt.Printf("done: %s\n", out)
			return nil
		},
	}
	cmd.Flags().Float64VarP(&duration, "duration", "d", 0, "render duration in seconds")
	cmd.Flags().BoolVar(&normalize, "normalize", true, "peak-normalize the output")
	cmd.Flags().BoolVar(&stems, "stems", false, "render separate stems (planned)")
	cmd.Flags().StringVar(&format, "format", "wav", "output format: wav (flac|mp3 planned)")
	return cmd
}

// ---------- eval ----------

func evalCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "eval [code]",
		Short: "Evaluate Codang code inline and print the resulting Haps",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			combined, cps, err := collectPatterns(args[0])
			if err != nil {
				return err
			}
			haps := combined.FirstCycle()
			fmt.Printf("cps=%.2f  haps(in first cycle)=%d\n", cps, len(haps))
			for i, h := range haps {
				if i >= 20 {
					fmt.Printf("  ... (%d more)\n", len(haps)-20)
					break
				}
				fmt.Printf("  [%d] %s\n", i, h.Show())
			}
			return nil
		},
	}
}

// ---------- run ----------

func runCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "run [file.cdc]",
		Short: "Execute a complete .cdc file and report its output",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			code, err := os.ReadFile(args[0])
			if err != nil {
				return fmt.Errorf("reading %s: %w", args[0], err)
			}
			prog, err := codang.Parse(string(code))
			if err != nil {
				return fmt.Errorf("parse error: %w", err)
			}
			for k, v := range prog.Metadata {
				fmt.Printf("  @%s %s\n", k, v)
			}
			var all []pattern.Pattern
			eval := codang.NewEvaluator(func(p pattern.Pattern) { all = append(all, p) })
			if err := eval.Eval(prog); err != nil {
				return fmt.Errorf("eval error: %w", err)
			}
			proj := eval.Project()
			fmt.Printf("patterns: %d | project: %q | tracks: %d\n", len(all), proj.Name, len(proj.Tracks))
			combined, cps, err := collectPatterns(string(code))
			if err == nil {
				fmt.Printf("first-cycle haps: %d (cps=%.2f)\n", len(combined.FirstCycle()), cps)
			}
			return nil
		},
	}
}

// ---------- watch ----------

func watchCmd() *cobra.Command {
	var duration float64
	var noOpen bool
	cmd := &cobra.Command{
		Use:   "watch [file.cdc]",
		Short: "Watch a .cdc file and re-render on every save",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path, err := filepath.Abs(args[0])
			if err != nil {
				return err
			}
			w, err := fsnotify.NewWatcher()
			if err != nil {
				return err
			}
			defer w.Close()
			if err := w.Add(filepath.Dir(path)); err != nil {
				return err
			}
			tmp, err := tempWAV("watch_preview.wav")
			if err != nil {
				return err
			}
			render := func() {
				fmt.Printf("[%s] change detected, rendering...\n", time.Now().Format("15:04:05"))
				if err := renderToFile(path, tmp, duration, true); err != nil {
					fmt.Fprintf(os.Stderr, "  render error: %v\n", err)
					return
				}
				fmt.Printf("  rendered -> %s\n", tmp)
				if !noOpen {
					_ = openPlayer(tmp)
				}
			}
			fmt.Printf("watching %s (Ctrl+C to stop)\n", path)
			render()
			for {
				select {
				case ev, ok := <-w.Events:
					if !ok {
						return nil
					}
					if filepath.Clean(ev.Name) == filepath.Clean(path) &&
						(ev.Op&fsnotify.Write == fsnotify.Write || ev.Op&fsnotify.Create == fsnotify.Create) {
						render()
					}
				case err, ok := <-w.Errors:
					if !ok {
						return nil
					}
					fmt.Fprintf(os.Stderr, "watch error: %v\n", err)
				}
			}
		},
	}
	cmd.Flags().Float64VarP(&duration, "duration", "d", 0, "render duration in seconds")
	cmd.Flags().BoolVar(&noOpen, "no-open", false, "do not open the player on each change")
	return cmd
}

// ---------- serve ----------

func serveCmd() *cobra.Command {
	var port int
	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Start an HTTP server (render API + version)",
		RunE: func(cmd *cobra.Command, args []string) error {
			mux := http.NewServeMux()
			mux.HandleFunc("/version", func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprintf(w, "codic %s\n", Version)
			})
			mux.HandleFunc("/render", func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPost {
					http.Error(w, "use POST with ?file=path.cdc", http.StatusMethodNotAllowed)
					return
				}
				f := r.URL.Query().Get("file")
				if f == "" {
					http.Error(w, "missing ?file=", http.StatusBadRequest)
					return
				}
				tmp, err := tempWAV("serve_preview.wav")
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				dur := viper.GetFloat64("default_duration")
				if err := renderToFile(f, tmp, dur, true); err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				http.ServeFile(w, r, tmp)
			})
			addr := fmt.Sprintf(":%d", port)
			fmt.Printf("codic serve listening on http://localhost%s\n", addr)
			fmt.Println("  POST /render?file=track.cdc  -> WAV")
			fmt.Println("  GET  /version")
			return http.ListenAndServe(addr, mux)
		},
	}
	cmd.Flags().IntVar(&port, "port", 7331, "HTTP port")
	return cmd
}

// ---------- version ----------

func versionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version, git hash and engine info",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("codic %s\n", Version)
			if GitHash != "" {
				fmt.Printf("git: %s\n", GitHash)
			}
			fmt.Printf("engine: offline render -> WAV (SampleRate=%d)\n", audio.SampleRate)
			fmt.Printf("config: %s\n", viper.ConfigFileUsed())
		},
	}
}

// GitHash is injected at build time; empty otherwise.
var GitHash = ""

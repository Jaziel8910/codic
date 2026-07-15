package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/Jaziel8910/codic/internal/audio"
	"github.com/Jaziel8910/codic/internal/dj"
	"github.com/Jaziel8910/codic/internal/pattern"
)

const djSecondsPerBar = 240.0 // cps = bpm/240, so 1 bar = 240/bpm seconds

// renderPattern renders a pattern to a WAV file (normalized) and optionally
// opens the OS player.
func renderPatternToFile(pat pattern.Pattern, cps, seconds float64, out string, play bool) error {
	buf, err := audio.RenderPattern(pat, cps, seconds)
	if err != nil {
		return err
	}
	buf = normalizePeak(buf)
	if err := audio.WriteWAVFile(out, buf, audio.SampleRate); err != nil {
		return err
	}
	fmt.Printf("rendered -> %s (%.1fs)\n", out, seconds)
	if play {
		return openPlayer(out)
	}
	return nil
}

// ---- dj learn ----

func djLearnCmd() *cobra.Command {
	var genre string
	var tags []string
	cmd := &cobra.Command{
		Use:   "learn [files...]",
		Short: "Train a genre model from .cdc files",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if genre == "" {
				return fmt.Errorf("--genre is required")
			}
			if err := dj.Learn(genre, args, tags); err != nil {
				return err
			}
			fmt.Printf("learned genre %q from %d file(s)\n", genre, len(args))
			return nil
		},
	}
	cmd.Flags().StringVar(&genre, "genre", "", "genre name to train")
	cmd.Flags().StringSliceVar(&tags, "tags", nil, "only learn these layers (kick,bass,lead,pad,fx)")
	return cmd
}

// ---- dj forget ----

func djForgetCmd() *cobra.Command {
	var genre string
	cmd := &cobra.Command{
		Use:   "forget",
		Short: "Delete a learned genre model",
		RunE: func(cmd *cobra.Command, args []string) error {
			if genre == "" {
				return fmt.Errorf("--genre is required")
			}
			if err := dj.Forget(genre); err != nil {
				return err
			}
			fmt.Printf("forgot genre %q\n", genre)
			return nil
		},
	}
	cmd.Flags().StringVar(&genre, "genre", "", "genre name to forget")
	return cmd
}

// ---- dj list ----

func djListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List learned genres",
		RunE: func(cmd *cobra.Command, args []string) error {
			genres, err := dj.ListGenres()
			if err != nil {
				return err
			}
			if len(genres) == 0 {
				fmt.Println("(no learned genres — use `dj learn` or a built-in like techno/house/dnb)")
				return nil
			}
			fmt.Println("learned genres:")
			for _, g := range genres {
				fmt.Printf("  - %s\n", g)
			}
			return nil
		},
	}
}

// ---- dj stats ----

func djStatsCmd() *cobra.Command {
	var genre string
	cmd := &cobra.Command{
		Use:   "stats",
		Short: "Show statistics of a learned genre model",
		RunE: func(cmd *cobra.Command, args []string) error {
			if genre == "" {
				return fmt.Errorf("--genre is required")
			}
			g, err := dj.Stats(genre)
			if err != nil {
				g = dj.DefaultProfileFor(genre)
				g.Name = genre
			}
			fmt.Printf("genre:      %s\n", g.Name)
			fmt.Printf("bpm range:  %d-%d\n", g.BPMMin, g.BPMMax)
			fmt.Printf("key:        %s\n", g.Key)
			fmt.Printf("tracks:     %d\n", g.Count)
			fmt.Printf("layers:\n")
			fmt.Printf("  kick:  %d   snare: %d   clap: %d   hat: %d\n", len(g.Layers.Kick), len(g.Layers.Snare), len(g.Layers.Clap), len(g.Layers.Hat))
			fmt.Printf("  bass:  %d   lead:  %d   pad:  %d   fx:  %d\n", len(g.Layers.Bass), len(g.Layers.Lead), len(g.Layers.Pad), len(g.Layers.Fx))
			fmt.Printf("structures: %s\n", strings.Join(g.Structures, ", "))
			return nil
		},
	}
	cmd.Flags().StringVar(&genre, "genre", "", "genre name")
	return cmd
}

// ---- dj play ----

func djPlayCmd() *cobra.Command {
	var genre, key string
	var bpm int
	var endless bool
	cmd := &cobra.Command{
		Use:   "play",
		Short: "Generate and play an endless (or long) loop for a genre",
		RunE: func(cmd *cobra.Command, args []string) error {
			if genre == "" {
				return fmt.Errorf("--genre is required")
			}
			pat, cps, _, err := dj.GenerateLoop(genre, key, bpm, 8, int64(os.Getpid()))
			if err != nil {
				return err
			}
			secs := 60.0
			if endless {
				secs = 240.0
			}
			out := "dj_play.wav"
			return renderPatternToFile(pat, cps, secs, out, true)
		},
	}
	cmd.Flags().StringVar(&genre, "genre", "", "genre name")
	cmd.Flags().StringVar(&key, "key", "", "musical key (e.g. 'f# minor')")
	cmd.Flags().IntVar(&bpm, "bpm", 0, "tempo (defaults to genre range)")
	cmd.Flags().BoolVar(&endless, "endless", false, "render a long loop (4 minutes)")
	return cmd
}

// ---- dj loop ----

func djLoopCmd() *cobra.Command {
	var genre, key, out string
	var bpm, bars int
	var stems bool
	cmd := &cobra.Command{
		Use:   "loop",
		Short: "Generate a finite loop and render it to WAV",
		RunE: func(cmd *cobra.Command, args []string) error {
			if genre == "" {
				return fmt.Errorf("--genre is required")
			}
			if bars <= 0 {
				bars = 8
			}
			if out == "" {
				out = "dj_loop.wav"
			}
			if stems {
				return renderDJStems(genre, key, bpm, bars, out)
			}
			pat, cps, _, err := dj.GenerateLoop(genre, key, bpm, bars, int64(os.Getpid()))
			if err != nil {
				return err
			}
			secs := float64(bars) * djSecondsPerBar / float64(bpmFor(genre, bpm))
			return renderPatternToFile(pat, cps, secs, out, false)
		},
	}
	cmd.Flags().StringVar(&genre, "genre", "", "genre name")
	cmd.Flags().StringVar(&key, "key", "", "musical key")
	cmd.Flags().IntVar(&bpm, "bpm", 0, "tempo (defaults to genre range)")
	cmd.Flags().IntVar(&bars, "bars", 8, "number of bars")
	cmd.Flags().StringVar(&out, "out", "dj_loop.wav", "output WAV file")
	cmd.Flags().BoolVar(&stems, "stems", false, "render each layer as a separate stem file")
	return cmd
}

func bpmFor(genre string, bpm int) int {
	if bpm > 0 {
		return bpm
	}
	if g, err := dj.Stats(genre); err == nil {
		b := (g.BPMMin + g.BPMMax) / 2
		if b > 0 {
			return b
		}
	}
	return 120
}

func renderDJStems(genre, key string, bpm, bars int, out string) error {
	if bpm <= 0 {
		bpm = bpmFor(genre, 0)
	}
	secs := float64(bars) * djSecondsPerBar / float64(bpm)
	layers := []string{"kick", "snare", "hat", "bass", "lead", "pad"}
	dir := strings.TrimSuffix(out, filepath.Ext(out)) + "_stems"
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	for _, layer := range layers {
		pat, cps, _, err := dj.GenerateLayer(genre, layer, bars, 1, bpm, int64(os.Getpid()))
		if err != nil {
			return err
		}
		if pat.Query == nil {
			continue
		}
		path := filepath.Join(dir, layer+".wav")
		if err := renderPatternToFile(pat, cps, secs, path, false); err != nil {
			return err
		}
	}
	fmt.Printf("stems written to %s/\n", dir)
	return nil
}

// ---- dj song ----

func djSongCmd() *cobra.Command {
	var genre, structure, key, out string
	var bpm int
	cmd := &cobra.Command{
		Use:   "song",
		Short: "Generate a full song following a named structure",
		RunE: func(cmd *cobra.Command, args []string) error {
			if genre == "" {
				return fmt.Errorf("--genre is required")
			}
			if structure == "" {
				structure = "standard_edm"
			}
			if out == "" {
				out = "dj_song.wav"
			}
			sections, err := dj.GenerateSong(genre, structure, key, bpm, int64(os.Getpid()))
			if err != nil {
				return err
			}
			var all []float64
			for _, s := range sections {
				buf, e := audio.RenderPattern(s.Pattern, s.CPS, s.Seconds)
				if e != nil {
					return e
				}
				buf = normalizePeak(buf)
				all = append(all, buf...)
			}
			if err := audio.WriteWAVFile(out, all, audio.SampleRate); err != nil {
				return err
			}
			fmt.Printf("song (%s) -> %s  [%d sections]\n", structure, out, len(sections))
			return nil
		},
	}
	cmd.Flags().StringVar(&genre, "genre", "", "genre name")
	cmd.Flags().StringVar(&structure, "structure", "standard_edm", "song structure: standard_edm|dub_techno|minimal|breakbeat|ambient")
	cmd.Flags().StringVar(&key, "key", "", "musical key")
	cmd.Flags().IntVar(&bpm, "bpm", 0, "tempo (defaults to genre range)")
	cmd.Flags().StringVar(&out, "out", "dj_song.wav", "output WAV file")
	return cmd
}

// ---- dj layer ----

func djLayerCmd() *cobra.Command {
	var genre, layer, key, out string
	var bpm, bars, variations int
	cmd := &cobra.Command{
		Use:   "layer",
		Short: "Generate a single layer (kick/bass/lead/pad/...) for a genre",
		RunE: func(cmd *cobra.Command, args []string) error {
			if genre == "" {
				return fmt.Errorf("--genre is required")
			}
			if layer == "" {
				layer = "bass"
			}
			if bars <= 0 {
				bars = 4
			}
			if out == "" {
				out = "dj_layer.wav"
			}
			pat, cps, code, err := dj.GenerateLayer(genre, layer, bars, variations, bpm, int64(os.Getpid()))
			if err != nil {
				return err
			}
			fmt.Printf("code:\n%s\n", code)
			secs := float64(bars) * djSecondsPerBar / float64(bpmFor(genre, bpm))
			return renderPatternToFile(pat, cps, secs, out, false)
		},
	}
	cmd.Flags().StringVar(&genre, "genre", "", "genre name")
	cmd.Flags().StringVar(&layer, "layer", "bass", "layer: kick|snare|hat|bass|lead|pad")
	cmd.Flags().StringVar(&key, "key", "", "musical key")
	cmd.Flags().IntVar(&bpm, "bpm", 0, "tempo (defaults to genre range)")
	cmd.Flags().IntVar(&bars, "bars", 4, "number of bars")
	cmd.Flags().IntVar(&variations, "variations", 1, "number of variations")
	cmd.Flags().StringVar(&out, "out", "dj_layer.wav", "output WAV file")
	return cmd
}

// ---- dj morph ----

func djMorphCmd() *cobra.Command {
	var genreB, out string
	var bpm, steps int
	cmd := &cobra.Command{
		Use:   "morph <genreA> [genreB]",
		Short: "Interpolate between two genres by crossfading their loops",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			genreA := args[0]
			if genreB == "" && len(args) > 1 {
				genreB = args[1]
			}
			if genreB == "" {
				genreB = "house"
			}
			if steps <= 0 {
				steps = 8
			}
			if out == "" {
				out = "dj_morph.wav"
			}
			pa, pb, cps, err := dj.Morph(genreA, genreB, steps, bpm, int64(os.Getpid()))
			if err != nil {
				return err
			}
			secs := float64(steps) * djSecondsPerBar / float64(bpmFor(genreA, bpm))
			ba, err := audio.RenderPattern(pa, cps, secs)
			if err != nil {
				return err
			}
			bb, err := audio.RenderPattern(pb, cps, secs)
			if err != nil {
				return err
			}
			ba = normalizePeak(ba)
			bb = normalizePeak(bb)
			merged := crossfadeBuffers(ba, bb)
			if err := audio.WriteWAVFile(out, merged, audio.SampleRate); err != nil {
				return err
			}
			fmt.Printf("morphed %s -> %s -> %s\n", genreA, out, genreB)
			return nil
		},
	}
	cmd.Flags().StringVar(&genreB, "to", "", "second genre (or pass as 2nd arg)")
	cmd.Flags().IntVar(&bpm, "bpm", 128, "tempo")
	cmd.Flags().IntVar(&steps, "steps", 8, "number of bars (crossfade length)")
	cmd.Flags().StringVar(&out, "out", "dj_morph.wav", "output WAV file")
	return cmd
}

// ---- dj jam ----

func djJamCmd() *cobra.Command {
	var genre, key string
	var bpm int
	cmd := &cobra.Command{
		Use:   "jam",
		Short: "Generate a loop and play it (quick interactive-style session)",
		RunE: func(cmd *cobra.Command, args []string) error {
			if genre == "" {
				genre = "techno"
			}
			pat, cps, code, err := dj.GenerateLoop(genre, key, bpm, 8, int64(os.Getpid()))
			if err != nil {
				return err
			}
			fmt.Printf("jamming %s:\n%s\n", genre, code)
			out := "dj_jam.wav"
			return renderPatternToFile(pat, cps, 60, out, true)
		},
	}
	cmd.Flags().StringVar(&genre, "genre", "techno", "genre name")
	cmd.Flags().StringVar(&key, "key", "", "musical key")
	cmd.Flags().IntVar(&bpm, "bpm", 0, "tempo (defaults to genre range)")
	return cmd
}

// ---- dj export ----

func djExportCmd() *cobra.Command {
	var genre, format, out string
	var bpm int
	cmd := &cobra.Command{
		Use:   "export",
		Short: "Export a genre model (json) or a generated loop (midi)",
		RunE: func(cmd *cobra.Command, args []string) error {
			if genre == "" {
				return fmt.Errorf("--genre is required")
			}
			if out == "" {
				out = "genre_" + djName(genre) + "." + exportExt(format)
			}
			switch strings.ToLower(format) {
			case "midi":
				pat, cps, _, err := dj.GenerateLoop(genre, "", bpm, 8, int64(os.Getpid()))
				if err != nil {
					return err
				}
				notes := patternNotesMIDI(pat, cps, 16)
				if err := writeMIDIFile(out, notes); err != nil {
					return err
				}
				fmt.Printf("exported MIDI loop -> %s (%d notes)\n", out, len(notes))
				return nil
			case "dawproject":
				fmt.Fprintln(os.Stderr, "warning: dawproject export planned; writing model JSON instead")
				fallthrough
			default:
				g, err := dj.Stats(genre)
				if err != nil {
					g = dj.DefaultProfileFor(genre)
				}
				if err := saveJSONFile(out, g); err != nil {
					return err
				}
				fmt.Printf("exported model JSON -> %s\n", out)
				return nil
			}
		},
	}
	cmd.Flags().StringVar(&genre, "genre", "", "genre name")
	cmd.Flags().StringVar(&format, "format", "json", "export format: json|midi|dawproject")
	cmd.Flags().IntVar(&bpm, "bpm", 0, "tempo for MIDI loop export")
	cmd.Flags().StringVar(&out, "out", "", "output file")
	return cmd
}

func djName(g string) string {
	g = strings.ToLower(strings.TrimSpace(g))
	var b strings.Builder
	for _, r := range g {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
		} else {
			b.WriteRune('_')
		}
	}
	return b.String()
}

func exportExt(format string) string {
	switch strings.ToLower(format) {
	case "midi":
		return "mid"
	case "dawproject":
		return "dawproject"
	default:
		return "json"
	}
}

// ---- dj recommend ----

func djRecommendCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "recommend <track.cdc>",
		Short: "Recommend a genre for a given track",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			best, err := dj.Recommend(args[0])
			if err != nil {
				return err
			}
			fmt.Printf("recommended genre: %s\n", best)
			return nil
		},
	}
}

// crossfadeBuffers blends a into b linearly across the shared length.
func crossfadeBuffers(a, b []float64) []float64 {
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	out := make([]float64, n)
	for i := 0; i < n; i++ {
		t := float64(i) / float64(n)
		g := smoothstep(0, 1, t)
		out[i] = a[i]*(1-g)*0.7 + b[i]*g*0.7
	}
	return out
}

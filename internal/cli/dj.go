package cli

import (
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/Jaziel8910/codic/internal/audio"
)

// DJState holds the virtual decks and queue.
type DJState struct {
	Decks map[string]string `json:"decks"` // deck id -> .cdc file
	Queue []string          `json:"queue"` // queued .cdc files
}

func djLoadState() (*DJState, error) {
	var s DJState
	if err := registryLoad("dj", "state", &s); err != nil {
		s = DJState{Decks: map[string]string{}, Queue: []string{}}
	}
	if s.Decks == nil {
		s.Decks = map[string]string{}
	}
	return &s, nil
}

func djSaveState(s *DJState) error {
	return registrySave("dj", "state", s)
}

func djRender(file string, secs float64) ([]float64, error) {
	combined, cps, err := collectPatternsFromFile(file)
	if err != nil {
		return nil, err
	}
	if secs <= 0 {
		secs = 16
	}
	buf, err := audio.RenderPattern(combined, cps, secs)
	if err != nil {
		return nil, err
	}
	return normalizePeak(buf), nil
}

func djCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "dj",
		Short: "DJ mode: procedural generation, genre learning and live-style decks (offline)",
	}
	// Procedural generation (Fase 4)
	cmd.AddCommand(djLearnCmd())
	cmd.AddCommand(djForgetCmd())
	cmd.AddCommand(djListCmd())
	cmd.AddCommand(djStatsCmd())
	cmd.AddCommand(djPlayCmd())
	cmd.AddCommand(djLoopCmd())
	cmd.AddCommand(djSongCmd())
	cmd.AddCommand(djLayerCmd())
	cmd.AddCommand(djMorphCmd())
	cmd.AddCommand(djJamCmd())
	cmd.AddCommand(djExportCmd())
	cmd.AddCommand(djRecommendCmd())
	// Live-style decks (mixing, cueing, crossfade)
	cmd.AddCommand(djLiveCmd())
	return cmd
}

// djLiveCmd groups the deck/mixing style operations under `dj live` so they do
// not collide with the procedural-generation subcommands above.
func djLiveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "live",
		Short: "Live-style deck operations: cue, mix, crossfade, scratch",
	}
	cmd.AddCommand(djCueCmd())
	cmd.AddCommand(djPlayDeckCmd())
	cmd.AddCommand(djMixCmd())
	cmd.AddCommand(djCrossfadeCmd())
	cmd.AddCommand(djBeatmatchCmd())
	cmd.AddCommand(djLoopDeckCmd())
	cmd.AddCommand(djScratchCmd())
	cmd.AddCommand(djFxCmd())
	cmd.AddCommand(djDeckCmd())
	cmd.AddCommand(djQueueCmd())
	cmd.AddCommand(djRecordCmd())
	cmd.AddCommand(djBroadcastCmd())
	return cmd
}

func djCueCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "cue <deck> <file.cdc>",
		Short: "Load a track onto a deck",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			s, err := djLoadState()
			if err != nil {
				return err
			}
			s.Decks[args[0]] = args[1]
			if err := djSaveState(s); err != nil {
				return err
			}
			fmt.Printf("deck %s cued: %s\n", args[0], args[1])
			return nil
		},
	}
}

func djPlayDeckCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "play <deck>",
		Short: "Play the track on a deck",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			s, err := djLoadState()
			if err != nil {
				return err
			}
			file, ok := s.Decks[args[0]]
			if !ok {
				return fmt.Errorf("deck %s is empty", args[0])
			}
			buf, err := djRender(file, 16)
			if err != nil {
				return err
			}
			out := "deck_" + args[0] + ".wav"
			if err := audio.WriteWAVFile(out, buf, audio.SampleRate); err != nil {
				return err
			}
			fmt.Printf("playing deck %s -> %s\n", args[0], out)
			return openPlayer(out)
		},
	}
}

func djMixCmd() *cobra.Command {
	var secs float64
	cmd := &cobra.Command{
		Use:   "mix <deckA> <deckB> [out.wav]",
		Short: "Mix two decks into one continuous track",
		Args:  cobra.RangeArgs(2, 3),
		RunE: func(cmd *cobra.Command, args []string) error {
			s, err := djLoadState()
			if err != nil {
				return err
			}
			a, ok := s.Decks[args[0]]
			if !ok {
				return fmt.Errorf("deck %s empty", args[0])
			}
			b, ok := s.Decks[args[1]]
			if !ok {
				return fmt.Errorf("deck %s empty", args[1])
			}
			ba, err := djRender(a, secs)
			if err != nil {
				return err
			}
			bb, err := djRender(b, secs)
			if err != nil {
				return err
			}
			out := "mix.wav"
			if len(args) > 2 {
				out = args[2]
			}
			if err := audio.WriteWAVFile(out, mixBuffers(ba, bb), audio.SampleRate); err != nil {
				return err
			}
			fmt.Printf("mixed %s + %s -> %s\n", args[0], args[1], out)
			return nil
		},
	}
	cmd.Flags().Float64Var(&secs, "sec", 16, "seconds per deck")
	return cmd
}

func djCrossfadeCmd() *cobra.Command {
	var pos float64
	cmd := &cobra.Command{
		Use:   "crossfade <deckA> <deckB> [out.wav]",
		Short: "Crossfade two decks over the whole mix",
		Args:  cobra.RangeArgs(2, 3),
		RunE: func(cmd *cobra.Command, args []string) error {
			s, err := djLoadState()
			if err != nil {
				return err
			}
			a, ok := s.Decks[args[0]]
			if !ok {
				return fmt.Errorf("deck %s empty", args[0])
			}
			b, ok := s.Decks[args[1]]
			if !ok {
				return fmt.Errorf("deck %s empty", args[1])
			}
			ba, _ := djRender(a, 16)
			bb, _ := djRender(b, 16)
			out := "crossfade.wav"
			if len(args) > 2 {
				out = args[2]
			}
			n := len(ba)
			if len(bb) < n {
				n = len(bb)
			}
			res := make([]float64, n)
			for i := 0; i < n; i++ {
				t := float64(i) / float64(n)
				gainA := 1 - smoothstep(0, pos+0.3, t)
				gainB := smoothstep(pos-0.3, 1, t)
				res[i] = ba[i]*gainA*0.7 + bb[i]*gainB*0.7
			}
			if err := audio.WriteWAVFile(out, res, audio.SampleRate); err != nil {
				return err
			}
			fmt.Printf("crossfaded -> %s\n", out)
			return nil
		},
	}
	cmd.Flags().Float64Var(&pos, "pos", 0.5, "crossfade center (0..1)")
	return cmd
}

func djBeatmatchCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "beatmatch <deckA> <deckB>",
		Short: "Compare the tempos of two decks",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			s, err := djLoadState()
			if err != nil {
				return err
			}
			a, ok := s.Decks[args[0]]
			if !ok {
				return fmt.Errorf("deck %s empty", args[0])
			}
			b, ok := s.Decks[args[1]]
			if !ok {
				return fmt.Errorf("deck %s empty", args[1])
			}
			ba, _ := os.ReadFile(a)
			bb, _ := os.ReadFile(b)
			pa, _ := bpmOf(string(ba))
			pb, _ := bpmOf(string(bb))
			fmt.Printf("deck %s: %.0f BPM\n", args[0], pa)
			fmt.Printf("deck %s: %.0f BPM\n", args[1], pb)
			if pa > 0 && pb > 0 {
				diff := pa - pb
				fmt.Printf("difference: %+.1f BPM (nudge deck b by %.1f%%)\n", diff, -diff/pb*100)
			}
			return nil
		},
	}
}

func djLoopDeckCmd() *cobra.Command {
	var secs float64
	cmd := &cobra.Command{
		Use:   "loop <deck> [out.wav]",
		Short: "Render a looping section of a deck",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			s, err := djLoadState()
			if err != nil {
				return err
			}
			file, ok := s.Decks[args[0]]
			if !ok {
				return fmt.Errorf("deck %s empty", args[0])
			}
			buf, err := djRender(file, secs)
			if err != nil {
				return err
			}
			out := "deck_" + args[0] + "_loop.wav"
			if len(args) > 1 {
				out = args[1]
			}
			if err := audio.WriteWAVFile(out, buf, audio.SampleRate); err != nil {
				return err
			}
			fmt.Printf("looped deck %s -> %s\n", args[0], out)
			return nil
		},
	}
	cmd.Flags().Float64Var(&secs, "sec", 4, "loop length in seconds")
	return cmd
}

func djScratchCmd() *cobra.Command {
	var secs float64
	cmd := &cobra.Command{
		Use:   "scratch <deck> [out.wav]",
		Short: "Apply a scratch-like pitch envelope to a deck",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			s, err := djLoadState()
			if err != nil {
				return err
			}
			file, ok := s.Decks[args[0]]
			if !ok {
				return fmt.Errorf("deck %s empty", args[0])
			}
			buf, err := djRender(file, secs)
			if err != nil {
				return err
			}
			n := len(buf)
			out := make([]float64, n)
			for i := 0; i < n; i++ {
				t := float64(i) / float64(n)
				wob := 1.0 + 0.4*math.Sin(2*math.Pi*3*t)
				src := int(float64(i) * wob)
				if src >= 0 && src < n {
					out[i] = buf[src]
				}
			}
			wav := "deck_" + args[0] + "_scratch.wav"
			if len(args) > 1 {
				wav = args[1]
			}
			if err := audio.WriteWAVFile(wav, out, audio.SampleRate); err != nil {
				return err
			}
			fmt.Printf("scratched deck %s -> %s\n", args[0], wav)
			return nil
		},
	}
	cmd.Flags().Float64Var(&secs, "sec", 6, "length in seconds")
	return cmd
}

func djFxCmd() *cobra.Command {
	var secs float64
	cmd := &cobra.Command{
		Use:   "fx <deck> <fxname> [out.wav]",
		Short: "Apply an FX preset to a deck",
		Args:  cobra.RangeArgs(2, 3),
		RunE: func(cmd *cobra.Command, args []string) error {
			s, err := djLoadState()
			if err != nil {
				return err
			}
			file, ok := s.Decks[args[0]]
			if !ok {
				return fmt.Errorf("deck %s empty", args[0])
			}
			var f Fx
			if err := registryLoad("fx", args[1], &f); err != nil {
				return errNotFound("fx", args[1])
			}
			buf, err := djRender(file, secs)
			if err != nil {
				return err
			}
			if f.Type == "distortion" {
				for i := range buf {
					buf[i] = clampf(buf[i]*1.5, -0.95, 0.95)
				}
			}
			out := "deck_" + args[0] + "_" + args[1] + ".wav"
			if len(args) > 2 {
				out = args[2]
			}
			if err := audio.WriteWAVFile(out, buf, audio.SampleRate); err != nil {
				return err
			}
			fmt.Printf("applied fx %q to deck %s -> %s\n", args[1], args[0], out)
			return nil
		},
	}
	cmd.Flags().Float64Var(&secs, "sec", 8, "length in seconds")
	return cmd
}

func djDeckCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "deck [deck] [file.cdc]",
		Short: "Show a deck or set its track",
		Args:  cobra.RangeArgs(0, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			s, err := djLoadState()
			if err != nil {
				return err
			}
			if len(args) == 0 {
				if len(s.Decks) == 0 {
					fmt.Println("(no decks loaded)")
					return nil
				}
				for k, v := range s.Decks {
					fmt.Printf("  %s -> %s\n", k, v)
				}
				return nil
			}
			if len(args) == 1 {
				if v, ok := s.Decks[args[0]]; ok {
					fmt.Printf("deck %s -> %s\n", args[0], v)
				} else {
					fmt.Printf("deck %s is empty\n", args[0])
				}
				return nil
			}
			s.Decks[args[0]] = args[1]
			if err := djSaveState(s); err != nil {
				return err
			}
			fmt.Printf("deck %s set -> %s\n", args[0], args[1])
			return nil
		},
	}
}

func djQueueCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "queue [file.cdc]",
		Short: "Show or add to the play queue",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			s, err := djLoadState()
			if err != nil {
				return err
			}
			if len(args) > 0 {
				s.Queue = append(s.Queue, args[0])
				if err := djSaveState(s); err != nil {
					return err
				}
				fmt.Printf("queued %s (position %d)\n", args[0], len(s.Queue))
				return nil
			}
			if len(s.Queue) == 0 {
				fmt.Println("(queue empty)")
				return nil
			}
			for i, q := range s.Queue {
				fmt.Printf("  [%d] %s\n", i, q)
			}
			return nil
		},
	}
}

func djRecordCmd() *cobra.Command {
	var secs float64
	cmd := &cobra.Command{
		Use:   "record [out.wav]",
		Short: "Render the entire queue into one WAV",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			s, err := djLoadState()
			if err != nil {
				return err
			}
			if len(s.Queue) == 0 {
				return fmt.Errorf("queue is empty")
			}
			var all []float64
			for _, q := range s.Queue {
				buf, e := djRender(q, secs)
				if e != nil {
					return e
				}
				all = append(all, buf...)
				for i := 0; i < audio.SampleRate; i++ {
					all = append(all, 0)
				}
			}
			out := "recording.wav"
			if len(args) > 0 {
				out = args[0]
			}
			if err := audio.WriteWAVFile(out, all, audio.SampleRate); err != nil {
				return err
			}
			fmt.Printf("recorded %d tracks -> %s\n", len(s.Queue), out)
			return nil
		},
	}
	cmd.Flags().Float64Var(&secs, "sec", 8, "seconds per queued track")
	return cmd
}

func djBroadcastCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "broadcast",
		Short: "Information about live broadcasting",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("Broadcasting requires a streaming server (Icecast/RTMP).")
			fmt.Println("Render your set with `codic dj record set.wav` then stream it")
			fmt.Println("with a tool such as ffmpeg or butt.")
			return nil
		},
	}
}

func mixBuffers(a, b []float64) []float64 {
	n := len(a)
	if len(b) > n {
		n = len(b)
	}
	out := make([]float64, n)
	for i := 0; i < n; i++ {
		var va, vb float64
		if i < len(a) {
			va = a[i]
		}
		if i < len(b) {
			vb = b[i]
		}
		out[i] = clampf(va*0.6+vb*0.6, -0.95, 0.95)
	}
	return out
}

func smoothstep(edge0, edge1, x float64) float64 {
	if edge1 == edge0 {
		return 0
	}
	t := clampf((x-edge0)/(edge1-edge0), 0, 1)
	return t * t * (3 - 2*t)
}

func clampf(x, lo, hi float64) float64 {
	if x < lo {
		return lo
	}
	if x > hi {
		return hi
	}
	return x
}

func bpmOf(src string) (float64, error) {
	for _, line := range strings.Split(src, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "@bpm") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				if v, err := strconv.ParseFloat(fields[1], 64); err == nil {
					return v, nil
				}
			}
		}
	}
	return 0, nil
}

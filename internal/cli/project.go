package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/Jaziel8910/codic/internal/audio"
	"github.com/Jaziel8910/codic/internal/codang"
	"github.com/Jaziel8910/codic/internal/project"
)

// ProjectFile is the on-disk representation of an album/EP/playlist.
type ProjectFile struct {
	Project  string         `yaml:"project"`
	Type     string         `yaml:"type"`
	BPM      float64        `yaml:"bpm"`
	Key      string         `yaml:"key"`
	Tracks   []ProjectTrack `yaml:"tracks"`
	Metadata ProjectMeta    `yaml:"metadata"`
}

// ProjectTrack is one entry in the project track list.
type ProjectTrack struct {
	Name     string   `yaml:"name"`
	File     string   `yaml:"file"`
	BPM      float64  `yaml:"bpm"`
	Duration string   `yaml:"duration"`
	Tags     []string `yaml:"tags"`
	Muted    bool     `yaml:"muted"`
	Solo     bool     `yaml:"solo"`
	Color    string   `yaml:"color"`
}

// ProjectMeta holds release metadata.
type ProjectMeta struct {
	Artist      string `yaml:"artist"`
	Label       string `yaml:"label"`
	ReleaseDate string `yaml:"release_date"`
	Cover       string `yaml:"cover"`
}

func projectCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "project",
		Short: "Manage albums / EPs / playlists (project.yaml)",
	}
	cmd.AddCommand(projectInitCmd())
	cmd.AddCommand(projectInfoCmd())
	cmd.AddCommand(projectAddCmd())
	cmd.AddCommand(projectRmCmd())
	cmd.AddCommand(projectReorderCmd())
	cmd.AddCommand(projectPlayCmd())
	cmd.AddCommand(projectRenderCmd())
	cmd.AddCommand(projectExportCmd())
	cmd.AddCommand(projectValidateCmd())
	cmd.AddCommand(projectStatusCmd())
	cmd.AddCommand(projectTagCmd())
	cmd.AddCommand(projectMetadataCmd())
	cmd.AddCommand(projectSplitCmd())
	cmd.AddCommand(projectMergeCmd())
	cmd.AddCommand(projectPlaylistCmd())
	return cmd
}

func loadProjectFile(path string) (*ProjectFile, error) {
	var pf ProjectFile
	if err := loadJSONFile(path, &pf); err != nil {
		// try YAML
		data, e := os.ReadFile(path)
		if e != nil {
			return nil, e
		}
		if err := yaml.Unmarshal(data, &pf); err != nil {
			return nil, err
		}
	}
	return &pf, nil
}

func saveProjectFile(path string, pf *ProjectFile) error {
	data, err := yaml.Marshal(pf)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

func projectYAMLPath(dir string) string {
	if dir == "" {
		dir = "."
	}
	return filepath.Join(dir, "project.yaml")
}

// renderProjectBuffer renders every track sequentially (with gaps) into one
// interleaved stereo float buffer.
func renderProjectBuffer(pf *ProjectFile, gapSec float64, normalize bool) ([]float64, error) {
	var out []float64
	gap := make([]float64, int(gapSec*float64(audio.SampleRate))*2)
	for i, t := range pf.Tracks {
		code, err := os.ReadFile(t.File)
		if err != nil {
			return nil, fmt.Errorf("track %q: %w", t.File, err)
		}
		combined, cps, err := collectPatterns(string(code))
		if err != nil {
			return nil, fmt.Errorf("track %q: %w", t.Name, err)
		}
		secs := trackSeconds(t, cps)
		buf, err := audio.RenderPattern(combined, cps, secs)
		if err != nil {
			return nil, err
		}
		out = append(out, buf...)
		if i < len(pf.Tracks)-1 && gapSec > 0 {
			out = append(out, gap...)
		}
	}
	if normalize && len(out) > 0 {
		out = normalizePeak(out)
	}
	return out, nil
}

func trackSeconds(t ProjectTrack, cps float64) float64 {
	if t.Duration != "" {
		if s, ok := parseClock(t.Duration); ok {
			return s
		}
	}
	if cps > 0 {
		// reasonable default: ~16 cycles
		return 16 / cps
	}
	return 8.0
}

// parseClock parses "m:ss" or "h:mm:ss" into seconds.
func parseClock(s string) (float64, bool) {
	parts := strings.Split(s, ":")
	var secs float64
	mult := 1.0
	for i := len(parts) - 1; i >= 0; i-- {
		f, err := strconv.ParseFloat(strings.TrimSpace(parts[i]), 64)
		if err != nil {
			return 0, false
		}
		secs += f * mult
		mult *= 60
	}
	if mult == 1 {
		return 0, false
	}
	return secs, true
}

func buildEngineProject(pf *ProjectFile) (*project.Project, error) {
	p := project.New(pf.Project)
	if pf.BPM > 0 {
		p.SetBPM(pf.BPM)
	}
	for _, t := range pf.Tracks {
		code, err := os.ReadFile(t.File)
		if err != nil {
			return nil, err
		}
		combined, _, err := collectPatterns(string(code))
		if err != nil {
			return nil, err
		}
		opts := []project.TrackOption{}
		if t.BPM > 0 {
			// approximate length from duration or default
		}
		p.AddPattern(t.Name, combined, opts...)
	}
	return p, nil
}

// ---------- commands ----------

func projectInitCmd() *cobra.Command {
	var typ string
	cmd := &cobra.Command{
		Use:   "init [name]",
		Short: "Create a project structure + project.yaml",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := "Untitled"
			dir := "."
			if len(args) > 0 {
				name = args[0]
				if err := os.MkdirAll(filepath.Join(dir, "tracks"), 0o755); err != nil {
					return err
				}
			}
			pf := ProjectFile{
				Project:  name,
				Type:     typ,
				BPM:      120,
				Key:      "c minor",
				Tracks:   []ProjectTrack{},
				Metadata: ProjectMeta{},
			}
			if err := saveProjectFile(projectYAMLPath(dir), &pf); err != nil {
				return err
			}
			fmt.Printf("created project %q (%s) at %s\n", name, typ, projectYAMLPath(dir))
			return nil
		},
	}
	cmd.Flags().StringVar(&typ, "type", "album", "project type: album|ep|single|playlist")
	return cmd
}

func projectInfoCmd() *cobra.Command {
	var asJSON bool
	cmd := &cobra.Command{
		Use:   "info [dir]",
		Short: "Show project metadata",
		RunE: func(cmd *cobra.Command, args []string) error {
			dir := "."
			if len(args) > 0 {
				dir = args[0]
			}
			pf, err := loadProjectFile(projectYAMLPath(dir))
			if err != nil {
				return err
			}
			if asJSON {
				b, _ := marshalJSON(pf)
				fmt.Println(b)
				return nil
			}
			fmt.Printf("Project: %s\n", pf.Project)
			fmt.Printf("Type    : %s\n", pf.Type)
			fmt.Printf("BPM     : %.0f\n", pf.BPM)
			fmt.Printf("Key     : %s\n", pf.Key)
			fmt.Printf("Tracks  : %d\n", len(pf.Tracks))
			fmt.Printf("Artist  : %s\n", pf.Metadata.Artist)
			fmt.Printf("Label   : %s\n", pf.Metadata.Label)
			return nil
		},
	}
	cmd.Flags().BoolVar(&asJSON, "json", false, "output as JSON")
	return cmd
}

func projectAddCmd() *cobra.Command {
	var pos int
	cmd := &cobra.Command{
		Use:   "add <track.cdc> [dir]",
		Short: "Add a track to the project",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			dir := "."
			if len(args) > 1 {
				dir = args[1]
			}
			pf, err := loadProjectFile(projectYAMLPath(dir))
			if err != nil {
				return err
			}
			t := ProjectTrack{Name: strings.TrimSuffix(filepath.Base(args[0]), ".cdc"), File: args[0]}
			// try to read @bpm / @key metadata
			if data, e := os.ReadFile(args[0]); e == nil {
				if prog, e2 := codang.Parse(string(data)); e2 == nil {
					if v, ok := prog.Metadata["bpm"]; ok {
						if f, err := parseFloat(v); err == nil {
							t.BPM = f
						}
					}
					if v, ok := prog.Metadata["key"]; ok {
						pf.Key = v
					}
				}
			}
			if pos >= 0 && pos < len(pf.Tracks) {
				pf.Tracks = append(pf.Tracks[:pos+1], append([]ProjectTrack{t}, pf.Tracks[pos+1:]...)...)
			} else {
				pf.Tracks = append(pf.Tracks, t)
			}
			if err := saveProjectFile(projectYAMLPath(dir), pf); err != nil {
				return err
			}
			fmt.Printf("added track %q (total %d)\n", t.Name, len(pf.Tracks))
			return nil
		},
	}
	cmd.Flags().IntVar(&pos, "pos", -1, "insert position (default: append)")
	return cmd
}

func projectRmCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "rm <track> [dir]",
		Short: "Remove a track from the project",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			dir := "."
			if len(args) > 1 {
				dir = args[1]
			}
			pf, err := loadProjectFile(projectYAMLPath(dir))
			if err != nil {
				return err
			}
			idx := trackIndex(pf, args[0])
			if idx < 0 {
				return errNotFound("track", args[0])
			}
			pf.Tracks = append(pf.Tracks[:idx], pf.Tracks[idx+1:]...)
			if err := saveProjectFile(projectYAMLPath(dir), pf); err != nil {
				return err
			}
			fmt.Printf("removed track %q (total %d)\n", args[0], len(pf.Tracks))
			return nil
		},
	}
}

func projectReorderCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "reorder <from> <to> [dir]",
		Short: "Reorder a track",
		Args:  cobra.RangeArgs(2, 3),
		RunE: func(cmd *cobra.Command, args []string) error {
			dir := "."
			if len(args) > 2 {
				dir = args[2]
			}
			from, err1 := strconv.Atoi(args[0])
			to, err2 := strconv.Atoi(args[1])
			if err1 != nil || err2 != nil {
				return fmt.Errorf("from/to must be integers")
			}
			pf, err := loadProjectFile(projectYAMLPath(dir))
			if err != nil {
				return err
			}
			if from < 0 || from >= len(pf.Tracks) || to < 0 || to >= len(pf.Tracks) {
				return fmt.Errorf("index out of range")
			}
			t := pf.Tracks[from]
			pf.Tracks = append(pf.Tracks[:from], pf.Tracks[from+1:]...)
			pf.Tracks = append(pf.Tracks[:to], append([]ProjectTrack{t}, pf.Tracks[to:]...)...)
			if err := saveProjectFile(projectYAMLPath(dir), pf); err != nil {
				return err
			}
			fmt.Printf("moved track to position %d\n", to)
			return nil
		},
	}
}

func projectPlayCmd() *cobra.Command {
	var from, gap float64
	var loop, shuffle bool
	cmd := &cobra.Command{
		Use:   "play [dir]",
		Short: "Render the album and play it in the OS player",
		RunE: func(cmd *cobra.Command, args []string) error {
			dir := "."
			if len(args) > 0 {
				dir = args[0]
			}
			pf, err := loadProjectFile(projectYAMLPath(dir))
			if err != nil {
				return err
			}
			tracks := pf.Tracks
			if from > 0 && int(from) < len(tracks) {
				tracks = tracks[int(from):]
			}
			if shuffle {
				tracks = shuffleTracks(tracks)
			}
			pf.Tracks = tracks
			buf, err := renderProjectBuffer(pf, gap, true)
			if err != nil {
				return err
			}
			tmp, err := tempWAV("album_preview.wav")
			if err != nil {
				return err
			}
			if err := audio.WriteWAVFile(tmp, buf, audio.SampleRate); err != nil {
				return err
			}
			fmt.Printf("rendered %d tracks -> %s\n", len(pf.Tracks), tmp)
			if loop {
				fmt.Println("(loop mode is a no-op for offline preview)")
			}
			return openPlayer(tmp)
		},
	}
	cmd.Flags().Float64Var(&from, "from", 0, "start from track N")
	cmd.Flags().Float64Var(&gap, "gap", 0, "gap in seconds between tracks")
	cmd.Flags().BoolVar(&loop, "loop", false, "loop the album")
	cmd.Flags().BoolVar(&shuffle, "shuffle", false, "shuffle track order")
	return cmd
}

func projectRenderCmd() *cobra.Command {
	var gap float64
	var normalize bool
	cmd := &cobra.Command{
		Use:   "render [out.wav] [dir]",
		Short: "Render the whole album to one WAV",
		Args:  cobra.RangeArgs(0, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			dir := "."
			out := "album.wav"
			if len(args) > 0 {
				out = args[0]
			}
			if len(args) > 1 {
				dir = args[1]
			}
			pf, err := loadProjectFile(projectYAMLPath(dir))
			if err != nil {
				return err
			}
			buf, err := renderProjectBuffer(pf, gap, normalize)
			if err != nil {
				return err
			}
			if err := audio.WriteWAVFile(out, buf, audio.SampleRate); err != nil {
				return err
			}
			fmt.Printf("rendered album (%d tracks) -> %s\n", len(pf.Tracks), out)
			return nil
		},
	}
	cmd.Flags().Float64Var(&gap, "gap", 1.0, "gap in seconds between tracks")
	cmd.Flags().BoolVar(&normalize, "normalize", true, "peak-normalize the output")
	return cmd
}

func projectExportCmd() *cobra.Command {
	var format string
	cmd := &cobra.Command{
		Use:   "export [out] [dir]",
		Short: "Export the full album (dawproject/wav/stems/midi)",
		Args:  cobra.RangeArgs(0, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			dir := "."
			if len(args) > 1 {
				dir = args[1]
			}
			pf, err := loadProjectFile(projectYAMLPath(dir))
			if err != nil {
				return err
			}
			switch format {
			case "dawproject":
				out := "album.dawproject"
				if len(args) > 0 {
					out = args[0]
				}
				p, err := buildEngineProject(pf)
				if err != nil {
					return err
				}
				if err := p.ExportDAWProject(out); err != nil {
					return err
				}
				fmt.Printf("exported DAWproject -> %s (%d tracks)\n", out, len(p.Tracks))
			case "wav":
				out := "album.wav"
				if len(args) > 0 {
					out = args[0]
				}
				buf, err := renderProjectBuffer(pf, 0, true)
				if err != nil {
					return err
				}
				if err := audio.WriteWAVFile(out, buf, audio.SampleRate); err != nil {
					return err
				}
				fmt.Printf("exported WAV -> %s\n", out)
			default:
				return fmt.Errorf("format %q not supported yet (use dawproject|wav)", format)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&format, "format", "dawproject", "export format: dawproject|wav|stems|midi")
	return cmd
}

func projectValidateCmd() *cobra.Command {
	var fix bool
	cmd := &cobra.Command{
		Use:   "validate [dir]",
		Short: "Verify project integrity",
		RunE: func(cmd *cobra.Command, args []string) error {
			dir := "."
			if len(args) > 0 {
				dir = args[0]
			}
			pf, err := loadProjectFile(projectYAMLPath(dir))
			if err != nil {
				return err
			}
			bad := 0
			for i, t := range pf.Tracks {
				if _, err := os.Stat(t.File); err != nil {
					fmt.Printf("[FAIL] track %d (%s): file %q missing\n", i, t.Name, t.File)
					bad++
					if fix {
						pf.Tracks = append(pf.Tracks[:i], pf.Tracks[i+1:]...)
						i--
					}
				}
			}
			if bad == 0 {
				fmt.Println("project valid")
				return nil
			}
			if fix {
				if err := saveProjectFile(projectYAMLPath(dir), pf); err != nil {
					return err
				}
				fmt.Printf("fixed: removed %d missing tracks\n", bad)
			}
			return fmt.Errorf("%d problem(s) found", bad)
		},
	}
	cmd.Flags().BoolVar(&fix, "fix", false, "remove missing-track entries")
	return cmd
}

func projectStatusCmd() *cobra.Command {
	var tree bool
	cmd := &cobra.Command{
		Use:   "status [dir]",
		Short: "Show a tree of tracks with durations",
		RunE: func(cmd *cobra.Command, args []string) error {
			dir := "."
			if len(args) > 0 {
				dir = args[0]
			}
			pf, err := loadProjectFile(projectYAMLPath(dir))
			if err != nil {
				return err
			}
			fmt.Printf("%s/ (%s, %d tracks)\n", pf.Project, pf.Type, len(pf.Tracks))
			for i, t := range pf.Tracks {
				dur := t.Duration
				if dur == "" {
					dur = "?"
				}
				indent := "  "
				if tree {
					indent = "â”œâ”€ "
				}
				fmt.Printf("%s[%d] %s  (%s)\n", indent, i, t.Name, dur)
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&tree, "tree", false, "tree layout")
	return cmd
}

func projectTagCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "tag <track> <tag> [dir]",
		Short: "Tag a track",
		Args:  cobra.RangeArgs(2, 3),
		RunE: func(cmd *cobra.Command, args []string) error {
			dir := "."
			if len(args) > 2 {
				dir = args[2]
			}
			pf, err := loadProjectFile(projectYAMLPath(dir))
			if err != nil {
				return err
			}
			idx := trackIndex(pf, args[0])
			if idx < 0 {
				return errNotFound("track", args[0])
			}
			pf.Tracks[idx].Tags = appendUnique(pf.Tracks[idx].Tags, args[1])
			if err := saveProjectFile(projectYAMLPath(dir), pf); err != nil {
				return err
			}
			fmt.Printf("tagged %q with %q\n", args[0], args[1])
			return nil
		},
	}
}

func projectMetadataCmd() *cobra.Command {
	var artist, label, cover, release string
	cmd := &cobra.Command{
		Use:   "metadata [dir]",
		Short: "Edit release metadata",
		RunE: func(cmd *cobra.Command, args []string) error {
			dir := "."
			if len(args) > 0 {
				dir = args[0]
			}
			pf, err := loadProjectFile(projectYAMLPath(dir))
			if err != nil {
				return err
			}
			if artist != "" {
				pf.Metadata.Artist = artist
			}
			if label != "" {
				pf.Metadata.Label = label
			}
			if cover != "" {
				pf.Metadata.Cover = cover
			}
			if release != "" {
				pf.Metadata.ReleaseDate = release
			}
			if err := saveProjectFile(projectYAMLPath(dir), pf); err != nil {
				return err
			}
			fmt.Println("metadata updated")
			return nil
		},
	}
	cmd.Flags().StringVar(&artist, "artist", "", "artist name")
	cmd.Flags().StringVar(&label, "label", "", "label name")
	cmd.Flags().StringVar(&cover, "cover", "", "cover image path")
	cmd.Flags().StringVar(&release, "release", "", "release date")
	return cmd
}

func projectSplitCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "split <N> [dir]",
		Short: "Split an album into two projects at track N",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			dir := "."
			if len(args) > 1 {
				dir = args[0]
			}
			n, err := strconv.Atoi(args[len(args)-1])
			if err != nil {
				return fmt.Errorf("N must be an integer")
			}
			pf, err := loadProjectFile(projectYAMLPath(dir))
			if err != nil {
				return err
			}
			if n < 1 || n >= len(pf.Tracks) {
				return fmt.Errorf("N out of range (1..%d)", len(pf.Tracks)-1)
			}
			a := *pf
			a.Tracks = append([]ProjectTrack{}, pf.Tracks[:n]...)
			a.Project = pf.Project + " (A)"
			b := *pf
			b.Tracks = append([]ProjectTrack{}, pf.Tracks[n:]...)
			b.Project = pf.Project + " (B)"
			if err := saveProjectFile(filepath.Join(dir, "project_a.yaml"), &a); err != nil {
				return err
			}
			if err := saveProjectFile(filepath.Join(dir, "project_b.yaml"), &b); err != nil {
				return err
			}
			fmt.Println("split into project_a.yaml and project_b.yaml")
			return nil
		},
	}
}

func projectMergeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "merge <project2> [dir]",
		Short: "Merge another project's tracks into this one",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			dir := "."
			if len(args) > 1 {
				dir = args[1]
			}
			pf, err := loadProjectFile(projectYAMLPath(dir))
			if err != nil {
				return err
			}
			other, err := loadProjectFile(args[0])
			if err != nil {
				return err
			}
			pf.Tracks = append(pf.Tracks, other.Tracks...)
			if err := saveProjectFile(projectYAMLPath(dir), pf); err != nil {
				return err
			}
			fmt.Printf("merged %d tracks (total %d)\n", len(other.Tracks), len(pf.Tracks))
			return nil
		},
	}
}

// ---------- helpers ----------

func trackIndex(pf *ProjectFile, name string) int {
	for i, t := range pf.Tracks {
		if t.Name == name || t.File == name {
			return i
		}
	}
	return -1
}

func appendUnique(s []string, v string) []string {
	for _, e := range s {
		if e == v {
			return s
		}
	}
	return append(s, v)
}

func shuffleTracks(in []ProjectTrack) []ProjectTrack {
	out := make([]ProjectTrack, len(in))
	copy(out, in)
	r := newRand()
	for i := len(out) - 1; i > 0; i-- {
		j := r.Intn(i + 1)
		out[i], out[j] = out[j], out[i]
	}
	return out
}

// ---------- Playlist support ----------

// PlaylistFile represents a playlist.yaml.
type PlaylistFile struct {
	Name   string          `yaml:"name"`
	Tracks []PlaylistTrack `yaml:"tracks"`
}

// PlaylistTrack is one entry in a playlist.
type PlaylistTrack struct {
	File       string  `yaml:"file"`
	StartAt    string  `yaml:"start_at"`
	EndAt      string  `yaml:"end_at"`
	Transition string  `yaml:"transition"`
	Pitch      float64 `yaml:"pitch"`
	Tempo      float64 `yaml:"tempo"`
}

func loadPlaylistFile(path string) (*PlaylistFile, error) {
	var pl PlaylistFile
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if err := yaml.Unmarshal(data, &pl); err != nil {
		return nil, err
	}
	return &pl, nil
}

func savePlaylistFile(path string, pl *PlaylistFile) error {
	data, err := yaml.Marshal(pl)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

// renderPlaylistBuffer renders every playlist track sequentially with transitions.
func renderPlaylistBuffer(pl *PlaylistFile, normalize bool) ([]float64, error) {
	var out []float64
	sr := float64(audio.SampleRate)
	for _, t := range pl.Tracks {
		code, err := os.ReadFile(t.File)
		if err != nil {
			return nil, fmt.Errorf("track %q: %w", t.File, err)
		}
		combined, cps, err := collectPatterns(string(code))
		if err != nil {
			return nil, fmt.Errorf("track %q: %w", t.File, err)
		}
		secs := 8.0
		if t.EndAt != "" {
			if end, ok := parseClock(t.EndAt); ok {
				start := 0.0
				if t.StartAt != "" {
					if s, ok := parseClock(t.StartAt); ok {
						start = s
					}
				}
				secs = end - start
				if secs <= 0 {
					secs = 8
				}
			}
		}
		buf, err := audio.RenderPattern(combined, cps, secs)
		if err != nil {
			return nil, err
		}
		if t.Tempo > 0 && cps > 0 {
			ratio := (t.Tempo / 60) / cps
			buf = audio.ResampleAudio(buf, audio.SampleRate, int(float64(audio.SampleRate)*ratio))
		}
		if len(out) > 0 {
			xfLen := int(sr * 2)
			if n := len(buf); n < xfLen {
				xfLen = n
			}
			if n := len(out); n < xfLen {
				xfLen = n
			}
			if strings.HasPrefix(t.Transition, "crossfade") {
				fields := strings.Fields(t.Transition)
				if len(fields) > 1 {
					if d, err := parseFloat(fields[1]); err == nil {
						xfLen = int(sr * d)
						if xfLen > len(buf) {
							xfLen = len(buf)
						}
						if xfLen > len(out) {
							xfLen = len(out)
						}
					}
				}
				for j := 0; j < xfLen; j++ {
					t := float64(j) / float64(xfLen)
					out[len(out)-xfLen+j] = out[len(out)-xfLen+j]*(1-t) + buf[j]*t
				}
				out = append(out, buf[xfLen:]...)
			} else {
				out = append(out, buf...)
			}
		} else {
			out = append(out, buf...)
		}
	}
	if normalize && len(out) > 0 {
		out = normalizePeak(out)
	}
	return out, nil
}

// ---------- project playlist subcommand ----------

func projectPlaylistCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "playlist",
		Short: "Manage playlists (playlist.yaml) with transitions",
	}
	cmd.AddCommand(&cobra.Command{
		Use:   "init <name>",
		Short: "Create a new playlist.yaml",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			pl := &PlaylistFile{Name: args[0], Tracks: []PlaylistTrack{}}
			if err := savePlaylistFile("playlist.yaml", pl); err != nil {
				return err
			}
			fmt.Printf("created playlist %q -> playlist.yaml\n", args[0])
			return nil
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:   "add <file.cdc>",
		Short: "Add a track to the playlist",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			pl, err := loadPlaylistFile("playlist.yaml")
			if err != nil {
				return err
			}
			pl.Tracks = append(pl.Tracks, PlaylistTrack{File: args[0], Transition: "cut"})
			if err := savePlaylistFile("playlist.yaml", pl); err != nil {
				return err
			}
			fmt.Printf("added %q to playlist (total %d)\n", args[0], len(pl.Tracks))
			return nil
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List tracks in the playlist",
		RunE: func(cmd *cobra.Command, args []string) error {
			pl, err := loadPlaylistFile("playlist.yaml")
			if err != nil {
				return err
			}
			if len(pl.Tracks) == 0 {
				fmt.Println("(empty playlist)")
				return nil
			}
			for idx, t := range pl.Tracks {
				fmt.Printf("[%d] %-20s trans: %s", idx, filepath.Base(t.File), t.Transition)
				if t.Tempo > 0 {
					fmt.Printf(" tempo: %.0f", t.Tempo)
				}
				if t.Pitch != 0 {
					fmt.Printf(" pitch: %+.0f", t.Pitch)
				}
				fmt.Println()
			}
			return nil
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:   "render [out.wav]",
		Short: "Render the playlist to a single WAV with transitions",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			pl, err := loadPlaylistFile("playlist.yaml")
			if err != nil {
				return err
			}
			buf, err := renderPlaylistBuffer(pl, true)
			if err != nil {
				return err
			}
			out := "playlist.wav"
			if len(args) > 0 {
				out = args[0]
			}
			if err := audio.WriteWAVFile(out, buf, audio.SampleRate); err != nil {
				return err
			}
			fmt.Printf("rendered playlist (%d tracks) -> %s\n", len(pl.Tracks), out)
			return nil
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:   "play [out.wav]",
		Short: "Render the playlist and play it",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			pl, err := loadPlaylistFile("playlist.yaml")
			if err != nil {
				return err
			}
			buf, err := renderPlaylistBuffer(pl, true)
			if err != nil {
				return err
			}
			out := "playlist_preview.wav"
			if len(args) > 0 {
				out = args[0]
			}
			if err := audio.WriteWAVFile(out, buf, audio.SampleRate); err != nil {
				return err
			}
			fmt.Printf("rendered playlist -> %s\n", out)
			return openPlayer(out)
		},
	})
	return cmd
}

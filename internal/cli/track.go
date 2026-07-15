package cli

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/Jaziel8910/codic/internal/audio"
	"github.com/Jaziel8910/codic/internal/pattern"
)

func trackCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "track",
		Short: "Manage tracks inside a project (project.yaml)",
	}
	cmd.AddCommand(trackNewCmd())
	cmd.AddCommand(trackAddCmd())
	cmd.AddCommand(trackListCmd())
	cmd.AddCommand(trackShowCmd())
	cmd.AddCommand(trackInfoCmd())
	cmd.AddCommand(trackRemoveCmd())
	cmd.AddCommand(trackMuteCmd())
	cmd.AddCommand(trackSoloCmd())
	cmd.AddCommand(trackRenameCmd())
	cmd.AddCommand(trackMoveCmd())
	cmd.AddCommand(trackDuplicateCmd())
	cmd.AddCommand(trackEditCmd())
	cmd.AddCommand(trackSplitCmd())
	cmd.AddCommand(trackMergeCmd())
	cmd.AddCommand(trackBounceCmd())
	cmd.AddCommand(trackLengthCmd())
	cmd.AddCommand(trackColorCmd())
	cmd.AddCommand(trackBPMCmd())
	cmd.AddCommand(trackKeyCmd())
	cmd.AddCommand(trackTransposeCmd())
	cmd.AddCommand(trackTrimCmd())
	cmd.AddCommand(trackStemsCmd())
	cmd.AddCommand(trackAnalyzeCmd())
	return cmd
}

func loadCWDProject() (*ProjectFile, error) {
	return loadProjectFile(projectYAMLPath("."))
}

func saveCWDProject(pf *ProjectFile) error {
	return saveProjectFile(projectYAMLPath("."), pf)
}

func trackNewCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "new <name>",
		Short: "Create a new empty track and .cdc file",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			pf, err := loadCWDProject()
			if err != nil {
				return err
			}
			file := filepath.Join("tracks", sanitize(args[0])+".cdc")
			if err := os.MkdirAll("tracks", 0o755); err != nil {
				return err
			}
			if _, e := os.Stat(file); e != nil {
				skel := fmt.Sprintf("@name %s\n@bpm 120\n\ns(\"bd\").out()\n", args[0])
				if e2 := os.WriteFile(file, []byte(skel), 0o644); e2 != nil {
					return e2
				}
			}
			pf.Tracks = append(pf.Tracks, ProjectTrack{Name: args[0], File: file, Duration: "0:08"})
			if err := saveCWDProject(pf); err != nil {
				return err
			}
			fmt.Printf("created track %q -> %s\n", args[0], file)
			return nil
		},
	}
}

func trackAddCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "add <file.cdc>",
		Short: "Add an existing .cdc file as a track",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			pf, err := loadCWDProject()
			if err != nil {
				return err
			}
			name := strings.TrimSuffix(filepath.Base(args[0]), ".cdc")
			pf.Tracks = append(pf.Tracks, ProjectTrack{Name: name, File: args[0], Duration: "0:08"})
			if err := saveCWDProject(pf); err != nil {
				return err
			}
			fmt.Printf("added track %q\n", name)
			return nil
		},
	}
}

func trackListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List tracks in the project",
		RunE: func(cmd *cobra.Command, args []string) error {
			pf, err := loadCWDProject()
			if err != nil {
				return err
			}
			if len(pf.Tracks) == 0 {
				fmt.Println("(no tracks)")
				return nil
			}
			for i, t := range pf.Tracks {
				flags := ""
				if t.Muted {
					flags += " [muted]"
				}
				if t.Solo {
					flags += " [solo]"
				}
				fmt.Printf("[%d] %-18s %-22s %s%s\n", i, t.Name, t.File, t.Duration, flags)
			}
			return nil
		},
	}
}

func trackShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show <n>",
		Short: "Show details of track n",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			pf, err := loadCWDProject()
			if err != nil {
				return err
			}
			i, err := strconv.Atoi(args[0])
			if err != nil || i < 0 || i >= len(pf.Tracks) {
				return fmt.Errorf("invalid track index")
			}
			t := pf.Tracks[i]
			fmt.Printf("index   : %d\n", i)
			fmt.Printf("name    : %s\n", t.Name)
			fmt.Printf("file    : %s\n", t.File)
			fmt.Printf("bpm     : %.0f\n", t.BPM)
			fmt.Printf("duration: %s\n", t.Duration)
			fmt.Printf("muted   : %v\n", t.Muted)
			fmt.Printf("solo    : %v\n", t.Solo)
			fmt.Printf("color   : %s\n", t.Color)
			fmt.Printf("tags    : %v\n", t.Tags)
			return nil
		},
	}
}

func trackRemoveCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "remove <n>",
		Short: "Remove track n",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			pf, err := loadCWDProject()
			if err != nil {
				return err
			}
			i, err := strconv.Atoi(args[0])
			if err != nil || i < 0 || i >= len(pf.Tracks) {
				return fmt.Errorf("invalid track index")
			}
			name := pf.Tracks[i].Name
			pf.Tracks = append(pf.Tracks[:i], pf.Tracks[i+1:]...)
			if err := saveCWDProject(pf); err != nil {
				return err
			}
			fmt.Printf("removed track %q\n", name)
			return nil
		},
	}
}

func trackMuteCmd() *cobra.Command {
	var on bool
	cmd := &cobra.Command{
		Use:   "mute <n>",
		Short: "Mute/unmute track n",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			pf, err := loadCWDProject()
			if err != nil {
				return err
			}
			i, err := strconv.Atoi(args[0])
			if err != nil || i < 0 || i >= len(pf.Tracks) {
				return fmt.Errorf("invalid track index")
			}
			pf.Tracks[i].Muted = on
			if err := saveCWDProject(pf); err != nil {
				return err
			}
			fmt.Printf("track %d %s\n", i, map[bool]string{true: "muted", false: "unmuted"}[on])
			return nil
		},
	}
	cmd.Flags().BoolVar(&on, "off", false, "unmute instead")
	return cmd
}

func trackSoloCmd() *cobra.Command {
	var on bool
	cmd := &cobra.Command{
		Use:   "solo <n>",
		Short: "Solo/unsolo track n",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			pf, err := loadCWDProject()
			if err != nil {
				return err
			}
			i, err := strconv.Atoi(args[0])
			if err != nil || i < 0 || i >= len(pf.Tracks) {
				return fmt.Errorf("invalid track index")
			}
			pf.Tracks[i].Solo = on
			if err := saveCWDProject(pf); err != nil {
				return err
			}
			fmt.Printf("track %d %s\n", i, map[bool]string{true: "soloed", false: "unsoloed"}[on])
			return nil
		},
	}
	cmd.Flags().BoolVar(&on, "off", false, "unsolo instead")
	return cmd
}

func trackRenameCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "rename <n> <newname>",
		Short: "Rename track n",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			pf, err := loadCWDProject()
			if err != nil {
				return err
			}
			i, err := strconv.Atoi(args[0])
			if err != nil || i < 0 || i >= len(pf.Tracks) {
				return fmt.Errorf("invalid track index")
			}
			pf.Tracks[i].Name = args[1]
			if err := saveCWDProject(pf); err != nil {
				return err
			}
			fmt.Printf("renamed track %d -> %q\n", i, args[1])
			return nil
		},
	}
}

func trackMoveCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "move <n> <pos>",
		Short: "Move track n to position pos",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			pf, err := loadCWDProject()
			if err != nil {
				return err
			}
			i, err := strconv.Atoi(args[0])
			pos, err2 := strconv.Atoi(args[1])
			if err != nil || err2 != nil || i < 0 || i >= len(pf.Tracks) || pos < 0 || pos >= len(pf.Tracks) {
				return fmt.Errorf("invalid index")
			}
			t := pf.Tracks[i]
			pf.Tracks = append(pf.Tracks[:i], pf.Tracks[i+1:]...)
			pf.Tracks = append(pf.Tracks[:pos], append([]ProjectTrack{t}, pf.Tracks[pos:]...)...)
			if err := saveCWDProject(pf); err != nil {
				return err
			}
			fmt.Printf("moved track to %d\n", pos)
			return nil
		},
	}
}

func trackSplitCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "split <n>",
		Short: "Duplicate a track into two entries (A/B)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			pf, err := loadCWDProject()
			if err != nil {
				return err
			}
			i, err := strconv.Atoi(args[0])
			if err != nil || i < 0 || i >= len(pf.Tracks) {
				return fmt.Errorf("invalid track index")
			}
			t := pf.Tracks[i]
			a := t
			a.Name = t.Name + "_A"
			b := t
			b.Name = t.Name + "_B"
			pf.Tracks = append(pf.Tracks[:i], append([]ProjectTrack{a, b}, pf.Tracks[i+1:]...)...)
			if err := saveCWDProject(pf); err != nil {
				return err
			}
			fmt.Printf("split track %q into A/B\n", t.Name)
			return nil
		},
	}
}

func trackMergeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "merge <a> <b> [out.cdc]",
		Short: "Merge two tracks into a new .cdc",
		Args:  cobra.RangeArgs(2, 3),
		RunE: func(cmd *cobra.Command, args []string) error {
			pf, err := loadCWDProject()
			if err != nil {
				return err
			}
			ia, err := strconv.Atoi(args[0])
			ib, err2 := strconv.Atoi(args[1])
			if err != nil || err2 != nil || ia < 0 || ib < 0 || ia >= len(pf.Tracks) || ib >= len(pf.Tracks) {
				return fmt.Errorf("invalid track index")
			}
			ca, e1 := os.ReadFile(pf.Tracks[ia].File)
			cb, e2 := os.ReadFile(pf.Tracks[ib].File)
			if e1 != nil || e2 != nil {
				return fmt.Errorf("reading track files: %v %v", e1, e2)
			}
			out := "tracks/merged.cdc"
			if len(args) > 2 {
				out = args[2]
			}
			if err := os.MkdirAll(filepath.Dir(out), 0o755); err != nil {
				return err
			}
			merged := fmt.Sprintf("// merged %s + %s\n%s\n%s\n", pf.Tracks[ia].Name, pf.Tracks[ib].Name, ca, cb)
			if err := os.WriteFile(out, []byte(merged), 0o644); err != nil {
				return err
			}
			pf.Tracks = append(pf.Tracks, ProjectTrack{Name: pf.Tracks[ia].Name + "+" + pf.Tracks[ib].Name, File: out, Duration: "0:16"})
			if err := saveCWDProject(pf); err != nil {
				return err
			}
			fmt.Printf("merged -> %s\n", out)
			return nil
		},
	}
}

func trackBounceCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "bounce <n> [out.wav]",
		Short: "Render track n to a WAV file",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			pf, err := loadCWDProject()
			if err != nil {
				return err
			}
			i, err := strconv.Atoi(args[0])
			if err != nil || i < 0 || i >= len(pf.Tracks) {
				return fmt.Errorf("invalid track index")
			}
			code, err := os.ReadFile(pf.Tracks[i].File)
			if err != nil {
				return err
			}
			combined, cps, err := collectPatterns(string(code))
			if err != nil {
				return err
			}
			secs := trackSeconds(pf.Tracks[i], cps)
			buf, err := audio.RenderPattern(combined, cps, secs)
			if err != nil {
				return err
			}
			buf = normalizePeak(buf)
			out := sanitize(pf.Tracks[i].Name) + ".wav"
			if len(args) > 1 {
				out = args[1]
			}
			if err := audio.WriteWAVFile(out, buf, audio.SampleRate); err != nil {
				return err
			}
			fmt.Printf("bounced track %d -> %s\n", i, out)
			return nil
		},
	}
}

func trackLengthCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "length <n> <clock>",
		Short: "Set the duration of track n (e.g. 3:30)",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			pf, err := loadCWDProject()
			if err != nil {
				return err
			}
			i, err := strconv.Atoi(args[0])
			if err != nil || i < 0 || i >= len(pf.Tracks) {
				return fmt.Errorf("invalid track index")
			}
			if _, ok := parseClock(args[1]); !ok {
				return fmt.Errorf("invalid clock format (use m:ss)")
			}
			pf.Tracks[i].Duration = args[1]
			if err := saveCWDProject(pf); err != nil {
				return err
			}
			fmt.Printf("track %d duration -> %s\n", i, args[1])
			return nil
		},
	}
}

func trackColorCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "color <n> <hex>",
		Short: "Set the display color of track n (e.g. #ff0000)",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			pf, err := loadCWDProject()
			if err != nil {
				return err
			}
			i, err := strconv.Atoi(args[0])
			if err != nil || i < 0 || i >= len(pf.Tracks) {
				return fmt.Errorf("invalid track index")
			}
			pf.Tracks[i].Color = args[1]
			if err := saveCWDProject(pf); err != nil {
				return err
			}
			fmt.Printf("track %d color -> %s\n", i, args[1])
			return nil
		},
	}
}

func trackInfoCmd() *cobra.Command {
	var asJSON bool
	cmd := &cobra.Command{
		Use:   "info <n>",
		Short: "Show detailed info for track n (alias for show with --json)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			pf, err := loadCWDProject()
			if err != nil {
				return err
			}
			i, err := strconv.Atoi(args[0])
			if err != nil || i < 0 || i >= len(pf.Tracks) {
				return fmt.Errorf("invalid track index")
			}
			t := pf.Tracks[i]
			if asJSON {
				b, _ := marshalJSON(t)
				fmt.Println(b)
				return nil
			}
			fmt.Printf("name    : %s\n", t.Name)
			fmt.Printf("file    : %s\n", t.File)
			fmt.Printf("bpm     : %.0f\n", t.BPM)
			fmt.Printf("duration: %s\n", t.Duration)
			fmt.Printf("muted   : %v\n", t.Muted)
			fmt.Printf("solo    : %v\n", t.Solo)
			fmt.Printf("color   : %s\n", t.Color)
			fmt.Printf("tags    : %v\n", t.Tags)
			return nil
		},
	}
	cmd.Flags().BoolVar(&asJSON, "json", false, "output as JSON")
	return cmd
}

func trackDuplicateCmd() *cobra.Command {
	var name string
	cmd := &cobra.Command{
		Use:   "duplicate <n>",
		Short: "Duplicate track n",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			pf, err := loadCWDProject()
			if err != nil {
				return err
			}
			i, err := strconv.Atoi(args[0])
			if err != nil || i < 0 || i >= len(pf.Tracks) {
				return fmt.Errorf("invalid track index")
			}
			t := pf.Tracks[i]
			newName := name
			if newName == "" {
				newName = t.Name + "_copy"
			}
			dup := t
			dup.Name = newName
			pf.Tracks = append(pf.Tracks[:i+1], append([]ProjectTrack{dup}, pf.Tracks[i+1:]...)...)
			if err := saveCWDProject(pf); err != nil {
				return err
			}
			fmt.Printf("duplicated track %d -> %q\n", i, newName)
			return nil
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "name for the copy")
	return cmd
}

func trackEditCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "edit <n>",
		Short: "Open track n's .cdc file in $EDITOR",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			pf, err := loadCWDProject()
			if err != nil {
				return err
			}
			i, err := strconv.Atoi(args[0])
			if err != nil || i < 0 || i >= len(pf.Tracks) {
				return fmt.Errorf("invalid track index")
			}
			file := pf.Tracks[i].File
			editor := os.Getenv("EDITOR")
			if editor == "" {
				editor = defaultEditor()
			}
			c := exec.Command(editor, file)
			c.Stdin = os.Stdin
			c.Stdout = os.Stdout
			c.Stderr = os.Stderr
			return c.Run()
		},
	}
}

func trackBPMCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "bpm <n> <bpm>",
		Short: "Set the @bpm of track n's .cdc file",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			pf, err := loadCWDProject()
			if err != nil {
				return err
			}
			i, err := strconv.Atoi(args[0])
			if err != nil || i < 0 || i >= len(pf.Tracks) {
				return fmt.Errorf("invalid track index")
			}
			bpmVal, err := parseFloat(args[1])
			if err != nil {
				return fmt.Errorf("bpm must be a number")
			}
			file := pf.Tracks[i].File
			data, err := os.ReadFile(file)
			if err != nil {
				return err
			}
			code := updateMetaBPM(string(data), bpmVal)
			if err := os.WriteFile(file, []byte(code), 0o644); err != nil {
				return err
			}
			pf.Tracks[i].BPM = bpmVal
			if err := saveCWDProject(pf); err != nil {
				return err
			}
			fmt.Printf("track %d bpm -> %.0f\n", i, bpmVal)
			return nil
		},
	}
}

func updateMetaBPM(src string, bpm float64) string {
	lines := strings.Split(src, "\n")
	replaced := false
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "@bpm") {
			lines[i] = fmt.Sprintf("@bpm %.0f", bpm)
			replaced = true
			break
		}
	}
	if !replaced {
		lines = append([]string{fmt.Sprintf("@bpm %.0f", bpm)}, lines...)
	}
	return strings.Join(lines, "\n")
}

func trackKeyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "key <n> <key>",
		Short: "Set the @key of track n's .cdc file",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			pf, err := loadCWDProject()
			if err != nil {
				return err
			}
			i, err := strconv.Atoi(args[0])
			if err != nil || i < 0 || i >= len(pf.Tracks) {
				return fmt.Errorf("invalid track index")
			}
			keyVal := args[1]
			file := pf.Tracks[i].File
			data, err := os.ReadFile(file)
			if err != nil {
				return err
			}
			code := updateMetaKey(string(data), keyVal)
			if err := os.WriteFile(file, []byte(code), 0o644); err != nil {
				return err
			}
			if err := saveCWDProject(pf); err != nil {
				return err
			}
			fmt.Printf("track %d key -> %s\n", i, keyVal)
			return nil
		},
	}
	return cmd
}

func updateMetaKey(src string, key string) string {
	lines := strings.Split(src, "\n")
	replaced := false
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "@key") {
			lines[i] = fmt.Sprintf("@key %s", key)
			replaced = true
			break
		}
	}
	if !replaced {
		lines = append([]string{fmt.Sprintf("@key %s", key)}, lines...)
	}
	return strings.Join(lines, "\n")
}

func trackTransposeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "transpose <n> <semitones>",
		Short: "Transpose all notes in track n by N semitones (updates @key)",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			pf, err := loadCWDProject()
			if err != nil {
				return err
			}
			i, err := strconv.Atoi(args[0])
			if err != nil || i < 0 || i >= len(pf.Tracks) {
				return fmt.Errorf("invalid track index")
			}
			semi, err := parseIntSafe2(args[1])
			if err != nil {
				return fmt.Errorf("semitones must be an integer")
			}
			file := pf.Tracks[i].File
			data, err := os.ReadFile(file)
			if err != nil {
				return err
			}
			code := string(data)
			if semi != 0 {
				code = transposeCodang(code, semi)
			}
			if err := os.WriteFile(file, []byte(code), 0o644); err != nil {
				return err
			}
			fmt.Printf("track %d transposed by %d semitones\n", i, semi)
			return nil
		},
	}
	return cmd
}

func transposeCodang(src string, semi int) string {
	lines := strings.Split(src, "\n")
	noteNames := []string{"c", "c#", "d", "d#", "e", "f", "f#", "g", "g#", "a", "a#", "b"}
	for li, line := range lines {
		fields := strings.Fields(line)
		for fi, f := range fields {
			for ni, nn := range noteNames {
				if idx := strings.Index(f, nn); idx >= 0 {
					_ = idx
					rest := f[len(nn):]
					octave := 0
					if len(rest) > 0 && rest[0] >= '0' && rest[0] <= '9' {
						octave = int(rest[0] - '0')
						newOct := (ni+semi)/12 + octave + 1
						newNote := noteNames[((ni+semi)%12+12)%12]
						rest2 := rest[1:]
						newField := strings.Replace(f, f, newNote+fmt.Sprintf("%d", newOct-1)+rest2, 1)
						_ = newField
						fields[fi] = newNote + fmt.Sprintf("%d", newOct-1) + rest2
						break
					}
				}
			}
		}
		lines[li] = strings.Join(fields, " ")
	}
	return strings.Join(lines, "\n")
}

func trackTrimCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "trim <n> <from> <to>",
		Short: "Set the effective duration of track n (clock format)",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			pf, err := loadCWDProject()
			if err != nil {
				return err
			}
			i, err := strconv.Atoi(args[0])
			if err != nil || i < 0 || i >= len(pf.Tracks) {
				return fmt.Errorf("invalid track index")
			}
			from := args[1]
			to := args[2]
			pf.Tracks[i].Duration = fmt.Sprintf("%s-%s", from, to)
			if err := saveCWDProject(pf); err != nil {
				return err
			}
			fmt.Printf("track %d trimmed to %s..%s\n", i, from, to)
			return nil
		},
	}
	return cmd
}

func trackStemsCmd() *cobra.Command {
	var outDir string
	cmd := &cobra.Command{
		Use:   "stems <n>",
		Short: "Render track n's .cdc into separate stem WAVs",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			pf, err := loadCWDProject()
			if err != nil {
				return err
			}
			i, err := strconv.Atoi(args[0])
			if err != nil || i < 0 || i >= len(pf.Tracks) {
				return fmt.Errorf("invalid track index")
			}
			t := pf.Tracks[i]
			code, err := os.ReadFile(t.File)
			if err != nil {
				return err
			}
			dir := outDir
			if dir == "" {
				dir = "stems"
			}
			if err := os.MkdirAll(dir, 0o755); err != nil {
				return err
			}
			combined, cps, err := collectPatterns(string(code))
			if err != nil {
				return err
			}
			secs := trackSeconds(t, cps)
			buf, err := audio.RenderPattern(combined, cps, secs)
			if err != nil {
				return err
			}
			buf = normalizePeak(buf)
			out := filepath.Join(dir, sanitize(t.Name)+"_stem.wav")
			if err := audio.WriteWAVFile(out, buf, audio.SampleRate); err != nil {
				return err
			}
			fmt.Printf("stem for track %d -> %s\n", i, out)
			return nil
		},
	}
	cmd.Flags().StringVar(&outDir, "out", "stems", "output directory")
	return cmd
}

func trackAnalyzeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "analyze <n>",
		Short: "Analyze track n: density, tempo, pattern complexity",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			pf, err := loadCWDProject()
			if err != nil {
				return err
			}
			i, err := strconv.Atoi(args[0])
			if err != nil || i < 0 || i >= len(pf.Tracks) {
				return fmt.Errorf("invalid track index")
			}
			t := pf.Tracks[i]
			code, err := os.ReadFile(t.File)
			if err != nil {
				return err
			}
			combined, cps, err := collectPatterns(string(code))
			if err != nil {
				return err
			}
			haps := combined.FirstCycle()
			totalHaps := len(haps)
			fmt.Printf("Track     : %s\n", t.Name)
			fmt.Printf("File      : %s\n", t.File)
			fmt.Printf("BPM       : %.0f (cps=%.4f)\n", cps*60, cps)
			fmt.Printf("Duration  : %s\n", t.Duration)
			fmt.Printf("Events/cyc: %d\n", totalHaps)
			noteCount := 0
			for _, h := range haps {
				if cm, ok := h.Value.(pattern.ControlMap); ok {
					if _, hasNote := cm["note"]; hasNote {
						noteCount++
					} else if _, hasN := cm["n"]; hasN {
						noteCount++
					}
				}
			}
			fmt.Printf("Notes     : %d\n", noteCount)
			fmt.Printf("Other     : %d\n", totalHaps-noteCount)
			return nil
		},
	}
	return cmd
}

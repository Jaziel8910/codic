package dj

import (
	"fmt"
	"math"
	"os"
	"sort"

	"github.com/Jaziel8910/codic/internal/codang"
	"github.com/Jaziel8910/codic/internal/pattern"
)

// Recommend analyzes a .cdc track and returns the name of the genre it most
// closely resembles, based on tempo and rhythmic density.
func Recommend(file string) (string, error) {
	src, err := os.ReadFile(file)
	if err != nil {
		return "", fmt.Errorf("reading %s: %w", file, err)
	}
	prog, err := codang.Parse(string(src))
	if err != nil {
		return "", fmt.Errorf("parse error: %w", err)
	}
	var all []pattern.Pattern
	ev := codang.NewEvaluator(func(p pattern.Pattern) { all = append(all, p) })
	if err := ev.Eval(prog); err != nil {
		return "", fmt.Errorf("eval error: %w", err)
	}
	if len(all) == 0 {
		return "", fmt.Errorf("track produces no patterns")
	}
	combined := all[0]
	for _, p := range all[1:] {
		combined = pattern.Stack(combined, p)
	}
	cps := ev.GetCPS()
	if cps <= 0 {
		cps = 1.0
	}
	bpm := int(math.Round(cps * 240))
	haps := combined.FirstCycle()
	density := len(haps)

	// candidate genres: learned + built-in defaults
	candidates := map[string]bool{}
	if learned, err := ListGenres(); err == nil {
		for _, g := range learned {
			candidates[g] = true
		}
	}
	for g := range DefaultProfiles {
		candidates[g] = true
	}

	type scored struct {
		name  string
		score float64
	}
	var scores []scored
	for name := range candidates {
		g := defaultProfile(name)
		center := (g.BPMMin + g.BPMMax) / 2
		if center <= 0 {
			center = 120
		}
		bpmPenalty := math.Abs(float64(bpm-center)) / 10.0
		// density heuristic: denser tracks ~ faster/higher-energy genres
		densityPenalty := 0.0
		if density > 0 {
			rel := float64(density) / float64(center/4+8)
			if rel > 1.5 {
				densityPenalty = rel - 1.5
			}
		}
		score := bpmPenalty + densityPenalty
		scores = append(scores, scored{name: name, score: score})
	}
	sort.Slice(scores, func(i, j int) bool { return scores[i].score < scores[j].score })

	if len(scores) == 0 {
		return "", fmt.Errorf("no genres available")
	}
	best := scores[0]
	fmt.Printf("track: bpm=%d density=%d events/cycle\n", bpm, density)
	fmt.Println("candidates (best first):")
	for i, s := range scores {
		if i >= 5 {
			break
		}
		marker := "  "
		if i == 0 {
			marker = "* "
		}
		fmt.Printf("  %s%s (score %.2f)\n", marker, s.name, s.score)
	}
	return best.name, nil
}

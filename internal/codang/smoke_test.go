package codang

import (
	"testing"

	"github.com/Jaziel8910/codic/internal/pattern"
)

func evalCapture(t *testing.T, src string) pattern.Pattern {
	t.Helper()
	prog, err := Parse(src)
	if err != nil {
		t.Fatalf("parse error: %v\n%s", err, src)
	}
	var captured *pattern.Pattern
	ev := NewEvaluator(func(p pattern.Pattern) { captured = &p })
	if err := ev.Eval(prog); err != nil {
		t.Fatalf("eval error: %v\n%s", err, src)
	}
	if captured != nil {
		return *captured
	}
	lp := ev.LastPattern()
	if lp.Query == nil {
		t.Fatalf("no pattern produced\n%s", src)
	}
	return lp
}

func TestNewStrudelFunctions(t *testing.T) {
	cases := []struct {
		name string
		src  string
	}{
		{"choose", `p = choose("bd","sd","hh")`},
		{"run", `p = run(8)`},
		{"scale", `p = scale("major", 0, 1, 2, 3).note()`},
		{"chord", `p = chord("c", "minor")`},
		{"degrade", `p = sound("bd*4").degrade()`},
		{"degradeBy", `p = sound("bd*4").degradeBy(0.3)`},
		{"palindrome", `p = palindrome(sound("bd sd cp"))`},
		{"scramble", `p = scramble(sound("bd sd hh cp"))`},
		{"params", `p = sound("bd").gain(0.8).cutoff(1200).room(0.3).pan(-0.5).speed(2)`},
		{"filter", `p = sound("bd").hpf(200).lpf(4000).distort(0.2).crush(4)`},
		{"tremolo", `p = sound("bd").tremolo(0.5).phaser(0.3).chorus(0.2)`},
		{"vowel", `p = sound("vowel ah").vowel("a")`},
		{"zoom", `p = sound("bd*4").zoom(0, 0.5)`},
		{"fit", `p = sound("bd*4").fit(2)`},
		{"ply", `p = sound("bd").ply(3)`},
		{"swing", `p = sound("bd*4").swing(0.25)`},
		{"adsr", `p = sound("bd").adsr(0.01, 0.1, 0.5, 0.2)`},
		{"lfo", `p = sound("bd").lfo("pan", 1, 0.5)`},
		{"midi", `p = sound("x").midinote(60).midichan(1)`},
		{"fm", `p = sound("bd").fm(2).fmi(3)`},
		{"loopSample", `p = sound("amen").loopSample(0.25, 0.75)`},
		{"octave", `p = note(0).octave(1)`},
		{"within", `func dbl(x): x.fast(2)` + "\n" + `p = sound("bd*4").within(0, 0.5, "dbl")`},
		{"sometimes", `func dbl(x): x.fast(2)` + "\n" + `p = sound("bd*4").sometimes("dbl")`},
		{"every", `func dbl(x): x.fast(2)` + "\n" + `p = sound("bd*4").every(2, "dbl")`},
		{"layer", `func a(x): x.fast(2)` + "\n" + `func b(x): x.slow(2)` + "\n" + `p = sound("bd").layer("a", "b")`},
		{"when", `func rev(x): x.rev()` + "\n" + `p = sound("bd*4").when(sound("1"), "rev")`},
		{"stack", `p = stack(sound("bd"), sound("sd").gain(0.5))`},
		{"wchoose", `p = wchoose([0.7, "bd"], [0.3, "sd"])`},
		{"irand", `p = irand(5)`},
		{"rand2", `p = rand2(1)`},
		{"rangex", `p = rangex(0, 10)`},
		{"arp", `p = chord("c","major").arp("up")`},
		{"voicing", `p = note(0).voicing(0, 4, 7)`},
		{"edoScale", `p = edoScale(12, 0, 1, 2)`},
		{"brak", `p = brak(sound("bd sd cp"))`},
		{"spread", `p = sound("bd*4").spread(4)`},
		{"squeeze", `p = sound("bd*4").squeeze(sound("1"))`},
		{"euclid", `p = sound("bd").euclid(8, 3, 0)`},
		{"euclidFn", `p = euclid(8, 3, 1)`},
		{"euclidLegato", `p = sound("bd").euclidLegato(8, 3, 0)`},
		{"euclidish", `p = euclidish([3, 2], 5)`},
		{"juxBy", `func rev(x): x.rev()` + "\n" + `p = sound("bd*4").juxBy(0.5, "rev")`},
		{"juxFlip", `func rev(x): x.rev()` + "\n" + `p = sound("bd*4").juxFlip("rev")`},
		{"compress", `p = sound("bd*4").compress(0.25, 0.75)`},
		{"unison", `p = sound("bd").unison(3)`},
		{"toMidi", `p = sound("c e g").toMidi()`},
		{"fromMidi", `p = sound("60 64 67").fromMidi()`},
		{"striate", `p = sound("amen").striate(4)`},
		{"chop", `p = sound("amen").chop(8)`},
		{"bite", `p = sound("amen").bite(0.1, 0.9, 4)`},
		{"gap", `p = sound("bd*4").gap(0.2)`},
		{"brand", `p = sound("bd").brand("kick")`},
		{"inhabit", `p = sound("bd*4").inhabit(sound("1 0"))`},
		{"zip", `p = zip(sound("bd"), sound("sd"))`},
		{"xfade", `p = xfade(sound("bd"), sound("sd"), 0.5)`},
		{"stepcat", `p = stepcat(sound("bd"), sound("sd"), sound("hh"))`},
		{"stepwise", `p = stepwise("bd", "sd", "hh")`},
		{"spiral", `p = spiral("minor", 8, 1, 0)`},
		{"setcpm", `p = sound("bd").setcpm(120)`},
		{"defragmentHaps", `p = sound("bd*4").defragmentHaps()`},
		{"reset", `p = sound("amen").reset()`},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			pat := evalCapture(t, c.src)
			if pat.Query == nil {
				t.Fatalf("no pattern captured for %s", c.name)
			}
			haps := pat.FirstCycle()
			if len(haps) == 0 {
				t.Logf("[%s] warning: 0 haps in first cycle", c.name)
			}
			for i, h := range haps {
				if i > 5 {
					break
				}
				t.Logf("  [%s] hap %d: %s", c.name, i, h.Show())
			}
		})
	}
}

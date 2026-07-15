package cli

import "testing"

func TestCollectPatternsQuoted(t *testing.T) {
	cases := []string{
		`s("bd").euclid(8, 2, 0).out()`,
		`s("hh").euclid(8, 3, 1).speed(2).gain(0.4).out()`,
		`"c3 e3 g3".note().s("sawtooth").cutoff(600).gain(0.3).out()`,
	}
	for _, c := range cases {
		p, cps, err := collectPatterns(c)
		if err != nil {
			t.Fatalf("code %q: %v", c, err)
		}
		if len(p.FirstCycle()) == 0 {
			t.Fatalf("code %q: produced no haps", c)
		}
		t.Logf("ok cps=%.2f haps=%d", cps, len(p.FirstCycle()))
	}
}

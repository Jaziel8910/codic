package pattern

import "testing"

func TestSpanishNotes(t *testing.T) {
	cases := []struct {
		in   string
		want int
	}{
		{"do3", 48},
		{"re3", 50},
		{"mi3", 52},
		{"fa3", 53},
		{"sol3", 55},
		{"la3", 57},
		{"si3", 59},
		{"do4", 60},
		{"do#4", 61},
		{"reb4", 61},
		{"sol#5", 80},
		{"sib3", 58},
		// English still works
		{"c4", 60},
		{"fs4", 66},
		{"bb2", 46},
	}
	for _, c := range cases {
		got, ok := noteToMidi(c.in)
		if !ok {
			t.Fatalf("noteToMidi(%q) not recognized", c.in)
		}
		if int(got+0.5) != c.want {
			t.Fatalf("noteToMidi(%q) = %v, want %v", c.in, int(got+0.5), c.want)
		}
	}
}

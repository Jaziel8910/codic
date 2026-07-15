package audio

import (
	"path/filepath"
	"testing"
)

func TestSampleStoreLoads(t *testing.T) {
	root, err := filepath.Abs(filepath.Join("..", "..", "samples"))
	if err != nil {
		t.Fatal(err)
	}
	if err := SetSampleRoot(root); err != nil {
		t.Fatalf("SetSampleRoot: %v", err)
	}
	names := SampleNames()
	if len(names) == 0 {
		t.Fatal("no sample names loaded")
	}
	t.Logf("loaded %d sample names", len(names))

	// Pick a known sample and decode it.
	name := ""
	for _, n := range names {
		if n == "808bd" {
			name = n
		}
	}
	if name == "" {
		name = names[0]
	}
	n, num := NormalizeSampleName(name + ":1")
	path, ok := resolveSample(n, num)
	if !ok {
		t.Fatalf("resolveSample(%q) failed", n)
	}
	buf, err := loadSampleAudio(path)
	if err != nil {
		t.Fatalf("decode %s: %v", path, err)
	}
	if len(buf) == 0 {
		t.Fatal("decoded buffer empty")
	}
	t.Logf("%s -> %d samples (%.2fs)", filepath.Base(path), len(buf), float64(len(buf))/float64(SampleRate))
}

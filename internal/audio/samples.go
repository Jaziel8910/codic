package audio

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

var (
	sampleMu    sync.Mutex
	sampleRoot  string
	sampleIndex map[string][]string // name -> relative wav paths
	sampleCache map[string][]float64
	indexLoaded bool
)

// SetSampleRoot sets the directory that contains strudel.json and the .wav
// files. It also eagerly loads the sample index.
func SetSampleRoot(root string) error {
	sampleMu.Lock()
	sampleRoot = root
	sampleMu.Unlock()
	return LoadSampleIndex()
}

// LoadSampleIndex reads strudel.json from the sample root and builds the
// name -> file map. It is safe to call multiple times.
func LoadSampleIndex() error {
	sampleMu.Lock()
	defer sampleMu.Unlock()
	if sampleRoot == "" {
		return fmt.Errorf("sample root not set")
	}
	idxPath := filepath.Join(sampleRoot, "strudel.json")
	data, err := os.ReadFile(idxPath)
	if err != nil {
		// No index: samples simply won't resolve.
		sampleIndex = map[string][]string{}
		indexLoaded = true
		return fmt.Errorf("no strudel.json at %s: %w", idxPath, err)
	}
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("bad strudel.json: %w", err)
	}
	idx := map[string][]string{}
	for k, v := range raw {
		if k == "_base" {
			continue
		}
		arr, ok := v.([]interface{})
		if !ok {
			continue
		}
		var paths []string
		for _, e := range arr {
			if s, ok := e.(string); ok {
				paths = append(paths, s)
			}
		}
		if len(paths) > 0 {
			idx[k] = paths
		}
	}
	sampleIndex = idx
	sampleCache = map[string][]float64{}
	indexLoaded = true
	return nil
}

// HasSample reports whether a sample name is known.
func HasSample(name string) bool {
	sampleMu.Lock()
	defer sampleMu.Unlock()
	if !indexLoaded || sampleIndex == nil {
		return false
	}
	_, ok := sampleIndex[name]
	return ok
}

// resolveSample returns the absolute path for sample `name` variant `n`.
func resolveSample(name string, n int) (string, bool) {
	sampleMu.Lock()
	defer sampleMu.Unlock()
	paths, ok := sampleIndex[name]
	if !ok || len(paths) == 0 {
		return "", false
	}
	if n < 0 || n >= len(paths) {
		n = 0
	}
	return filepath.Join(sampleRoot, filepath.FromSlash(paths[n])), true
}

// loadSampleAudio decodes a wav file, resamples it to the engine rate, and
// caches the result keyed by absolute path.
func loadSampleAudio(absPath string) ([]float64, error) {
	sampleMu.Lock()
	if sampleCache != nil {
		if c, ok := sampleCache[absPath]; ok {
			sampleMu.Unlock()
			return c, nil
		}
	}
	sampleMu.Unlock()

	raw, rate, err := decodeWAV(absPath)
	if err != nil {
		return nil, err
	}
	out := resample(raw, rate, SampleRate)

	sampleMu.Lock()
	if sampleCache == nil {
		sampleCache = map[string][]float64{}
	}
	sampleCache[absPath] = out
	sampleMu.Unlock()
	return out, nil
}

// resample linearly resamples a mono buffer from srcRate to dstRate.
func resample(in []float64, srcRate, dstRate int) []float64 {
	if srcRate == dstRate || len(in) == 0 {
		return in
	}
	ratio := float64(dstRate) / float64(srcRate)
	outLen := int(float64(len(in)) * ratio)
	if outLen < 1 {
		outLen = 1
	}
	out := make([]float64, outLen)
	for i := 0; i < outLen; i++ {
		pos := float64(i) / ratio
		i0 := int(pos)
		if i0 >= len(in)-1 {
			out[i] = in[len(in)-1]
			continue
		}
		frac := pos - float64(i0)
		out[i] = in[i0]*(1-frac) + in[i0+1]*frac
	}
	return out
}

// SampleNames returns all known sample names (for inventory/docs).
func SampleNames() []string {
	sampleMu.Lock()
	defer sampleMu.Unlock()
	names := make([]string, 0, len(sampleIndex))
	for k := range sampleIndex {
		names = append(names, k)
	}
	return names
}

// SampleCount returns the number of files behind a sample name.
func SampleCount(name string) int {
	sampleMu.Lock()
	defer sampleMu.Unlock()
	return len(sampleIndex[name])
}

// NormalizeSampleName strips a trailing ":n" variant specifier.
func NormalizeSampleName(s string) (string, int) {
	if idx := strings.IndexByte(s, ':'); idx >= 0 {
		num := 0
		fmt.Sscanf(s[idx+1:], "%d", &num)
		return s[:idx], num
	}
	return s, 0
}

// InitSamples locates the samples directory (containing strudel.json) and
// loads the index. It searches several candidate locations so the app works
// regardless of where it is launched from.
func InitSamples() error {
	candidates := []string{}
	if env := os.Getenv("CODIC_SAMPLES"); env != "" {
		candidates = append(candidates, env)
	}
	candidates = append(candidates,
		"sounds",
		filepath.Join("..", "sounds"),
		filepath.Join("..", "..", "sounds"),
	)
	if exe, err := os.Executable(); err == nil {
		candidates = append(candidates, filepath.Join(filepath.Dir(exe), "sounds"))
	}
	if home := userProfileDir(); home != "" {
		candidates = append(candidates,
			filepath.Join(home, "CODIC", "sounds"),
			filepath.Join(home, ".codic", "samples"),
		)
	}
	for _, c := range candidates {
		if _, err := os.Stat(filepath.Join(c, "strudel.json")); err == nil {
			return SetSampleRoot(c)
		}
	}
	// No samples directory found; run with synthesis only.
	sampleMu.Lock()
	sampleRoot = ""
	sampleIndex = map[string][]string{}
	sampleCache = map[string][]float64{}
	indexLoaded = true
	sampleMu.Unlock()
	return fmt.Errorf("samples directory not found")
}

// SamplePath returns the absolute file path for sample `name` variant `n`, and
// whether it exists. It is the exported wrapper around resolveSample.
func SamplePath(name string, n int) (string, bool) {
	return resolveSample(name, n)
}

// SampleRoot returns the configured sample directory ("" if none).
func SampleRoot() string {
	sampleMu.Lock()
	defer sampleMu.Unlock()
	return sampleRoot
}

// userProfileDir returns the user's profile folder (C:\Users\<you>),
// preferring USERPROFILE over the OS home dir so the workspace never lands
// on the Desktop by accident.
func userProfileDir() string {
	if p := os.Getenv("USERPROFILE"); p != "" {
		return p
	}
	if home, err := os.UserHomeDir(); err == nil {
		return home
	}
	return ""
}

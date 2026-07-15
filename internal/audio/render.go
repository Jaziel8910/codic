package audio

import (
	"encoding/binary"
	"fmt"
	"math"
	"os"

	"github.com/Jaziel8910/codic/internal/pattern"
)

// RenderPattern renders a pattern to a stereo float buffer (interleaved L,R)
// at SampleRate, without any audio device. This is used to export .cdc songs
// to WAV files.
func RenderPattern(pat pattern.Pattern, cps, seconds float64) ([]float64, error) {
	if pat.Query == nil {
		return nil, fmt.Errorf("pattern has no query function")
	}
	if cps <= 0 {
		cps = 1.0
	}
	if seconds <= 0 {
		seconds = 8
	}

	sr := float64(SampleRate)
	totalSamples := int(seconds * sr)
	totalCycles := seconds * cps

	type event struct {
		v     *Voice
		start int // start sample index
		dur   int // duration in samples
	}
	var events []event

	span := pattern.TimeSpan{
		Begin: pattern.FracFloat(0),
		End:   pattern.FracFloat(totalCycles),
	}
	haps := pat.Query(pattern.State{Span: span})
	for _, h := range haps {
		durCycles := h.Part.End.Sub(h.Part.Begin).Float64()
		if durCycles <= 0 {
			durCycles = 1.0 / cps
		}
		startCycles := h.Part.Begin.Float64()
		durSec := durCycles / cps
		v := buildVoice(pattern.ToControlMap(h.Value), durSec)
		if v == nil {
			continue
		}
		start := int(startCycles / cps * sr)
		if start < 0 {
			start = 0
		}
		dur := int(durSec * sr)
		if dur < 1 {
			dur = 1
		}
		events = append(events, event{v: v, start: start, dur: dur})
	}

	out := make([]float64, totalSamples*2) // stereo interleaved
	master := 0.8
	for n := 0; n < totalSamples; n++ {
		t := float64(n) / sr
		var l, r float64
		for i := range events {
			ev := &events[i]
			sampleStart := float64(ev.start) / sr
			sampleEnd := float64(ev.start+ev.dur) / sr
			if t < sampleStart {
				continue
			}
			if !ev.v.looping && t >= sampleEnd {
				continue
			}
			ev.v.elapsed = t - sampleStart
			if ev.v.sample != nil && len(ev.v.sample) > 0 {
				pos := (t - sampleStart) * ev.v.speed * sr
				if ev.v.looping {
					pos = math.Mod(pos, float64(len(ev.v.sample)))
				}
				ev.v.sampleFloatPos = pos
			}
			s := ev.v.nextSample()
			ll, rr := ev.v.panMix(s)
			l += ll
			r += rr
		}
		out[2*n] = clamp(l*master, -1, 1)
		out[2*n+1] = clamp(r*master, -1, 1)
	}
	return out, nil
}

// writeWAV encodes an interleaved stereo float buffer as a 16-bit PCM WAV file.
func writeWAV(path string, samples []float64, sr int) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	numFrames := len(samples) / 2
	numChannels := 2
	bits := 16
	dataBytes := numFrames * numChannels * (bits / 8)
	byteRate := sr * numChannels * (bits / 8)
	blockAlign := numChannels * (bits / 8)

	header := make([]byte, 44)
	copy(header[0:4], "RIFF")
	binary.LittleEndian.PutUint32(header[4:8], uint32(36+dataBytes))
	copy(header[8:12], "WAVE")
	copy(header[12:16], "fmt ")
	binary.LittleEndian.PutUint32(header[16:20], 16)
	binary.LittleEndian.PutUint16(header[20:22], 1) // PCM
	binary.LittleEndian.PutUint16(header[22:24], uint16(numChannels))
	binary.LittleEndian.PutUint32(header[24:28], uint32(sr))
	binary.LittleEndian.PutUint32(header[28:32], uint32(byteRate))
	binary.LittleEndian.PutUint16(header[32:34], uint16(blockAlign))
	binary.LittleEndian.PutUint16(header[34:36], uint16(bits))
	copy(header[36:40], "data")
	binary.LittleEndian.PutUint32(header[40:44], uint32(dataBytes))

	if _, err := f.Write(header); err != nil {
		return err
	}
	buf := make([]byte, 2)
	for _, s := range samples {
		v := int16(clamp(s, -1, 1) * 32767)
		binary.LittleEndian.PutUint16(buf, uint16(v))
		if _, err := f.Write(buf); err != nil {
			return err
		}
	}
	return nil
}

// WriteWAVFile encodes an interleaved stereo float buffer as a 16-bit PCM WAV.
func WriteWAVFile(path string, samples []float64, sr int) error {
	return writeWAV(path, samples, sr)
}

// RenderToWAV evaluates is not done here; callers pass an already-evaluated
// pattern. This helper renders and writes in one step.
func RenderToWAV(pat pattern.Pattern, cps, seconds float64, path string) error {
	buf, err := RenderPattern(pat, cps, seconds)
	if err != nil {
		return err
	}
	return writeWAV(path, buf, SampleRate)
}

package audio

import (
	"github.com/Jaziel8910/codic/internal/pattern"
)

// buildVoice turns a control map into a ready-to-play Voice, or nil if the
// control map describes nothing playable. Shared by the offline WAV renderer.
func buildVoice(cm pattern.ControlMap, durSec float64) *Voice {
	if durSec <= 0 {
		durSec = 0.2
	}

	// Extract params
	soundName := ""
	if s, ok := cm["s"]; ok {
		soundName = toString(s)
	}
	noteName := ""
	if n, ok := cm["note"]; ok {
		noteName = toString(n)
	}
	gain := 0.5
	if g, ok := cm["gain"]; ok {
		gain = toFloatVal(g)
	}
	pan := 0.5
	if pn, ok := cm["pan"]; ok {
		pan = clamp(toFloatVal(pn), 0, 1)
	}
	speed := 1.0
	if sp, ok := cm["speed"]; ok {
		speed = toFloatVal(sp)
		if speed <= 0 {
			speed = 1.0
		}
	}
	loop := false
	if lp, ok := cm["loop"]; ok {
		loop = toBoolVal(lp)
	}

	// Build the effects chain from the control parameters.
	chain := buildFXChain(cm)

	// Normalize "name:n" sample variant syntax.
	sampleName, sampleNum := NormalizeSampleName(soundName)

	// Decide what to play
	switch {
	case IsValidDrum(soundName):
		// Play synthesized drum
		sample := GetDrumSample(soundName)
		return &Voice{
			sample:     sample,
			sampleRate: SampleRate,
			gain:       gain,
			pan:        pan,
			duration:   durSec,
			elapsed:    0,
			env:        NewEnvelope(0.002, 0.05, 1.0, 0.03),
			fx:         chain,
		}

	case HasSample(sampleName):
		// Play a file-based sample from the Strudel/Dirt-Samples library.
		absPath, ok := resolveSample(sampleName, sampleNum)
		if ok {
			if buf, err := loadSampleAudio(absPath); err == nil && len(buf) > 0 {
				dur := float64(len(buf)) / float64(SampleRate) / speed
				if loop {
					dur = durSec
				} else if dur < durSec {
					dur = durSec
				}
				return &Voice{
					sample:     buf,
					sampleRate: SampleRate,
					speed:      speed,
					looping:    loop,
					gain:       gain,
					pan:        pan,
					duration:   dur,
					elapsed:    0,
					env:        NewEnvelope(0.002, 0.02, 1.0, 0.06),
					fx:         chain,
				}
			}
		}

	case IsOscillatorType(soundName) && noteName != "":
		// Play oscillator with note
		freq := noteToFreq(noteName)
		v := NewVoice(soundName, freq, durSec, gain, pan)
		v.fx = chain
		return v

	case noteName != "":
		// Default to sine oscillator if note but no sound type
		freq := noteToFreq(noteName)
		v := NewVoice("sine", freq, durSec, gain, pan)
		v.fx = chain
		return v

	case soundName != "":
		// Try to treat as a note name
		if freq, ok := noteNameToMidi(soundName); ok {
			v := NewVoice("sine", midiToFreq(freq), durSec, gain, pan)
			v.fx = chain
			return v
		}
	}
	return nil
}

// buildFXChain assembles the effect processors from the control map, in a
// sensible signal order: filters -> distortion -> crush -> vowel -> phaser ->
// chorus -> delay -> reverb.
func buildFXChain(cm pattern.ControlMap) []Processor {
	var chain []Processor

	// Filters
	if c, ok := cm["cutoff"]; ok {
		freq := toFloatVal(c)
		if freq > 0 {
			q := 0.7
			if qv, ok := cm["lpq"]; ok {
				q = toFloatVal(qv)
			}
			chain = append(chain, NewBiquad(filterLowpass, freq, q))
		}
	}
	if h, ok := cm["hpf"]; ok {
		freq := toFloatVal(h)
		if freq > 0 {
			q := 0.7
			if qv, ok := cm["hpq"]; ok {
				q = toFloatVal(qv)
			}
			chain = append(chain, NewBiquad(filterHighpass, freq, q))
		}
	}
	if b, ok := cm["bpf"]; ok {
		freq := toFloatVal(b)
		if freq > 0 {
			q := 1.0
			if qv, ok := cm["bpq"]; ok {
				q = toFloatVal(qv)
			}
			chain = append(chain, NewBiquad(filterBandpass, freq, q))
		}
	}

	// Distortion
	if d, ok := cm["distort"]; ok {
		amt := toFloatVal(d)
		if amt > 0 {
			chain = append(chain, NewDistortion(amt))
		}
	}

	// Bitcrush
	if c, ok := cm["crush"]; ok {
		bits := toFloatVal(c)
		if bits > 0 {
			chain = append(chain, NewCrush(bits, bits))
		}
	}

	// Vowel / formant
	if vw, ok := cm["vowel"]; ok {
		if s, ok := vw.(string); ok && s != "" {
			dry := 0.4
			chain = append(chain, NewFormant(s, dry))
		}
	}

	// Phaser
	if ph, ok := cm["phaser"]; ok {
		depth := toFloatVal(ph)
		if depth > 0 {
			if pd, ok := cm["phaserdepth"]; ok {
				depth = toFloatVal(pd)
			}
			chain = append(chain, NewPhaser(0.5, depth, 0.5))
		}
	}

	// Chorus
	if ch, ok := cm["chorus"]; ok {
		amt := toFloatVal(ch)
		if amt > 0 {
			chain = append(chain, NewChorus(0.7, amt, 0.5))
		}
	}

	// Delay / echo
	if dl, ok := cm["delay"]; ok {
		dt := toFloatVal(dl)
		if dt > 0 {
			fb := 0.4
			if fv, ok := cm["delayfeedback"]; ok {
				fb = toFloatVal(fv)
			}
			chain = append(chain, NewDelay(dt, fb, 0.5))
		}
	}

	// Tremolo
	if t, ok := cm["tremolo"]; ok {
		depth := toFloatVal(t)
		if depth > 0 {
			rate := 5.0
			if rv, ok := cm["tremolorate"]; ok {
				rate = toFloatVal(rv)
			}
			if dv, ok := cm["tremolodepth"]; ok {
				depth = toFloatVal(dv)
			}
			chain = append(chain, NewTremolo(rate, depth))
		}
	}

	// Reverb (also triggered by room/size/roomsize)
	for _, key := range []string{"reverb", "roomsize", "size", "room"} {
		if r, ok := cm[key]; ok {
			amt := toFloatVal(r)
			if amt > 0 {
				chain = append(chain, NewReverb(amt))
				break
			}
		}
	}

	return chain
}

func toString(v interface{}) string {
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}

func toFloatVal(v interface{}) float64 {
	switch t := v.(type) {
	case float64:
		return t
	case int:
		return float64(t)
	case int64:
		return float64(t)
	case string:
		if f, ok := noteNameToMidi(t); ok {
			return f
		}
	}
	return 0
}

func toBoolVal(v interface{}) bool {
	switch t := v.(type) {
	case bool:
		return t
	case string:
		return t == "true" || t == "1"
	case float64:
		return t != 0
	}
	return false
}

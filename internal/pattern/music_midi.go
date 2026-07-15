package pattern

// ============================================================
// Funciones Musicales y MIDI (port de strudel)
// ============================================================

// --- Escalas ---

// Scale genera un patrón de grados de escala (semitonos desde la raíz).
// name: "major","minor","dorian","phrygian","lydian","mixolydian","locrian",
//
//	"penta","minorpenta","majorpenta","blues","harmonicMinor", etc.
//
// Scale es un método de patrón: p.scale("major", 0, 1, 2).
func (p Pattern) Scale(name string, degrees ...interface{}) Pattern {
	steps := scaleSteps(name)
	if len(steps) == 0 {
		steps = scaleSteps("major")
	}
	if len(degrees) == 0 {
		// genera la escala completa
		vals := make([]interface{}, 0, len(steps)+1)
		vals = append(vals, 0)
		for _, s := range steps {
			vals = append(vals, s)
		}
		return Sequence(vals...)
	}
	vals := make([]interface{}, len(degrees))
	for i, d := range degrees {
		idx := toNumber(d)
		oct := int(idx) / len(steps)
		deg := int(idx) % len(steps)
		if deg < 0 {
			deg += len(steps)
		}
		vals[i] = float64(steps[deg]) + float64(oct*12)
	}
	return Sequence(vals...)
}

// EdoScale escala en n-tono igual (edo). nPor defecto 12.
func EdoScale(edo int, degrees ...interface{}) Pattern {
	if edo <= 0 {
		edo = 12
	}
	if len(degrees) == 0 {
		return Sequence(0.0, float64(12*100/edo)) // aproximación
	}
	vals := make([]interface{}, len(degrees))
	for i, d := range degrees {
		idx := int(toNumber(d))
		vals[i] = float64(idx * (1200 / edo))
	}
	return Sequence(vals...)
}

// ScaleTranspose transposición en grados de escala.
func (p Pattern) ScaleTranspose(scale string, steps float64) Pattern {
	sc := scaleSteps(scale)
	if len(sc) == 0 {
		return p
	}
	n := len(sc)
	oct := int(steps) / n
	deg := int(steps) % n
	if deg < 0 {
		deg += n
	}
	offset := float64(sc[deg]) + float64(oct*12)
	return p.Add(Pure(offset))
}

// Mode fija el modo musical (para visualización / acordes).
func (p Pattern) Mode(v string) Pattern { return p.AddParam("mode", v) }

// Tune microafinación en semitonos.
func (p Pattern) Tune(v float64) Pattern { return p.AddParam("tune", v) }

// --- Acordes ---

// Chord construye un acorde a partir de una raíz y un nombre de acorde.
// name: "major","minor","maj7","min7","sus4","dim","aug","7","m7b5", etc.
func Chord(root string, name string) Pattern {
	rootMidi, ok := noteToMidi(root)
	if !ok {
		rootMidi = 0
	}
	intervals := chordIntervals(name)
	vals := make([]interface{}, len(intervals))
	for i, iv := range intervals {
		vals[i] = rootMidi + float64(iv)
	}
	return Sequence(vals...).Add(Pure(0)) // nota base
}

// chordIntervals devuelve los intervalos semitónicos de un acorde.
func chordIntervals(name string) []int {
	switch name {
	case "major", "maj", "":
		return []int{0, 4, 7}
	case "minor", "min", "m":
		return []int{0, 3, 7}
	case "maj7", "M7":
		return []int{0, 4, 7, 11}
	case "min7", "m7":
		return []int{0, 3, 7, 10}
	case "7", "dom7":
		return []int{0, 4, 7, 10}
	case "dim", "diminished":
		return []int{0, 3, 6}
	case "aug", "augmented":
		return []int{0, 4, 8}
	case "sus2":
		return []int{0, 2, 7}
	case "sus4":
		return []int{0, 5, 7}
	case "m7b5", "halfdim":
		return []int{0, 3, 6, 10}
	case "maj9":
		return []int{0, 4, 7, 11, 14}
	case "min9", "m9":
		return []int{0, 3, 7, 10, 14}
	case "9":
		return []int{0, 4, 7, 10, 14}
	case "add9":
		return []int{0, 4, 7, 14}
	case "6":
		return []int{0, 4, 7, 9}
	case "m6":
		return []int{0, 3, 7, 9}
	default:
		return []int{0, 4, 7}
	}
}

// Voicing aplica una disposición de acorde (lista de semitonos) al patrón.
func (p Pattern) Voicing(intervals ...float64) Pattern {
	if len(intervals) == 0 {
		return p
	}
	pats := make([]Pattern, len(intervals))
	for i, iv := range intervals {
		pats[i] = p.Add(Pure(iv))
	}
	return Stack(pats...)
}

// AddVoicings añade varias voces (lista de listas de semitonos).
func (p Pattern) AddVoicings(voicings [][]float64) Pattern {
	if len(voicings) == 0 {
		return p
	}
	pats := make([]Pattern, 0, len(voicings))
	for _, v := range voicings {
		pats = append(pats, p.Voicing(v...))
	}
	return Stack(pats...)
}

// scaleSteps devuelve los intervalos semitónicos de una escala.
func scaleSteps(name string) []int {
	switch name {
	case "major", "ionian", "":
		return []int{0, 2, 4, 5, 7, 9, 11}
	case "minor", "aeolian", "natminor":
		return []int{0, 2, 3, 5, 7, 8, 10}
	case "dorian":
		return []int{0, 2, 3, 5, 7, 9, 10}
	case "phrygian":
		return []int{0, 1, 3, 5, 7, 8, 10}
	case "lydian":
		return []int{0, 2, 4, 6, 7, 9, 11}
	case "mixolydian":
		return []int{0, 2, 4, 5, 7, 9, 10}
	case "locrian":
		return []int{0, 1, 3, 5, 6, 8, 10}
	case "harmonicminor", "harmminor":
		return []int{0, 2, 3, 5, 7, 8, 11}
	case "melodicminor", "melminor":
		return []int{0, 2, 3, 5, 7, 9, 11}
	case "penta", "majorpenta":
		return []int{0, 2, 4, 7, 9}
	case "minorpenta", "minpenta":
		return []int{0, 3, 5, 7, 10}
	case "blues":
		return []int{0, 3, 5, 6, 7, 10}
	case "egyptian":
		return []int{0, 2, 5, 7, 10}
	case "japanese", "in":
		return []int{0, 2, 5, 7, 9}
	case "chromatic":
		return []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11}
	default:
		return []int{0, 2, 4, 5, 7, 9, 11}
	}
}

// Arp arpegia un patrón de notas (reproduce una tras otra).
func (p Pattern) Arp(kind string) Pattern {
	return p.ArpWith(kind, p)
}

// ArpWith combina un patrón de notas con un patrón de arte (orden de arpegio).
func (p Pattern) ArpWith(kind string, notes Pattern) Pattern {
	// kind: "up","down","updown","downup","coin","random","ratchet"
	switch kind {
	case "down":
		return p.Rev()
	case "updown":
		return Stack(p, p.Rev().Early(FracInt(1)))
	case "downup":
		return Stack(p.Rev(), p.Early(FracInt(1)))
	case "random", "rand":
		return p.Scramble()
	default:
		return p
	}
}

// Partial añade armónicos (síntesis aditiva). n = número de armónicos.
func (p Pattern) Partial(n int) Pattern {
	if n <= 1 {
		return p
	}
	pats := make([]Pattern, n)
	for i := 0; i < n; i++ {
		pats[i] = p.Add(Pure(float64(i * 12)))
	}
	return Stack(pats...)
}

// --- MIDI ---

// Midinote fija la nota MIDI directamente.
func (p Pattern) Midinote(v float64) Pattern { return p.AddParam("midinote", v) }

// Midichan canal MIDI.
func (p Pattern) Midichan(v float64) Pattern { return p.AddParam("midichan", v) }

// Midicmd comando MIDI ("noteon","noteoff","cc","prog", etc).
func (p Pattern) Midicmd(v string) Pattern { return p.AddParam("midicmd", v) }

// Ccn número de control (CC).
func (p Pattern) Ccn(v float64) Pattern { return p.AddParam("ccn", v) }

// Ccv valor de control (CC).
func (p Pattern) Ccv(v float64) Pattern { return p.AddParam("ccv", v) }

// Control define un mensaje de control MIDI [ccn, ccv].
func (p Pattern) Control(ccn, ccv float64) Pattern {
	return p.AddParam("ccn", ccn).AddParam("ccv", ccv)
}

// Sysex mensaje SysEx.
func (p Pattern) Sysex(v []float64) Pattern { return p.AddParam("sysex", v) }

// Sysexdata datos de SysEx.
func (p Pattern) Sysexdata(v []float64) Pattern { return p.AddParam("sysexdata", v) }

// Sysexid identificador de SysEx.
func (p Pattern) Sysexid(v []float64) Pattern { return p.AddParam("sysexid", v) }

// ProgNum número de programa (patch).
func (p Pattern) ProgNum(v float64) Pattern { return p.AddParam("progNum", v) }

// Pitchwheel rueda de pitch.
func (p Pattern) Pitchwheel(v float64) Pattern { return p.AddParam("pitchwheel", v) }

// Nrpnn NRPN número.
func (p Pattern) Nrpnn(v float64) Pattern { return p.AddParam("nrpnn", v) }

// Nrpv NRPN valor.
func (p Pattern) Nrpv(v float64) Pattern { return p.AddParam("nrpv", v) }

// Midi marca el patrón como salida MIDI.
func (p Pattern) Midi(v string) Pattern { return p.AddParam("midi", v) }

// --- Síntesis FM ---

// Fm cantidad de modulación de frecuencia.
func (p Pattern) Fm(v float64) Pattern { return p.AddParam("fm", v) }

// Fmi índice de modulación FM.
func (p Pattern) Fmi(v float64) Pattern { return p.AddParam("fmi", v) }

// Fmh armónico del modulador FM.
func (p Pattern) Fmh(v float64) Pattern { return p.AddParam("fmh", v) }

// Fmw forma de onda del modulador FM.
func (p Pattern) Fmw(v float64) Pattern { return p.AddParam("fmwave", v) }

// Fmattack ataque del envelope FM.
func (p Pattern) Fmattack(v float64) Pattern { return p.AddParam("fmattack", v) }

// Fmdecay decaimiento del envelope FM.
func (p Pattern) Fmdecay(v float64) Pattern { return p.AddParam("fmdecay", v) }

// Fmsustain sustain del envelope FM.
func (p Pattern) Fmsustain(v float64) Pattern { return p.AddParam("fmsustain", v) }

// Fmrelease liberación del envelope FM.
func (p Pattern) Fmrelease(v float64) Pattern { return p.AddParam("fmrelease", v) }

// Fmenv envelope FM completo.
func (p Pattern) Fmenv(v []float64) Pattern { return p.AddParam("fmenv", v) }

// --- Wavetable ---

// Wt posición en la wavetable (0..1).
func (p Pattern) Wt(v float64) Pattern { return p.AddParam("wt", v) }

// Wtattack ataque del warp de wavetable.
func (p Pattern) Wtattack(v float64) Pattern { return p.AddParam("wtattack", v) }

// Wtdecay decaimiento del warp de wavetable.
func (p Pattern) Wtdecay(v float64) Pattern { return p.AddParam("wtdecay", v) }

// Wtdepth profundidad del warp de wavetable.
func (p Pattern) Wtdepth(v float64) Pattern { return p.AddParam("wtdepth", v) }

// Wtrate velocidad del barrido de wavetable.
func (p Pattern) Wtrate(v float64) Pattern { return p.AddParam("wtrate", v) }

// Wtrelease liberación del warp de wavetable.
func (p Pattern) Wtrelease(v float64) Pattern { return p.AddParam("wtrelease", v) }

// Wtshape forma del warp de wavetable.
func (p Pattern) Wtshape(v float64) Pattern { return p.AddParam("wtshape", v) }

// Wtsustain sustain del warp de wavetable.
func (p Pattern) Wtsustain(v float64) Pattern { return p.AddParam("wtsustain", v) }

// Wtphaserand aleatoriedad de fase de wavetable.
func (p Pattern) Wtphaserand(v float64) Pattern { return p.AddParam("wtphaserand", v) }

// --- Bandas de filtro dinámico (envelopes de filtro) ---

// Lpattack ataque del envelope de corte.
func (p Pattern) Lpattack(v float64) Pattern { return p.AddParam("lpattack", v) }

// Lpdecay decaimiento del envelope de corte.
func (p Pattern) Lpdecay(v float64) Pattern { return p.AddParam("lpdecay", v) }

// Lpdepth profundidad del envelope de corte.
func (p Pattern) Lpdepth(v float64) Pattern { return p.AddParam("lpdepth", v) }

// Lprate velocidad del barrido de corte.
func (p Pattern) Lprate(v float64) Pattern { return p.AddParam("lprate", v) }

// Lpsustain sustain del envelope de corte.
func (p Pattern) Lpsustain(v float64) Pattern { return p.AddParam("lpsustain", v) }

// Lpsync sincronía del barrido de corte.
func (p Pattern) Lpsync(v float64) Pattern { return p.AddParam("lpsync", v) }

// Hpattack ataque del envelope paso-alto.
func (p Pattern) Hpattack(v float64) Pattern { return p.AddParam("hpattack", v) }

// Hpdecay decaimiento del envelope paso-alto.
func (p Pattern) Hpdecay(v float64) Pattern { return p.AddParam("hpdecay", v) }

// Hpdepth profundidad del envelope paso-alto.
func (p Pattern) Hpdepth(v float64) Pattern { return p.AddParam("hpdepth", v) }

// Hprate velocidad del barrido paso-alto.
func (p Pattern) Hprate(v float64) Pattern { return p.AddParam("hprate", v) }

// Hpsustain sustain del envelope paso-alto.
func (p Pattern) Hpsustain(v float64) Pattern { return p.AddParam("hpsustain", v) }

// Hpsync sincronía del barrido paso-alto.
func (p Pattern) Hpsync(v float64) Pattern { return p.AddParam("hpsync", v) }

// Bpattack ataque del envelope paso-banda.
func (p Pattern) Bpattack(v float64) Pattern { return p.AddParam("bpattack", v) }

// Bpdecay decaimiento del envelope paso-banda.
func (p Pattern) Bpdecay(v float64) Pattern { return p.AddParam("bpdecay", v) }

// Bpdepth profundidad del envelope paso-banda.
func (p Pattern) Bpdepth(v float64) Pattern { return p.AddParam("bpdepth", v) }

// Bprate velocidad del barrido paso-banda.
func (p Pattern) Bprate(v float64) Pattern { return p.AddParam("bprate", v) }

// Bpsustain sustain del envelope paso-banda.
func (p Pattern) Bpsustain(v float64) Pattern { return p.AddParam("bpsustain", v) }

// Bpsync sincronía del barrido paso-banda.
func (p Pattern) Bpsync(v float64) Pattern { return p.AddParam("bpsync", v) }

// --- Efectos extra ---

// Squiz (twist) efecto de "squeeze".
func (p Pattern) Squiz(v float64) Pattern { return p.AddParam("squiz", v) }

// Real cantidad de "real" (gate).
func (p Pattern) Real(v float64) Pattern { return p.AddParam("real", v) }

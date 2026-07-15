package pattern

// ============================================================
// Parámetros de Control de Audio (port de strudel params)
// ============================================================
// Cada método añade/clave un parámetro al mapa de control del evento.
// Encadenables: p.gain(0.8).cutoff(1200).room(0.3)
//
// NOTA: gain, pan, note, n, s, octave, coarse, attack, sustain, release,
// shape, cutoff, crush, delay, reverb, room, begin, end, speed ya existen
// en params.go con firma (...interface{}). Aquí solo se añaden los nuevos.

// --- Ganancia / mezcla ---

// Postgain ganancia aplicada tras los efectos.
func (p Pattern) Postgain(v float64) Pattern { return p.AddParam("postgain", v) }

// PanWidth dispersa el sonido alrededor del panorama actual.
func (p Pattern) PanWidth(v float64) Pattern { return p.AddParam("panwidth", v) }

// Octaves es alias de Octave (desplazamiento por octavas).
func (p Pattern) Octaves(v float64) Pattern { return p.AddParam("octave", v) }

// --- Pitch / afinación ---

// Detune desplazamiento fino de tono en semitonos.
func (p Pattern) Detune(v float64) Pattern { return p.AddParam("detune", v) }

// Freq fuerza la frecuencia en Hz.
func (p Pattern) Freq(v float64) Pattern { return p.AddParam("freq", v) }

// Ftranspose transposición de frecuencia en semitonos.
func (p Pattern) Ftranspose(v float64) Pattern { return p.AddParam("ftranspose", v) }

// --- Envolvente (ADSR) ---

// Decay tiempo de decaimiento (segundos).
func (p Pattern) Decay(v float64) Pattern { return p.AddParam("decay", v) }

// Adsr combina attack/decay/sustain/release de una vez.
func (p Pattern) Adsr(attack, decay, sustain, release float64) Pattern {
	return p.Attack(attack).Decay(decay).Sustain(sustain).Release(release)
}

// Env define la envolvente completa (vector de puntos).
func (p Pattern) Env(e []float64) Pattern { return p.AddParam("env", e) }

// --- Filtros ---

// Lpq resonancia (Q) del filtro paso-bajo.
func (p Pattern) Lpq(v float64) Pattern { return p.AddParam("lpq", v) }

// Hpf frecuencia de corte del filtro paso-alto (Hz).
func (p Pattern) Hpf(v float64) Pattern { return p.AddParam("hpf", v) }

// Hpq resonancia (Q) del filtro paso-alto.
func (p Pattern) Hpq(v float64) Pattern { return p.AddParam("hpq", v) }

// Bpf frecuencia de corte del filtro paso-banda (Hz).
func (p Pattern) Bpf(v float64) Pattern { return p.AddParam("bpf", v) }

// Bpq resonancia (Q) del filtro paso-banda.
func (p Pattern) Bpq(v float64) Pattern { return p.AddParam("bpq", v) }

// Bandf alias de Bpf.
func (p Pattern) Bandf(v float64) Pattern { return p.AddParam("bpf", v) }

// Bandq alias de Bpq.
func (p Pattern) Bandq(v float64) Pattern { return p.AddParam("bpq", v) }

// Djf filtro DJ (barrido -1..1: -1 paso-alto, +1 paso-bajo).
func (p Pattern) Djf(v float64) Pattern { return p.AddParam("djf", v) }

// --- Distorsión / bitcrush ---

// Distort cantidad de distorsión.
func (p Pattern) Distort(v float64) Pattern { return p.AddParam("distort", v) }

// Dist alias de Distort.
func (p Pattern) Dist(v float64) Pattern { return p.AddParam("distort", v) }

// Drive alias de Distort.
func (p Pattern) Drive(v float64) Pattern { return p.AddParam("distort", v) }

// Distorttype tipo de distorsión.
func (p Pattern) Distorttype(v string) Pattern { return p.AddParam("distorttype", v) }

// --- Vibrato ---

// Vib cantidad de vibrato.
func (p Pattern) Vib(v float64) Pattern { return p.AddParam("vib", v) }

// Vibmod velocidad de vibrato.
func (p Pattern) Vibmod(v float64) Pattern { return p.AddParam("vibmod", v) }

// --- Vowel / formante ---

// Vowel filtro de formantes ("a","e","i","o","u").
func (p Pattern) Vowel(v string) Pattern { return p.AddParam("vowel", v) }

// Ribbon controlador ribbon (0..1).
func (p Pattern) Ribbon(v float64) Pattern { return p.AddParam("ribbon", v) }

// --- Tremolo ---

// Tremolo cantidad de tremolo.
func (p Pattern) Tremolo(v float64) Pattern { return p.AddParam("tremolo", v) }

// Tremolorate velocidad de tremolo.
func (p Pattern) Tremolorate(v float64) Pattern { return p.AddParam("tremolorate", v) }

// Tremolodepth profundidad de tremolo.
func (p Pattern) Tremolodepth(v float64) Pattern { return p.AddParam("tremolodepth", v) }

// --- Phaser ---

// Phaser cantidad de phaser.
func (p Pattern) Phaser(v float64) Pattern { return p.AddParam("phaser", v) }

// Phasercenter frecuencia central del phaser.
func (p Pattern) Phasercenter(v float64) Pattern { return p.AddParam("phasercenter", v) }

// Phaserdepth profundidad del phaser.
func (p Pattern) Phaserdepth(v float64) Pattern { return p.AddParam("phaserdepth", v) }

// Phasersweep barrido del phaser.
func (p Pattern) Phasersweep(v float64) Pattern { return p.AddParam("phasersweep", v) }

// --- Chorus ---

// Chorus cantidad de chorus.
func (p Pattern) Chorus(v float64) Pattern { return p.AddParam("chorus", v) }

// --- Comb (filtro de peine) ---

// Comb cantidad de comb filter.
func (p Pattern) Comb(v float64) Pattern { return p.AddParam("comb", v) }

// --- LFO (oscilador de baja frecuencia) ---

// Lfo aplica un LFO a un parámetro dado (param, rate, depth).
func (p Pattern) Lfo(param string, rate, depth float64) Pattern {
	return p.SetParam(param, LfoSignal(rate, depth))
}

// --- Delay ---

// Delaytime tiempo del eco (segundos).
func (p Pattern) Delaytime(v float64) Pattern { return p.AddParam("delaytime", v) }

// Delayfeedback realimentación del eco (0..1).
func (p Pattern) Delayfeedback(v float64) Pattern { return p.AddParam("delayfeedback", v) }

// Delaysync sincronía del eco (0 o 1).
func (p Pattern) Delaysync(v float64) Pattern { return p.AddParam("delaysync", v) }

// Echo alias de Delay.
func (p Pattern) Echo(v float64) Pattern { return p.AddParam("delay", v) }

// --- Room / reverb ---

// Roomsize tamaño de la sala.
func (p Pattern) Roomsize(v float64) Pattern { return p.AddParam("roomsize", v) }

// Roomdim dimensión de la sala.
func (p Pattern) Roomdim(v float64) Pattern { return p.AddParam("roomdim", v) }

// Roomfade decaimiento de la sala (segundos).
func (p Pattern) Roomfade(v float64) Pattern { return p.AddParam("roomfade", v) }

// Roomlp filtro paso-bajo de la sala.
func (p Pattern) Roomlp(v float64) Pattern { return p.AddParam("roomlp", v) }

// Size tamaño (alias de roomsize).
func (p Pattern) Size(v float64) Pattern { return p.AddParam("roomsize", v) }

// --- Sample / fuente ---

// Source fuente de audio (alias de s).
func (p Pattern) Source(v string) Pattern { return p.AddParam("s", v) }

// Src alias de S.
func (p Pattern) Src(v string) Pattern { return p.AddParam("s", v) }

// Accelerate aceleración del sample (cambio de velocidad).
func (p Pattern) Accelerate(v float64) Pattern { return p.AddParam("accelerate", v) }

// Rate velocidad de reproducción (alias de speed).
func (p Pattern) Rate(v float64) Pattern { return p.AddParam("speed", v) }

// LoopBegin inicio del loop del sample.
func (p Pattern) LoopBegin(v float64) Pattern { return p.AddParam("loopbegin", v) }

// LoopEnd fin del loop del sample.
func (p Pattern) LoopEnd(v float64) Pattern { return p.AddParam("loopend", v) }

// LoopSample hace loop del sample entre begin y end.
func (p Pattern) LoopSample(begin, end float64) Pattern {
	return p.AddParam("loopbegin", begin).AddParam("loopend", end)
}

// Cut detiene sonidos con el mismo nombre de corte (choking).
func (p Pattern) Cut(v string) Pattern { return p.AddParam("cut", v) }

// --- Bank / aliases de sonidos ---

// Bank banco de samples.
func (p Pattern) Bank(v string) Pattern { return p.AddParam("bank", v) }

// AliasBank banco de alias de sonidos.
func (p Pattern) AliasBank(v string) Pattern { return p.AddParam("aliasbank", v) }

// SoundAlias alias de sonido.
func (p Pattern) SoundAlias(v string) Pattern { return p.AddParam("soundalias", v) }

// --- Enrutamiento ---

// Orbit pista de salida (bus).
func (p Pattern) Orbit(v float64) Pattern { return p.AddParam("orbit", v) }

// Channel canal de salida.
func (p Pattern) Channel(v float64) Pattern { return p.AddParam("channel", v) }

// Channels alias de Channel.
func (p Pattern) Channels(v float64) Pattern { return p.AddParam("channel", v) }

// --- Etiquetado / visualización ---

// Label etiqueta del evento.
func (p Pattern) Label(v string) Pattern { return p.AddParam("label", v) }

// Tag etiqueta (alias).
func (p Pattern) Tag(v string) Pattern { return p.AddParam("tag", v) }

// Color color de visualización.
func (p Pattern) Color(v string) Pattern { return p.AddParam("color", v) }

// Colour alias de Color.
func (p Pattern) Colour(v string) Pattern { return p.AddParam("color", v) }

// --- Multiparam helpers ---

// AddParam añade un parámetro escalar al patrón.
func (p Pattern) AddParam(key string, value interface{}) Pattern {
	return p.WithValue(func(v interface{}) interface{} {
		return UnionControls(ToControlMap(v), ControlMap{key: value})
	})
}

// SetParam usa otro patrón como fuente del parámetro.
func (p Pattern) SetParam(key string, other Pattern) Pattern {
	return p.WithValue(func(v interface{}) interface{} {
		return UnionControls(ToControlMap(v), ControlMap{key: other})
	})
}

// LfoSignal crea un patrón de señal LFO (usado por .lfo()).
func LfoSignal(rate, depth float64) Pattern {
	return Sine2.Mul(Pure(depth)).Add(Pure(depth))
}

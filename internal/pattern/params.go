package pattern

// ControlParam creates a standalone param function and a method-style wrapper.
// When called standalone: Note("c3") → pattern with {note: "c3"}
// When called as method: p.Note("c3") → p unioned with {note: "c3"}

// makeParam returns a standalone function that creates a param pattern.
func makeParam(key string) func(args ...interface{}) Pattern {
	return func(args ...interface{}) Pattern {
		pat := Sequence(args...)
		return pat.WithValue(func(v interface{}) interface{} {
			return ControlMap{key: v}
		})
	}
}

// makeParamMethod returns a method on Pattern that unions the param.
func makeParamMethod(key string) func(p Pattern, args ...interface{}) Pattern {
	paramFn := makeParam(key)
	return func(p Pattern, args ...interface{}) Pattern {
		return p.Union(paramFn(args...))
	}
}

// --- Standalone param functions ---

var NoteFn = makeParam("note")
var NFn = makeParam("n")
var SFn = makeParam("s")
var GainFn = makeParam("gain")
var PanFn = makeParam("pan")
var CutoffFn = makeParam("cutoff")
var ResonanceFn = makeParam("resonance")
var DelayFn = makeParam("delay")
var DelayTimeFn = makeParam("delaytime")
var ReverbFn = makeParam("reverb")
var AttackFn = makeParam("attack")
var ReleaseFn = makeParam("release")
var SustainFn = makeParam("sustain")
var ShapeFn = makeParam("shape")
var BeginFn = makeParam("begin")
var EndFn = makeParam("end")
var LegatoFn = makeParam("legato")
var CoarseFn = makeParam("coarse")
var HPFFn = makeParam("hpf")
var LPFFn = makeParam("lpf")
var BandpassFn = makeParam("bandf")
var CrushFn = makeParam("crush")
var LoopFn = makeParam("loop")
var ChannelFn = makeParam("channel")
var AccelFn = makeParam("accelerate")
var SpeedFn = makeParam("speed")
var UpFn = makeParam("up")
var DownFn = makeParam("down")
var OctaveFn = makeParam("octave")

// --- Method-style param wrappers on Pattern ---

// Note fija la nota. Sin argumentos, usa los valores propios del patrón.
func (p Pattern) Note(args ...interface{}) Pattern {
	if len(args) == 0 {
		return p.WithValue(func(v interface{}) interface{} {
			return UnionControls(ToControlMap(v), ControlMap{"note": v})
		})
	}
	return makeParamMethod("note")(p, args...)
}

// N es alias de Note.
func (p Pattern) N(args ...interface{}) Pattern {
	if len(args) == 0 {
		return p.WithValue(func(v interface{}) interface{} {
			return UnionControls(ToControlMap(v), ControlMap{"n": v})
		})
	}
	return makeParamMethod("n")(p, args...)
}

// S selecciona el sonido. Sin argumentos, usa los valores propios.
func (p Pattern) S(args ...interface{}) Pattern {
	if len(args) == 0 {
		return p.WithValue(func(v interface{}) interface{} {
			return UnionControls(ToControlMap(v), ControlMap{"s": v})
		})
	}
	return makeParamMethod("s")(p, args...)
}
func (p Pattern) Gain(args ...interface{}) Pattern   { return makeParamMethod("gain")(p, args...) }
func (p Pattern) Pan(args ...interface{}) Pattern    { return makeParamMethod("pan")(p, args...) }
func (p Pattern) Cutoff(args ...interface{}) Pattern { return makeParamMethod("cutoff")(p, args...) }
func (p Pattern) Resonance(args ...interface{}) Pattern {
	return makeParamMethod("resonance")(p, args...)
}
func (p Pattern) Delay(args ...interface{}) Pattern { return makeParamMethod("delay")(p, args...) }
func (p Pattern) DelayTime(args ...interface{}) Pattern {
	return makeParamMethod("delaytime")(p, args...)
}
func (p Pattern) Reverb(args ...interface{}) Pattern  { return makeParamMethod("reverb")(p, args...) }
func (p Pattern) Attack(args ...interface{}) Pattern  { return makeParamMethod("attack")(p, args...) }
func (p Pattern) Release(args ...interface{}) Pattern { return makeParamMethod("release")(p, args...) }
func (p Pattern) Sustain(args ...interface{}) Pattern { return makeParamMethod("sustain")(p, args...) }
func (p Pattern) Shape(args ...interface{}) Pattern   { return makeParamMethod("shape")(p, args...) }
func (p Pattern) Begin(args ...interface{}) Pattern   { return makeParamMethod("begin")(p, args...) }
func (p Pattern) End(args ...interface{}) Pattern     { return makeParamMethod("end")(p, args...) }

// NOTE: Legato is declared in pattern.go with a different signature (Fraction param)
func (p Pattern) Coarse(args ...interface{}) Pattern   { return makeParamMethod("coarse")(p, args...) }
func (p Pattern) HPF(args ...interface{}) Pattern      { return makeParamMethod("hpf")(p, args...) }
func (p Pattern) LPF(args ...interface{}) Pattern      { return makeParamMethod("lpf")(p, args...) }
func (p Pattern) Bandpass(args ...interface{}) Pattern { return makeParamMethod("bandf")(p, args...) }
func (p Pattern) Crush(args ...interface{}) Pattern    { return makeParamMethod("crush")(p, args...) }
func (p Pattern) Speed(args ...interface{}) Pattern    { return makeParamMethod("speed")(p, args...) }
func (p Pattern) Up(args ...interface{}) Pattern       { return makeParamMethod("up")(p, args...) }
func (p Pattern) Down(args ...interface{}) Pattern     { return makeParamMethod("down")(p, args...) }
func (p Pattern) Octave(args ...interface{}) Pattern   { return makeParamMethod("octave")(p, args...) }
func (p Pattern) Room(args ...interface{}) Pattern     { return makeParamMethod("reverb")(p, args...) }

// NOTE: Legato is defined in pattern.go
// NOTE: Scale is defined in music_midi.go (Strudel-correct signature)

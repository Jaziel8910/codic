package pattern

// ============================================================
// Más funciones de Strudel (ritmos, slicing, mezcla, MIDI)
// ============================================================

// --- Ritmos euclídeos ---

// euclidBools calcula la distribución euclídea de pulses en steps.
func euclidBools(pulses, steps int) []bool {
	res := make([]bool, steps)
	if pulses <= 0 {
		return res
	}
	if pulses >= steps {
		for i := range res {
			res[i] = true
		}
		return res
	}
	for i := 0; i < pulses; i++ {
		pos := int((float64(i) * float64(steps)) / float64(pulses))
		if pos >= steps {
			pos = steps - 1
		}
		res[pos] = true
	}
	return res
}

// Euclid genera un ritmo euclídeo (patrón de booleanos) sobre un ciclo.
// steps = subdivisiones, pulses = golpes, rot = rotación.
func Euclid(steps, pulses, rot int) Pattern {
	b := euclidBools(pulses, steps)
	if steps > 0 {
		rot = ((rot % steps) + steps) % steps
		if rot != 0 {
			b = append(append([]bool{}, b[rot:]...), b[:rot]...)
		}
	}
	pats := make([]interface{}, len(b))
	for i, v := range b {
		pats[i] = v
	}
	return Sequence(pats...)
}

// EuclidRot = Euclid con rotación.
func EuclidRot(steps, pulses, rot int) Pattern { return Euclid(steps, pulses, rot) }

// EuclidLegato igual que Euclid pero con legato (notas sostenidas).
func EuclidLegato(steps, pulses, rot int) Pattern {
	return Euclid(steps, pulses, rot).Legato(FracFloat(1))
}

// Euclidish aproximación euclídea con pasos/longitudes variables.
func Euclidish(groups []int, pulses int) Pattern {
	var pats []interface{}
	total := 0
	for _, g := range groups {
		total += g
	}
	b := euclidBools(pulses, total)
	idx := 0
	for _, g := range groups {
		grp := make([]interface{}, 0, g)
		for i := 0; i < g && idx < len(b); i++ {
			grp = append(grp, b[idx])
			idx++
		}
		pats = append(pats, Sequence(grp...))
	}
	return Sequence(pats...)
}

// --- Método euclid en patrón (compuerta el patrón con el ritmo) ---

// Euclid aplica un ritmo euclídeo como máscara sobre el patrón.
func (p Pattern) Euclid(steps, pulses, rot int) Pattern {
	return p.Struct(Euclid(steps, pulses, rot))
}

// EuclidRot método con rotación.
func (p Pattern) EuclidRot(steps, pulses, rot int) Pattern {
	return p.Struct(EuclidRot(steps, pulses, rot))
}

// EuclidLegato método con legato.
func (p Pattern) EuclidLegato(steps, pulses, rot int) Pattern {
	return p.Struct(EuclidLegato(steps, pulses, rot))
}

// Euclidish método con grupos variables.
func (p Pattern) Euclidish(groups []int, pulses int) Pattern {
	return p.Struct(Euclidish(groups, pulses))
}

// --- Jux con parámetros ---

// JuxBy aplica f al canal derecho con un desplazamiento de panorama dado.
func (p Pattern) JuxBy(by float64, f func(Pattern) Pattern) Pattern {
	left := p.WithValue(func(v interface{}) interface{} {
		return UnionControls(ToControlMap(v), ControlMap{"pan": 0.5 - by})
	})
	right := p.WithValue(func(v interface{}) interface{} {
		return UnionControls(ToControlMap(v), ControlMap{"pan": 0.5 + by})
	})
	return Stack(left, f(right))
}

// JuxFlip aplica f al canal izquierdo en vez del derecho.
func (p Pattern) JuxFlip(f func(Pattern) Pattern) Pattern {
	return p.JuxBy(0.5, func(sub Pattern) Pattern {
		return f(sub).WithValue(func(v interface{}) interface{} {
			return UnionControls(ToControlMap(v), ControlMap{"pan": -0.5})
		})
	})
}

// JuxFlipBy como JuxFlip pero con desplazamiento dado.
func (p Pattern) JuxFlipBy(by float64, f func(Pattern) Pattern) Pattern {
	return p.JuxBy(by, func(sub Pattern) Pattern {
		return f(sub).WithValue(func(v interface{}) interface{} {
			return UnionControls(ToControlMap(v), ControlMap{"pan": -by})
		})
	})
}

// --- Compress (función) ---

// Compress comprime este patrón dentro de la sub-porción [begin,end] del ciclo.
func (p Pattern) Compress(begin, end Fraction) Pattern {
	return p.CompressSpan(TimeSpan{Begin: begin, End: end})
}

// --- Unison ---

// Unison apila n copias del patrón con pequeña desafinación (detune).
func (p Pattern) Unison(n int) Pattern {
	if n <= 1 {
		return p
	}
	pats := make([]Pattern, n)
	for i := 0; i < n; i++ {
		det := float64(i-(n/2)) * 0.07
		pats[i] = p.Detune(det)
	}
	return Stack(pats...)
}

// --- Crossfade / xfade ---

// Xfade mezcla dos patrones con proporción t (0 = a, 1 = b).
func Xfade(a, b Pattern, t float64) Pattern {
	return Stack(a.Gain(1-t), b.Gain(t))
}

// Crossfade alias de Xfade.
func Crossfade(a, b Pattern, t float64) Pattern { return Xfade(a, b, t) }

// --- Conversión MIDI ---

// ToMidi convierte nombres de nota a números MIDI en los valores del patrón.
func (p Pattern) ToMidi() Pattern {
	return p.asNumber().WithValue(func(v interface{}) interface{} {
		if s, ok := v.(string); ok {
			if m, ok2 := noteToMidi(s); ok2 {
				return m
			}
		}
		return v
	})
}

// FromMidi convierte números MIDI a nombres de nota en los valores.
func (p Pattern) FromMidi() Pattern {
	return p.asNumber().WithValue(func(v interface{}) interface{} {
		if f, ok := v.(float64); ok {
			return midiToNote(int(f + 0.5))
		}
		return v
	})
}

// midiToNote convierte un número MIDI a nombre (p.ej. 60 -> "c5").
func midiToNote(m int) string {
	names := []string{"c", "cs", "d", "ds", "e", "f", "fs", "g", "gs", "a", "as", "b"}
	oct := (m / 12) - 1
	return names[((m%12)+12)%12] + itoa(oct)
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	digits := ""
	for n > 0 {
		digits = string(rune('0'+n%10)) + digits
		n /= 10
	}
	if neg {
		digits = "-" + digits
	}
	return digits
}

// --- Branding (etiquetado para resaltado) ---

// Brand marca los eventos con un identificador (para resaltado en vivo).
func (p Pattern) Brand(id string) Pattern { return p.AddParam("brand", id) }

// BrandBy marca con un id derivado.
func (p Pattern) BrandBy(id string) Pattern { return p.AddParam("brand", id) }

// --- Zip / combinación ---

// Zip combina dos patrones intercalando sus valores.
func Zip(a, b Pattern) Pattern {
	return Stack(a, b)
}

// Unzip separa (identidad útil en composiciones).
func (p Pattern) Unzip() Pattern { return p }

// --- Step sequencing ---

// Stepcat concatena patrones "paso a paso" (cada uno un paso).
func Stepcat(pats ...Pattern) Pattern { return Fastcat(pats...) }

// Step crea un paso único (un evento por ciclo).
func Step(v interface{}) Pattern { return Sequence(v) }

// Stepwise encadena valores como pasos.
func Stepwise(vals ...interface{}) Pattern { return Sequence(vals...) }

// SeqPLoop repite una secuencia en bucle (alias de Sequence).
func SeqPLoop(vals ...interface{}) Pattern { return Sequence(vals...) }

// --- Spiral (arpegio en espiral) ---

// Spiral genera un patrón en espiral sobre una escala.
// scale = nombre, steps = pasos, rotary = salto, root = nota raíz.
func Spiral(scale string, steps, rotary, root int) Pattern {
	stepsArr := scaleSteps(scale)
	if len(stepsArr) == 0 {
		stepsArr = scaleSteps("major")
	}
	pats := make([]interface{}, 0, steps)
	for i := 0; i < steps; i++ {
		deg := (i*rotary + root) % len(stepsArr)
		if deg < 0 {
			deg += len(stepsArr)
		}
		oct := (i*rotary + root) / len(stepsArr)
		pats = append(pats, float64(stepsArr[deg])+float64(oct*12))
	}
	return Sequence(pats...)
}

// --- Setcpm (alias de cps en patrones) ---

// Setcpm fija los ciclos por minuto como parámetro de control.
func (p Pattern) Setcpm(cpm float64) Pattern { return p.AddParam("cpm", cpm) }

// --- Slicing de samples ---

// Striate divide el sample en n rebanadas distribuidas en el ciclo.
func (p Pattern) Striate(n int) Pattern {
	if n <= 1 {
		return p
	}
	pats := make([]Pattern, n)
	for i := 0; i < n; i++ {
		begin := float64(i) / float64(n)
		end := float64(i+1) / float64(n)
		pats[i] = p.Begin(begin).End(end).Early(FracFloat(begin))
	}
	return Stack(pats...).Fast(FracInt(int64(n)))
}

// Stripe alias de Striate (variante de canal).
func (p Pattern) Stripe(n int) Pattern { return p.Striate(n) }

// Chop rebanada secuencial del sample en n partes.
func (p Pattern) Chop(n int) Pattern {
	if n <= 1 {
		return p
	}
	pats := make([]Pattern, n)
	for i := 0; i < n; i++ {
		begin := float64(i) / float64(n)
		end := float64(i+1) / float64(n)
		pats[i] = p.Begin(begin).End(end)
	}
	return Fastcat(pats...)
}

// Slice alias de Chop.
func (p Pattern) Slice(n int) Pattern { return p.Chop(n) }

// Bite toma n rebanadas del sample entre from y to (0..1).
func (p Pattern) Bite(from, to float64, n int) Pattern {
	if n <= 0 {
		n = 1
	}
	pats := make([]Pattern, n)
	span := to - from
	for i := 0; i < n; i++ {
		begin := from + (span * float64(i) / float64(n))
		end := from + (span * float64(i+1) / float64(n))
		pats[i] = p.Begin(begin).End(end)
	}
	return Fastcat(pats...)
}

// Splice rebanada con solapamiento (alias de Chop con n mayor).
func (p Pattern) Splice(n int) Pattern { return p.Chop(n) }

// --- Gap introduce huecos (silencio) entre eventos ---

// Gap inserta silencio relativo entre eventos.
func (p Pattern) Gap(amount float64) Pattern {
	return p.WithEventSpan(func(ts TimeSpan) TimeSpan {
		dur := ts.End.Sub(ts.Begin)
		newDur := dur.Mul(FracFloat(1 - amount))
		return TimeSpan{Begin: ts.Begin, End: ts.Begin.Add(newDur)}
	})
}

// --- Reset / restart (marcado de control) ---

// Reset marca el patrón para reinicio de muestra.
func (p Pattern) Reset() Pattern { return p.AddParam("reset", true) }

// Restart alias de Reset.
func (p Pattern) Restart() Pattern { return p.Reset() }

// --- Defragment (une haps contiguos) ---

// DefragmentHaps une haps adyacentes con el mismo valor.
func (p Pattern) DefragmentHaps() Pattern {
	return p.WithEvents(func(haps []Hap) []Hap {
		if len(haps) == 0 {
			return haps
		}
		out := []Hap{haps[0]}
		for _, h := range haps[1:] {
			last := out[len(out)-1]
			if sameValue(last.Value, h.Value) && last.Part.End.Equals(h.Part.Begin) {
				out[len(out)-1] = Hap{Whole: last.Whole, Part: TimeSpan{Begin: last.Part.Begin, End: h.Part.End}, Value: h.Value, Context: last.Context}
			} else {
				out = append(out, h)
			}
		}
		return out
	})
}

func sameValue(a, b interface{}) bool {
	return toNumber(a) == toNumber(b)
}

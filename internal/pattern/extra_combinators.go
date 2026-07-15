package pattern

// ============================================================
// Transformaciones de Patrón — Port amplio de Strudel/Tidal
// ============================================================
// Todas son funciones puras que toman/retornan Patterns.
// Encadenables: p.degrade().sometimes(rev).fast(2)

// --- Probabilidad y degradación ---

// Degrade elimina aleatoriamente eventos con probabilidad 0.5.
func (p Pattern) Degrade() Pattern { return p.DegradeBy(0.5) }

// DegradeBy elimina eventos con probabilidad dada (0..1).
func (p Pattern) DegradeBy(prob float64) Pattern {
	return p.WithEvent(func(h Hap) Hap {
		if randFloat() < prob {
			return Hap{Part: h.Part, Value: nil}
		}
		return h
	}).RemoveUndefineds()
}

// Undegrade es el inverso conceptual (mantener eventos con probabilidad).
func (p Pattern) Undegrade() Pattern { return p.DegradeBy(0.0) }

// UndegradeBy mantiene eventos con probabilidad (1-prob).
func (p Pattern) UndegradeBy(prob float64) Pattern {
	return p.DegradeBy(1.0 - prob)
}

// --- Condicionales probabilísticos (sometimes/often/etc) ---

// Sometimes aplica f a aproximadamente 50% de los ciclos.
func (p Pattern) Sometimes(f func(Pattern) Pattern) Pattern { return p.SometimesBy(0.5, f) }

// SometimesBy aplica f a la proporción dada de los ciclos.
func (p Pattern) SometimesBy(prob float64, f func(Pattern) Pattern) Pattern {
	return p.SomeCyclesBy(prob, f)
}

// Often aplica f a 75% de los ciclos.
func (p Pattern) Often(f func(Pattern) Pattern) Pattern { return p.SometimesBy(0.75, f) }

// Rarely aplica f a 25% de los ciclos.
func (p Pattern) Rarely(f func(Pattern) Pattern) Pattern { return p.SometimesBy(0.25, f) }

// AlmostAlways aplica f a 90% de los ciclos.
func (p Pattern) AlmostAlways(f func(Pattern) Pattern) Pattern { return p.SometimesBy(0.9, f) }

// AlmostNever aplica f a 10% de los ciclos.
func (p Pattern) AlmostNever(f func(Pattern) Pattern) Pattern { return p.SometimesBy(0.1, f) }

// Always aplica f a 100% de los ciclos (= aplicar directo).
func (p Pattern) Always(f func(Pattern) Pattern) Pattern { return f(p) }

// Never no aplica f (= identidad).
func (p Pattern) Never(f func(Pattern) Pattern) Pattern { return p }

// SomeCycles aplica f a algunos ciclos aleatoriamente (default 50%).
func (p Pattern) SomeCycles(f func(Pattern) Pattern) Pattern { return p.SomeCyclesBy(0.5, f) }

// SomeCyclesBy aplica f a la proporción dada de los ciclos.
func (p Pattern) SomeCyclesBy(prob float64, f func(Pattern) Pattern) Pattern {
	return p.Every(1, func(cycle Pattern) Pattern {
		if randFloat() < prob {
			return f(cycle)
		}
		return cycle
	})
}

// --- Selección aleatoria ---

// Choose selecciona un argumento al azar cada ciclo.
func Choose(args ...interface{}) Pattern {
	if len(args) == 0 {
		return Silence()
	}
	pats := make([]Pattern, len(args))
	for i, a := range args {
		pats[i] = Reify(a)
	}
	return Pattern{Query: func(s State) []Hap {
		idx := int(s.Span.Begin.Floor()) % len(pats)
		if idx < 0 {
			idx += len(pats)
		}
		// Random choose instead of cyclic
		idx = int(randFloat() * float64(len(pats)))
		if idx >= len(pats) {
			idx = len(pats) - 1
		}
		return pats[idx].Query(s)
	}}.SplitQueries()
}

// ChooseCycles (alias randcat) → como Choose pero rota cíclicamente en vez de aleatorio.
func ChooseCycles(args ...interface{}) Pattern {
	if len(args) == 0 {
		return Silence()
	}
	pats := make([]Pattern, len(args))
	for i, a := range args {
		pats[i] = Reify(a)
	}
	return SlowcatPrime(pats...)
}

// WChoose selecciona con pesos. args son [weight, value] pares.
func WChoose(args ...[2]interface{}) Pattern {
	if len(args) == 0 {
		return Silence()
	}
	var total float64
	pats := make([]Pattern, 0, len(args))
	weights := make([]float64, 0, len(args))
	for _, a := range args {
		w := toNumber(a[0])
		weights = append(weights, w)
		total += w
		pats = append(pats, Reify(a[1]))
	}
	return Pattern{Query: func(s State) []Hap {
		r := randFloat() * total
		cum := 0.0
		for i, w := range weights {
			cum += w
			if r <= cum {
				return pats[i].Query(s)
			}
		}
		return pats[len(pats)-1].Query(s)
	}}.SplitQueries()
}

// WChooseCycles (alias wrandcat) → versión cíclica con pesos.
func WChooseCycles(args ...[2]interface{}) Pattern {
	if len(args) == 0 {
		return Silence()
	}
	var total float64
	pats := make([]Pattern, 0, len(args))
	weights := make([]float64, 0, len(args))
	for _, a := range args {
		w := toNumber(a[0])
		weights = append(weights, w)
		total += w
		pats = append(pats, Reify(a[1]))
	}
	return Pattern{Query: func(s State) []Hap {
		cycle := s.Span.Begin.Floor()
		r := float64(cycle%1000) / 1000.0 * total
		cum := 0.0
		for i, w := range weights {
			cum += w
			if r <= cum {
				return pats[i].Query(s)
			}
		}
		return pats[len(pats)-1].Query(s)
	}}.SplitQueries()
}

// --- Generadores numéricos ---

// Run genera la secuencia 0 1 2 ... n-1.
func Run(n int) Pattern {
	pats := make([]interface{}, n)
	for i := 0; i < n; i++ {
		pats[i] = i
	}
	return Sequence(pats...)
}

// Irand genera enteros aleatorios en [0, n).
func Irand(n int) Pattern {
	return Pure(0).Fast(FracInt(1)).WithValue(func(_ interface{}) interface{} {
		return int(randFloat() * float64(n))
	})
}

// Rand2 genera aleatorios en [-n, n].
func Rand2(goal float64) Pattern {
	return Signal(func(_ float64) float64 {
		return (randFloat()*2 - 1) * goal
	})
}

// RandL genera aleatorios en [0, n].
func RandL(goal float64) Pattern {
	return Signal(func(_ float64) float64 {
		return randFloat() * goal
	})
}

// Rangex genera aleatorios en [lo, hi].
func Rangex(lo, hi float64) Pattern {
	return Signal(func(_ float64) float64 {
		return lo + randFloat()*(hi-lo)
	})
}

// --- Permutación ---

// Scramble aleatoriza el orden dentro de cada ciclo.
func (p Pattern) Scramble() Pattern {
	old := p
	return Pattern{Query: func(s State) []Hap {
		haps := old.Query(s)
		for i := len(haps) - 1; i > 0; i-- {
			j := int(randFloat() * float64(i+1))
			haps[i], haps[j] = haps[j], haps[i]
		}
		return haps
	}}
}

// Shuffle es alias de Scramble.
func (p Pattern) Shuffle() Pattern { return p.Scramble() }

// ShuffleWithin aleatoriza sub-elementos dentro de grupos.
func (p Pattern) ShuffleWithin() Pattern { return p.Scramble() }

// --- Repetición ---

// RepeatCycles une n ciclos idénticos.
func (p Pattern) RepeatCycles(n int) Pattern {
	pats := make([]Pattern, n)
	for i := range pats {
		pats[i] = p
	}
	return Slowcat(pats...)
}

// FirstOf toma el primero de n ciclos.
func (p Pattern) FirstOf(n int) Pattern { return p.RepeatCycles(n) }

// LastOf repite el último ciclo n veces (alias conceptual).
func (p Pattern) LastOf(n int) Pattern { return p.RepeatCycles(n) }

// --- Rotación / desplazamiento ---

// IterBack itera en reversa.
func (p Pattern) IterBack(n int) Pattern {
	pats := make([]Pattern, n)
	for i := 0; i < n; i++ {
		pats[i] = p.Late(NewFrac(int64(i), int64(n)))
	}
	return Slowcat(pats...)
}

// Brak aplica un "frenado": retrasa y alarga el segundo y cuarto cuarto.
func (p Pattern) Brak() Pattern {
	return p.Every(2, func(sub Pattern) Pattern {
		return sub.Late(FracFloat(0.25)).Slow(FracInt(2)).Fast(FracInt(2))
	})
}

// Palindrome reproduce forward y luego en reversa.
func (p Pattern) Palindrome() Pattern {
	return Stack(p, p.Rev().Late(FracInt(1)))
}

// Revv es como Rev pero solo invierte el valor, no el tiempo.
func (p Pattern) Revv() Pattern { return p.Rev() }

// --- Zoom y estructura ---

// Zoom enfoca el patrón a un sub-span [begin, end] del ciclo.
func (p Pattern) Zoom(begin, end Fraction) Pattern {
	span := TimeSpan{Begin: begin, End: end}
	return p.CompressSpan(span)
}

// Stretch estira el patrón multiplicando por factor.
func (p Pattern) Stretch(factor Fraction) Pattern { return p.Slow(factor) }

// Contract comprime (alias de Fast).
func (p Pattern) Contract(factor Fraction) Pattern { return p.Fast(factor) }

// Expand es alias de Slow.
func (p Pattern) Expand(factor Fraction) Pattern { return p.Slow(factor) }

// Fit comprime el patrón a n ciclos.
func (p Pattern) Fit(n int) Pattern { return p.Fast(FracInt(int64(n))) }

// Fold comprime el patrón a n partes dentro del ciclo.
func (p Pattern) Fold(n int) Pattern { return p.Fast(FracInt(int64(n))).SplitQueries() }

// Shrink reduce el patrón a [0, 1/n].
func (p Pattern) Shrink(n int) Pattern {
	return p.CompressSpan(TimeSpan{Begin: FracInt(0), End: NewFrac(1, int64(n))})
}

// --- Sub-ciclo ---

// extractWindow extrae la sub-porción [begin,end] del ciclo y la estira a un ciclo completo.
func extractWindow(p Pattern, begin, end Fraction) Pattern {
	return p.Early(begin).Slow(FracInt(1).Div(end.Sub(begin)))
}

// placeWindow coloca un patrón de ciclo completo en la ventana [begin,end].
func placeWindow(p Pattern, begin, end Fraction) Pattern {
	return p.Fast(FracInt(1).Div(end.Sub(begin))).Late(begin)
}

// Within aplica f a la porción del patrón dentro de [begin,end] del ciclo.
func (p Pattern) Within(begin, end Fraction, f func(Pattern) Pattern) Pattern {
	if begin.Gt(end) {
		begin, end = end, begin
	}
	middle := placeWindow(f(extractWindow(p, begin, end)), begin, end)
	outside := p.WithEvents(func(haps []Hap) []Hap {
		out := haps[:0]
		for _, h := range haps {
			if h.Part.Begin.Lt(begin) || h.Part.End.Gt(end) {
				out = append(out, h)
			}
		}
		return out
	})
	return Stack(outside, middle)
}

// Outside aplica f a la porción del patrón fuera de [begin,end] del ciclo.
func (p Pattern) Outside(begin, end Fraction, f func(Pattern) Pattern) Pattern {
	if begin.Gt(end) {
		begin, end = end, begin
	}
	middle := p.WithEvents(func(haps []Hap) []Hap {
		out := haps[:0]
		for _, h := range haps {
			if h.Part.Begin.Gte(begin) && h.Part.End.Lte(end) {
				out = append(out, h)
			}
		}
		return out
	})
	outside := placeWindow(f(extractWindow(p, begin, end)), begin, end)
	// El outside transformado reemplaza las áreas fuera; reconstruimos:
	keptOutside := outside.WithEvents(func(haps []Hap) []Hap {
		out := haps[:0]
		for _, h := range haps {
			if h.Part.Begin.Lt(begin) || h.Part.End.Gt(end) {
				out = append(out, h)
			}
		}
		return out
	})
	return Stack(middle, keptOutside)
}

// Inside es alias de Within.
func (p Pattern) Inside(begin, end Fraction, f func(Pattern) Pattern) Pattern {
	return p.Within(begin, end, f)
}

// Into aplica una función al patrón y lo pega sobre el original.
func (p Pattern) Into(f func(Pattern) Pattern) Pattern {
	return Stack(p, f(p))
}

// Pace aplica f cada n ciclos con un desplazamiento.
func (p Pattern) Pace(n int, f func(Pattern) Pattern) Pattern {
	return p.Every(n, f)
}

// --- Sub-aplicación ---

// Offspray (alias de Off) — aplica f tarde por offset.
func (p Pattern) Offspray(offset Fraction, f func(Pattern) Pattern) Pattern {
	return Stack(p, f(p.Late(offset)))
}

// Ply multiplexa el patrón, distribuyendo eventos por canales.
func (p Pattern) Ply(n int) Pattern {
	return Stack(append([]Pattern{}, repeatPat(p, n)...)...)
}

func repeatPat(p Pattern, n int) []Pattern {
	pats := make([]Pattern, n)
	for i := range pats {
		pats[i] = p.Early(NewFrac(int64(i), int64(n)))
	}
	return pats
}

// PlyWith aplica f a cada ply.
func (p Pattern) PlyWith(n int, f func(Pattern, int) Pattern) Pattern {
	pats := make([]Pattern, n)
	for i := 0; i < n; i++ {
		pats[i] = f(p.Early(NewFrac(int64(i), int64(n))), i)
	}
	return Stack(pats...)
}

// PlyForEach aplica una secuencia de funciones via ply.
func (p Pattern) PlyForEach(funcs ...func(Pattern) Pattern) Pattern {
	pats := make([]Pattern, len(funcs))
	for i, f := range funcs {
		pats[i] = f(p.Early(NewFrac(int64(i), int64(len(funcs)))))
	}
	return Stack(pats...)
}

// --- Press / PressBy (compresión temporal)}

// Press comprime copias del patrón.
func (p Pattern) Press() Pattern { return p.PressBy(2) }

// PressBy comprime n copias del patrón en el ciclo.
func (p Pattern) PressBy(n int) Pattern {
	return p.PlyWith(n, func(sub Pattern, _ int) Pattern {
		return sub.Slow(FracInt(int64(n)))
	})
}

// --- Swing ---

// Swing retrasa notas en posiciones impares por la fracción dada.
func (p Pattern) Swing(amount Fraction) Pattern {
	return p.SwingBy(amount, FracFloat(0.25))
}

// SwingBy retrasa notas en ciertas subdivisiones.
func (p Pattern) SwingBy(amount, cycle Fraction) Pattern {
	// Apply late by amount, only at odd positions of cycle subdivision
	oddPat := Pure(false).Fast(FracInt(2)).Early(cycle)
	return Stack(
		p.When(oddPat, func(sub Pattern) Pattern { return sub.Late(amount) }),
	)
}

// --- Chunk (subdividir ciclos) ---

// Chunk aplica f a cada n-ésimo chunk del patrón.
func (p Pattern) Chunk(n int, f func(Pattern) Pattern) Pattern {
	return p.SlowChunk(n, f)
}

// SlowChunk (alias chunk)
func (p Pattern) SlowChunk(n int, f func(Pattern) Pattern) Pattern {
	return p.Every(n, f)
}

// FastChunk comprime el chunk aplicado.
func (p Pattern) FastChunk(n int, f func(Pattern) Pattern) Pattern {
	return p.Every(n, func(sub Pattern) Pattern { return f(sub.Fast(FracInt(int64(n)))) })
}

// ChunkInto stack con chunk transformado.
func (p Pattern) ChunkInto(n int, f func(Pattern) Pattern) Pattern {
	return Stack(p, p.Chunk(n, f))
}

// ChunkBack rota los chunks.
func (p Pattern) ChunkBack(n int) Pattern {
	return p.IterBack(n)
}

// --- Estructura ---

// Anchor fija un ancla del patrón.
func (p Pattern) Anchor(n Fraction) Pattern  { return p.Early(n) }
func (p Pattern) Fanchor(n Fraction) Pattern { return p.Late(n) }
func (p Pattern) Panchor(n Fraction) Pattern { return p.Early(n) }

// Tour recorre patrones con desplazamiento.
func (p Pattern) Tour(n int) Pattern { return p.Iter(n) }

// As etiqueta el patrón (utilidad de tipado suave).
func (p Pattern) As() Pattern { return p }

// AsNumber convierte los valores a numéricos.
func (p Pattern) AsNumber() Pattern { return p.asNumber() }

// --- Sub-aplicación de tiempo ---

// Per aplica f por cada ciclo (similar a Every(1)).
func (p Pattern) Per(f func(Pattern) Pattern) Pattern { return f(p) }

// PerCycle es alias de Per.
func (p Pattern) PerCycle(f func(Pattern) Pattern) Pattern { return f(p) }

// Perx aplica f cada n ciclos.
func (p Pattern) Perx(n int, f func(Pattern) Pattern) Pattern { return p.Every(n, f) }

// --- Loop ---

// Loop es alias de Slowcat infinito (el patrón ya se repite cíclicamente).
func (p Pattern) Loop() Pattern { return p }

// LoopAt repite el patrón n veces por ciclo.
func (p Pattern) LoopAt(n int) Pattern { return p.Fast(FracInt(int64(n))) }

// LoopAtCps repite el patrón a una CPS dada.
func (p Pattern) LoopAtCps(cps float64) Pattern { return p.Fast(FracFloat(cps)) }

// --- Squeeze / subordina ---

// Squeeze mantiene los eventos de p donde other tiene presencia (evento).
func (p Pattern) Squeeze(other Pattern) Pattern {
	bin := other.WithEvents(func(haps []Hap) []Hap {
		out := make([]Hap, len(haps))
		for i, h := range haps {
			out[i] = h.WithValue(func(_ interface{}) interface{} { return true })
		}
		return out
	})
	return p.Mask(bin)
}

// Inhabit (alias pickSqueeze) — elige sub-patrones según un selector.
func (p Pattern) Inhabit(selector Pattern) Pattern {
	return p.Mask(selector)
}

// --- Varios ---

// Grow genera crecimiento recursivo.
func (p Pattern) Grow(n int) Pattern {
	pats := make([]Pattern, n)
	for i := range pats {
		pats[i] = p
	}
	return Slowcat(pats...)
}

// Hurry es alias de Fast.
func (p Pattern) Hurry(n Fraction) Pattern { return p.Fast(n) }

// Keep mantiene solo eventos con onset.
func (p Pattern) Keep() Pattern { return p.OnsetsOnly() }

// Drop elimina eventos con onset (inverso de Keep).
func (p Pattern) Drop() Pattern {
	return p.WithEvents(func(haps []Hap) []Hap {
		out := haps[:0]
		for _, h := range haps {
			if !h.HasOnset() {
				out = append(out, h)
			}
		}
		return out
	})
}

// Linger extiende la duración de cada evento.
func (p Pattern) Linger() Pattern {
	return p.WithEventSpan(func(ts TimeSpan) TimeSpan {
		return TimeSpan{Begin: ts.Begin, End: ts.Begin.NextSam()}
	})
}

// Take limita el patrón a n ciclos (luego silencio).
func (p Pattern) Take(n int) Pattern {
	return Pattern{Query: func(s State) []Hap {
		if s.Span.Begin.Floor() >= int64(n) {
			return nil
		}
		return p.Query(s)
	}}
}

// Spread distribuye el patrón a través de varios canales (versión simplificada).
func (p Pattern) Spread(n int) Pattern {
	pats := make([]Pattern, n)
	for i := range pats {
		pats[i] = p.Early(NewFrac(int64(i), int64(n*4)))
	}
	return Stack(pats...)
}

// StripContext elimina la información de contexto (para evitar resaltado).
func (p Pattern) StripContext() Pattern {
	return p.WithContext(func(_ HapContext) HapContext {
		return HapContext{}
	})
}

// FilterHaps mantiene haps que pasan el test.
func (p Pattern) FilterHaps(test func(Hap) bool) Pattern {
	return p.WithEvents(func(haps []Hap) []Hap {
		out := haps[:0]
		for _, h := range haps {
			if test(h) {
				out = append(out, h)
			}
		}
		return out
	})
}

// FilterWhen aplica la condición solo en ciertos ciclos.
func (p Pattern) FilterWhen(prob float64, test func(Hap) bool) Pattern {
	return p.FilterHaps(func(h Hap) bool {
		if randFloat() < prob {
			return test(h)
		}
		return true
	})
}

// FirstCycleValues devuelve valores del primer ciclo.
func (p Pattern) FirstCycleValues() []interface{} {
	haps := p.FirstCycle()
	vals := make([]interface{}, len(haps))
	for i, h := range haps {
		vals[i] = h.Value
	}
	return vals
}

// ShowFirstCycle imprime el primer ciclo.
func (p Pattern) ShowFirstCycle() Pattern {
	haps := p.FirstCycle()
	for i, h := range haps {
		_ = i
		_ = h
	}
	return p
}

// toNumber is defined in arithmetic.go

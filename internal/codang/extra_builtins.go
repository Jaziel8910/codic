package codang

import (
	"fmt"
	"github.com/Jaziel8910/codic/internal/pattern"
)

// paramSpec describe un método de patrón que simplemente fija una clave de parámetro.
type paramSpec struct {
	key string
	str bool // true si el valor es string, false si es numérico
}

// paramMethodSpecs mapea nombres de método (usados en Codang) a su clave de
// parámetro en el motor. El switch de callPatternMethod consulta este mapa en
// su caso `default`, de modo que añadir un parámetro nuevo es solo agregar
// una línea aquí.
var paramMethodSpecs = map[string]paramSpec{
	// mezcla / panorama
	"postgain": {"postgain", false},
	"panwidth": {"panwidth", false},
	"octaves":  {"octave", false},
	// afinación
	"detune":     {"detune", false},
	"freq":       {"freq", false},
	"ftranspose": {"ftranspose", false},
	// envolvente
	"decay": {"decay", false},
	// filtros
	"lpq":   {"lpq", false},
	"hpf":   {"hpf", false},
	"hpq":   {"hpq", false},
	"lpf":   {"cutoff", false},
	"bpf":   {"bpf", false},
	"bpq":   {"bpq", false},
	"bandf": {"bpf", false},
	"bandq": {"bpq", false},
	"djf":   {"djf", false},
	// distorsión / bitcrush
	"distort":     {"distort", false},
	"dist":        {"distort", false},
	"drive":       {"distort", false},
	"distorttype": {"distorttype", true},
	// vibrato
	"vib":    {"vib", false},
	"vibmod": {"vibmod", false},
	// vowel / formante
	"vowel":  {"vowel", true},
	"ribbon": {"ribbon", false},
	// tremolo
	"tremolo":      {"tremolo", false},
	"tremolorate":  {"tremolorate", false},
	"tremolodepth": {"tremolodepth", false},
	// phaser
	"phaser":       {"phaser", false},
	"phasercenter": {"phasercenter", false},
	"phaserdepth":  {"phaserdepth", false},
	"phasersweep":  {"phasersweep", false},
	// chorus / comb
	"chorus": {"chorus", false},
	"comb":   {"comb", false},
	// delay
	"delaytime":     {"delaytime", false},
	"delayfeedback": {"delayfeedback", false},
	"delaysync":     {"delaysync", false},
	"echo":          {"delay", false},
	// room / reverb
	"roomsize": {"roomsize", false},
	"roomdim":  {"roomdim", false},
	"roomfade": {"roomfade", false},
	"roomlp":   {"roomlp", false},
	"size":     {"roomsize", false},
	"room":     {"roomsize", false},
	// sample
	"source":       {"s", true},
	"src":          {"s", true},
	"sintetizador": {"s", true},
	"accelerate":   {"accelerate", false},
	"rate":         {"speed", false},
	"loopbegin":    {"loopbegin", false},
	"loopend":      {"loopend", false},
	"cut":          {"cut", true},
	// bank / aliases
	"bank":       {"bank", true},
	"aliasbank":  {"aliasbank", true},
	"soundalias": {"soundalias", true},
	// enrutamiento
	"orbit":    {"orbit", false},
	"channel":  {"channel", false},
	"channels": {"channel", false},
	// etiquetas
	"label":  {"label", true},
	"tag":    {"tag", true},
	"color":  {"color", true},
	"colour": {"color", true},
	// musical
	"mode": {"mode", true},
	"tune": {"tune", false},
	// MIDI
	"midinote":   {"midinote", false},
	"midichan":   {"midichan", false},
	"midicmd":    {"midicmd", true},
	"ccn":        {"ccn", false},
	"ccv":        {"ccv", false},
	"progNum":    {"progNum", false},
	"pitchwheel": {"pitchwheel", false},
	"nrpnn":      {"nrpnn", false},
	"nrpv":       {"nrpv", false},
	"midi":       {"midi", true},
	// FM
	"fm":        {"fm", false},
	"fmi":       {"fmi", false},
	"fmh":       {"fmh", false},
	"fmw":       {"fmwave", false},
	"fmattack":  {"fmattack", false},
	"fmdecay":   {"fmdecay", false},
	"fmsustain": {"fmsustain", false},
	"fmrelease": {"fmrelease", false},
	// wavetable
	"wt":          {"wt", false},
	"wtattack":    {"wtattack", false},
	"wtdecay":     {"wtdecay", false},
	"wtdepth":     {"wtdepth", false},
	"wtrate":      {"wtrate", false},
	"wtrelease":   {"wtrelease", false},
	"wtshape":     {"wtshape", false},
	"wtsustain":   {"wtsustain", false},
	"wtphaserand": {"wtphaserand", false},
	// envelopes de filtro
	"lpattack":  {"lpattack", false},
	"lpdecay":   {"lpdecay", false},
	"lpdepth":   {"lpdepth", false},
	"lprate":    {"lprate", false},
	"lpsustain": {"lpsustain", false},
	"lpsync":    {"lpsync", false},
	"hpattack":  {"hpattack", false},
	"hpdecay":   {"hpdecay", false},
	"hpdepth":   {"hpdepth", false},
	"hprate":    {"hprate", false},
	"hpsustain": {"hpsustain", false},
	"hpsync":    {"hpsync", false},
	"bpattack":  {"bpattack", false},
	"bpdecay":   {"bpdecay", false},
	"bpdepth":   {"bpdepth", false},
	"bprate":    {"bprate", false},
	"bpsustain": {"bpsustain", false},
	"bpsync":    {"bpsync", false},
	// efectos extra
	"squiz": {"squiz", false},
	"real":  {"real", false},
}

// funcArgMethods son métodos que aceptan una función (por nombre) como argumento.
var funcArgMethods = map[string]bool{
	"sometimes": true, "often": true, "rarely": true, "almostAlways": true,
	"almostNever": true, "always": true, "never": true, "someCycles": true,
	"someCyclesBy": true, "when": true, "off": true, "layer": true,
	"superimpose": true, "within": true, "outside": true, "into": true,
	"pace": true, "offspray": true, "plyWith": true, "plyForEach": true,
	"press": true, "pressBy": true, "chunk": true, "slowChunk": true,
	"fastChunk": true, "chunkInto": true, "swing": true, "swingBy": true,
	"per": true, "perCycle": true, "perx": true, "jux": true,
}

// transformName resuelve un nombre de función (definida por el usuario o
// transformación integrada) a un closure sobre patrones.
func (ev *Evaluator) resolveTransform(name string) func(pattern.Pattern) pattern.Pattern {
	if fn, ok := ev.env.getFunc(name); ok {
		return func(sub pattern.Pattern) pattern.Pattern {
			res, _ := ev.callUserFuncVals(fn, []interface{}{sub})
			if rp, ok := res.(pattern.Pattern); ok {
				return rp
			}
			return sub
		}
	}
	// transformaciones integradas sin argumentos
	switch name {
	case "rev":
		return func(p pattern.Pattern) pattern.Pattern { return p.Rev() }
	case "palindrome":
		return func(p pattern.Pattern) pattern.Pattern { return p.Palindrome() }
	case "brak":
		return func(p pattern.Pattern) pattern.Pattern { return p.Brak() }
	case "degrade":
		return func(p pattern.Pattern) pattern.Pattern { return p.Degrade() }
	case "scramble", "shuffle":
		return func(p pattern.Pattern) pattern.Pattern { return p.Scramble() }
	case "keep":
		return func(p pattern.Pattern) pattern.Pattern { return p.Keep() }
	case "drop":
		return func(p pattern.Pattern) pattern.Pattern { return p.Drop() }
	case "linger":
		return func(p pattern.Pattern) pattern.Pattern { return p.Linger() }
	}
	return nil
}

// callUserFuncVals llama una función de usuario pasando valores ya evaluados
// (no nodos AST), útil para pasar patrones a métodos de orden superior.
func (ev *Evaluator) callUserFuncVals(fn *FuncDef, vals []interface{}) (interface{}, error) {
	scope := ev.env.child()
	for i, param := range fn.Params {
		if i < len(vals) {
			scope.setVar(param, vals[i])
		} else {
			scope.setVar(param, nil)
		}
	}
	oldEnv := ev.env
	ev.env = scope
	defer func() { ev.env = oldEnv }()
	var last interface{}
	for _, stmt := range fn.Body {
		val, err := ev.evalNode(stmt)
		if err != nil {
			return nil, err
		}
		if rv, ok := val.(*returnValue); ok {
			return rv.value, nil
		}
		last = val
	}
	return last, nil
}

// funcArg resuelve el argumento i como una función de transformación (por nombre).
func (ev *Evaluator) funcArg(args []interface{}, i int) func(pattern.Pattern) pattern.Pattern {
	if i < len(args) {
		if nm, ok := args[i].(string); ok {
			return ev.resolveTransform(nm)
		}
	}
	return nil
}

// funcArgs resuelve todos los argumentos como transformaciones.
func (ev *Evaluator) funcArgs(args []interface{}) []func(pattern.Pattern) pattern.Pattern {
	var out []func(pattern.Pattern) pattern.Pattern
	for _, a := range args {
		if nm, ok := a.(string); ok {
			if t := ev.resolveTransform(nm); t != nil {
				out = append(out, t)
			}
		}
	}
	return out
}

// argsToStrings convierte los argumentos a []string.
func argsToStrings(args []interface{}) []string {
	out := make([]string, 0, len(args))
	for _, a := range args {
		out = append(out, toStringVal(a))
	}
	return out
}

// argsToFloats convierte los argumentos a []float64.
func argsToFloats(args []interface{}) []float64 {
	out := make([]float64, 0, len(args))
	for _, a := range args {
		out = append(out, toNumber(a))
	}
	return out
}

// argsToVoicings convierte argumentos anidados a [][]float64.
func argsToVoicings(args []interface{}) [][]float64 {
	var out [][]float64
	for _, a := range args {
		if arr, ok := a.([]interface{}); ok {
			row := make([]float64, 0, len(arr))
			for _, e := range arr {
				row = append(row, toNumber(e))
			}
			out = append(out, row)
		}
	}
	return out
}

// --- Builtins independientes (funciones de nivel superior en Codang) ---

func builtinChoose(ev *Evaluator, args []Node) (interface{}, error) {
	vals, err := evalArgs(ev, args)
	if err != nil {
		return nil, err
	}
	return pattern.Choose(vals...), nil
}

func builtinChooseCycles(ev *Evaluator, args []Node) (interface{}, error) {
	vals, err := evalArgs(ev, args)
	if err != nil {
		return nil, err
	}
	pats := make([]interface{}, len(vals))
	copy(pats, vals)
	return pattern.ChooseCycles(pats...), nil
}

func builtinWChoose(ev *Evaluator, args []Node) (interface{}, error) {
	vals, err := evalArgs(ev, args)
	if err != nil {
		return nil, err
	}
	var pairs [][2]interface{}
	for _, a := range vals {
		if arr, ok := a.([]interface{}); ok && len(arr) >= 2 {
			pairs = append(pairs, [2]interface{}{arr[0], arr[1]})
		}
	}
	return pattern.WChoose(pairs...), nil
}

func builtinWChooseCycles(ev *Evaluator, args []Node) (interface{}, error) {
	vals, err := evalArgs(ev, args)
	if err != nil {
		return nil, err
	}
	var pairs [][2]interface{}
	for _, a := range vals {
		if arr, ok := a.([]interface{}); ok && len(arr) >= 2 {
			pairs = append(pairs, [2]interface{}{arr[0], arr[1]})
		}
	}
	return pattern.WChooseCycles(pairs...), nil
}

func builtinRun(ev *Evaluator, args []Node) (interface{}, error) {
	vals, err := evalArgs(ev, args)
	if err != nil {
		return nil, err
	}
	n := 8
	if len(vals) > 0 {
		n = int(toNumber(vals[0]))
	}
	return pattern.Run(n), nil
}

func builtinIrand(ev *Evaluator, args []Node) (interface{}, error) {
	vals, err := evalArgs(ev, args)
	if err != nil {
		return nil, err
	}
	n := 16
	if len(vals) > 0 {
		n = int(toNumber(vals[0]))
	}
	return pattern.Irand(n), nil
}

func builtinRand2(ev *Evaluator, args []Node) (interface{}, error) {
	vals, err := evalArgs(ev, args)
	if err != nil {
		return nil, err
	}
	g := 1.0
	if len(vals) > 0 {
		g = toNumber(vals[0])
	}
	return pattern.Rand2(g), nil
}

func builtinRandL(ev *Evaluator, args []Node) (interface{}, error) {
	vals, err := evalArgs(ev, args)
	if err != nil {
		return nil, err
	}
	g := 1.0
	if len(vals) > 0 {
		g = toNumber(vals[0])
	}
	return pattern.RandL(g), nil
}

func builtinRangex(ev *Evaluator, args []Node) (interface{}, error) {
	vals, err := evalArgs(ev, args)
	if err != nil {
		return nil, err
	}
	lo, hi := 0.0, 1.0
	if len(vals) > 0 {
		lo = toNumber(vals[0])
	}
	if len(vals) > 1 {
		hi = toNumber(vals[1])
	}
	return pattern.Rangex(lo, hi), nil
}

func builtinEdoScale(ev *Evaluator, args []Node) (interface{}, error) {
	vals, err := evalArgs(ev, args)
	if err != nil {
		return nil, err
	}
	edo := 12
	if len(vals) > 0 {
		edo = int(toNumber(vals[0]))
	}
	rest := vals[1:]
	pats := make([]interface{}, len(rest))
	copy(pats, rest)
	return pattern.EdoScale(edo, pats...), nil
}

func builtinChord(ev *Evaluator, args []Node) (interface{}, error) {
	vals, err := evalArgs(ev, args)
	if err != nil {
		return nil, err
	}
	root, name := "c", "major"
	if len(vals) > 0 {
		root = toStringVal(vals[0])
	}
	if len(vals) > 1 {
		name = toStringVal(vals[1])
	}
	return pattern.Chord(root, name), nil
}

func builtinScaleFn(ev *Evaluator, args []Node) (interface{}, error) {
	vals, err := evalArgs(ev, args)
	if err != nil {
		return nil, err
	}
	name := "major"
	if len(vals) > 0 {
		name = toStringVal(vals[0])
	}
	rest := []interface{}{}
	if len(vals) > 1 {
		rest = vals[1:]
	}
	return pattern.Pure(0).Scale(name, rest...), nil
}

func builtinPalindrome(ev *Evaluator, args []Node) (interface{}, error) {
	vals, err := evalArgs(ev, args)
	if err != nil {
		return nil, err
	}
	if len(vals) == 0 {
		return pattern.Silence(), nil
	}
	return toPattern(vals[0]).Palindrome(), nil
}

func builtinBrak(ev *Evaluator, args []Node) (interface{}, error) {
	vals, err := evalArgs(ev, args)
	if err != nil {
		return nil, err
	}
	if len(vals) == 0 {
		return pattern.Silence(), nil
	}
	return toPattern(vals[0]).Brak(), nil
}

func builtinScramble(ev *Evaluator, args []Node) (interface{}, error) {
	vals, err := evalArgs(ev, args)
	if err != nil {
		return nil, err
	}
	if len(vals) == 0 {
		return pattern.Silence(), nil
	}
	return toPattern(vals[0]).Scramble(), nil
}

func builtinDegrade(ev *Evaluator, args []Node) (interface{}, error) {
	vals, err := evalArgs(ev, args)
	if err != nil {
		return nil, err
	}
	if len(vals) == 0 {
		return pattern.Silence(), nil
	}
	return toPattern(vals[0]).Degrade(), nil
}

func builtinDegradeBy(ev *Evaluator, args []Node) (interface{}, error) {
	vals, err := evalArgs(ev, args)
	if err != nil {
		return nil, err
	}
	p := toPattern(vals[0])
	prob := 0.5
	if len(vals) > 1 {
		prob = toNumber(vals[1])
	}
	return p.DegradeBy(prob), nil
}

func builtinIrad(ev *Evaluator, args []Node) (interface{}, error) {
	return pattern.Pure(0), nil
}

// registerExtraBuiltins añade las funciones independientes nuevas.
func registerExtraBuiltins(b map[string]builtinFunc) {
	b["choose"] = builtinChoose
	b["chooseCycles"] = builtinChooseCycles
	b["randcat"] = builtinChooseCycles
	b["wchoose"] = builtinWChoose
	b["wrandcat"] = builtinWChooseCycles
	b["run"] = builtinRun
	b["irand"] = builtinIrand
	b["rand2"] = builtinRand2
	b["randL"] = builtinRandL
	b["rangex"] = builtinRangex
	b["edoScale"] = builtinEdoScale
	b["chord"] = builtinChord
	b["scale"] = builtinScaleFn
	b["palindrome"] = builtinPalindrome
	b["brak"] = builtinBrak
	b["scramble"] = builtinScramble
	b["shuffle"] = builtinScramble
	b["degrade"] = builtinDegrade
	b["degradeBy"] = builtinDegradeBy
	b["zip"] = builtinZip
	b["xfade"] = builtinXfade
	b["crossfade"] = builtinXfade
	b["stepcat"] = builtinStepcat
	b["step"] = builtinStep
	b["stepwise"] = builtinStepwise
	b["seqPLoop"] = builtinStepwise
	b["spiral"] = builtinSpiral
	b["euclid"] = builtinEuclid
	b["euclidish"] = builtinEuclidish
	b["euclidExtended"] = builtinEuclidExtended
	b["seed"] = builtinSeed
	b["reseed"] = builtinReseed
	b["deserialize"] = builtinDeserialize
	b["deserializeJSON"] = builtinDeserialize
	b["analyze"] = builtinAnalyze
}

func builtinZip(ev *Evaluator, args []Node) (interface{}, error) {
	vals, err := evalArgs(ev, args)
	if err != nil {
		return nil, err
	}
	if len(vals) < 2 {
		return pattern.Silence(), nil
	}
	return pattern.Zip(toPattern(vals[0]), toPattern(vals[1])), nil
}

func builtinXfade(ev *Evaluator, args []Node) (interface{}, error) {
	vals, err := evalArgs(ev, args)
	if err != nil {
		return nil, err
	}
	if len(vals) < 3 {
		return pattern.Silence(), nil
	}
	return pattern.Xfade(toPattern(vals[0]), toPattern(vals[1]), toNumber(vals[2])), nil
}

func builtinStepcat(ev *Evaluator, args []Node) (interface{}, error) {
	vals, err := evalArgs(ev, args)
	if err != nil {
		return nil, err
	}
	pats := make([]pattern.Pattern, 0, len(vals))
	for _, v := range vals {
		pats = append(pats, toPattern(v))
	}
	return pattern.Stepcat(pats...), nil
}

func builtinStep(ev *Evaluator, args []Node) (interface{}, error) {
	vals, err := evalArgs(ev, args)
	if err != nil {
		return nil, err
	}
	if len(vals) == 0 {
		return pattern.Step(nil), nil
	}
	return pattern.Step(vals[0]), nil
}

func builtinStepwise(ev *Evaluator, args []Node) (interface{}, error) {
	vals, err := evalArgs(ev, args)
	if err != nil {
		return nil, err
	}
	return pattern.Stepwise(vals...), nil
}

func builtinSpiral(ev *Evaluator, args []Node) (interface{}, error) {
	vals, err := evalArgs(ev, args)
	if err != nil {
		return nil, err
	}
	scale := "major"
	steps, rotary, root := 8, 1, 0
	if len(vals) > 0 {
		scale = toStringVal(vals[0])
	}
	if len(vals) > 1 {
		steps = int(toNumber(vals[1]))
	}
	if len(vals) > 2 {
		rotary = int(toNumber(vals[2]))
	}
	if len(vals) > 3 {
		root = int(toNumber(vals[3]))
	}
	return pattern.Spiral(scale, steps, rotary, root), nil
}

func builtinEuclid(ev *Evaluator, args []Node) (interface{}, error) {
	vals, err := evalArgs(ev, args)
	if err != nil {
		return nil, err
	}
	steps, pulses, rot := 8, 3, 0
	if len(vals) > 0 {
		steps = int(toNumber(vals[0]))
	}
	if len(vals) > 1 {
		pulses = int(toNumber(vals[1]))
	}
	if len(vals) > 2 {
		rot = int(toNumber(vals[2]))
	}
	return pattern.Euclid(steps, pulses, rot), nil
}

func builtinEuclidExtended(ev *Evaluator, args []Node) (interface{}, error) {
	vals, err := evalArgs(ev, args)
	if err != nil {
		return nil, err
	}
	steps, pulses, rot, epp := 8, 3, 0, 1
	if len(vals) > 0 {
		steps = int(toNumber(vals[0]))
	}
	if len(vals) > 1 {
		pulses = int(toNumber(vals[1]))
	}
	if len(vals) > 2 {
		rot = int(toNumber(vals[2]))
	}
	if len(vals) > 3 {
		epp = int(toNumber(vals[3]))
	}
	return pattern.EuclidExtended(steps, pulses, rot, epp), nil
}

func builtinSeed(ev *Evaluator, args []Node) (interface{}, error) {
	vals, err := evalArgs(ev, args)
	if err != nil {
		return nil, err
	}
	var s int64
	if len(vals) > 0 {
		s = int64(toNumber(vals[0]))
	}
	pattern.SetSeed(s)
	return nil, nil
}

func builtinReseed(ev *Evaluator, args []Node) (interface{}, error) {
	pattern.Reseed()
	return nil, nil
}

func builtinDeserialize(ev *Evaluator, args []Node) (interface{}, error) {
	vals, err := evalArgs(ev, args)
	if err != nil {
		return nil, err
	}
	if len(vals) > 0 {
		j := toStringVal(vals[0])
		p, err := pattern.DeserializeJSON(j)
		if err != nil {
			return nil, fmt.Errorf("deserialize: %v", err)
		}
		return p, nil
	}
	return pattern.Silence(), nil
}

func builtinAnalyze(ev *Evaluator, args []Node) (interface{}, error) {
	vals, err := evalArgs(ev, args)
	if err != nil {
		return nil, err
	}
	if len(vals) == 0 {
		return "no pattern", nil
	}
	p := toPattern(vals[0])
	m := p.Analyze()
	return fmt.Sprintf("events=%.0f notes=%.0f range=%.0f poly=%.0f",
		m["events"], m["notes"], m["note_range"], m["polyphony"]), nil
}

func builtinEuclidish(ev *Evaluator, args []Node) (interface{}, error) {
	vals, err := evalArgs(ev, args)
	if err != nil {
		return nil, err
	}
	if len(vals) < 2 {
		return pattern.Silence(), nil
	}
	groups := make([]int, 0)
	if arr, ok := vals[0].([]interface{}); ok {
		for _, n := range arr {
			groups = append(groups, int(toNumber(n)))
		}
	}
	return pattern.Euclidish(groups, int(toNumber(vals[1]))), nil
}

package codang

import (
	"github.com/Jaziel8910/codic/internal/pattern"
)

// builtinFunc is a native function callable from Codang.
type builtinFunc func(ev *Evaluator, args []Node) (interface{}, error)

// --- Eval helpers for builtin arg parsing ---

func evalArgs(ev *Evaluator, args []Node) ([]interface{}, error) {
	vals := make([]interface{}, 0, len(args))
	for _, arg := range args {
		v, err := ev.evalNode(arg)
		if err != nil {
			return nil, err
		}
		vals = append(vals, v)
	}
	return vals, nil
}

// --- Pattern constructor builtins ---

func builtinNote(ev *Evaluator, args []Node) (interface{}, error) {
	vals, err := evalArgs(ev, args)
	if err != nil {
		return nil, err
	}
	return pattern.NoteFn(vals...), nil
}

func builtinN(ev *Evaluator, args []Node) (interface{}, error) {
	vals, err := evalArgs(ev, args)
	if err != nil {
		return nil, err
	}
	return pattern.NFn(vals...), nil
}

func builtinS(ev *Evaluator, args []Node) (interface{}, error) {
	vals, err := evalArgs(ev, args)
	if err != nil {
		return nil, err
	}
	return pattern.SFn(vals...), nil
}

func builtinGain(ev *Evaluator, args []Node) (interface{}, error) {
	vals, err := evalArgs(ev, args)
	if err != nil {
		return nil, err
	}
	return pattern.GainFn(vals...), nil
}

func builtinFreq(ev *Evaluator, args []Node) (interface{}, error) {
	vals, err := evalArgs(ev, args)
	if err != nil {
		return nil, err
	}
	return makeParamPattern("freq", vals), nil
}

func builtinPan(ev *Evaluator, args []Node) (interface{}, error) {
	vals, err := evalArgs(ev, args)
	if err != nil {
		return nil, err
	}
	return pattern.PanFn(vals...), nil
}

func builtinCutoff(ev *Evaluator, args []Node) (interface{}, error) {
	vals, err := evalArgs(ev, args)
	if err != nil {
		return nil, err
	}
	return pattern.CutoffFn(vals...), nil
}

func builtinReverb(ev *Evaluator, args []Node) (interface{}, error) {
	vals, err := evalArgs(ev, args)
	if err != nil {
		return nil, err
	}
	return pattern.ReverbFn(vals...), nil
}

func builtinDelay(ev *Evaluator, args []Node) (interface{}, error) {
	vals, err := evalArgs(ev, args)
	if err != nil {
		return nil, err
	}
	return pattern.DelayFn(vals...), nil
}

// --- Composition builtins ---

func builtinStack(ev *Evaluator, args []Node) (interface{}, error) {
	vals, err := evalArgs(ev, args)
	if err != nil {
		return nil, err
	}
	pats := make([]pattern.Pattern, 0, len(vals))
	for _, v := range vals {
		pats = append(pats, toPattern(v))
	}
	return pattern.Stack(pats...), nil
}

func builtinCat(ev *Evaluator, args []Node) (interface{}, error) {
	vals, err := evalArgs(ev, args)
	if err != nil {
		return nil, err
	}
	pats := make([]pattern.Pattern, 0, len(vals))
	for _, v := range vals {
		pats = append(pats, toPattern(v))
	}
	return pattern.Cat(pats...), nil
}

func builtinSlowcat(ev *Evaluator, args []Node) (interface{}, error) {
	vals, err := evalArgs(ev, args)
	if err != nil {
		return nil, err
	}
	pats := make([]pattern.Pattern, 0, len(vals))
	for _, v := range vals {
		pats = append(pats, toPattern(v))
	}
	return pattern.Slowcat(pats...), nil
}

func builtinFastcat(ev *Evaluator, args []Node) (interface{}, error) {
	vals, err := evalArgs(ev, args)
	if err != nil {
		return nil, err
	}
	pats := make([]pattern.Pattern, 0, len(vals))
	for _, v := range vals {
		pats = append(pats, toPattern(v))
	}
	return pattern.Fastcat(pats...), nil
}

func builtinSequence(ev *Evaluator, args []Node) (interface{}, error) {
	vals, err := evalArgs(ev, args)
	if err != nil {
		return nil, err
	}
	return pattern.Sequence(vals...), nil
}

func builtinPolymeter(ev *Evaluator, args []Node) (interface{}, error) {
	vals, err := evalArgs(ev, args)
	if err != nil {
		return nil, err
	}
	if len(vals) == 0 {
		return pattern.Silence(), nil
	}
	steps := 0
	startIdx := 0
	if n, ok := vals[0].(float64); ok {
		steps = int(n)
		startIdx = 1
	}
	if startIdx >= len(vals) {
		return pattern.Silence(), nil
	}
	return pattern.Polymeter(steps, vals[startIdx:]...), nil
}

func builtinPolyrhythm(ev *Evaluator, args []Node) (interface{}, error) {
	vals, err := evalArgs(ev, args)
	if err != nil {
		return nil, err
	}
	return pattern.Polyrhythm(vals...), nil
}

func builtinTimecat(ev *Evaluator, args []Node) (interface{}, error) {
	vals, err := evalArgs(ev, args)
	if err != nil {
		return nil, err
	}
	var pairs [][2]interface{}
	for i := 0; i+1 < len(vals); i += 2 {
		pairs = append(pairs, [2]interface{}{vals[i], vals[i+1]})
	}
	return pattern.TimeCat(pairs...), nil
}

// --- Combinator standalone builtins ---

func builtinFast(ev *Evaluator, args []Node) (interface{}, error) {
	vals, err := evalArgs(ev, args)
	if err != nil {
		return nil, err
	}
	if len(vals) < 2 {
		return nil, nil
	}
	frac := toFraction(vals[0])
	pat := toPattern(vals[1])
	return pat.Fast(frac), nil
}

func builtinSlow(ev *Evaluator, args []Node) (interface{}, error) {
	vals, err := evalArgs(ev, args)
	if err != nil {
		return nil, err
	}
	if len(vals) < 2 {
		return nil, nil
	}
	frac := toFraction(vals[0])
	pat := toPattern(vals[1])
	return pat.Slow(frac), nil
}

func builtinRev(ev *Evaluator, args []Node) (interface{}, error) {
	vals, err := evalArgs(ev, args)
	if err != nil {
		return nil, err
	}
	if len(vals) < 1 {
		return pattern.Silence(), nil
	}
	return toPattern(vals[0]).Rev(), nil
}

func builtinEarly(ev *Evaluator, args []Node) (interface{}, error) {
	vals, err := evalArgs(ev, args)
	if err != nil {
		return nil, err
	}
	if len(vals) < 2 {
		return nil, nil
	}
	frac := toFraction(vals[0])
	pat := toPattern(vals[1])
	return pat.Early(frac), nil
}

func builtinLate(ev *Evaluator, args []Node) (interface{}, error) {
	vals, err := evalArgs(ev, args)
	if err != nil {
		return nil, err
	}
	if len(vals) < 2 {
		return nil, nil
	}
	frac := toFraction(vals[0])
	pat := toPattern(vals[1])
	return pat.Late(frac), nil
}

// bpm(120) → sets the tempo in beats per minute
func builtinBPM(ev *Evaluator, args []Node) (interface{}, error) {
	vals, err := evalArgs(ev, args)
	if err != nil {
		return nil, err
	}
	if len(vals) > 0 {
		bpm := toNumber(vals[0])
		cps := bpm / 60.0 / 4.0 // assuming 4 beats per cycle
		ev.env.setVar("__cps", cps)
		ev.proj.SetBPM(bpm)
	}
	return nil, nil
}

// Signal builtins: sine, saw, tri, square, etc.
func builtinSine(ev *Evaluator, args []Node) (interface{}, error)    { return pattern.Sine, nil }
func builtinCosine(ev *Evaluator, args []Node) (interface{}, error)  { return pattern.Cosine, nil }
func builtinSaw(ev *Evaluator, args []Node) (interface{}, error)     { return pattern.Saw, nil }
func builtinIsaw(ev *Evaluator, args []Node) (interface{}, error)    { return pattern.Isaw, nil }
func builtinTri(ev *Evaluator, args []Node) (interface{}, error)     { return pattern.Tri, nil }
func builtinSquare(ev *Evaluator, args []Node) (interface{}, error)  { return pattern.Square, nil }
func builtinRand(ev *Evaluator, args []Node) (interface{}, error)    { return pattern.Rand, nil }
func builtinSine2(ev *Evaluator, args []Node) (interface{}, error)   { return pattern.Sine2, nil }
func builtinSaw2(ev *Evaluator, args []Node) (interface{}, error)    { return pattern.Saw2, nil }
func builtinTri2(ev *Evaluator, args []Node) (interface{}, error)    { return pattern.Tri2, nil }
func builtinSilence(ev *Evaluator, args []Node) (interface{}, error) { return pattern.Silence(), nil }

// makeParamPattern wraps values in a control map with the given key.
func makeParamPattern(key string, vals []interface{}) pattern.Pattern {
	var pats []interface{}
	for _, v := range vals {
		pats = append(pats, v)
	}
	pat := pattern.Sequence(pats...)
	return pat.WithValue(func(val interface{}) interface{} {
		return pattern.ControlMap{key: val}
	})
}

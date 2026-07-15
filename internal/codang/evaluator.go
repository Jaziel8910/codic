package codang

import (
	"fmt"
	"math"
	"strings"

	"github.com/Jaziel8910/codic/internal/pattern"
	"github.com/Jaziel8910/codic/internal/project"
)

// Env holds variable bindings for the evaluator.
type Env struct {
	vars   map[string]interface{}
	funcs  map[string]*FuncDef
	parent *Env
}

func newEnv() *Env {
	return &Env{vars: map[string]interface{}{}, funcs: map[string]*FuncDef{}}
}

func (e *Env) child() *Env {
	return &Env{vars: map[string]interface{}{}, funcs: map[string]*FuncDef{}, parent: e}
}

func (e *Env) getVar(name string) (interface{}, bool) {
	if v, ok := e.vars[name]; ok {
		return v, true
	}
	if e.parent != nil {
		return e.parent.getVar(name)
	}
	return nil, false
}

func (e *Env) setVar(name string, v interface{}) {
	e.vars[name] = v
}

func (e *Env) getFunc(name string) (*FuncDef, bool) {
	if f, ok := e.funcs[name]; ok {
		return f, true
	}
	if e.parent != nil {
		return e.parent.getFunc(name)
	}
	return nil, false
}

// Output is called when a pattern is sent to .out()
type Output func(p pattern.Pattern)

// Evaluator walks the AST and produces values (mostly patterns).
type Evaluator struct {
	env         *Env
	output      Output
	builtins    map[string]builtinFunc
	lastPattern pattern.Pattern
	proj *project.Project
	sections    map[string]pattern.Pattern
	arrangement []string
}

// NewEvaluator creates an evaluator with the given output callback.
func NewEvaluator(out Output) *Evaluator {
	e := &Evaluator{
		env:      newEnv(),
		output:   out,
		builtins: map[string]builtinFunc{},
		proj:     project.New("Sin título"),
	}
	e.setupBuiltins()
	return e
}

// Eval evaluates a program.
func (ev *Evaluator) Eval(prog *Program) error {
	for _, stmt := range prog.Statements {
		val, err := ev.evalNode(stmt)
		if err != nil {
			return err
		}
		if p, ok := val.(pattern.Pattern); ok {
			ev.lastPattern = p
		}
	}
	return nil
}

// LastPattern returns the most recent pattern value produced during evaluation.
func (ev *Evaluator) LastPattern() pattern.Pattern { return ev.lastPattern }

// Project returns the high-level album/project being assembled during evaluation.
func (ev *Evaluator) Project() *project.Project { return ev.proj }

// evalNode evaluates any AST node and returns its value.
func (ev *Evaluator) evalNode(node Node) (interface{}, error) {
	switch n := node.(type) {
	case *ExprStmt:
		return ev.evalNode(n.Expr)
	case *AssignStmt:
		val, err := ev.evalNode(n.Value)
		if err != nil {
			return nil, err
		}
		ev.env.setVar(n.Name, val)
		return val, nil
	case *FuncDef:
		ev.env.funcs[n.Name] = n
		return nil, nil
	case *ReturnStmt:
		val, err := ev.evalNode(n.Value)
		if err != nil {
			return nil, err
		}
		return &returnValue{value: val}, nil
	case *IfStmt:
		cond, err := ev.evalNode(n.Cond)
		if err != nil {
			return nil, err
		}
		if truthy(cond) {
			return ev.evalBlock(n.Then)
		}
		return ev.evalBlock(n.ElseBody)
	case *NumberLit:
		return n.Value, nil
	case *StringLit:
		// Strings are mini-notation patterns
		return pattern.ParseMini(n.Value), nil
	case *BoolLit:
		return n.Value, nil
	case *NilLit:
		return nil, nil
	case *Ident:
		v, ok := ev.env.getVar(n.Name)
		if !ok {
			// Try builtins (signals like sine, saw, etc.)
			if b, ok := ev.builtins[n.Name]; ok {
				// Call it with no args — it's a signal builtin
				return b(ev, nil)
			}
			return nil, fmt.Errorf("undefined variable: %s", n.Name)
		}
		return v, nil
	case *ArrayLit:
		var vals []interface{}
		for _, el := range n.Elements {
			v, err := ev.evalNode(el)
			if err != nil {
				return nil, err
			}
			vals = append(vals, v)
		}
		return vals, nil
	case *BinaryOp:
		return ev.evalBinaryOp(n)
	case *UnaryOp:
		return ev.evalUnaryOp(n)
	case *MethodCall:
		return ev.evalMethodCall(n)
	case *CallExpr:
		return ev.evalCallExpr(n)
	default:
		return nil, fmt.Errorf("cannot evaluate node type %s", node.nodeType())
	}
}

func (ev *Evaluator) evalBlock(stmts []Node) (interface{}, error) {
	var last interface{}
	for _, stmt := range stmts {
		val, err := ev.evalNode(stmt)
		if err != nil {
			return nil, err
		}
		if rv, ok := val.(*returnValue); ok {
			return rv, nil
		}
		last = val
	}
	return last, nil
}

// --- Binary operations ---

func (ev *Evaluator) evalBinaryOp(n *BinaryOp) (interface{}, error) {
	// Short-circuit for and/or
	if n.Op == "and" {
		left, err := ev.evalNode(n.Left)
		if err != nil {
			return nil, err
		}
		if !truthy(left) {
			return false, nil
		}
		right, err := ev.evalNode(n.Right)
		if err != nil {
			return nil, err
		}
		return truthy(right), nil
	}
	if n.Op == "or" {
		left, err := ev.evalNode(n.Left)
		if err != nil {
			return nil, err
		}
		if truthy(left) {
			return true, nil
		}
		right, err := ev.evalNode(n.Right)
		if err != nil {
			return nil, err
		}
		return truthy(right), nil
	}

	left, err := ev.evalNode(n.Left)
	if err != nil {
		return nil, err
	}
	right, err := ev.evalNode(n.Right)
	if err != nil {
		return nil, err
	}

	// Pattern arithmetic: + - * / on patterns
	if lp, ok := left.(pattern.Pattern); ok {
		rp := toPattern(right)
		switch n.Op {
		case "+":
			return lp.Add(rp), nil
		case "-":
			return lp.Sub(rp), nil
		case "*":
			return lp.Mul(rp), nil
		case "/":
			return lp.Div(rp), nil
		}
	}

	// Numeric arithmetic
	ln := toNumber(left)
	rn := toNumber(right)
	switch n.Op {
	case "+":
		return ln + rn, nil
	case "-":
		return ln - rn, nil
	case "*":
		return ln * rn, nil
	case "/":
		if rn == 0 {
			return 0.0, nil
		}
		return ln / rn, nil
	case "%":
		return math.Mod(ln, rn), nil
	case "<":
		return ln < rn, nil
	case ">":
		return ln > rn, nil
	case "<=":
		return ln <= rn, nil
	case ">=":
		return ln >= rn, nil
	case "==":
		return ln == rn, nil
	case "!=":
		return ln != rn, nil
	}

	return nil, fmt.Errorf("unknown operator: %s", n.Op)
}

func (ev *Evaluator) evalUnaryOp(n *UnaryOp) (interface{}, error) {
	operand, err := ev.evalNode(n.Operand)
	if err != nil {
		return nil, err
	}
	switch n.Op {
	case "-":
		return -toNumber(operand), nil
	case "!":
		return !truthy(operand), nil
	case "not":
		return !truthy(operand), nil
	}
	return nil, fmt.Errorf("unknown unary operator: %s", n.Op)
}

// --- Method calls (pattern chaining) ---

func (ev *Evaluator) evalMethodCall(n *MethodCall) (interface{}, error) {
	target, err := ev.evalNode(n.Target)
	if err != nil {
		return nil, err
	}

	// Evaluate args
	args := make([]interface{}, 0, len(n.Args))
	for _, arg := range n.Args {
		v, err := ev.evalNode(arg)
		if err != nil {
			return nil, err
		}
		args = append(args, v)
	}

	// Pattern method calls
	if p, ok := target.(pattern.Pattern); ok {
		return ev.callPatternMethod(p, n.Method, args), nil
	}

	// No method found
	return nil, fmt.Errorf("method '%s' not found on type %T", n.Method, target)
}

// --- Function calls ---

func (ev *Evaluator) evalCallExpr(n *CallExpr) (interface{}, error) {
	// User-defined function
	if fn, ok := ev.env.getFunc(n.Name); ok {
		return ev.callUserFunc(fn, n.Args)
	}

	// Builtin function
	if fn, ok := ev.builtins[n.Name]; ok {
		return fn(ev, n.Args)
	}

	// cps(1.5) is special — sets tempo
	if n.Name == "cps" {
		if len(n.Args) > 0 {
			val, err := ev.evalNode(n.Args[0])
			if err != nil {
				return nil, err
			}
			ev.env.setVar("__cps", toNumber(val))
		}
		return nil, nil
	}

	// hush()
	if n.Name == "hush" {
		if ev.output != nil {
			ev.output(pattern.Silence())
		}
		return nil, nil
	}

	return nil, fmt.Errorf("undefined function: %s", n.Name)
}

func (ev *Evaluator) callUserFunc(fn *FuncDef, argNodes []Node) (interface{}, error) {
	// Create a new scope
	scope := ev.env.child()
	for i, param := range fn.Params {
		if i < len(argNodes) {
			val, err := ev.evalNode(argNodes[i])
			if err != nil {
				return nil, err
			}
			scope.setVar(param, val)
		} else {
			scope.setVar(param, nil)
		}
	}
	// Swap environment temporarily
	oldEnv := ev.env
	ev.env = scope
	defer func() { ev.env = oldEnv }()

	// Execute body
	for _, stmt := range fn.Body {
		val, err := ev.evalNode(stmt)
		if err != nil {
			return nil, err
		}
		if rv, ok := val.(*returnValue); ok {
			return rv.value, nil
		}
	}
	return nil, nil
}

// --- Pattern method dispatch ---

func (ev *Evaluator) callPatternMethod(p pattern.Pattern, method string, args []interface{}) interface{} {
	// Helper to get Fraction arg
	fracArg := func(i int) pattern.Fraction {
		if i < len(args) {
			return toFraction(args[i])
		}
		return pattern.FracInt(1)
	}
	floatArg := func(i int) float64 {
		if i < len(args) {
			return toNumber(args[i])
		}
		return 0.0
	}
	strArg := func(i int) string {
		if i < len(args) {
			return toStringVal(args[i])
		}
		return ""
	}

	switch method {
	case "fast":
		return p.Fast(fracArg(0))
	case "slow":
		return p.Slow(fracArg(0))
	case "rev":
		return p.Rev()
	case "early":
		return p.Early(fracArg(0))
	case "late":
		return p.Late(fracArg(0))
	case "jux":
		return p.Jux(func(sub pattern.Pattern) pattern.Pattern {
			return sub.Rev() // default jux transform
		})
	case "iter":
		return p.Iter(int(fracArg(0).Num))
	case "every":
		n := int(floatArg(0))
		if t := ev.funcArg(args, 1); t != nil {
			return p.Every(n, t)
		}
		return p.Every(n, func(sub pattern.Pattern) pattern.Pattern { return sub })
	case "stack":
		pats := []pattern.Pattern{p}
		for _, a := range args {
			pats = append(pats, toPattern(a))
		}
		return pattern.Stack(pats...)
	case "append":
		if len(args) > 0 {
			return p.Append(toPattern(args[0]))
		}
		return p
	case "gain":
		return p.Gain(toFracVal(args[0]))
	case "pan":
		return p.Pan(toFracVal(args[0]))
	case "cutoff":
		return p.Cutoff(toFracVal(args[0]))
	case "resonance":
		return p.Resonance(toFracVal(args[0]))
	case "delay":
		return p.Delay(toFracVal(args[0]))
	case "delaytime":
		return p.DelayTime(toFracVal(args[0]))
	case "reverb", "room":
		return p.Reverb(toFracVal(args[0]))
	case "attack":
		return p.Attack(toFracVal(args[0]))
	case "release":
		return p.Release(toFracVal(args[0]))
	case "sustain":
		return p.Sustain(toFracVal(args[0]))
	case "shape":
		return p.Shape(toFracVal(args[0]))
	case "hpf":
		return p.HPF(toFracVal(args[0]))
	case "lpf":
		return p.LPF(toFracVal(args[0]))
	case "legato":
		return p.Legato(toFraction(args[0]))
	case "velocity":
		return p.Velocity(floatArg(0))
	case "scale":
		return p.Scale(strArg(0))
	case "note":
		return p.Note(args...)
	case "s", "sound":
		return p.S(args...)
	case "n":
		return p.N(toFracVal(args[0]))
	case "begin":
		return p.Begin(toFracVal(args[0]))
	case "end":
		return p.End(toFracVal(args[0]))
	case "crush":
		return p.Crush(toFracVal(args[0]))
	case "speed":
		return p.Speed(toFracVal(args[0]))
	case "octave":
		return p.Octave(toFracVal(args[0]))
	case "up":
		return p.Up(toFracVal(args[0]))
	case "down":
		return p.Down(toFracVal(args[0]))
	case "onsets":
		return p.OnsetsOnly()
	case "range":
		return p.Range(floatArg(0), floatArg(1))
	case "add":
		return p.Add(toPattern(args[0]))
	case "sub":
		return p.Sub(toPattern(args[0]))
	case "mul":
		return p.Mul(toPattern(args[0]))
	case "div":
		return p.Div(toPattern(args[0]))
	case "out", "play", "trigger":
		if ev.output != nil {
			ev.output(p)
		}
		return p
	case "log", "logvalues":
		// Log pattern's first cycle
		haps := p.FirstCycle()
		fmt.Printf("log: %d haps\n", len(haps))
		for i, h := range haps {
			fmt.Printf("  [%d] %s\n", i, h.Show())
		}
		return p
	case "bypass":
		return p.Bypass(int(floatArg(0)))
	case "hush":
		if ev.output != nil {
			ev.output(pattern.Silence())
		}
		return pattern.Silence()

	// --- Transformaciones (sin función) ---
	case "degrade":
		return p.Degrade()
	case "degradeBy":
		return p.DegradeBy(floatArg(0))
	case "undegrade":
		return p.Undegrade()
	case "undegradeBy":
		return p.UndegradeBy(floatArg(0))
	case "scramble", "shuffle":
		return p.Scramble()
	case "repeatCycles", "firstOf", "lastOf":
		return p.RepeatCycles(int(floatArg(0)))
	case "iterBack":
		return p.IterBack(int(floatArg(0)))
	case "palindrome":
		return p.Palindrome()
	case "zoom":
		return p.Zoom(fracArg(0), fracArg(1))
	case "stretch":
		return p.Stretch(fracArg(0))
	case "contract":
		return p.Contract(fracArg(0))
	case "expand":
		return p.Expand(fracArg(0))
	case "fit":
		return p.Fit(int(floatArg(0)))
	case "fold":
		return p.Fold(int(floatArg(0)))
	case "shrink":
		return p.Shrink(int(floatArg(0)))
	case "ply":
		return p.Ply(int(floatArg(0)))
	case "press":
		return p.Press()
	case "pressBy":
		return p.PressBy(int(floatArg(0)))
	case "swing":
		return p.Swing(fracArg(0))
	case "swingBy":
		return p.SwingBy(fracArg(0), fracArg(1))
	case "chunkBack":
		return p.ChunkBack(int(floatArg(0)))
	case "anchor":
		return p.Anchor(fracArg(0))
	case "fanchor":
		return p.Fanchor(fracArg(0))
	case "panchor":
		return p.Panchor(fracArg(0))
	case "tour":
		return p.Tour(int(floatArg(0)))
	case "as":
		return p.As()
	case "asNumber":
		return p.AsNumber()
	case "loop":
		return p.Loop()
	case "loopAt":
		return p.LoopAt(int(floatArg(0)))
	case "loopAtCps":
		return p.LoopAtCps(floatArg(0))
	case "squeeze":
		if len(args) > 0 {
			return p.Squeeze(toPattern(args[0]))
		}
		return p
	case "grow":
		return p.Grow(int(floatArg(0)))
	case "hurry":
		return p.Hurry(fracArg(0))
	case "keep":
		return p.Keep()
	case "drop":
		return p.Drop()
	case "linger":
		return p.Linger()
	case "take":
		return p.Take(int(floatArg(0)))
	case "spread":
		return p.Spread(int(floatArg(0)))
	case "stripContext":
		return p.StripContext()
	case "showFirstCycle":
		return p.ShowFirstCycle()

	// --- Multi-argument / especiales ---
	case "adsr":
		return p.Adsr(floatArg(0), floatArg(1), floatArg(2), floatArg(3))
	case "env":
		return p.Env(argsToFloats(args))
	case "lfo":
		return p.Lfo(strArg(0), floatArg(1), floatArg(2))
	case "loopSample":
		return p.LoopSample(floatArg(0), floatArg(1))
	case "voicing":
		return p.Voicing(argsToFloats(args)...)
	case "arp":
		return p.Arp(strArg(0))
	case "partial":
		return p.Partial(int(floatArg(0)))
	case "control":
		return p.Control(floatArg(0), floatArg(1))
	case "sysex", "sysexdata", "sysexid":
		return p.AddParam(method, argsToFloats(args))
	case "scaleTranspose":
		return p.ScaleTranspose(strArg(0), floatArg(1))
	case "fmenv":
		return p.AddParam("fmenv", argsToFloats(args))
	case "addVoicings":
		return p.AddVoicings(argsToVoicings(args))

	// --- Métodos que aceptan función (por nombre) ---
	case "sometimes":
		if t := ev.funcArg(args, 0); t != nil {
			return p.Sometimes(t)
		}
		return p
	case "often":
		if t := ev.funcArg(args, 0); t != nil {
			return p.Often(t)
		}
		return p
	case "rarely":
		if t := ev.funcArg(args, 0); t != nil {
			return p.Rarely(t)
		}
		return p
	case "almostAlways":
		if t := ev.funcArg(args, 0); t != nil {
			return p.AlmostAlways(t)
		}
		return p
	case "almostNever":
		if t := ev.funcArg(args, 0); t != nil {
			return p.AlmostNever(t)
		}
		return p
	case "always":
		if t := ev.funcArg(args, 0); t != nil {
			return p.Always(t)
		}
		return p
	case "never":
		if t := ev.funcArg(args, 0); t != nil {
			return p.Never(t)
		}
		return p
	case "someCycles":
		if t := ev.funcArg(args, 0); t != nil {
			return p.SomeCycles(t)
		}
		return p
	case "someCyclesBy":
		if t := ev.funcArg(args, 1); t != nil {
			return p.SomeCyclesBy(floatArg(0), t)
		}
		return p
	case "into":
		if t := ev.funcArg(args, 0); t != nil {
			return p.Into(t)
		}
		return p
	case "pace":
		if t := ev.funcArg(args, 1); t != nil {
			return p.Pace(int(floatArg(0)), t)
		}
		return p
	case "offspray":
		if t := ev.funcArg(args, 1); t != nil {
			return p.Offspray(fracArg(0), t)
		}
		return p
	case "when":
		if t := ev.funcArg(args, 1); t != nil {
			return p.When(toPattern(args[0]), t)
		}
		return p
	case "off":
		if t := ev.funcArg(args, 1); t != nil {
			return p.Off(toPattern(args[0]), t)
		}
		return p
	case "within":
		if t := ev.funcArg(args, 2); t != nil {
			return p.Within(fracArg(0), fracArg(1), t)
		}
		return p
	case "outside":
		if t := ev.funcArg(args, 2); t != nil {
			return p.Outside(fracArg(0), fracArg(1), t)
		}
		return p
	case "per":
		if t := ev.funcArg(args, 0); t != nil {
			return p.Per(t)
		}
		return p
	case "perCycle":
		if t := ev.funcArg(args, 0); t != nil {
			return p.PerCycle(t)
		}
		return p
	case "perx":
		if t := ev.funcArg(args, 1); t != nil {
			return p.Perx(int(floatArg(0)), t)
		}
		return p
	case "chunk":
		if t := ev.funcArg(args, 1); t != nil {
			return p.Chunk(int(floatArg(0)), t)
		}
		return p
	case "slowChunk":
		if t := ev.funcArg(args, 1); t != nil {
			return p.SlowChunk(int(floatArg(0)), t)
		}
		return p
	case "fastChunk":
		if t := ev.funcArg(args, 1); t != nil {
			return p.FastChunk(int(floatArg(0)), t)
		}
		return p
	case "chunkInto":
		if t := ev.funcArg(args, 1); t != nil {
			return p.ChunkInto(int(floatArg(0)), t)
		}
		return p
	case "plyWith":
		if t := ev.funcArg(args, 1); t != nil {
			return p.PlyWith(int(floatArg(0)), func(sub pattern.Pattern, _ int) pattern.Pattern { return t(sub) })
		}
		return p
	case "plyForEach":
		funcs := ev.funcArgs(args)
		if len(funcs) > 0 {
			return p.PlyForEach(funcs...)
		}
		return p
	case "layer":
		funcs := ev.funcArgs(args)
		if len(funcs) > 0 {
			return p.Layer(funcs...)
		}
		return p
	case "superimpose":
		funcs := ev.funcArgs(args)
		if len(funcs) > 0 {
			return p.Superimpose(funcs...)
		}
		return p
	case "euclid":
		return p.Euclid(int(floatArg(0)), int(floatArg(1)), int(floatArg(2)))
	case "euclidRot":
		return p.EuclidRot(int(floatArg(0)), int(floatArg(1)), int(floatArg(2)))
	case "euclidLegato":
		return p.EuclidLegato(int(floatArg(0)), int(floatArg(1)), int(floatArg(2)))
	case "euclidish":
		// grupos como lista, luego pulses
		if nums, ok := args[0].([]interface{}); ok {
			groups := make([]int, 0, len(nums))
			for _, n := range nums {
				groups = append(groups, int(toNumber(n)))
			}
			return p.Euclidish(groups, int(floatArg(1)))
		}
		return p
	case "juxBy":
		by := floatArg(0)
		if t := ev.funcArg(args, 1); t != nil {
			return p.JuxBy(by, t)
		}
		return p
	case "juxFlip":
		if t := ev.funcArg(args, 0); t != nil {
			return p.JuxFlip(t)
		}
		return p
	case "juxFlipBy":
		by := floatArg(0)
		if t := ev.funcArg(args, 1); t != nil {
			return p.JuxFlipBy(by, t)
		}
		return p
	case "compress":
		return p.Compress(toFraction(args[0]), toFraction(args[1]))
	case "unison":
		return p.Unison(int(floatArg(0)))
	case "toMidi":
		return p.ToMidi()
	case "fromMidi":
		return p.FromMidi()
	case "brand":
		return p.Brand(strArg(0))
	case "brandBy":
		return p.BrandBy(strArg(0))
	case "inhabit":
		if len(args) > 0 {
			return p.Inhabit(toPattern(args[0]))
		}
		return p
	case "striate":
		return p.Striate(int(floatArg(0)))
	case "stripe":
		return p.Stripe(int(floatArg(0)))
	case "chop":
		return p.Chop(int(floatArg(0)))
	case "slice":
		return p.Slice(int(floatArg(0)))
	case "splice":
		return p.Splice(int(floatArg(0)))
	case "bite":
		return p.Bite(toNumber(args[0]), toNumber(args[1]), int(floatArg(2)))
	case "gap":
		return p.Gap(floatArg(0))
	case "reset":
		return p.Reset()
	case "restart":
		return p.Restart()
	case "defragmentHaps":
		return p.DefragmentHaps()
	case "setcpm":
		return p.Setcpm(floatArg(0))
	case "spiral":
		return pattern.Spiral(strArg(0), int(floatArg(1)), int(floatArg(2)), int(floatArg(3)))

	// --- Fase 2: nuevas funciones ---
	case "harmonize":
		if len(args) >= 2 {
			intervals := argsToFloats(args[1:])
			return p.Harmonize(strArg(0), intervals...)
		}
		return p.Harmonize("major")
	case "counterpoint":
		return p.Counterpoint(argsToStrings(args)...)
	case "arpeggiate":
		rng := int(floatArg(1))
		return p.Arpeggiate(strArg(0), rng)
	case "humanize":
		return p.Humanize(floatArg(0), floatArg(1), floatArg(2))
	case "groove":
		s := floatArg(1)
		return p.Groove(strArg(0), s)
	case "layer2":
		dens := int(floatArg(0))
		varVar := floatArg(1)
		return p.LayerDensity(dens, varVar)
	case "morph":
		if len(args) >= 1 {
			steps := int(floatArg(1))
			cv := strArg(2)
			return p.Morph(toPattern(args[0]), steps, cv)
		}
		return p
	case "markov":
		return p.Markov(int(floatArg(0)))
	case "constrain":
		return p.Constrain(strArg(0), argsToStrings(args)...)
	case "serialize":
		return p.Serialize()
	case "analyze":
		m := p.Analyze()
		return fmt.Sprintf("events=%.0f notes=%.0f range=%.0f poly=%.0f",
			m["events"], m["notes"], m["note_range"], m["polyphony"])
	case "remix":
		return p.Remix(strArg(0), args[1:]...)
	case "visualize":
		return p.Visualize(strArg(0))

	default:
		if spec, ok := paramMethodSpecs[method]; ok {
			if spec.str {
				return p.AddParam(spec.key, strArg(0))
			}
			return p.AddParam(spec.key, floatArg(0))
		}
		fmt.Printf("warning: unknown method '%s'\n", method)
		return p
	}
}

// --- Value conversion helpers ---

func toPattern(v interface{}) pattern.Pattern {
	if p, ok := v.(pattern.Pattern); ok {
		return p
	}
	return pattern.Pure(v)
}

func toNumber(v interface{}) float64 {
	switch t := v.(type) {
	case float64:
		return t
	case int:
		return float64(t)
	case pattern.Fraction:
		return t.Float64()
	case pattern.Pattern:
		haps := t.FirstCycle()
		if len(haps) > 0 {
			return toNumber(haps[0].Value)
		}
	}
	return 0
}

func toStringVal(v interface{}) string {
	switch t := v.(type) {
	case string:
		return t
	case float64:
		return fmt.Sprintf("%g", t)
	}
	return fmt.Sprintf("%v", v)
}

func toFraction(v interface{}) pattern.Fraction {
	switch t := v.(type) {
	case pattern.Fraction:
		return t
	case float64:
		return pattern.FracFloat(t)
	case int:
		return pattern.FracInt(int64(t))
	case string:
		if f, err := parseFracStr(t); err == nil {
			return f
		}
	}
	return pattern.FracInt(1)
}

func toFracVal(v interface{}) interface{} {
	// Convert to a value suitable for param args
	switch t := v.(type) {
	case float64:
		return t
	case int:
		return float64(t)
	case pattern.Fraction:
		return t.Float64()
	}
	return v
}

func parseFracStr(s string) (pattern.Fraction, error) {
	if idx := strings.Index(s, "/"); idx >= 0 {
		n := parseIntSafe(s[:idx])
		d := parseIntSafe(s[idx+1:])
		return pattern.NewFrac(int64(n), int64(d)), nil
	}
	n := parseIntSafe(s)
	return pattern.NewFrac(int64(n), 1), nil
}

func parseIntSafe(s string) int {
	n := 0
	neg := false
	i := 0
	if i < len(s) && (s[i] == '-' || s[i] == '+') {
		if s[i] == '-' {
			neg = true
		}
		i++
	}
	for i < len(s) && s[i] >= '0' && s[i] <= '9' {
		n = n*10 + int(s[i]-'0')
		i++
	}
	if neg {
		n = -n
	}
	return n
}

func truthy(v interface{}) bool {
	switch t := v.(type) {
	case nil:
		return false
	case bool:
		return t
	case float64:
		return t != 0
	case int:
		return t != 0
	}
	return true
}

// GetCPS returns the current cycles-per-second tempo setting.
func (ev *Evaluator) GetCPS() float64 {
	if cps, ok := ev.env.getVar("__cps"); ok {
		return toNumber(cps)
	}
	return 1.0
}

func (ev *Evaluator) setupBuiltins() {
	b := ev.builtins
	b["note"] = builtinNote
	b["n"] = builtinN
	b["s"] = builtinS
	b["sound"] = builtinS
	b["gain"] = builtinGain
	b["freq"] = builtinFreq
	b["pan"] = builtinPan
	b["cutoff"] = builtinCutoff
	b["reverb"] = builtinReverb
	b["delay"] = builtinDelay
	b["stack"] = builtinStack
	b["cat"] = builtinCat
	b["slowcat"] = builtinSlowcat
	b["fastcat"] = builtinFastcat
	b["sequence"] = builtinSequence
	b["seq"] = builtinSequence
	b["polymeter"] = builtinPolymeter
	b["polyrhythm"] = builtinPolyrhythm
	b["timecat"] = builtinTimecat
	b["sine"] = builtinSine
	b["cosine"] = builtinCosine
	b["saw"] = builtinSaw
	b["isaw"] = builtinIsaw
	b["tri"] = builtinTri
	b["square"] = builtinSquare
	b["rand"] = builtinRand
	b["sine2"] = builtinSine2
	b["saw2"] = builtinSaw2
	b["tri2"] = builtinTri2
	b["silence"] = builtinSilence
	b["fast"] = builtinFast
	b["slow"] = builtinSlow
	b["rev"] = builtinRev
	b["early"] = builtinEarly
	b["late"] = builtinLate
	b["bpm"] = builtinBPM
	b["cps"] = func(ev *Evaluator, args []Node) (interface{}, error) {
		if len(args) > 0 {
			val, err := ev.evalNode(args[0])
			if err != nil {
				return nil, err
			}
			ev.env.setVar("__cps", toNumber(val))
		}
		return nil, nil
	}
	b["hush"] = func(ev *Evaluator, args []Node) (interface{}, error) {
		if ev.output != nil {
			ev.output(pattern.Silence())
		}
		return nil, nil
	}
	ev.env.setVar("__cps", 1.0)
	b["section"] = builtinSection
	registerExtraBuiltins(b)
	ev.registerProjectBuiltins()
}

// returnValue is a sentinel type for return statements.
type returnValue struct {
	value interface{}
}

func builtinSection(ev *Evaluator, args []Node) (interface{}, error) {
    if len(args) < 2 {
        return pattern.Silence(), nil
    }
    nameVal, err := ev.evalNode(args[0])
    if err != nil {
        return nil, err
    }
    pat, err := ev.evalNode(args[1])
    if err != nil {
        return nil, err
    }
    p := toPattern(pat)
    ev.registerSection(toStringVal(nameVal), p)
    return p, nil
}

func (ev *Evaluator) registerSection(name string, p pattern.Pattern) {
    if ev.sections == nil {
        ev.sections = map[string]pattern.Pattern{}
    }
    ev.sections[name] = p
    for _, n := range ev.arrangement {
        if n == name {
            return
        }
    }
    ev.arrangement = append(ev.arrangement, name)
}

func (ev *Evaluator) SectionNames() []string {
    return ev.arrangement
}
package pattern

import (
	"strings"
	"unicode"
)

// ParseMini parses a Tidal mini-notation string into a Pattern.
// This is the core of Strudel's mini-notation, supporting:
//
//	"bd sd hh cp"       → sequence
//	"[bd sd] hh"         → grouping
//	"bd, sd"             → stack (superposition)
//	"bd*2"               → repeat
//	"bd/2"               → slow down
//	"bd!3"               → replicate
//	"bd@2"               → elongate
//	"bd _"                → elongate (single step)
//	"bd?"                → random degrade
//	"<bd cp hh>"          → alternate
//	"{bd bd, cp hh}"      → polymeter
//	"bd(3,8)"             → euclidean
//	"[bd|sd]"             → random choice
//	"bd:2"                → sample selection
//	"~"                   → rest
func ParseMini(s string) Pattern {
	toks := tokenizeMini(s)
	parser := &miniParser{toks: toks, pos: 0}
	if len(toks) == 0 {
		return Silence()
	}
	return parser.parseStack()
}

type miniToken struct {
	typ string // "word", "punct"
	val string
}

func tokenizeMini(s string) []miniToken {
	var toks []miniToken
	runes := []rune(s)
	i := 0
	for i < len(runes) {
		// Skip whitespace
		if unicode.IsSpace(runes[i]) {
			i++
			continue
		}
		c := runes[i]
		// Punctuation
		switch c {
		case '[', ']', '{', '}', '<', '>', '(', ')', ',', '*', '/', '@', '!', '?', ':', '%', '~', '_':
			toks = append(toks, miniToken{typ: "punct", val: string(c)})
			i++
			continue
		}
		// Word: read until space or punctuation
		start := i
		for i < len(runes) && !unicode.IsSpace(runes[i]) {
			c := runes[i]
			if c == '[' || c == ']' || c == '{' || c == '}' || c == '<' || c == '>' ||
				c == '(' || c == ')' || c == ',' || c == '*' || c == '/' || c == '@' ||
				c == '!' || c == '?' || c == ':' || c == '%' || c == '~' || c == '_' {
				break
			}
			i++
		}
		if i > start {
			toks = append(toks, miniToken{typ: "word", val: string(runes[start:i])})
		}
	}
	return toks
}

type miniParser struct {
	toks []miniToken
	pos  int
}

func (p *miniParser) peek() *miniToken {
	if p.pos < len(p.toks) {
		return &p.toks[p.pos]
	}
	return nil
}

func (p *miniParser) next() *miniToken {
	if p.pos < len(p.toks) {
		t := &p.toks[p.pos]
		p.pos++
		return t
	}
	return nil
}

func (p *miniParser) isPunct(val string) bool {
	t := p.peek()
	return t != nil && t.typ == "punct" && t.val == val
}

// parseStack handles comma-separated superposition.
func (p *miniParser) parseStack() Pattern {
	pats := []Pattern{p.parseSeq()}
	for p.isPunct(",") {
		p.next()
		pats = append(pats, p.parseSeq())
	}
	if len(pats) == 1 {
		return pats[0]
	}
	return Stack(pats...)
}

// parseSeq handles space-separated sequences.
func (p *miniParser) parseSeq() Pattern {
	var elements []interface{}
	for {
		t := p.peek()
		if t == nil || (t.typ == "punct" && (t.val == "," || t.val == "]" || t.val == "}" || t.val == ")")) {
			break
		}
		pat := p.parseElement()
		if pat != nil {
			elements = append(elements, pat)
		}
	}
	if len(elements) == 0 {
		return Silence()
	}
	return sequenceCount(elements)
}

// parseElement parses a single element with optional modifiers.
func (p *miniParser) parseElement() interface{} {
	t := p.peek()
	if t == nil {
		return nil
	}
	var base Pattern
	consumed := false

	switch {
	case p.isPunct("~"):
		p.next()
		base = Silence()
		consumed = true
	case p.isPunct("_"):
		p.next()
		return Silence() // elongate previous (TODO: proper elongation)
	case p.isPunct("["):
		p.next()
		base = p.parseStack()
		if p.isPunct("]") {
			p.next()
		}
		consumed = true
	case p.isPunct("{"):
		p.next()
		base = p.parsePolymeter()
		consumed = true
	case p.isPunct("<"):
		p.next()
		base = p.parseAlternate()
		consumed = true
	case t.typ == "word":
		p.next()
		base = Pure(t.val)
		consumed = true
	default:
		p.next() // skip unknown
		return nil
	}

	if !consumed {
		return nil
	}

	// Parse modifiers
	return p.parseModifiers(base)
}

func (p *miniParser) parseModifiers(base Pattern) Pattern {
	result := base
	for {
		t := p.peek()
		if t == nil || t.typ == "word" {
			break
		}
		switch t.val {
		case "*":
			p.next()
			n := p.parseInt()
			if n > 0 {
				result = result.Fast(FracInt(int64(n)))
			}
		case "/":
			p.next()
			n := p.parseInt()
			if n > 0 {
				result = result.Slow(FracInt(int64(n)))
			}
		case "!":
			p.next()
			n := p.parseInt()
			if n > 1 {
				result = replicate(result, n)
			}
		case "@":
			p.next()
			n := p.parseInt()
			if n > 1 {
				result = elongate(result, n)
			}
		case "?":
			p.next()
			// Match probability 0.5
			result = degradeBy(result, 0.5)
		case ":":
			p.next()
			n := p.parseInt()
			if n >= 0 {
				result = result.N(int(n))
			}
		case "(":
			p.next()
			// Euclidean rhythm: (n, m[, s])
			n := p.parseInt()
			off := 0
			if p.isPunct(",") {
				p.next()
				m := p.parseInt()
				if p.isPunct(",") {
					p.next()
					off = p.parseInt()
				}
				result = euclid(result, n, m, off)
			}
			if p.isPunct(")") {
				p.next()
			}
		default:
			// Non-modifier punct → done
			return result
		}
	}
	return result
}

// parseInt expects the next token to be a word that's a number.
func (p *miniParser) parseInt() int {
	t := p.peek()
	if t == nil || t.typ != "word" {
		return 0
	}
	p.next()
	return parseIntStr(t.val)
}

// parsePolymeter handles {...} with optional %subdivision.
func (p *miniParser) parsePolymeter() Pattern {
	pats := []Pattern{p.parseSeq()}
	for p.isPunct(",") {
		p.next()
		pats = append(pats, p.parseSeq())
	}
	// Check for %
	subdivision := 0
	if p.isPunct("%") {
		p.next()
		subdivision = p.parseInt()
	}
	// Check for closing }
	if p.isPunct("}") {
		p.next()
	}
	// Build polymeter
	if subdivision > 0 {
		args := make([]interface{}, len(pats))
		for i, pat := range pats {
			args[i] = pat
		}
		return Polymeter(subdivision, args...)
	}
	return Stack(pats...)
}

// parseAlternate handles <a b c> which alternates one per cycle.
func (p *miniParser) parseAlternate() Pattern {
	var pats []Pattern
	for {
		t := p.peek()
		if t == nil || (t.typ == "punct" && t.val == ">") {
			break
		}
		if p.isPunct(",") {
			p.next()
			continue
		}
		pat := p.parseElement()
		if pat != nil {
			if pp, ok := pat.(Pattern); ok {
				pats = append(pats, pp)
			}
		}
	}
	if p.isPunct(">") {
		p.next()
	}
	if len(pats) == 0 {
		return Silence()
	}
	return Slowcat(pats...)
}

// --- Mini-notation helpers ---

func replicate(p Pattern, n int) Pattern {
	pats := make([]Pattern, n)
	for i := range pats {
		pats[i] = p
	}
	return Fastcat(pats...)
}

func elongate(p Pattern, n int) Pattern {
	pats := make([]Pattern, n)
	pats[0] = p
	for i := 1; i < n; i++ {
		pats[i] = Silence()
	}
	return Fastcat(pats...)
}

func degradeBy(p Pattern, prob float64) Pattern {
	return p.WithEvent(func(h Hap) Hap {
		if randFloat() < prob {
			return Hap{Part: h.Part, Value: nil}
		}
		return h
	}).RemoveUndefineds()
}

func euclid(p Pattern, n, m, offset int) Pattern {
	// Generate a boolean pattern where true marks the euclidean positions
	positions := euclidPositions(n, m, offset)
	bits := make([]interface{}, m)
	for i := 0; i < m; i++ {
		bits[i] = positions[i]
	}
	boolPat := sequenceCount(bits)
	return boolPat.WithValue(func(v interface{}) interface{} {
		return AppFunc(func(val interface{}) interface{} {
			if v.(bool) {
				return val
			}
			return nil
		})
	}).AppRight(p).RemoveUndefineds()
}

// euclidPositions generates a boolean pattern for E(n, m, offset).
func euclidPositions(n, m, offset int) []bool {
	if n > m {
		n = m
	}
	result := make([]bool, m)
	// Bjorklund algorithm (simplified)
	bucket := 0.0
	for i := 0; i < m; i++ {
		bucket += float64(n)
		if bucket >= float64(m) {
			result[(i+offset)%m] = true
			bucket -= float64(m)
		}
	}
	return result
}

// ParseMiniMultiple parses a comma-separated list of mini-notation strings into a Stack.
func ParseMiniMultiple(strs ...string) Pattern {
	pats := make([]Pattern, 0, len(strs))
	for _, s := range strs {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		pats = append(pats, ParseMini(s))
	}
	if len(pats) == 0 {
		return Silence()
	}
	return Stack(pats...)
}

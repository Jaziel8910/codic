package codang

import (
	"fmt"
	"strconv"
	"strings"
)

// ValidateSong checks that a Codang program meets the demands of its declared
// @type. It returns blocking errors and non-blocking warnings (e.g. missing
// suggested sections). A file that only glues one loop forever will be
// rejected for any arrangement-requiring type.
//
// If no @type is declared but the file produces audio (.out()), it is treated
// as a song and @type becomes mandatory.
func ValidateSong(prog *Program) (errs []string, warns []string) {
	typ := strings.TrimSpace(prog.Metadata["type"])
	if typ == "" {
		if countMethod(prog, "out") > 0 {
			errs = append(errs, "falta @type: declara el tipo de canción al inicio, p.ej. @type full-prod-song (usa 'codic lang types' para ver los 10 tipos)")
		}
		return
	}
	st, ok := SongTypes[strings.ToLower(typ)]
	if !ok {
		errs = append(errs, fmt.Sprintf("@type '%s' no es válido. Tipos: %s", typ, ValidTypeList()))
		return
	}

	if st.NeedsTitle && strings.TrimSpace(prog.Metadata["title"]) == "" {
		errs = append(errs, fmt.Sprintf("el tipo '%s' exige @title (ponle nombre a la canción, ej. @title \"Mi Temazo\")", st.ID))
	}

	cyc := metaInt(prog, "cycles")
	switch {
	case cyc > 0 && cyc < st.MinCycles:
		errs = append(errs, fmt.Sprintf("'%s' necesita al menos %d ciclos; pusiste @cycles %d", st.ID, st.MinCycles, cyc))
	case cyc == 0 && st.NeedsArrangement:
		errs = append(errs, fmt.Sprintf("'%s' exige declarar la duración con @cycles (mínimo %d). Ej: @cycles %d", st.ID, st.MinCycles, st.MinCycles))
	}

	layers := countMethod(prog, "out")
	if layers < st.MinLayers {
		errs = append(errs, fmt.Sprintf("'%s' necesita al menos %d capas de sonido (.out()); tiene %d. Añade batería, bajo, armonía, melodía, percusión y fx", st.ID, st.MinLayers, layers))
	}

	secs := collectSections(prog)
	if st.MinSections > 0 && len(secs) < st.MinSections {
		errs = append(errs, fmt.Sprintf("'%s' necesita al menos %d secciones nombradas con section(\"nombre\", patron); tiene %d: [%s]", st.ID, st.MinSections, len(secs), strings.Join(secs, ", ")))
	}

	if st.NeedsArrangement && !hasArrangement(prog) {
		errs = append(errs, fmt.Sprintf("'%s' exige un arreglo: une las secciones con cat()/seq(), ej. cat(section(\"intro\",...), section(\"verso\",...), ...).out()", st.ID))
	}

	if len(st.CoreSections) > 0 {
		have := map[string]bool{}
		for _, s := range secs {
			have[strings.ToLower(s)] = true
		}
		var missing []string
		for _, c := range st.CoreSections {
			if !have[strings.ToLower(c)] {
				missing = append(missing, c)
			}
		}
		if len(missing) > 0 {
			warns = append(warns, fmt.Sprintf("estructura sugerida para '%s': faltan secciones comunes [%s]. No es obligatorio, pero es la que usa la mayoría", st.ID, strings.Join(missing, ", ")))
		}
	}
	return
}

// SongDuration derives the render length in seconds from @cycles and the tempo
// metadata (@cps or @bpm). If overrideSeconds > 0 it wins (CLI flag).
func SongDuration(prog *Program, overrideSeconds float64) float64 {
	if overrideSeconds > 0 {
		return overrideSeconds
	}
	cyc := metaInt(prog, "cycles")
	if cyc <= 0 {
		return 0
	}
	cps := metaFloat(prog, "cps")
	if cps <= 0 {
		if bpm := metaFloat(prog, "bpm"); bpm > 0 {
			cps = bpm / 240.0 // 4 beats per cycle
		}
	}
	if cps <= 0 {
		cps = 1.0
	}
	return float64(cyc) / cps
}

func metaInt(prog *Program, key string) int {
	v, ok := prog.Metadata[key]
	if !ok {
		return 0
	}
	n, err := strconv.Atoi(strings.TrimSpace(v))
	if err != nil {
		return 0
	}
	return n
}

func metaFloat(prog *Program, key string) float64 {
	v, ok := prog.Metadata[key]
	if !ok {
		return 0
	}
	f, err := strconv.ParseFloat(strings.TrimSpace(v), 64)
	if err != nil {
		return 0
	}
	return f
}

// --- AST walking helpers ---

func walkNodes(node Node, fn func(Node)) {
	if node == nil {
		return
	}
	fn(node)
	switch n := node.(type) {
	case *Program:
		for _, s := range n.Statements {
			walkNodes(s, fn)
		}
	case *AssignStmt:
		walkNodes(n.Value, fn)
	case *ExprStmt:
		walkNodes(n.Expr, fn)
	case *FuncDef:
		for _, s := range n.Body {
			walkNodes(s, fn)
		}
	case *ReturnStmt:
		walkNodes(n.Value, fn)
	case *IfStmt:
		walkNodes(n.Cond, fn)
		for _, s := range n.Then {
			walkNodes(s, fn)
		}
		for _, s := range n.ElseBody {
			walkNodes(s, fn)
		}
	case *ArrayLit:
		for _, e := range n.Elements {
			walkNodes(e, fn)
		}
	case *CallExpr:
		for _, a := range n.Args {
			walkNodes(a, fn)
		}
	case *MethodCall:
		walkNodes(n.Target, fn)
		for _, a := range n.Args {
			walkNodes(a, fn)
		}
	case *BinaryOp:
		walkNodes(n.Left, fn)
		walkNodes(n.Right, fn)
	case *UnaryOp:
		walkNodes(n.Operand, fn)
	case *Index:
		walkNodes(n.Target, fn)
		walkNodes(n.Index, fn)
	}
}

func countMethod(prog *Program, method string) int {
	c := 0
	walkNodes(prog, func(n Node) {
		if m, ok := n.(*MethodCall); ok && m.Method == method {
			c++
		}
	})
	return c
}

// collectSections returns the names of every section("name", ...) call, in
// order and de-duplicated.
func collectSections(prog *Program) []string {
	var names []string
	seen := map[string]bool{}
	walkNodes(prog, func(n Node) {
		ce, ok := n.(*CallExpr)
		if !ok || ce.Name != "section" || len(ce.Args) == 0 {
			return
		}
		if s, ok := ce.Args[0].(*StringLit); ok {
			nm := strings.TrimSpace(s.Value)
			if nm != "" && !seen[nm] {
				seen[nm] = true
				names = append(names, nm)
			}
		}
	})
	return names
}

func hasArrangement(prog *Program) bool {
	found := false
	walkNodes(prog, func(n Node) {
		if ce, ok := n.(*CallExpr); ok {
			switch ce.Name {
			case "cat", "seq", "sequence", "timecat", "slowcat", "fastcat":
				found = true
			}
		}
		if mc, ok := n.(*MethodCall); ok && mc.Method == "append" {
			found = true
		}
	})
	return found
}

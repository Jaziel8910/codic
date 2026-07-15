package codang

import (
	"fmt"

	"github.com/Jaziel8910/codic/internal/project"
)

// projectBuiltins registra las funciones de "álbum / stems / exportar".
func (ev *Evaluator) registerProjectBuiltins() {
	ev.builtins["album"] = ev.builtinAlbum
	ev.builtins["stem"] = ev.builtinStem
	ev.builtins["track"] = ev.builtinTrack
	ev.builtins["export"] = ev.builtinExportDAW
	ev.builtins["export_dawproject"] = ev.builtinExportDAW
}

func (ev *Evaluator) builtinAlbum(_ *Evaluator, args []Node) (interface{}, error) {
	vals, err := evalArgs(ev, args)
	if err != nil {
		return nil, err
	}
	if len(vals) > 0 {
		ev.proj.SetName(toStringVal(vals[0]))
	}
	return nil, nil
}

func (ev *Evaluator) builtinStem(_ *Evaluator, args []Node) (interface{}, error) {
	vals, err := evalArgs(ev, args)
	if err != nil {
		return nil, err
	}
	if len(vals) < 2 {
		return nil, errArgCount("stem", 2, len(vals))
	}
	name := toStringVal(vals[0])
	path := toStringVal(vals[1])
	opts := []project.TrackOption{}
	if len(vals) > 2 {
		opts = append(opts, project.WithGain(toNumber(vals[2])))
	}
	if len(vals) > 3 {
		opts = append(opts, project.WithPan(toNumber(vals[3])))
	}
	ev.proj.AddStem(name, path, opts...)
	return nil, nil
}

func (ev *Evaluator) builtinTrack(_ *Evaluator, args []Node) (interface{}, error) {
	vals, err := evalArgs(ev, args)
	if err != nil {
		return nil, err
	}
	if len(vals) < 2 {
		return nil, errArgCount("track", 2, len(vals))
	}
	name := toStringVal(vals[0])
	pat := toPattern(vals[1])
	opts := []project.TrackOption{}
	if len(vals) > 2 {
		opts = append(opts, project.WithGain(toNumber(vals[2])))
	}
	if len(vals) > 3 {
		opts = append(opts, project.WithPan(toNumber(vals[3])))
	}
	ev.proj.AddPattern(name, pat, opts...)
	return nil, nil
}

func (ev *Evaluator) builtinExportDAW(_ *Evaluator, args []Node) (interface{}, error) {
	vals, err := evalArgs(ev, args)
	if err != nil {
		return nil, err
	}
	if len(vals) < 1 {
		return nil, errArgCount("export_dawproject", 1, len(vals))
	}
	path := toStringVal(vals[0])
	if err := ev.proj.ExportDAWProject(path); err != nil {
		return nil, err
	}
	return nil, nil
}

func errArgCount(name string, want, got int) error {
	return fmt.Errorf("la función %s necesita %d argumentos, recibió %d", name, want, got)
}

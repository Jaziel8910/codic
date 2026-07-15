package codang

import (
	"fmt"
	"sort"
	"strings"
)

// SongType describes one of the 10 canonical kinds of Codang song, from a tiny
// loop up to a fully produced track. The higher the Rank, the more the language
// DEMANDS: more cycles, more layers, more named sections and an explicit
// arrangement (so an AI can't get away with gluing one loop forever).
type SongType struct {
	ID               string
	Name             string
	Rank             int
	MinCycles        int
	MinLayers        int
	MinSections      int
	NeedsArrangement bool
	NeedsTitle       bool
	// CoreSections are the sections most songs of this type use. They are only
	// WARNED about (not hard errors): "no tiene que obligatoriamente tener esa
	// estructura, pero es la que la mayoría usa".
	CoreSections []string
	Description  string
}

// SongTypeList is ordered by Rank (mini-loop -> full-prod-song).
var SongTypeList = []SongType{
	{
		ID: "mini-loop", Name: "Mini loop", Rank: 1,
		MinCycles: 2, MinLayers: 1, MinSections: 0,
		NeedsArrangement: false, NeedsTitle: false,
		Description: "Una idea mínima de 1-2 capas. Para probar un sonido o un ritmo suelto.",
	},
	{
		ID: "loop", Name: "Loop sólido", Rank: 2,
		MinCycles: 8, MinLayers: 2, MinSections: 0,
		NeedsArrangement: false, NeedsTitle: false,
		Description: "Un loop cerrado y contundente, con algo de variación rítmica.",
	},
	{
		ID: "riff", Name: "Riff / gancho", Rank: 3,
		MinCycles: 16, MinLayers: 3, MinSections: 1,
		NeedsArrangement: false, NeedsTitle: false,
		CoreSections:  []string{"riff"},
		Description: "Una idea con gancho que se repite y muta.",
	},
	{
		ID: "groove", Name: "Groove", Rank: 4,
		MinCycles: 24, MinLayers: 4, MinSections: 2,
		NeedsArrangement: true, NeedsTitle: false,
		CoreSections: []string{"intro", "main"},
		Description: "Una base con personalidad rítmica y varias capas.",
	},
	{
		ID: "beat", Name: "Beat", Rank: 5,
		MinCycles: 32, MinLayers: 5, MinSections: 3,
		NeedsArrangement: true, NeedsTitle: true,
		CoreSections: []string{"intro", "verse", "chorus"},
		Description: "La base de una canción: intro, verso, coro, outro.",
	},
	{
		ID: "sketch", Name: "Sketch de canción", Rank: 6,
		MinCycles: 40, MinLayers: 5, MinSections: 3,
		NeedsArrangement: true, NeedsTitle: true,
		CoreSections: []string{"intro", "verse", "chorus"},
		Description: "Un borrador de canción con estructura y dinámica.",
	},
	{
		ID: "track", Name: "Track completo", Rank: 7,
		MinCycles: 48, MinLayers: 6, MinSections: 4,
		NeedsArrangement: true, NeedsTitle: true,
		CoreSections: []string{"intro", "verse", "chorus", "bridge"},
		Description: "Un track terminado: intro, versos, coros, puente y outro.",
	},
	{
		ID: "song", Name: "Canción", Rank: 8,
		MinCycles: 64, MinLayers: 7, MinSections: 5,
		NeedsArrangement: true, NeedsTitle: true,
		CoreSections: []string{"intro", "verse", "pre-chorus", "chorus", "bridge"},
		Description: "Una canción completa con pre-coro, dinámica y transiciones.",
	},
	{
		ID: "epic", Name: "Épica", Rank: 9,
		MinCycles: 96, MinLayers: 8, MinSections: 6,
		NeedsArrangement: true, NeedsTitle: true,
		CoreSections: []string{"intro", "verse", "chorus", "bridge", "breakdown", "outro"},
		Description: "Una pieza larga, en capas, con builds y caídas.",
	},
	{
		ID: "full-prod-song", Name: "Canción full producción", Rank: 10,
		MinCycles: 128, MinLayers: 8, MinSections: 7,
		NeedsArrangement: true, NeedsTitle: true,
		CoreSections: []string{"intro", "verse", "pre-chorus", "chorus", "bridge", "breakdown", "outro"},
		Description: "Lo máximo: canción totalmente producida, con movimientos, rellenos y ear-candy.",
	},
}

// SongTypes maps the type id to its definition.
var SongTypes map[string]SongType

func init() {
	SongTypes = make(map[string]SongType, len(SongTypeList))
	for _, t := range SongTypeList {
		SongTypes[strings.ToLower(t.ID)] = t
	}
}

// ValidTypeList returns a comma-separated list of valid type ids.
func ValidTypeList() string {
	names := make([]string, 0, len(SongTypeList))
	for _, t := range SongTypeList {
		names = append(names, t.ID)
	}
	sort.Strings(names)
	return strings.Join(names, ", ")
}

// Summary returns a human-readable one-block description of the type.
func (t SongType) Summary() string {
	extra := ""
	if t.NeedsTitle {
		extra += ", @title obligatorio"
	}
	if t.NeedsArrangement {
		extra += ", arreglo obligatorio"
	}
	return fmt.Sprintf("%s — %s\n  mínimo: %d ciclos, %d capas, %d secciones%s\n  %s",
		t.ID, t.Name, t.MinCycles, t.MinLayers, t.MinSections, extra, t.Description)
}

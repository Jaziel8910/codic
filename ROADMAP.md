# ROADMAP.md (v1) — OBSOLETO

> Este roadmap original (DAW con TUI + audio en tiempo real) ha sido **superseded** por ROADMAP-v2.md, que es el plan unificado y activo. El c←digo ya elimin← la TUI (internal/tui) y el scheduler/engine realtime, y usa Cobra/Viper en lugar de Charm/oto. Lo ≠nico del v1 que se conserv← (lo no-TUI) est← integrado en ROADMAP-v2.md.

---

# Codic â€” Roadmap

> **Codic** es un DAW de terminal que usa cÃ³digo. Inspirado en [Strudel](https://strudel.cc) / TidalCycles, potenciado por Go + [Charm](https://charm.land) (Bubble Tea v2, Lip Gloss v2, Bubbles v2).  
> Lenguaje propio: **Codang** (`.cdc`) â€” Python-like, diseÃ±ado para humanos e IA.

---

## Fase 0 â€” FundaciÃ³n (sprint 0)

- [x] Investigar Strudel (pattern engine, mini-notation, scheduler, webaudio params)
- [x] Investigar Charm suite (bubbletea, lipgloss, bubbles)
- [x] Investigar Go audio (oto, beep, gomidi)
- [ ] Crear `go mod init github.com/codic/codic`
- [ ] Decidir estructura de directorios
- [ ] README.md + LICENSE

---

## Fase 1 â€” Core Engine (sprint 1)

### PatrÃ³n, Tiempo y Eventos

| Archivo | Lo que hace |
|---------|------------|
| `internal/pattern/fraction.go` | Fracciones exactas con `int64` numerador/denominador. Operaciones `Add`, `Sub`, `Mul`, `Div`, `Sam` (ciclo actual), `NextSam` |
| `internal/pattern/timespan.go` | `TimeSpan{Begin, End Fraction}` â€” `SpanCycles()`, `Intersection()`, `Midpoint()` |
| `internal/pattern/hap.go` | `Hap{Whole, Part, Value, Context}` â€” un evento musical. `WithValue()`, `WithSpan()`, `HasOnset()` |
| `internal/pattern/state.go` | `State{Span, Controls}` â€” entrada de query |
| `internal/pattern/pattern.go` | `Pattern{Query func(State) []Hap}` â€” el core. Pura funciÃ³n de tiempo â†’ eventos |

### Constructores de Patrones

| FunciÃ³n | Comportamiento |
|---------|---------------|
| `Pure(v)` | Valor que se repite cada ciclo |
| `Steady(v)` | Valor continuo |
| `Signal(f)` | SeÃ±al continua desde funciÃ³n matemÃ¡tica |
| `Silence()` | PatrÃ³n vacÃ­o |
| `Stack(pats...)` | Superpone patrones |
| `Cat(pats...)` (fastcat) | Concatena un ciclo de cada patrÃ³n |
| `Slowcat(pats...)` | Concatena un ciclo por patrÃ³n, sin comprimir |
| `Sequence(xs...)` | Secuencia, auto-anida arrays |
| `Polymeter(steps, args...)` | Alinea patrones de distinta longitud |
| `Polyrhythm(xs...)` | Polirritmia pura |

### Transformaciones

| MÃ©todo | Tidal equiv | Efecto |
|--------|-------------|--------|
| `.Fast(f)` | `fast` | Comprime tiempo Ã—f |
| `.Slow(f)` | `slow` | Expande tiempo Ã·f |
| `.Early(o)` | `<~` | Desplaza atrÃ¡s |
| `.Late(o)` | `~>` | Desplaza adelante |
| `.Rev()` | `rev` | Invierte orden |
| `.Jux(f)` | `jux` | Aplica f al canal derecho |
| `.Struct(b)` | `struct` | Filtra con patrÃ³n booleano |
| `.Mask(b)` | `mask` | Enmascara con patrÃ³n booleano |
| `.Every(n, f)` | `every` | Aplica f cada n ciclos |
| `.When(b, f)` | `when` | Aplica f cuando b es true |
| `.Off(t, f)` | `off` | Aplica f desplazado t |
| `.Iter(n)` | `iter` | Itera rotaciones |
| `.Stut(n, fb, t)` | `stut` | Stutter con feedback |
| `.Superimpose(fs...)` | `superimpose` | Apila transformaciones |

### AritmÃ©tica de Patrones

| OperaciÃ³n | Efecto |
|-----------|--------|
| `.Add(n)` | Suma n a valores |
| `.Sub(n)` | Resta n |
| `.Mul(n)` | Multiplica por n |
| `.Div(n)` | Divide por n |
| `.Union(p)` | Fusiona con otro patrÃ³n |

### SeÃ±ales Continuas

```go
Sine    // 0â†’1â†’0 senoidal unipolar
Saw     // 0â†’1 sawtooth
Tri     // 0â†’1â†’0 triangular
Square  // 0 o 1
Cosine  // sine desfasado
Rand    // 0â€“1 aleatorio
```

### ParÃ¡metros de Control (port de strudel-webaudio)

```go
Note(p)     â†’ aÃ±ade {note: valor}
N(p)        â†’ aÃ±ade {n: valor}
S(p)        â†’ aÃ±ade {s: valor}  (selecciona sample/instrumento)
Gain(p)     â†’ volumen
Pan(p)      â†’ paneo
Cutoff(p)   â†’ filtro pasabajos
Resonance(p)â†’ resonancia del filtro
Delay(p)    â†’ feedback delay
DelayTime(p)â†’ tiempo de delay
Reverb(p)   â†’ reverb (room size)
Attack(p)   â†’ ADSR attack
Release(p)  â†’ ADSR release
Sustain(p)  â†’ ADSR sustain
Shape(p)    â†’ distorsiÃ³n / waveshaping
End(p)      â†’ truncar sample
Begin(p)    â†’ offset de sample
Legato(p)   â†’ duraciÃ³n relativa
Velocity(p) â†’ velocidad MIDI
```

### Scheduler

| Componente | DescripciÃ³n |
|------------|-------------|
| `internal/scheduler/scheduler.go` | Loop cada 50ms, query al patrÃ³n, trigger eventos con minLatency 100ms. Reloj independiente |

### Audio Engine

| Componente | DescripciÃ³n |
|------------|-------------|
| `internal/audio/engine.go` | Wrapper sobre `oto.Context`. Inicializa 44100Hz, stereo, `FormatFloat32LE` |
| `internal/audio/oscillator.go` | Osciladores: sine, square, saw, triangle, noise |
| `internal/audio/envelope.go` | ADSR |
| `internal/audio/fx.go` | Reverb (Schroeder), delay, filter (biquad), gain |
| `internal/audio/samples.go` | Sample bank embebido via `embed.FS`. Samples de baterÃ­a, sintes |
| `internal/audio/player.go` | Convierte eventos del scheduler â†’ sonido. Cada Hap con `s`, `note`, `gain` etc se traduce a oscilador+efectos |

### Entregable Fase 1

```
codic eval "s('bd sd hh').fast(2).out()"
â†’ Se escucha un ritmo 4/4 a 120bpm
```

---

## Fase 2 â€” Codang Language (sprint 2)

### Componentes del Lenguaje

| Archivo | DescripciÃ³n |
|---------|-------------|
| `internal/codang/lexer.go` | Tokeniza `.cdc` â†’ tokens |
| `internal/codang/ast.go` | Nodos del AST |
| `internal/codang/parser.go` | Recursive descent parser â†’ AST |
| `internal/codang/evaluator.go` | Walk del AST â†’ Patrones Go + side effects |
| `internal/codang/builtins.go` | Funciones nativas expuestas a Codang |

### Sintaxis de Codang

```python
# Esto es Codang (.cdc)

# AsignaciÃ³n de variables
melody = note("c3 d3 e3 g3")
beat = s("bd sd hh cp")

# Operadores de patrÃ³n (todo pipeline)
track = beat.fast(2).rev().gain(0.8)

# Funciones
func arp(base, interval):
    return note(base + "<0 4 7 12>").s("square").slow(2)

# Uso
arp("c3").out()

# Polirritmia
poly = polymeter([4, note("c2 e2 g2")], [3, note("d2 f#2 a2")])
poly.out()

# Efectos encadenados con seÃ±ales
synth = note("<c3 e3 g3>")
    .s("sawtooth")
    .cutoff(sine.range(200, 3000).slow(4))
    .reverb(0.3)
synth.out()

# BPM
cps(1.5)  # 90 bpm

# Metadata (auto-gestionada)
# @title Mi Tema
# @bpm 120
```

### Mini-Notation (heredada de Tidal)

| SÃ­mbolo | Significado | Ejemplo |
|---------|-------------|---------|
| `" "` | Secuencia | `"bd sd hh"` |
| `[ ]` | AgrupaciÃ³n | `"[bd sd] hh"` |
| `,` | SuperposiciÃ³n | `"[bd sd, hh hh]"` |
| `*` | Repetir | `"bd*3 sd"` |
| `/` | Ralentizar | `"bd/2"` |
| `!` | Replicar | `"bd!3"` |
| `_` | Alargar | `"bd _ _"` |
| `@` | Alargar num | `"bd@3"` |
| `?` | Aleatorio | `"bd?"` |
| `< >` | Alternar | `"<bd cp hh>"` |
| `{ }` | PolimÃ©trica | `"{c3 d3, e3 f3 g3}"` |
| `{ }%` | SubdivisiÃ³n | `"{c3 d3 e3}%8"` |
| `( )` | Euclidiana | `"bd(3,8)"` |
| `:` | SelecciÃ³n | `"bd:2"` |
| `~` | Silencio | `"bd ~ hh"` |
| `\|` | Aleatorio | `"[bd\|cp]"` |

### Funciones Built-in de Codang

**Constructores:**
`note()`, `n()`, `s()`, `freq()`, `gain()`, `pan()`, `cutoff()`, `resonance()`, `delay()`, `delaytime()`, `reverb()`, `attack()`, `release()`, `sustain()`, `shape()`, `begin()`, `end()`, `legato()`, `velocity()`

**Combinadores:**
`fast()`, `slow()`, `rev()`, `early()`, `late()`, `jux()`, `struct()`, `mask()`, `every()`, `when()`, `off()`, `iter()`, `stut()`, `stutwith()`, `superimpose()`, `layer()`, `edit()`, `pipe()`

**SeÃ±ales:**
`sine`, `saw`, `tri`, `square`, `cosine`, `rand`, `isaw`, `saw2`, `tri2`, `square2`

**ComposiciÃ³n:**
`stack()`, `cat()`, `slowcat()`, `sequence()`, `polymeter()`, `polyrhythm()`, `timecat()`

**AritmÃ©tica:**
`.add()`, `.sub()`, `.mul()`, `.div()`, `.union()`, `.range()`, `.scale()`

**Salida:**
`.out()`, `.log()`, `.logvalues()`

---

## Fase 3 â€” TUI Hermosa (sprint 3)

### Layout de la Terminal

```
â”Œâ”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”
â”‚  Codic v0.1                   â— 120 BPM    âµ â¸ â¹          â”‚
â”œâ”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”¤
â”‚  â”Œâ”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”  â”Œâ”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â” â”‚
â”‚  â”‚  EDITOR              â”‚  â”‚  TIMELINE                    â”‚ â”‚
â”‚  â”‚                      â”‚  â”‚  â–ˆâ–ˆâ–ˆâ–ˆâ–ˆâ–ˆâ–ˆâ–ˆâ–‘â–‘â–‘â–‘â–‘â–‘â–‘â–‘â–‘â–‘â–‘â–‘â–ˆâ–ˆâ–ˆ   â”‚ â”‚
â”‚  â”‚  melody = note(...)  â”‚  â”‚  â–ˆâ–ˆâ–‘â–‘â–‘â–‘â–ˆâ–ˆâ–ˆâ–ˆâ–‘â–‘â–‘â–‘â–‘â–‘â–ˆâ–ˆâ–‘â–‘â–‘â–ˆâ–ˆ   â”‚ â”‚
â”‚  â”‚  beat = s("...")     â”‚  â”‚  â–‘â–‘â–‘â–‘â–‘â–‘â–‘â–‘â–‘â–‘â–‘â–‘â–‘â–‘â–‘â–‘â–‘â–‘â–‘â–‘â–‘â–‘â–‘   â”‚ â”‚
â”‚  â”‚  synth = note(...)   â”‚  â”‚  â–ˆâ–ˆâ–ˆâ–ˆâ–ˆâ–ˆâ–‘â–‘â–‘â–‘â–‘â–‘â–ˆâ–ˆâ–ˆâ–ˆâ–‘â–‘â–‘â–‘â–‘â–‘â–‘â–‘   â”‚ â”‚
â”‚  â”‚    .s("saw")         â”‚  â”‚                              â”‚ â”‚
â”‚  â”‚    .cutoff(sine...)  â”‚  â”‚  [timeline en tiempo real]   â”‚ â”‚
â”‚  â”‚                      â”‚  â”‚                              â”‚ â”‚
â”‚  â”‚  [codigo .cdc]       â”‚  â””â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”˜ â”‚
â”‚  â””â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”˜                                   â”‚
â”œâ”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”¤
â”‚  MIXER                                                    â”‚
â”‚  â”Œâ”€â”€â”€â”€â”€â”€â” â”Œâ”€â”€â”€â”€â”€â”€â” â”Œâ”€â”€â”€â”€â”€â”€â” â”Œâ”€â”€â”€â”€â”€â”€â”                     â”‚
â”‚  â”‚melodyâ”‚ â”‚ beat â”‚ â”‚synth â”‚ â”‚ poly â”‚                     â”‚
â”‚  â”‚â–“â–“â–“â–“â–“â–“â”‚ â”‚â–“â–“â–“â–‘â–‘â–‘â”‚ â”‚â–“â–“â–“â–“â–“â–“â”‚ â”‚â–“â–‘â–‘â–‘â–‘â–‘â”‚                     â”‚
â”‚  â”‚-12dB â”‚ â”‚ -3dB â”‚ â”‚ -6dB â”‚ â”‚ -infâ”‚                     â”‚
â”‚  â””â”€â”€â”€â”€â”€â”€â”˜ â””â”€â”€â”€â”€â”€â”€â”˜ â””â”€â”€â”€â”€â”€â”€â”˜ â””â”€â”€â”€â”€â”€â”€â”˜                     â”‚
â”œâ”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”¤
â”‚  LOGS / OUTPUT                                               â”‚
â”‚  [1.0] note:c3 | s:sawtooth | gain:0.8                    â”‚
â”‚  [1.5] note:e3 | s:sawtooth | gain:0.8                    â”‚
â”‚  [2.0] note:g3 | s:sawtooth | gain:0.8                    â”‚
â””â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”˜
```

### Componentes TUI

| Componente | Archivo | Lo que hace |
|------------|---------|-------------|
| **Modelo ppal** | `internal/tui/model.go` | Estado global, mensajes, submodelos |
| **Editor** | `internal/tui/editor.go` | Bubbles textarea con sintaxis .cdc + line numbers |
| **Timeline** | `internal/tui/timeline.go` | VisualizaciÃ³n de eventos en tiempo real. Barras que se iluminan |
| **Mixer** | `internal/tui/mixer.go` | Sliders de volumen/pan por pista |
| **Transporte** | `internal/tui/transport.go` | Play/Stop/Pause + BPM tap + beat counter |
| **Output log** | `internal/tui/output.go` | Log de eventos disparados con colores |
| **Paleta ayuda** | `internal/tui/help.go` | Keybindings superpuestos (modal) |

### Keybindings

| Tecla | AcciÃ³n |
|-------|--------|
| `Ctrl+Enter` | Evaluar cÃ³digo (update pattern) |
| `Space` | Play / Pause |
| `Ctrl+S` | Guardar .cdc |
| `Ctrl+O` | Abrir .cdc |
| `Tab` | Cambiar foco entre paneles |
| `Ctrl+H` | Help |
| `Ctrl+Q` | Salir |
| `Flechas` | Navegar editor / mixer |
| `+ / -` | Cambiar BPM |

---

## Fase 4 â€” Proyectos y Archivos .cdc (sprint 4)

### Formato .cdc

```
# @title Mi Cancion
# @artist: "Artista"
# @bpm: 120
# @cps: 2.0
# @instruments: ["sawtooth", "bd", "sd", "hh"]
# @duration: 32 ciclos
# @codic-version: "0.1.0"
# @created: 2026-07-13
# @modified: 2026-07-13
# @hash: a1b2c3d4...  (auto, cambia si editas el cÃ³digo)

# El cÃ³digo despuÃ©s de los metadatos
melody = note("c3 d3 e3 g3").s("sawtooth").out()
beat = s("bd sd hh cp").fast(2).out()
```

**Propiedades:**
- Metadata auto-gestionada (se actualiza sola al guardar)
- Hash del contenido para detecciÃ³n de cambios
- Lista de instrumentos usados (para precarga)
- BPM y CPS embebidos (al cargar, se restaura el tempo exacto)

### Sistema de Proyectos

| Comando | AcciÃ³n |
|---------|--------|
| `codic new nombre.cdc` | Nuevo proyecto con template |
| `codic open nombre.cdc` | Abrir proyecto |
| `codic save` | Guardar + actualizar metadata |
| `codic export nombre.wav` | Exportar a WAV offline |
| `codic eval "s('bd').out()"` | Evaluar una lÃ­nea (modo REPL rÃ¡pido) |

---

## Fase 5 â€” Ampliaciones (sprint 5+)

### MIDI Out
- Enviar eventos MIDI a hardware/DAW via `gomidi/midi`
- Cada `note()` se traduce a MIDI Note On/Off
- Soporte para controladores MIDI CC

### LibrerÃ­a de Instrumentos Custom
- Instrumentos definidos en Codang (archivos `.cdc` en `~/.codic/instruments/`)
- SÃ­ntesis por tabla de ondas, FM bÃ¡sica, granular
- Ejemplo:
```python
# ~/.codic/instruments/bass.cdc
func bass(note_freq):
    return note(note_freq)
        .s("square")
        .cutoff(saw.range(200, 800).slow(2))
        .attack(0.01)
        .release(0.3)
        .gain(0.7)
```

### Efectos de Audio
- Reverb (Schroeder + convolution)
- Delay (ping-pong, multi-tap)
- Filtros (LPF, HPF, BPF, notch)
- DistorsiÃ³n / waveshaping
- Compresor
- Phaser / Flanger / Chorus

### ExportaciÃ³n
- WAV stereo 44.1kHz/48kHz
- MIDI file (.mid)
- Ableton Live Set (.als) â€” opcional

### Visualizaciones
- Forma de onda del master
- Espectrograma (FFT en terminal)
- VU meter por pista

### ColaboraciÃ³n
- Codic Server: sesiones multi-usuario vÃ­a WebSocket
- Compartir patrones en tiempo real
- Codic Hub: repositorio comunitario de instrumentos .cdc

---

## Timeline Estimado

| Fase | Tiempo | Entregable |
|------|--------|------------|
| Fase 0 | 1 dÃ­a | Repo, go.mod, README |
| Fase 1 | 2 semanas | Pattern engine, scheduler, audio, REPL bÃ¡sico |
| Fase 2 | 1 semana | Lexer, parser, evaluator de Codang, builtins |
| Fase 3 | 2 semanas | TUI completa con editor, timeline, mixer |
| Fase 4 | 3 dÃ­as | Proyectos .cdc, metadata, export |
| Fase 5 | Ongoing | MIDI, efectos, instrumentos custom, colaboraciÃ³n |

**Total MVP: ~5-6 semanas**

---

## Dependencias Go

```
github.com/charmbracelet/bubbletea/v2   â†’ TUI framework
github.com/charmbracelet/lipgloss/v2     â†’ Estilos terminal
github.com/charmbracelet/bubbles/v2      â†’ Componentes TUI
github.com/ebitengine/oto/v3             â†’ Audio out
gitlab.com/gomidi/midi/v2                â†’ MIDI (Fase 5)
github.com/gopxl/beep                    â†’ Sample decode (WAV/MP3/OGG)
```

---

## InspiraciÃ³n

- **[Strudel](https://strudel.cc)** / TidalCycles â€” pattern engine, mini-notation, control params
- **[Charm](https://charm.land)** â€” Bubble Tea, Lip Gloss, Bubbles
- **[Oto](https://github.com/ebitengine/oto)** â€” Audio multiplataforma
- **[SuperCollider](https://supercollider.github.io/)** / SuperDirt â€” parÃ¡metros de control
- **[Hydra](https://hydra.ojack.xyz/)** â€” live coding visual

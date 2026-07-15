# Codic — ROADMAP v2 (unificado con v1)

> **Codic**: CLI-first music production studio. Lenguaje **Codang** (`.cdc`) — Python-like, diseñado para humanos e IA.  
> Sin TUI, sin audio engine en tiempo real. Render offline a WAV + reproductor del OS.

> **Unificación**: este documento fusiona `ROADMAP.md` (v1) y `ROADMAP-v2.md`. Se incorporó todo lo del v1 que **no** es TUI ni audio en tiempo real: referencia del lenguaje Codang (sintaxis, mini-notation, parámetros de control, builtins), formato de archivo `.cdc`, catálogo de efectos, instrumentos custom en Codang, MIDI Out y colaboración. El plan de v1 basado en TUI (Bubble Tea) y realtime (oto) **queda descartado** — sus dependencias ya fueron eliminadas del `go.mod` y `internal/tui` + `internal/scheduler` borrados.

---

## Tabla de Contenido

- [Fase 0 — Pure CLI Foundation](#fase-0--pure-cli-foundation)
- [Fase 1 — 115 CLI Commands](#fase-1--115-cli-commands)
- [Referencia del Lenguaje Codang](#referencia-del-lenguaje-codang)
- [Formato de Archivo `.cdc`](#formato-de-archivo-cdc)
- [Fase 2 — 15 New Codang Functions](#fase-2--15-new-codang-functions)
- [Fase 3 — Albums / EP / Playlists](#fase-3--albums--ep--playlists)
- [Fase 4 — DJ Mode (Procedural Generation)](#fase-4--dj-mode-procedural-generation)
- [Fase 5 — Codang como Lenguaje Formal](#fase-5--codang-como-lenguaje-formal)
- [Fase 6 — Packaging / Ecosystem](#fase-6--packaging--ecosystem)
- [Fase 7 — Colaboración (de v1)](#fase-7--colaboración-de-v1)
- [Estructura de Directorios Final](#estructura-de-directorios-final)
- [Timeline Estimado](#timeline-estimado)
- [Dependencias Go](#dependencias-go)

---

## Fase 0 — Pure CLI Foundation

### Eliminar
```
internal/tui/*              → 8 archivos (theme, editor, timeline, mixer, transport, output, help, model)
internal/audio/engine.go    → Wrapper oto.Context (real-time audio out)
internal/audio/player.go    → Convierte eventos scheduler → sonido
internal/scheduler/*        → Scheduler en tiempo real
```

### Mantener
```
internal/pattern/*          → 17 archivos. Core del pattern engine (intacto)
internal/codang/*           → 8 archivos. Lenguaje Codang (intacto)
internal/audio/render.go    → Offline WAV rendering (clave)
internal/audio/wav.go       → WAV file I/O
internal/audio/oscillator.go → Síntesis
internal/audio/envelope.go  → ADSR
internal/audio/fx.go        → Efectos
internal/audio/effects.go   → Más efectos
internal/audio/drums.go     → Drum synthesis
internal/audio/samples.go   → Sample bank
internal/project/project.go → DAWproject export
```

### Framework CLI
- **Cobra** + **Viper** para comandos y configuración
- `~/.codic/config.yaml` como archivo de config
- Sistema de play: render → WAV temporal → abrir reproductor OS (`rundll32`, `open`, `xdg-open`)
- Watch mode con `fsnotify`

---

## Fase 1 — 115 CLI Commands

### 1. CORE — 8 cmd

| Comando | Acción |
|---------|--------|
| `codic new [tipo] [nombre]` | Scaffold: track\|project\|dj\|template\|instrument |
| `codic play [path] [flags]` | Render + abrir reproductor OS |
| `codic render [path] [out.wav]` | Render permanente (--stems --normalize --format flac\|mp3) |
| `codic eval [code]` | Evalúa Codang inline, imprime Hap |
| `codic run [file.cdc]` | Ejecuta archivo .cdc completo |
| `codic watch [file.cdc]` | fsnotify + re-render automático |
| `codic serve [--port]` | Servidor HTTP (REST API, WebSocket live preview) |
| `codic version` | Versión + git hash + engine info |

### 2. PROJECT / ALBUM / EP / PLAYLIST — 14 cmd

> Del v1: `codic new`/`open`/`save` se cubren con `project init` + los comandos de edición de metadata; `open` implícito al leer `project.yaml`, `save` implícito al escribir. El sistema de proyectos completo vive en `project.yaml` (ver [Formato de Archivo `.cdc`](#formato-de-archivo-cdc) para la metadata por pista).

| Comando | Acción |
|---------|--------|
| `codic project init [--type album\|ep\|single\|playlist]` | Crea structure + project.yaml |
| `codic project info [--json]` | Muestra metadata, duración, BPM, tracks |
| `codic project add <track.cdc> [--pos N]` | Añade track |
| `codic project rm <track>` | Elimina track |
| `codic project reorder <from> <to>` | Reordena |
| `codic project play [--from N] [--loop] [--shuffle]` | Render álbum + play secuencia |
| `codic project render [out.wav] [--gap 2s] [--normalize]` | Render álbum completo |
| `codic project export [--format dawproject\|wav\|stems\|midi]` | Export full album |
| `codic project validate [--fix]` | Verifica integridad |
| `codic project status [tree]` | Árbol de tracks con duración |
| `codic project tag <track> <tag>` | Taggea tracks |
| `codic project metadata [--artist] [--label] [--cover]` | Edita metadatos |
| `codic project split <N>` | Divide álbum en 2 proyectos |
| `codic project merge <project2>` | Fusiona 2 proyectos |

### 3. TRACK — 14 cmd

| Comando | Acción |
|---------|--------|
| `codic track new <name> [--bpm] [--key] [--bars]` | Crea .cdc + template |
| `codic track edit <track>` | Abre en $EDITOR |
| `codic track rm <track>` | Elimina |
| `codic track list [--project] [--glob] [--tags]` | Lista tracks |
| `codic track info <track> [--json]` | Duración, BPM, key, samples |
| `codic track duplicate <track> [--name]` | Duplica |
| `codic track split <track> <bar>` | Parte un track |
| `codic track join <track1> <track2>` | Une 2 tracks |
| `codic track transpose <track> <semitones>` | Transpone notas |
| `codic track bpm <track> <bpm>` | Time-stretch |
| `codic track key <track> <key>` | Re-armoniza |
| `codic track trim <track> <from> <to>` | Recorta |
| `codic track stems <track> [--out dir]` | Extrae stems |
| `codic track analyze <track>` | Análisis: densidad, entropía, rango |

### 4. DJ — GENERACIÓN PROCEDURAL — 12 cmd

| Comando | Acción |
|---------|--------|
| `codic dj learn --genre <g> <paths...> [--tags]` | Entrena desde .cdc |
| `codic dj forget --genre <g>` | Borra modelo |
| `codic dj list [--genre]` | Lista géneros |
| `codic dj stats --genre <g>` | Estadísticas del modelo |
| `codic dj play --genre <g> [--bpm] [--key] [--endless]` | Streaming infinito |
| `codic dj loop --genre <g> --bars N [--out] [--stems]` | Loop finito |
| `codic dj song --genre <g> --structure <s> [--duration]` | Canción completa |
| `codic dj layer --genre <g> --layer <l> --bars N [--variations]` | Solo una capa |
| `codic dj morph --genre <g1> <g2> --steps 8` | Interpola géneros |
| `codic dj jam [--genre] [--instruments...]` | Sesión interactiva |
| `codic dj export --genre <g> [--format dawproject\|midi]` | Exporta modelo |
| `codic dj recommend <track.cdc>` | Recomienda género |

### 5. SAMPLE — 10 cmd

| Comando | Acción |
|---------|--------|
| `codic sample list [--dir] [--tags] [--json]` | Lista samples |
| `codic sample info <sample>` | Metadata: duración, SR, bits, peak |
| `codic sample import <path> [--name] [--tags]` | Importa + copia |
| `codic sample trim <sample> <from> <to>` | Recorta |
| `codic sample normalize <sample> [--lufs -14]` | Normaliza loudness |
| `codic sample reverse <sample>` | Invierte |
| `codic sample fade <sample> [--in] [--out] [--ms]` | Fade in/out |
| `codic sample slice <sample> --beats N` | Rebana en slices |
| `codic sample pack <files...> --name <pack>` | Empaqueta |
| `codic sample convert <sample> <format>` | Convierte formato |

### 6. INSTRUMENT — 8 cmd

| Comando | Acción |
|---------|--------|
| `codic instrument list [--type synth\|drum\|fx] [--json]` | Lista instrumentos |
| `codic instrument info <inst>` | Parámetros, rango, samples |
| `codic instrument new <name> --type synth\|drum\|fx` | Crea definición |
| `codic instrument edit <inst>` | Edita parámetros |
| `codic instrument preset <inst> <preset>` | Guarda/carga presets |
| `codic instrument save <inst>` | Guarda como .cdc reusable |
| `codic instrument import <path>` | Importa SoundFont/etc |
| `codic instrument rm <inst>` | Elimina |

**Instrumentos custom en Codang** (fusionado desde v1 Fase 5): los instrumentos se definen como archivos `.cdc` en `~/.codic/instruments/`, usando síntesis por tabla de ondas, FM básica o granular. Ejemplo:
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

### 7. FX — 8 cmd

| Comando | Acción |
|---------|--------|
| `codic fx list [--type]` | Lista efectos |
| `codic fx info <fx>` | Parámetros, descripción |
| `codic fx chain <track> <fx1,fx2,...>` | Aplica cadena FX |
| `codic fx param <fx> <param> <value>` | Ajusta parámetro |
| `codic fx preset <fx> <preset>` | Guarda/carga presets |
| `codic fx new <name> --chain <fx1,fx2,...>` | FX compuesto |
| `codic fx automate <track> <fx> <param> <pattern>` | Automatiza |
| `codic fx bypass <track> <fx> [on\|off]` | Bypass |

**Catálogo de efectos de audio** (objetivo del motor, fusionado desde v1 Fase 5):
- **Reverb**: Schroeder + convolution
- **Delay**: ping-pong, multi-tap
- **Filtros**: LPF, HPF, BPF, notch (biquad)
- **Distorsión / waveshaping**
- **Compresor**
- **Phaser / Flanger / Chorus**
- **Tremolo / LFO** (modulación de parámetros)

### 8. PATTERN (Codang) — 10 cmd

| Comando | Acción |
|---------|--------|
| `codic pattern new <name> [--code]` | Crea patrón reusable |
| `codic pattern eval <code>` | Evalúa expresión |
| `codic pattern info <pattern>` | Estructura interna |
| `codic pattern render <pattern> [--bars 4] [--out]` | Render directo |
| `codic pattern seed <N>` | Fija seed global |
| `codic pattern scale <name> <root>` | Info de escala |
| `codic pattern chord <root> <type>` | Notas del acorde |
| `codic pattern arp <chord> --style up\|down\|random` | Arpegio |
| `codic pattern harmonize <pattern> --scale <s>` | Armoniza |
| `codic pattern analyze <pattern>` | Métricas |

### 9. EXPORT — 8 cmd

| Comando | Acción |
|---------|--------|
| `codic export wav <track/project> [out.wav]` | WAV |
| `codic export mp3 <track/project> [out.mp3]` | MP3 |
| `codic export flac <track/project> [out.flac]` | FLAC |
| `codic export midi <track/project> [out.mid]` | MIDI |
| `codic export dawproject <track/project> [out.dawproject]` | DAWproject |
| `codic export ableton <track/project> [out.als]` | Ableton Live |
| `codic export flstudio <track/project> [out.flp]` | FL Studio |
| `codic export stems <track/project> --out dir` | Stems separados |

### 10. MIDI — 8 cmd

| Comando | Acción |
|---------|--------|
| `codic midi import <file.mid> [--track N]` | Importa MIDI |
| `codic midi export <track> <out.mid>` | Exporta MIDI |
| `codic midi learn [--device]` | Aprende mapeo |
| `codic midi map <track> --cc <N> <param>` | Mapea control |
| `codic midi monitor [--device]` | Monitoriza entrada |
| `codic midi clock [--bpm] [--out]` | MIDI clock out |
| `codic midi list [--devices]` | Lista dispositivos |
| `codic midi route <input> <output>` | Rutea MIDI |

**MIDI Out** (fusionado desde v1 Fase 5): cada `note()` se traduce a MIDI Note On/Off; soporte para controladores MIDI CC (mapeo vía `codic midi map`). Envío a hardware/DAW vía `gomidi/midi`.

### 11. CONFIG — 8 cmd

| Comando | Acción |
|---------|--------|
| `codic config init` | Crea config default |
| `codic config get <key>` | Muestra valor |
| `codic config set <key> <value>` | Fija valor |
| `codic config list [--json]` | Config completa |
| `codic config reset [key]` | Resetea |
| `codic config edit` | Abre en $EDITOR |
| `codic config import <path>` | Importa config |
| `codic config export <path>` | Exporta config |

### 12. DOCTOR — 6 cmd

| Comando | Acción |
|---------|--------|
| `codic doctor check` | Diagnóstico completo |
| `codic doctor deps` | Verifica dependencias |
| `codic doctor paths` | Estructura de directorios |
| `codic doctor samples` | Samples corruptos |
| `codic doctor repair` | Repara issues |
| `codic doctor env` | Variables de entorno |

### 13. LANG (Codang Language Tooling) — 8 cmd

| Comando | Acción |
|---------|--------|
| `codic lang check [file]` | Type-check + lint |
| `codic lang format [file]` | Formatea código |
| `codic lang docs [--serve]` | Genera docs HTML |
| `codic lang spec` | Imprime EBNF grammar |
| `codic lang ast [file]` | Muestra AST como JSON |
| `codic lang tokens [file]` | Muestra tokens |
| `codic lang repl` | REPL interactivo |
| `codic lang module list\|info\|new` | Gestión de módulos |

### 14. EXTRAS — Tooling — 8 cmd

| Comando | Acción |
|---------|--------|
| `codic completion bash\|zsh\|fish\|powershell` | Completion scripts |
| `codic help [cmd]` | Ayuda con ejemplos |
| `codic man` | Genera página man |
| `codic tutorial [--module]` | Tutorial interactivo |
| `codic examples [--search]` | Ejemplos incluidos |
| `codic template list\|new\|rm\|download` | Templates |
| `codic palette [--format]` | Paleta de funciones/params |
| `codic bench [--pattern] [--cycles]` | Benchmark |

### 15. PKG (Package Manager) — 5 cmd

| Comando | Acción |
|---------|--------|
| `codic pkg install <url/github>` | Instala paquete |
| `codic pkg update [name]` | Actualiza |
| `codic pkg list` | Lista paquetes |
| `codic pkg info <name>` | Info de paquete |
| `codic pkg publish` | Publica (future) |

---

## Referencia del Lenguaje Codang

> Fusionado desde `ROADMAP.md` (v1). El objetivo sigue siendo CLI-first (sin TUI, sin tiempo real); esto es la especificación del lenguaje que faltaba en v2.

### Sintaxis

```python
# Asignación de variables
melody = note("c3 d3 e3 g3")
beat = s("bd sd hh cp")

# Operadores de patrón (pipeline)
track = beat.fast(2).rev().gain(0.8)

# Funciones (definición con `:` y `return`)
func arp(base, interval):
    return note(base + "<0 4 7 12>").s("square").slow(2)
arp("c3").out()

# Polirritmia
poly = polymeter([4, note("c2 e2 g2")], [3, note("d2 f#2 a2")])
poly.out()

# Efectos encadenados con señales continuas
synth = note("<c3 e3 g3>").s("sawtooth").cutoff(sine.range(200, 3000).slow(4)).reverb(0.3)
synth.out()

# Tempo
cps(1.5)   # 90 bpm

# Metadata (directivas, auto-gestionadas)
@title Mi Tema
@bpm 120
```

### Mini-Notation (heredada de Tidal)

| Símbolo | Significado | Ejemplo |
|---------|-------------|---------|
| `" "` | Secuencia | `"bd sd hh"` |
| `[ ]` | Agrupación | `"[bd sd] hh"` |
| `,` | Superposición | `"[bd sd, hh hh]"` |
| `*` | Repetir | `"bd*3 sd"` |
| `/` | Ralentizar | `"bd/2"` |
| `!` | Replicar | `"bd!3"` |
| `_` | Alargar | `"bd _ _"` |
| `@` | Alargar num | `"bd@3"` |
| `?` | Aleatorio | `"bd?"` |
| `< >` | Alternar | `"<bd cp hh>"` |
| `{ }` | Polimétrica | `"{c3 d3, e3 f3 g3}"` |
| `{ }%` | Subdivisión | `"{c3 d3 e3}%8"` |
| `( )` | Euclidiana | `"bd(3,8)"` |
| `:` | Selección | `"bd:2"` |
| `~` | Silencio | `"bd ~ hh"` |
| `\|` | Aleatorio | `"[bd\|cp]"` |

### Parámetros de Control

```go
Note(p)      → añade {note: valor}
N(p)         → añade {n: valor}
S(p)         → añade {s: valor}  (selecciona sample/instrumento)
Gain(p)      → volumen
Pan(p)       → paneo
Cutoff(p)    → filtro pasabajos
Resonance(p) → resonancia del filtro
Delay(p)     → feedback delay
DelayTime(p) → tiempo de delay
Reverb(p)    → reverb (room size)
Attack(p)    → ADSR attack
Release(p)   → ADSR release
Sustain(p)   → ADSR sustain
Shape(p)     → distorsión / waveshaping
End(p)       → truncar sample
Begin(p)     → offset de sample
Legato(p)    → duración relativa
Velocity(p)  → velocidad MIDI
```

### Funciones Built-in

**Constructores:** `note()`, `n()`, `s()`, `freq()`, `gain()`, `pan()`, `cutoff()`, `resonance()`, `delay()`, `delaytime()`, `reverb()`, `attack()`, `release()`, `sustain()`, `shape()`, `begin()`, `end()`, `legato()`, `velocity()`

**Combinadores:** `fast()`, `slow()`, `rev()`, `early()`, `late()`, `jux()`, `struct()`, `mask()`, `every()`, `when()`, `off()`, `iter()`, `stut()`, `stutwith()`, `superimpose()`, `layer()`, `edit()`, `pipe()`

**Señales:** `sine`, `saw`, `tri`, `square`, `cosine`, `rand`, `isaw`, `saw2`, `tri2`, `square2`

**Composición:** `stack()`, `cat()`, `slowcat()`, `sequence()`, `polymeter()`, `polyrhythm()`, `timecat()`

**Aritmética:** `.add()`, `.sub()`, `.mul()`, `.div()`, `.union()`, `.range()`, `.scale()`

**Salida:** `.out()`, `.log()`, `.logvalues()`

---

## Formato de Archivo `.cdc`

```cdc
@title Mi Cancion
@artist "Artista"
@bpm 120
@cps 2.0
@instruments ["sawtooth", "bd", "sd", "hh"]
@duration 32 ciclos
@codic-version "0.1.0"
@created 2026-07-13
@modified 2026-07-13
@hash a1b2c3d4...   (auto, cambia si editas el código)

melody = note("c3 d3 e3 g3").s("sawtooth").out()
beat = s("bd sd hh cp").fast(2).out()
```

**Propiedades:**
- Metadata auto-gestionada (se actualiza sola al guardar)
- Hash del contenido para detección de cambios
- Lista de instrumentos usados (para precarga)
- BPM y CPS embebidos (al cargar, se restaura el tempo exacto)

> **Convención de metadata**: v2 usa la forma **`@bpm 120`** (espacio, sin dos puntos). El v1 usaba `@bpm: 120` (con dos puntos) — se estandariza a la forma **sin dos puntos** en todo el codebase.

---

## Fase 2 — 15 New Codang Functions

Nuevas funciones built-in del lenguaje Codang:

| # | Función | Descripción |
|---|---------|-------------|
| 1 | `harmonize(pattern, scale, intervals...)` | Armoniza notas: `note("c3").harmonize("minor", 3, 5, 7)` → acorde |
| 2 | `counterpoint(cantus, rules?)` | Genera contrapunto especie 1-5. `rules: "strict" \| "free" \| "florid"` |
| 3 | `arpeggiate(pattern, style, range)` | Arpegios: `up`, `down`, `updown`, `random`, `chord`, `broken`, `alberti` |
| 4 | `humanize(pattern, timing, velocity, pitch)` | Micro-variaciones: ±20ms, ±15% vel, ±20 cent |
| 5 | `groove(pattern, template, strength)` | Groove templates: `swing`, `shuffle`, `drag`, `rush`, `latin`, `funk`, `dnb`, `hiphop` |
| 6 | `layer(pattern, density, variation)` | Capas derivadas: octava, quinta, inversiones, rhythm displacement |
| 7 | `morph(pattern1, pattern2, steps, curve)` | Interpola 2 patrones en N pasos. `curve: linear\|exp\|log\|sigmoid\|elastic` |
| 8 | `seed(value)` / `reseed()` | Control determinista de aleatoriedad. `seed(42)` → reproducible |
| 9 | `markov(pattern, order)` | Cadena de Markov orden N. Aprende transiciones notas/ritmos |
| 10 | `euclid(steps, pulses, rotation, events)` | Ritmos euclidianos extendidos: cada pulso = patrón |
| 11 | `constrain(pattern, scale, mode, range)` | Fuerza notas a escala. `mode: clip\|fold\|nearest` |
| 12 | `serialize(pattern)` / `deserialize(json)` | Patrón ↔ JSON portable |
| 13 | `analyze(pattern)` | Métricas: densidad, entropía rítmica, centro tonal, rango, polirritmia |
| 14 | `remix(stems..., strategy)` | Remezcla: `shuffle`, `stutter`, `reverse`, `filter`, `rearrange`, `dub` |
| 15 | `visualize(pattern, type, output?)` | ASCII/ANSI: `pianoroll`, `drumgrid`, `waveform`, `spectrum`, `circle5ths` |

---

## Fase 3 — Albums / EP / Playlists

### Estructura de datos

```
.codic/project.yaml          → Metadata del álbum/EP/playlist
.codic/playlist.yaml         → Playlist con transiciones
tracks/
├── 01-intro.cdc
├── 02-track.cdc
└── 03-outro.cdc
```

### project.yaml

```yaml
project: "Mi Álbum"
type: album          # album | ep | single | playlist
bpm: 128
key: "f# minor"
tracks:
  - name: "Opening"
    file: tracks/01-opening.cdc
    bpm: 126
    duration: "4:32"
  - name: "Drive"
    file: tracks/02-drive.cdc
    bpm: 128
    duration: "6:14"
metadata:
  artist: "Nombre"
  label: "Mi Label"
  release_date: "2026-07-15"
  cover: assets/cover.jpg
```

### playlist.yaml

```yaml
name: "DJ Set"
tracks:
  - file: album1/tracks/02-drive.cdc
    start_at: "0:00"
    end_at: "4:30"
    transition: crossfade 8s
  - file: album2/tracks/01-opening.cdc
    transition: cut
    pitch: +2
    tempo: 130
```

---

## Fase 4 — DJ Mode (Procedural Generation)

### Arquitectura

```
~/.codic/dj/
├── genres/
│   ├── techno.json       # Patrones, BPM range, instrumentos, estructura
│   ├── house.json
│   ├── dnb.json
│   ├── ambient.json
│   ├── trance.json
│   ├── dub.json
│   ├── minimal.json
│   ├── industrial.json
│   ├── breakbeat.json
│   └── custom/           # Géneros aprendidos por el usuario
├── instruments/
│   ├── kick_techno.cdc
│   ├── bass_acid.cdc
│   ├── lead_detroit.cdc
│   └── percussion/
└── structures/
    ├── standard_edm.yaml     # intro-buildup-drop-breakdown-drop-outro
    ├── dub_techno.yaml       # dub-dub-break-dub
    └── minimal.yaml          # additive-subtractive
```

### Aprendizaje

```bash
# Entrena desde archivos .cdc etiquetados
codic dj learn --genre techno tracks/*.cdc --tag kick,bass,lead,pad,fx

# Desde proyecto existente
codic dj learn --from project.yaml --genre techno

# Estadísticas
codic dj stats --genre techno
# → Kick patterns: 47 | Bass lines: 23 | Lead motifs: 18
# → Harmonic vocabulary: minor, phrygian, dorian
```

### Generación

```bash
# Loop infinito (streaming a reproductor)
codic dj play --genre techno --bpm 130 --key "f# minor" --endless

# Loop finito
codic dj loop --genre techno --bars 16 --bpm 128 --out loop.wav --stems

# Canción completa con estructura
codic dj song --genre techno --structure standard_edm --duration "6:00" --out song.wav

# Solo una capa
codic dj layer --genre techno --layer bass --bars 32 --variations 4

# Interpolar entre géneros
codic dj morph --genre techno dnb --steps 8
```

---

## Fase 5 — Codang como Lenguaje Formal

### 5.1 Tree-sitter Grammar

```
grammars/tree-sitter-codang/
├── grammar.js
├── src/scanner.c
├── queries/
│   ├── highlights.scm     → Syntax highlighting (VS Code, Neovim, Helix, Zed)
│   ├── folds.scm           → Code folding
│   ├── indents.scm         → Indent rules
│   ├── locals.scm          → Scope locals
│   └── tags.scm            → Symbol tags
├── binding.gyp
├── Cargo.toml
└── package.json
```

### 5.2 LSP Server (Language Server Protocol)

```
cmd/codic-lsp/main.go
internal/lsp/
├── server.go              → Initialize, TextDocument*, Workspace*
├── handler.go             → Dispatch por método LSP
├── documents.go           → Document state (URI ↔ AST cache)
├── diag.go                → Errores de parseo + tipo en tiempo real
├── completions.go         → Autocompletado (builtins, vars, métodos)
├── hover.go               → Tipo + docstring + ejemplo
├── goto_definition.go     → Go to definition
├── find_references.go     → Find references
├── rename.go              → Rename symbol
├── folding.go             → Folding ranges
├── symbols.go             → Document/workspace symbols
├── semantic_tokens.go     → Highlighting semántico avanzado
├── signature_help.go      → Firma de función/método
├── code_actions.go        → Quick fixes
└── types.go               → Type definitions + checker
```

**Servicios LSP:**
| Método | Status |
|--------|--------|
| `textDocument/diagnostic` | ✅ |
| `textDocument/completion` | ✅ |
| `textDocument/hover` | ✅ |
| `textDocument/definition` | ✅ |
| `textDocument/references` | ✅ |
| `textDocument/rename` | ✅ |
| `textDocument/foldingRange` | ✅ |
| `textDocument/documentSymbol` | ✅ |
| `textDocument/semanticTokens` | ✅ |
| `textDocument/signatureHelp` | ✅ |
| `textDocument/codeAction` | ✅ |

### 5.3 Module System / Imports

```codang
import "std/rhythms" as r
import "std/harmony"

kicks = r.fourOnFloor()
harmony.harmonize(melody, "minor", 3, 5, 7)
```

**Rutas de resolución:**

| Scheme | Path |
|--------|------|
| `std://` | `~/.codic/stdlib/` |
| `file://rel` | `./relative.cdc` |
| `file://abs` | `/absolute/path.cdc` |
| (implícito) | `./<name>.cdc` → `./<name>/mod.cdc` |

**Standard Library (`stdlib/`):**

```
rhythms.cdc     → fourOnFloor, breakbeat, halfTime, shuffle, swing, dnb, funk
scales.cdc      → major, minor, dorian, phrygian, lydian, mixolydian, locrian
chords.cdc      → major, minor, dom7, maj7, min7, dim, aug, sus2, sus4
arpeggios.cdc   → up, down, updown, random, chord, broken, alberti
euclid.cdc      → Euclidean rhythm generators avanzados
markov.cdc      → Markov chain generators
harmony.cdc     → Harmonization, counterpoint, voice leading
grooves.cdc     → Groove templates MIDI-based
kits.cdc        → Drum kit mappings (808, 909, linn, cr78, etcd)
fx.cdc          → FX chain presets
midi.cdc        → MIDI utilities
modulation.cdc  → LFO, envelope, automation generators
random.cdc      → Constrained random generators
```

### 5.4 Type System (Gradual Typing)

```codang
func add(a: number, b: number) -> pattern:
    note(a) + note(b)
end

func process(p: pattern, amount: number = 0.5) -> pattern:
    p.degradeBy(amount).rev()
end
```

**Tipos:**
- Primitivos: `number`, `string`, `bool`, `nil`
- Compuestos: `pattern`, `array[T]`, `fn(T) -> R`
- Inferencia local
- Duck typing opcional (modo estricto desactivable)

```
internal/codang/types.go         → Type definitions
internal/codang/type_checker.go  → Visitor-based type checker
internal/codang/unify.go         → Type unification
```

### 5.5 Code Formatter (`.cdc fmt`)

```
codic lang format [file] --write   → gofmt-style in-place
codic lang format [file] --check   → CI mode
```

Reglas: indent 2 espacios, alineación `=`, espaciado operadores, max line length, imports ordenados, metadata al inicio.

### 5.6 REPL Interactivo

```
$ codic lang repl
Codang v0.1 — live coding music REPL
>>> bpm 128
>>> kicks = s("bd").euclid(8, 4, 0)
>>> hats = s("hh").euclid(8, 3, 1).speed(2)
>>> stack(kicks, hats).out()
◉ Playing... (Ctrl+C to stop)
>>>
```

Comandos internos: `:save`, `:load`, `:info`, `:help`, `:clear`, `:export`

### 5.7 VS Code Extension

```
extensions/vscode/
├── package.json
├── syntaxes/codang.tmLanguage.json    → TextMate grammar (fallback)
├── language-configuration.json        → Comments, brackets, onEnter
└── README.md
```

---

## Fase 6 — Packaging / Ecosystem

### Paquetes (`.codic/packages/`)

```bash
codic pkg install github.com/user/codic-package
codic pkg update [name]
codic pkg list
codic pkg publish  # future: registry público
```

### Standard Library (`.codic/stdlib/`)

Instalada automáticamente con `codic init`:

```bash
~/.codic/
├── config.yaml
├── stdlib/            → Módulos estándar (.cdc)
├── packages/          → Paquetes instalados
├── dj/                → Modelos DJ entrenados
├── instruments/       → Instrumentos custom
├── samples/           → Sample packs
└── templates/         → Project templates
```

### Distribución

```yaml
# codic.yaml — Package descriptor
name: "my-package"
version: "1.0.0"
description: "Paquete de ritmos techno"
author: "Henny"
modules:
  - rhythms
  - percussion
  - basslines
dependencies:
  - "stdlib/chords@^1.0"
```

---

## Fase 7 — Colaboración (de v1)

Fusionado desde v1 Fase 5. Extiende el `serve` de Fase 1 de API de render a sesiones multi-usuario.

- **Codic Server**: sesiones multi-usuario vía WebSocket — compartir patrones en tiempo real.
- **Codic Hub**: repositorio comunitario de instrumentos `.cdc` y paquetes.
- Cada cliente renderiza offline (CLI-first); el servidor sincroniza el estado de los patrones, no el audio.

---

## Estructura de Directorios Final

```
cmd/
├── codic/main.go           → CLI principal (Cobra)
└── codic-lsp/main.go       → LSP server

internal/
├── audio/
│   ├── drums.go            → Drum synthesis (keep)
│   ├── effects.go          → Efectos (keep)
│   ├── envelope.go         → ADSR (keep)
│   ├── fx.go               → FX processors (keep)
│   ├── oscillator.go       → Osciladores (keep)
│   ├── render.go           → Offline WAV render (keep)
│   ├── samples.go          → Sample bank (keep)
│   └── wav.go              → WAV I/O (keep)
├── codang/
│   ├── ast.go              → AST nodes (+ ImportStmt, TypeAnnotation)
│   ├── builtins.go         → Builtins base
│   ├── docs.go             → Docstring parser
│   ├── evaluator.go        → Evaluator (+ imports)
│   ├── extra_builtins.go   → Builtins extra
│   ├── formatter.go        → Code formatter
│   ├── imports.go          → Module resolution
│   ├── lexer.go            → Lexer (+ import keyword)
│   ├── module.go           → Module type
│   ├── module_resolver.go  → Path resolver
│   ├── parser.go           → Parser (+ import, types)
│   ├── project_builtins.go → Album/stem/export builtins
│   ├── repl.go             → REPL loop
│   ├── smoke_test.go       → Tests
│   ├── type_checker.go     → Type checker
│   ├── types.go            → Type definitions
│   └── unify.go            → Type unification
├── dj/
│   ├── engine.go           → DJ generation engine
│   ├── genres.go           → Genre profiles
│   ├── learn.go            → Pattern learning
│   ├── markov.go           → Markov chains
│   ├── rhythm.go           → Drum pattern generation
│   ├── harmony.go          → Harmonic progression gen
│   └── structure.go        → Song structure templates
├── lsp/
│   ├── server.go
│   ├── handler.go
│   ├── documents.go
│   ├── diag.go
│   ├── completions.go
│   ├── hover.go
│   ├── goto_definition.go
│   ├── find_references.go
│   ├── rename.go
│   ├── folding.go
│   ├── symbols.go
│   ├── semantic_tokens.go
│   ├── signature_help.go
│   ├── code_actions.go
│   └── types.go
├── pattern/                → (17 files, sin cambios)
├── project/                → DAWproject export (sin cambios)
└── cmd/                    → Implementación de CLI commands

grammars/
└── tree-sitter-codang/
    ├── grammar.js
    ├── src/scanner.c
    └── queries/*.scm

stdlib/                     → Standard Library (.cdc modules)

extensions/
└── vscode/
    ├── package.json
    ├── syntaxes/codang.tmLanguage.json
    └── language-configuration.json

~/.codic/
├── config.yaml
├── stdlib/
├── packages/
├── dj/genres/
├── instruments/
├── samples/
└── templates/
```

---

## Timeline Estimado

| Fase | Tiempo | Entregable |
|------|--------|------------|
| **Fase 0** — Pure CLI | 3 días | Cobra framework, render+play, `codic new`, `codic render`, watch mode, eliminar TUI |
| **Fase 1** — 115 commands | 2 semanas | Todos los comandos implementados y documentados |
| **Fase 2** — 15 funciones | 1 semana | `harmonize`, `counterpoint`, `arpeggiate`, `humanize`, `groove`, `layer`, `morph`, `seed`, `markov`, `euclid`, `constrain`, `serialize`, `analyze`, `remix`, `visualize` |
| **Fase 3** — Albums/EP | 3 días | `project.yaml`, tracks, render secuencia, crossfade, metadata |
| **Fase 4** — DJ Mode | 2 semanas | Engine de aprendizaje, generación procedural, loops infinitos, morphing |
| **Fase 5.1** — Tree-sitter | 2 días | Grammar + queries + highlights en todos los editores |
| **Fase 5.2** — LSP | 1 semana | Servidor LSP completo con todos los métodos |
| **Fase 5.3** — Modules | 3 días | Sistema de imports, module resolver |
| **Fase 5.4** — Type system | 4 días | Type checker, unificación, inferencia |
| **Fase 5.5** — Stdlib | 1 semana | 12+ módulos estándar documentados |
| **Fase 5.6** — Formatter | 2 días | `codic lang format` |
| **Fase 5.7** — REPL | 2 días | REPL interactivo con historial |
| **Fase 5.8** — VS Code ext | 2 días | Extensión + syntax highlighting |
| **Fase 6** — Packages | 3 días | `codic pkg install`, upgrade, publish |
| **Fase 7** — Colaboración | 1 semana | Codic Server (WebSocket), Codic Hub |

**Total MVP: ~9-11 semanas**

---

## Dependencias Go

```
github.com/spf13/cobra              → CLI framework
github.com/spf13/viper               → Config management
github.com/fsnotify/fsnotify         → Watch mode
 github.com/sourcegraph/go-lsp        → LSP protocol
 gopkg.in/yaml.v3                     → Project YAML
 gitlab.com/gomidi/midi/v2            → MIDI In/Out (MIDI Out, Fase 7 colaboración)
 github.com/go-audio/audio            → WAV I/O avanzado
github.com/icza/bitio                → WAV bit-level
github.com/inconshreveable/mousetrap → CLI helpers
github.com/mattn/go-runewidth        → Terminal width
github.com/pelletier/go-toml/v2      → TOML config alternativo
github.com/charmbracelet/glamour     → Markdown en terminal (help)
```

---

## Inspiración

- **[Strudel](https://strudel.cc)** / TidalCycles — pattern engine, mini-notation
- **[Hydra](https://hydra.ojack.xyz/)** — live coding visual
- **[SuperCollider](https://supercollider.github.io/)** — síntesis procedural
- **[Orca](https://github.com/hundredrabbits/Orca)** — live coding TUI → CLI
- **[Go LSP](https://github.com/gopls/gopls)** — referencia de LSP en Go
- **[Tree-sitter](https://tree-sitter.github.io/)** — parsing incremental
- **[Bitwig DAWproject](https://www.bitwig.com/dawproject/)** — formato abierto de intercambio

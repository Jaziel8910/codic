# Codic

**Codic** is a CLI-first music production studio. Write songs in **Codang** (`.cdc`),
a Python-like music language, and render them offline to WAV — then open them in your OS
player. No TUI, no real-time audio engine.

Codic ports the pattern engine of [Strudel](https://strudel.cc) / TidalCycles to Go and
extends it: offline rendering, polyphony, an embedded sample library, custom instruments,
and a self-describing `.cdc` file format.

## Features

- **Pattern engine** — faithful port of Tidal/Strudel: `Pattern` as a pure function of time,
  `Fraction`-based cycles, mini-notation, signals, combinators.
- **Codang language** — Python-like syntax, `note("c3 d3 e3")`, `stack()`, `fast(2)`, encadenado.
- **Mini-notation** — `"bd*2 [sd hh] ~"` inherited from Tidal.
- **Embedded samples + synthesis** — oscillators, envelopes, effects, no external deps.
- **Offline WAV rendering** — pure render path, then play in your OS player.
- **Self-describing `.cdc`** — metadata (BPM, author, instruments) lives in the file, updates itself.

## Install

**Option A — Windows installer (one file).** Download `codic-windows.exe` from the
[latest release](https://github.com/Jaziel8910/codic/releases/latest), run it, then:

```bat
codic install --samples
```

This drops the embedded stdlib/templates/examples into `%USERPROFILE%\.codic` and
fetches the sample bank.

**Option B — npm / npx (cross-platform).**

```bash
npx @jaziel8910/codic          # runs the right binary for your OS/arch
# or
npm i -g @jaziel8910/codic
```

**Option C — Go.** Requires a Go toolchain:

```bash
go install github.com/Jaziel8910/codic/cmd/codic@latest
```

or build from source:

```bash
git clone https://github.com/Jaziel8910/codic
cd codic
go build ./cmd/codic
```

## Usage

```bash
# Scaffold a new track / project
codic new track my_song
codic new project my_album

# Render a .cdc file to a permanent WAV
codic render my_song.cdc out.wav -d 8

# Render and open it in your OS player
codic play my_song.cdc

# Evaluate Codang inline (prints the resulting Haps)
codic eval 's("bd sd hh cp").fast(2).out()'

# Run a complete .cdc file
codic run my_song.cdc

# Re-render on every save
codic watch my_song.cdc

# Manage config (~/.codic/config.yaml)
codic config init
codic config set default_duration 16
```

### Example `.cdc`

```python
# @bpm 120
melody = note("c3 d3 e3 g3").s("sawtooth").out()
beat   = s("bd sd hh cp").fast(2).out()
```

## Roadmap

See [ROADMAP-v2.md](./ROADMAP-v2.md) for the v2 plan: pure-CLI foundation, 115 CLI
commands, albums/EPs, DJ mode, Codang as a formal language, and packaging.

## License

MIT — see [LICENSE](./LICENSE).

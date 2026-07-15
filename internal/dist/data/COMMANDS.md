# COMMANDS.md — Codic CLI

Resumen de los comandos disponibles. Sintaxis general:

```
codic [comando] [argumentos] [flags]
```

## Workspace / instalación

- `codic install [--samples]` — crea el workspace global `CODIC` (config,
  stdlib, templates, ejemplos, docs, `QUICKSTART.docx`). Con `--samples`
  descarga el banco de sonidos a `sounds/`.
- `codic backup` — comprime el workspace en `CODIC/backups/CODIC-AAAAmmdd-HHMMSS.zip`
  (omite `backups/` y `out/`).
- `codic doctor check|deps|paths|samples|repair|env` — diagnóstico.

## Crear

- `codic new track <nombre>` — nueva canción.
- `codic new project <nombre>` — nuevo álbum/EP/playlist.
- `codic new dj <nombre>` — set de DJ.
- `codic new instrument <nombre>` — instrumento (synth/sampler).
- `codic new template <nombre>` — plantilla.

## Render / reproducir

- `codic render <archivo.cdc>` — renderiza a WAV (en `out/` por defecto).
- `codic play <archivo.cdc>` — renderiza y abre el reproductor del SO.
- `codic run <archivo.cdc>` — ejecuta el `.cdc` y reporta su salida.
- `codic eval "<codigo>"` — evalúa Codang inline.
- `codic watch <archivo.cdc>` — re-renderiza en cada guardado.
- `codic serve` — servidor HTTP (API de render + versión).

## Proyectos y contenido

- `codic project ...` — gestionar álbumes/EPs/playlists (`project.yaml`).
- `codic track ...` — gestionar tracks dentro de un proyecto.
- `codic instrument ...` — gestionar instrumentos.
- `codic fx ...` — cadenas de efectos.
- `codic sample ...` — inspeccionar la librería de samples.
- `codic pattern ...` — crear/manipular patrones `.cdc`.
- `codic export ...` — exportar audio/proyectos a varios formatos.
- `codic midi ...` — importar/exportar/mapear MIDI.
- `codic pkg ...` — gestor de paquetes de patrones/librerías Codang.
- `codic dj ...` — modo DJ (generación, aprendizaje de género, decks).
- `codic lang ...` — herramientas del lenguaje (repl, format, lint, docs).
  - `codic lang types` — lista los 10 tipos de canción (`@type`) y sus mínimos.
  - `codic lang lint <archivo.cdc>` — valida `@type`, capas, secciones y arreglo.
- `codic extras ...` — metrónomo, afinador, tap tempo, teoría.
- `codic config ...` — leer/escribir `config.yaml`.
- `codic version` — versión, hash git e info del motor.

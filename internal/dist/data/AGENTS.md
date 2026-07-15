# AGENTS.md — Codic

Guía para agentes de IA (o humanos técnicos) que trabajen con **Codic**.
Codic es un estudio de música "CLI-first": escribes música en el lenguaje
**Codang** (archivos `.cdc`) y la renderizas offline a WAV.

## La carpeta CODIC (workspace global)

Codic instala todo en una carpeta visible llamada **CODIC** dentro de la
carpeta de usuario (`~/CODIC` en Linux/macOS, `C:\Users\<tú>\CODIC` en Windows).
Es el workspace global del usuario: aquí viven álbumes, proyectos, sonidos,
documentación y respaldos.

Estructura:

```
CODIC/
  AGENTS.md          # este archivo (guía para IA)
  COMMANDS.md        # lista de comandos del CLI
  QUICKSTART.docx    # guía rápida para humanos (Word)
  config.yaml        # configuración del usuario
  sounds/            # banco de samples (strudel.json + .wav)
  docs/              # documentación (llms.txt + 00..11 .md)
  finals/            # renders finales exportados
  projects/          # álbumes / EPs / playlists (project.yaml)
  backups/           # copias de seguridad automáticas (zip)
  stdlib/            # librería estándar de Codang
  templates/         # plantillas de track / project
  examples/          # ejemplos (basic.cdc, etc.)
  instruments/ dj/ packages/ out/ tmp/   # datos de runtime
  <carpetas-del-usuario>/   # álbumes y canciones que el usuario cree
```

## Archivos a leer PRIMERO (contexto)

1. `AGENTS.md` (este).
2. `COMMANDS.md` (qué hace cada comando).
3. `docs/llms.txt` (resumen denso para IA).
4. `docs/09-referencia-completa.md` y `docs/10-catalogo-sonidos.md`.

## Cómo trabajar con el usuario

- Para **renderizar** un archivo: `codic render archivo.cdc` (guarda WAV) o
  `codic play archivo.cdc` (genera y abre el reproductor del SO).
- Para **crear** cosas: `codic new track <nombre>`, `codic new project <nombre>`,
  `codic new dj <nombre>`, `codic new instrument <nombre>`.
- Los **samples** viven en `sounds/`. Si faltan: `codic install --samples`.
- Para **respaldar** todo el workspace: `codic backup` (zip en `backups/`).
- Para **diagnosticar**: `codic doctor check`.
- La config del usuario está en `config.yaml` (`codic config get/set`).

## Convenciones

- Los archivos de música son `.cdc` (Codang). Sintaxis tipo Python:
  `note("c3 d3 e3")`, `stack()`, `fast(2)`, mini-notación `"bd*2 [sd hh] ~"`.
- **Todo `.cdc` empieza declarando `@type`** (uno de los 10 tipos: `mini-loop`,
  `loop`, `riff`, `groove`, `beat`, `sketch`, `track`, `song`, `epic`,
  `full-prod-song`). El tipo OBLIGA a una complejidad mínima: ciclos, capas
  (`.out()`), secciones (`section("nombre", patron)`) y arreglo (`cat(...)`).
  Codang **rechaza** el render si no se cumple. Consulta `codic lang types`.
- NO entregues loops sueltos pegados: una canción de verdad tiene intro, versos,
  coros, puente/outro. Nombra cada parte con `section(...)` y únelas en un
  arreglo. La duración sale de `@cycles` + `@bpm`/`@cps` (no de un valor fijo).
- Un `.cdc` debe terminar cada capa con `.out()` para producir audio.

Cabecera mínima de una canción completa:

```
@type full-prod-song
@title "Nombre"
@cycles 128
@bpm 120

section("intro",  s("bd ~ ~ ~"))
section("verse",  s("bd sd"))
section("chorus", s("bd*2 sd"))
# … más secciones (pre-chorus, bridge, breakdown, outro) …
cat(s("intro"), s("verse"), s("chorus")).out()   # arreglo
s("bd*4").out()                                   # más capas
note("c2 e2 g2").out()
```
- Los proyectos (álbumes) usan `project.yaml` y se gestionan con
  `codic project` / `codic track`.
- No borres `sounds/`, `stdlib/`, `config.yaml` ni `docs/`; son parte del programa.
- El contenido del usuario (sus álbumes/canciones) vive en subcarpetas que él
  mismos crea dentro de `CODIC/` o en `projects/`. Presérvalo siempre.

## Notas técnicas

- El motor de patrones es un port de Tidal/Strudel (Fraction, mini-notación,
  señales, combinadores). El render es offline (sin audio en tiempo real).
- El binario del instalador descarga el banco de samples desde el release de
  GitHub (`v0.1.0/samples.zip`) y lo extrae en `sounds/`.
- El workspace se define en `cli.CodicDir()` = `~/CODIC`.

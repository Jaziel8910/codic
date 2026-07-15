# Referencia completa de Codang

Esta es la lista de todo lo que existe en el lenguaje. Si ya sabes lo básico,
esto es el "diccionario". Los métodos se aplican con punto (`.`) después de un
patrón; las funciones sueltas se escriben solas o entre paréntesis.

---

## Tipos de canción (obligatorio: `@type`)

Toda canción **empieza declarando su tipo** con `@type`. El tipo fija cuánta
complejidad exige Codang: no puedes declarar una `full-prod-song` y entregar un
loop suelto de 4 golpes. Si el `.cdc` produce audio (`.out()`) y no declara
`@type`, el render se rechaza. Consulta los 10 tipos con `codic lang types`.

| Tipo | Ciclos mín. | Capas mín. | Secciones mín. | Título | Arreglo |
|------|:-----------:|:----------:|:--------------:|:------:|:-------:|
| `mini-loop` | 2 | 1 | 0 | no | no |
| `loop` | 8 | 2 | 0 | no | no |
| `riff` | 16 | 3 | 1 | no | no |
| `groove` | 24 | 4 | 2 | no | sí |
| `beat` | 32 | 5 | 3 | sí | sí |
| `sketch` | 40 | 5 | 3 | sí | sí |
| `track` | 48 | 6 | 4 | sí | sí |
| `song` | 64 | 7 | 5 | sí | sí |
| `epic` | 96 | 8 | 6 | sí | sí |
| `full-prod-song` | 128 | 8 | 7 | sí | sí |

- **Capas** = cada `.out()` (batería, bajo, armonía, melodía, percusión, fx…).
- **Secciones** = cada `section("nombre", patron)` (intro, verse, chorus, bridge…).
- **Arreglo** = unir las secciones con `cat(...)`/`seq(...)` y sacarlas con `.out()`.
- Cada tipo **sugiere** sus secciones típicas (intro, verse, pre-chorus, chorus,
  bridge, breakdown, outro). No son obligatorias, pero es la estructura que usa
  la mayoría; si faltan, Codang lo **avisa** (no bloquea).

### Cabecera de una canción

```
@type full-prod-song
@title "Mi Temazo"
@cycles 128          # duración en ciclos (define lo que dura el render)
@bpm 120             # o @cps 0.5 (ciclos por segundo)

section("intro",  s("bd ~ ~ ~"))
section("verse",  s("bd sd"))
section("chorus", s("bd*2 sd"))
# … resto de secciones …

cat(s("intro"), s("verse"), s("chorus")).out()   # arreglo
s("bd*4").out()                                   # capas
note("c2 e2 g2").out()
```

La **duración** del render sale de `@cycles` y el tempo (`@cps`, o `@bpm/240`).
Con `@cycles 128` y `@bpm 120` → 128 / 0.5 = 256 s. La bandera `-d` la sobreescribe.

---

## Fuentes de sonido (crean patrones)

| Nombre | Qué devuelve |
|--------|--------------|
| `sonido("…")` / `s("…")` | patrón de samples (batería, etc.) |
| `nota(…)` / `n(…)` | patrón de notas (número o nombre como `"do3"`) |
| `frecuencia(hz)` / `freq` | patrón a una frecuencia en Hz |
| `seno` / `coseno` / `sierra` / `isaw` / `triangulo` / `cuadrada` | ondas (señales) |
| `aleatorio` / `rand` | valor al azar (señal) |
| `silencio()` | patrón vacío |

## Construcción de patrones

`pila(a,b,…)`, `secuencia(a,b,…)` / `cat`, `secuenciaRapida` / `fastcat`,
`secuenciaLenta` / `slowcat`, `polimetro`, `polirritmia`, `timecat`,
`secuencia` / `seq`, `silencio()`.

`section("nombre", patron)` — nombra una parte de la canción (intro, verse,
chorus…). Cuenta como sección para el `@type` y devuelve el patrón para poder
unirlo en el arreglo con `cat(...)`.

## Mini-notación (dentro de `"…"`)

espacio = orden · `,` = a la vez · `[ ]` = grupo · `{ }` = polímetro (`%n`) ·
`< >` = alternar · `*` = repetir · `/` = alargar · `?` = azar 50% ·
`:n` = nota n · `(n,m,s)` = euclídeo · `~` = silencio.

---

## Transformaciones (métodos)

**Tiempo:** `rapido(n)` `lento(n)` `adelantar(n)` `atrasar(n)` `iter(n)` `iterBack(n)`
**Orden:** `revertir()` `palindromo()` `brak()` `mezclar()` `scramble()` `shuffle()`
**Recorte:** `zoom(a,b)` `estirar(n)` `contraer(n)` `expandir(n)` `encajar(n)` `doblar(n)` `encoger(n)`
**Dentro:** `dentro(a,b,f)` `fuera(a,b,f)` `adentro(f)` `hacia(f)` `ritmo(f)` `off(f)` `offspray(f)`
**Capas:** `capas(f…)` `superponer(f…)` `golpear(n)` `golpearCon(n,f)` `golpearPorCada(n,f)` `prensar(n)`
**Swing:** `swing(n)` `swingBy(n,f)`
**Trozos:** `trozo(n,f)` `trozoLento(n,f)` `trozoRapido(n,f)` `trozoEn(n,f)` `trozoAtras(n,f)`
**Ancla:** `ancla(n,f)` `fancla(n,f)` `pancla(n,f)` `recorrido(f)`
**Espacio:** `crecer(n)` `apurar(n)` `mantener(n)` `soltar(n)` `linger` `tomar(n)` `esparcir(n)` `exprimir(f)` `como(t)` `comoNumero()`
**Ritmo:** `euclid(p,g,r)` `euclidRot(p,g,r)` `euclidLegato(p,g,r)` `euclidish(grupos,g)`
**Canales:** `jux(f)` `juxBy(pan,f)` `juxFlip(f)` `juxFlipBy(pan,f)`
**Varias:** `repetirCiclos(n)` `comprimir(a,b)` `unison(n)` `zip(a,b)` `degradar()` `degradarPor(p)`

---

## Parámetros de sonido (métodos)

**Mezcla:** `volumen(n)` `paneo(n)` `velocidad(n)` `room(n)` `roomsize(n)` `roomfade(n)` `roomlp(n)` `delay(n)` `delaytime(n)` `delayfeedback(n)` `delaysync(n)` `crush(n)` `ancho_pan(n)`
**Filtros:** `filtro(n)` `resonancia(n)` `lpf(n)` `hpf(n)` `lpq(n)` `hpq(n)` `bpf(n)` `bpq(n)` `djf(n)` `bandf(n)` `bandq(n)`
**Distorsión/carácter:** `distorsion(n)` `dist(n)` `drive(n)` `distorttype(t)` `vib(n)` `vibmod(n)` `vocal(v)` `ribbon(n)` `tremolo(n)` `tremolorate(n)` `tremolodepth(n)` `phaser(n)` `phasercenter(n)` `phaserdepth(n)` `phasersweep(n)` `chorus(n)` `comb(n)`
**Env.:** `ataque(n)` `sostenido(n)` `liberacion(n)` `forma(n)` `decay(n)` `adsr(a,s,d,r)`
**Tono:** `frecuencia(n)` `transponer(n)` `detune(n)` `octava(n)` `subir(n)` `bajar(n)`
**Sample:** `acelerar(n)` `loop(n)` `loopBegin(n)` `loopEnd(n)` `loopSample(a,b)` `cut(n)` `begin(n)` `end(n)` `speed(n)`
**Salida:** `orbit(n)` `canal(n)` `canales(n)` `etiqueta(t)` `label(t)` `tag(t)` `color(t)` `brand(t)`
**LFO:** `lfo("mando", vel, cant)` `lfoSignal(…)`

---

## Música y MIDI

`escala(nombre, …notas)` · `edoScale(div, …)` · `scaleTranspose(n)` · `modo(t)` ·
`afinar(n)` · `acorde(nombre, tipo)` · `voicing(…)` · `addVoicings(…)` ·
`arp(tipo)` · `arpWith(f)` · `partial(n)` · `midinote(n)` · `midichan(n)` ·
`midicmd(t)` · `ccn(n)` · `ccv(n)` · `control(…)` · `sysex(…)` · `progNum(n)` ·
`pitchwheel(n)` · `nrpnn(n)` · `nrpv(n)` · `fm(…)` `fmi` `fmh` `fmw` `fmattack` … ·
`wt…` (wavetable) · `lpattac`/`hpat…`/`bpat…` (envolventes de filtro) ·
`toMidi()` `fromMidi()`.

---

## Generación y azar (funciones sueltas y métodos)

Funciones sueltas: `elegir(…)` `choose` · `elegirCiclos(…)` `chooseCycles` ·
`elegirPesos(…)` `wchoose` · `randcat` · `corrido(n)` `run` · `irand(n)` ·
`rand2()` · `randL()` · `rangoX(a,b)` `rangex` · `escala(…)` `scale` ·
`acorde(…)` `chord` · `edoScale` · `palindromo()` `palindrome` · `brak()` ·
`mezclar()` · `euclid(p,g,r)` · `euclidish(…)` · `espiral(esc, pasos, salto, raiz)` `spiral` ·
`secuenciaP(…)` `seqPLoop` · `stepcat(…)` · `step(…)` · `stepwise(…)`.

Métodos de "vida": `degradar()` `degrade` · `degradarPor(p)` `degradeBy` ·
`a veces(f)` `sometimes` · `a menudo(f)` `often` · `rara vez(f)` `rarely` ·
`casi siempre(f)` `almostAlways` · `casi nunca(f)` `almostNever` · `siempre(f)` `always` ·
`nunca(f)` `never` · `algunosCiclos(f)` `someCycles` · `algunosCiclosPor(p,f)` `someCyclesBy` ·
`cada(n,f)` `every` · `dentro(…)` · `fuera(…)` · `unison(n)` · `comprimir(a,b)` ·
`zip(a,b)` · `xfade(a,b,t)` `crossfade` · `inhabitar(f)` `inhabit` · `defragmentHaps()`.

Stems/sample: `striate(n)` `stripe(n)` `chop(n)` `slice(n)` `bite(a,b,n)` `splice(n)` · `gap(n)` · `reset()` `restart()`.

---

## Álbum y exportación

`album("nombre")` · `bpm(n)` · `stem("nombre","ruta",vol?,pan?)` ·
`track("nombre", patron, vol?, pan?)` · `export("archivo.dawproject")` ·
`export_dawproject("archivo.dawproject")`.

Desde terminal: `codic export archivo.cdc [salida.dawproject]`.

---

## Reproducir (CLI)

```
codic demo                                   # suena una demo
codic eval "sonido(\"bd sd\").out()"         # evalúa código
codic file cancion.cdc                       # reproduce un archivo
codic export cancion.cdc salida.dawproject   # exporta sin sonar
codic lang types                             # los 10 tipos de canción
codic lang lint cancion.cdc                  # valida tipo, capas, secciones…
```

---

## Notas importantes

- Empieza toda canción declarando `@type` (ver "Tipos de canción"); Codang
  exige la complejidad del tipo elegido y rechaza el render si no se cumple.
- Siempre termina la idea que quieres oír con `.out()`.
- Los nombres de función se pasan **entre comillas** a los métodos de azar:
  `sonido("bd").a veces("MiFunc")`.
- Los números de volumen/paneo van de 0 a 1; el paneo de −1 (izq) a +1 (der).
- Todo vive en un ciclo que se repite para siempre; tú decides qué pasa en él.

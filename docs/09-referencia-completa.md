# Referencia completa de Codang

Esta es la lista de todo lo que existe en el lenguaje. Si ya sabes lo básico,
esto es el "diccionario". Los métodos se aplican con punto (`.`) después de un
patrón; las funciones sueltas se escriben solas o entre paréntesis.

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
```

---

## Notas importantes

- Siempre termina la idea que quieres oír con `.out()`.
- Los nombres de función se pasan **entre comillas** a los métodos de azar:
  `sonido("bd").a veces("MiFunc")`.
- Los números de volumen/paneo van de 0 a 1; el paneo de −1 (izq) a +1 (der).
- Todo vive en un ciclo que se repite para siempre; tú decides qué pasa en él.

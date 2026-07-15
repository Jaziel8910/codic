# 10 · Catálogo completo de sonidos 🎛️

Esta es **toda** la lista de lo que puedes usar en Codic para hacer ruido.
Dividido en 4 familias:

1. **Sintetizadores** (ondas generadas por código, siempre disponibles)
2. **Batería sintetizada** (bombo, caja, hi-hat… hechos en código)
3. **Samples de Dirt-Samples** (archivos de audio reales, 218 nombres)
4. **Notas musicales** (para melodías)

---

## 1 · Sintetizadores (ondas)

Se usan con `.sintetizador("...")` o `.sonido("...")`. Codic genera la onda en
vivo. Funcionan con `nota(...)` para darles tono:

```python
nota("do3 mi3 sol3").sintetizador("sierra").out()
nota("la2").sonido("seno").gain(0.4).out()
```

| Nombre | Alias | Tipo de onda |
|--------|-------|--------------|
| `sine` | `seno` | Senoidal (suave, pura) |
| `sawtooth` | `saw`, `sierra` | Sierra (rítmica, áspera) |
| `square` | `cuadrada` | Cuadrada (8-bit, vidriosa) |
| `triangle` | `tri`, `triangulo` | Triangular (entre seno y sierra) |
| `noise` | `ruido` | Ruido blanco |
| `cosine` | `coseno` | Coseno (igual que seno, desfasada) |
| `rand` | — | Ruido aleatorio (suena como `ruido`) |
| `isaw` | — | Sierra invertida |

> Consejo: combina con `cutoff(800)` para un filtro, o `distort(0.3)` para suciedad.

---

## 2 · Batería sintetizada (siempre disponible)

Estos no necesitan archivos: Codic los fabrica. Usa `sonido("...")`:

```python
sonido("bd sd hh cp").out()          # bombo caja hihat palmas
sonido("bd").gain(0.9).out()
```

| Nombre | También como | Sonido |
|--------|--------------|--------|
| `bd` | `kick`, `bt` | Bombo / kick |
| `sd` | `snare`, `sn` | Caja / snare |
| `hh` | `hihat`, `ch` | Hi-hat cerrado |
| `oh` | — | Hi-hat abierto |
| `cp` | `clap` | Palmas / clap |

Para aún más baterías reales, usa los samples de abajo (`808bd`, `909sd`, etc.).

---

## 3 · Samples de Dirt-Samples (218 nombres)

Codic trajo la **librería completa de Strudel** (`Dirt-Samples`): 2038 archivos
de audio reales en `samples/`. Cada nombre tiene varias **variantes**; elige
una con `:n` (número, empezando en 0):

```python
sonido("808bd").out()          # variante 0 (la primera)
sonido("808bd:3").out()        # variante 3
sonido("808bd*4, hh*8").out()  # patrón con repeticiones
```

Si escribes un número mayor al disponible, Codic usa la variante 0.

### Baterías por banco (las más útiles)
- **808**: `808` `808bd` `808cy` `808hc` `808ht` `808lc` `808lt` `808mc` `808mt` `808oh` `808sd`
- **909**: `909` `909bd`

### Batería y percusión genérica
`bd`(24) · `sd`(2) · `hh`(13) · `hh27`(13) · `ht`(16) · `lt`(16) · `mt`(16) ·
`oh`(6) · `cp`(2) · `cr`(6) · `cb`(1) · `rm`(2) · `rs`(1) · `sn`(52) ·
`perc`(6) · `metal`(10) · `tok`(4) · `tink`(5) · `click`(4) · `clak`(2) ·
`bin`(2) · `coins`(1) · `can`(14) · `bottle`(13) · `pebbles`(1) ·
`glass`→`glasstap`(3) · `tabla`(26) · `tabla2`(46) · `tablex`(3) ·
`circus`(3) · `tacscan`(22) · `sundance`(6) · `stomp`(10) · `hand`(17)

### Bajos (bass)
`bass`(4) · `bass0`(3) · `bass1`(30) · `bass2`(5) · `bass3`(11) ·
`bassdm`(24) · `bassfoo`(3) · `jungbass`(20) · `jvbass`(13) ·
`moog`(7) · `subroc3d`(11) · `ul`(10) · `ulgab`(5)

### Guitarras / cuerdas / teclados
`gtr`(3) · `psr`(30) · `casio`(3) · `sid`(12) · `juno`(12) ·
`sitar`(8) · `pad`(3) · `padlong`(1) · `space`(18) · `synth`→`synth`
`mund`→`world`(3) · `sax`(22) · `trump`(11) · `v`(6) · `ul`(10)

### Plucks / stabs / arps
`pluck`(17) · `stab`(23) · `arp`(2) · `arpy`(11) · `jab`→`gab`(10) ·
`gabba`(4) · `gabbaloud`(4) · `gabbalouder`(4) · `flick`(17) ·
`future`(17) · `print`(11) · `proc`(2) · `procshort`(8)

### Beats / breaks / drum machines
`dr`(42) · `dr2`(6) · `dr55`(4) · `dr_few`(8) · `drum`(6) ·
`drumtraks`(13) · `breaks125`(2) · `breaks152`(1) · `breaks157`(1) ·
`breaks165`(1) · `tech`(13) · `techno`(7) · `house`(8) ·
`electro1`(13) · `rave`(8) · `rave2`(4) · `ravemono`(2) ·
`hardcore`(12) · `gabba`(4) · `industrial`(32) · `armora`(7) ·
`made`(7) · `made2`(1) · `mash`(2) · `mash2`(4) · `miniyeah`(4) ·
`monsterb`(6) · `kicklinn`(1) · `clubkick`(5) · `hardkick`(6) ·
`popkick`(10) · `reverbkick`(1) · `koy`(2) · `tok`(4)

### Voces / palabras / efectos de voz
`baa`(7) · `baa2`(7) · `hmm`(1) · `mouth`(15) · `speech`(7) ·
`speechless`(10) · `speakspell`(12) · `alphabet`(26) · `numbers`(9) ·
`num`(21) · `msg`(9) · `fest`(1) · `sugar`(2) · `yeah`(31) ·
`voodoo`(5) · `wind`(10)

### Ruido / atmósferas / objetos
`noise`(1) · `noise2`(8) · `birds`(10) · `birds3`(19) · `bubble`(8) ·
`fire`(1) · `water`→`outdoor`(6) · `seawolf`(3) · `invaders`(18) ·
`incoming`(8) · `creak`→`crow`(4) · `fx`→`feelfx`(8) · `flick`(17) ·
`glitch`(8) · `glitch2`(8) · `wobble`(1) · `voodoo`(5) · `latibro`(8) ·
`world`(3) · `ul`(10)

### Otros (nombres sueltos)
`ab`(12) · `ade`(10) · `ades2`(9) · `ades3`(7) · `ades4`(6) · `alex`(2) ·
`amencutup`(32) · `auto`(11) · `bev`(2) · `blue`(2) · `bend`(4) ·
`blip`(2) · `bleep`(13) · `cc`(6) · `chin`(4) · `control`(2) ·
`cosmicg`(15) · `d`(4) · `db`(13) · `dist`(16) · `dork2`(4) ·
`dorkbot`(2) · `e`(8) · `east`(9) · `em2`(6) · `erk`(1) · `f`(1) ·
`feel`(7) · `fm`(17) · `foo`(27) · `gab`(10) · `glasstap`(3) ·
`hc`(6) · `haw`(6) · `h`(7) · `hit`(6) · `ho`(6) · `hoover`(6) ·
`if`(5) · `ifdrums`(3) · `kurt`(7) · `led`(1) · `less`(4) ·
`linnhats`(6) · `mute`(28) · `newnotes`(15) · `notes`(15) ·
`oc`(4) · `odx`(15) · `off`(1) · `peri`(15) · `realclaps`(4) ·
`sheffield`(1) · `short`(5) · `sequential`(8) · `sf`(18) ·
`sheffield`(1) · `synth` → no listado · `tabla`(26) · `tink`(5) ·
`toys`(13) · `uxay`(3) · `xmas`(1) · `diphone`(38) · `diphone2`(12)

### Lista completa y exacta (218 nombres, con nº de variantes)

```
808(6) 808bd(25) 808cy(25) 808hc(5) 808ht(5) 808lc(5) 808lt(5) 808mc(5)
808mt(5) 808oh(5) 808sd(25) 909(1) ab(12) ade(10) ades2(9) ades3(7)
ades4(6) alex(2) alphabet(26) amencutup(32) armora(7) arp(2) arpy(11)
auto(11) baa(7) baa2(7) bass(4) bass0(3) bass1(30) bass2(5) bass3(11)
bassdm(24) bassfoo(3) battles(2) bd(24) bend(4) bev(2) bin(2) birds(10)
birds3(19) bleep(13) blip(2) blue(2) bottle(13) breaks125(2) breaks152(1)
breaks157(1) breaks165(1) breath(1) bubble(8) can(14) casio(3) cb(1)
cc(6) chin(4) circus(3) clak(2) click(4) clubkick(5) co(4) coins(1)
control(2) cosmicg(15) cp(2) cr(6) crow(4) d(4) db(13) diphone(38)
diphone2(12) dist(16) dork2(4) dorkbot(2) dr(42) dr2(6) dr55(4) dr_few(8)
drum(6) drumtraks(13) e(8) east(9) electro1(13) em2(6) erk(1) f(1)
feel(7) feelfx(8) fest(1) fire(1) flick(17) fm(17) foo(27) future(17)
gab(10) gabba(4) gabbaloud(4) gabbalouder(4) glasstap(3) glitch(8)
glitch2(8) gretsch(24) gtr(3) h(7) hand(17) hardcore(12) hardkick(6)
haw(6) hc(6) hh(13) hh27(13) hit(6) hmm(1) ho(6) hoover(6) house(8)
ht(16) if(5) ifdrums(3) incoming(8) industrial(32) insect(3) invaders(18)
jazz(8) jungbass(20) jungle(13) juno(12) jvbass(13) kicklinn(1) koy(2)
kurt(7) latibro(8) led(1) less(4) lighter(33) linnhats(6) lt(16) made(7)
made2(1) mash(2) mash2(4) metal(10) miniyeah(4) monsterb(6) moog(7)
mouth(15) mp3(4) msg(9) mt(16) mute(28) newnotes(15) noise(1) noise2(8)
notes(15) num(21) numbers(9) oc(4) odx(15) off(1) outdoor(6) pad(3)
padlong(1) pebbles(1) perc(6) peri(15) pluck(17) popkick(10) print(11)
proc(2) procshort(8) psr(30) rave(8) rave2(4) ravemono(2) realclaps(4)
reverbkick(1) rm(2) rs(1) sax(22) sd(2) seawolf(3) sequential(8) sf(18)
sheffield(1) short(5) sid(12) simplesine(6) sitar(8) sn(52) space(18)
speakspell(12) speech(7) speechless(10) speedupdown(9) stab(23) stomp(10)
subroc3d(11) sugar(2) sundance(6) tabla(26) tabla2(46) tablex(3) tacscan(22)
tech(13) techno(7) tink(5) tok(4) toys(13) trump(11) ul(10) ulgab(5)
uxay(3) v(6) voodoo(5) wind(10) wobble(1) world(3) xmas(1) yeah(31)
```

> ¿Quieres añadir tus propios samples? Pon archivos `.wav` en `samples/` y
> regístralos con `samples({ minombre: ['carpeta/archivo.wav'] })` (función
> avanzada). El archivo `samples/strudel.json` es el índice oficial.

---

## 4 · Notas musicales

Para melodías usa `nota("...")`. Acepta **español** y **inglés**:

```python
nota("do re mi fa sol la si").sintetizador("sierra").out()
nota("c d e f g a b").sonido("seno").out()      # igual, en inglés
nota("do3 mi3 sol3 do4").out()                   # con octavas
nota("do#3 reb4 solb5").out()                    # sostenidos (#) y bemoles (b)
```

- **Notas**: `do re mi fa sol la si` (o `c d e f g a b`).
- **Sostenido**: `#` o `s` (ej. `do#3`, `fs4`). **Bemol**: `b` (ej. `reb4`, `bb2`).
- **Octavas**: del `-2` al `10`. Por defecto la 4 (`do3` = Do en octava 3,
  `do4` = Do central).
- También acepta **números MIDI** directos con `n("60 64 67")` (60 = Do4).

---

## Cómo elegir el "instrumento"

| Quieres… | Usa |
|----------|-----|
| Una onda sintetizada | `sonido("sierra")` o `sintetizador("sierra")` |
| Un sample de batería real | `sonido("808bd")` / `sonido("808bd:3")` |
| Un sample de instrumento | `sonido("bass1")` / `sonido("gtr")` |
| Una melodía con onda | `nota("do3").sintetizador("cuadrada")` |
| Una melodía sampleada | `nota("do3").sonido("bass1")` (suena al tono del sample) |

Y no olvides los **modificadores** (volumen, filtro, pan, reverb, etc.) que
están en la `09-referencia-completa.md`.

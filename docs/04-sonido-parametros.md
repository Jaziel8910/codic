# Sonido y parámetros

Aquí está la lista de todo lo que puedes usar para dar forma a tu sonido. Cada
mando se aplica con un punto: `sonido("bd").volumen(0.8)`.

Recuerda: los números van del **0 al 1** para cosas como volumen y paneo, y en
**hertz (Hz)** para frecuencias. No te preocupes por las unidades: juega con los
números hasta que suene bien.

---

## Sonidos de batería (samples)

Estos se usan con `sonido("…")` o `s("…")`:

`bd` (bombo), `sd` (caja), `rim` (borde), `sn` (snare), `lt` `mt` `ht`
(toms), `hh` (hi-hat cerrado), `oh` (hi-hat abierto), `cy`/`cym` (platillo),
`cp` (palmas), `cl` (clave), `cb`, `tom`, `perc`, `shaker`, `cowbell`, `kick`,
`hat`, `openhat`.

## Sintetizadores (ondas)

Para melodías usas `nota(…)` con un "tipo de onda":

- `seno` / `sine` — suave y redondo
- `sierra` / `saw` — brillante y áspero
- `triangulo` / `tri` — medio
- `cuadrada` / `square` — fuerte y chillonas
- `isaw`, `coseno` — variantes

```
nota("do3 re3 mi3").sintetizador("sierra")
```

O también puedes usar las ondas directamente como patrones de "señal": `seno`,
`sierra`, `triangulo`, `cuadrada` (suelen servir como moduladores, ver `lfo`).

---

## Parámetros de mezcla

| Mando | Qué hace | Rango |
|-------|----------|-------|
| `volumen(n)` / `gain` | fuerza del sonido | 0–1 |
| `paneo(n)` / `pan` | izquierda(−1) ↔ derecha(+1) | −1 a 1 |
| `velocidad` / `speed` | velocidad de reproducción del sample | >0 |
| `room` / `reverb` | eco/ambiente | 0–1 |
| `roomsize` | tamaño de la sala (eco) | 0–1 |
| `roomfade` | tiempo de desvanecido del eco | 0–1 |
| `delay` | eco corto | 0–1 |
| `delaytime` | tiempo del eco | segundos |
| `delayfeedback` | cuánto rebota el eco | 0–1 |
| `delaysync` | sincroniza el eco al tempo | 0–1 |
| `crush` | distorsión de "videojuego viejo" (bitcrush) | >0 |

Ejemplo:

```
sonido("bd sd").volumen(0.8).paneo(-0.3).room(0.4)
```

---

## Filtros (quitan o dejan frecuencias)

| Mando | Qué hace | Rango |
|-------|----------|-------|
| `filtro(n)` / `cutoff` | corte del filtro (brillo) | Hz |
| `resonancia` / `res` | énfasis del filtro | 0–1 |
| `lpf` | filtro paso-bajo | Hz |
| `hpf` | filtro paso-alto | Hz |
| `hpq` / `lpq` | "calidad" (resonancia) de esos filtros | 0–1 |
| `bpf` / `bpq` | filtro paso-banda | Hz |
| `djf` | filtro tipo DJ (barrido) | 0–1 |

```
sonido("hh*8").filtro(8000).resonancia(0.2)
```

---

## Distorsión y carácter

| Mando | Qué hace |
|-------|----------|
| `distorsion(n)` / `dist` / `drive` | distorsión |
| `distorttype` | tipo de distorsión |
| `vib` / `vibracion` | vibrato (temblor de tono) |
| `vibmod` | cantidad de vibrato |
| `vowel` / `vocal` | formante vocal (`a`, `e`, `i`, `o`, `u`) |
| `ribbon` | barrido de tono tipo "cinta" |
| `tremolo` | temblor de volumen |
| `tremolorate` / `tremolodepth` | velocidad/profundidad del tremolo |
| `phaser` / `phasercenter` / `phaserdepth` / `phasersweep` | efecto phaser |
| `chorus` | efecto chorus (grosor) |
| `comb` | resonancia tipo peine |
| `lfp` (lfo) | ver abajo |

---

## Envolvente (cómo ataca y se apaga el sonido)

| Mando | Qué hace |
|-------|----------|
| `ataque` / `attack` | cuánto tarda en aparecer |
| `sostenido` / `sustain` | cuánto se mantiene |
| `liberacion` / `release` | cuánto tarda en apagarse |
| `forma` / `shape` | forma de la onda de volumen |
| `decay` | decaimiento |
| `adsr(a, s, d, r)` | los cuatro de una vez |

```
sonido("bd").adsr(0.01, 0.1, 0.5, 0.2)
```

---

## Otros mandos útiles

| Mando | Qué hace |
|-------|----------|
| `frecuencia(n)` / `freq` | fija la frecuencia en Hz |
| `transponer(n)` / `ftranspose` | sube/baja el tono (semitonos) |
| `detune` | desafinación sutil (grosor) |
| `octava(n)` | sube/baja octavas enteras |
| `ancho_pan` / `panwidth` | ancho del paneo estéreo |
| `acelerar` / `accelerate` | acelera el sample |
| `loop` / `loopBegin` / `loopEnd` / `loopSample` | reproduce solo un trozo del sample |
| `cut` | recorte (vinyl cut) |
| `orbit` | envía a un "canal" de efectos |
| `canal` / `channel` | canal de salida |
| `etiqueta` / `label` / `tag` / `color` | marca para organizar |

---

## LFO (movimiento automático)

Un **LFO** hace que un mando suba y baje solo, en bucle, como un robot moviendo
el potenciómetro:

```
sonido("bd").lfo("pan", 1, 0.5)     # el paneo va de lado a lado 1 vez por ciclo
sonido("hh").lfo("cutoff", 2, 4000) # el filtro sube y baja
```

`lfo("mando", velocidad, cantidad)`.

---

## Modo "señal" de los mandos

Casi cualquier mando puede recibir una **señal** en vez de un número, para que
cambie con el tiempo. Por ejemplo `volumen(seno)` hace que el volumen suba y
baje con una onda. Esto es avanzado; juega con `lfo` primero.

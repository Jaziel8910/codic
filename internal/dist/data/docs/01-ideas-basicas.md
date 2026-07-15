# Ideas básicas (sin saber programar)

Aquí explicamos las cuatro palabras que usaremos todo el tiempo. No son
términos de computadora complicados: son ideas musicales sencillas.

---

## 1. Patrón (pattern)

Un **patrón** es simplemente "una idea musical que se repite". Por ejemplo:
"un kick en cada vuelta" es un patrón. "La nota do, luego re, luego mi" es otro
patrón.

En Codic, cuando escribes algo como `sonido("bd")`, obtienes un patrón: la idea
de "tocar un kick, una y otra vez".

Piénsalo como una cinta magnética en bucle: la música está grabada en la cinta
y se reproduce sin parar.

---

## 2. Sonido (sound)

Un **sonido** es el "instrumento" o "sample" que suena. Codic trae varios
sonidos de batería ya incluidos, por su nombre corto:

| Nombre | Qué es          |
|--------|-----------------|
| `bd`   | bombo (kick)    |
| `sd`   | caja (snare)    |
| `hh`   | hi-hat cerrado  |
| `oh`   | hi-hat abierto  |
| `cp`   | palmas (clap)   |
| `rim`  | golpe de borde  |
| `cy`   | platillo (cymbal)|

Y muchos más. También puedes usar sonidos con forma de onda (sintetizadores):
`seno` (sine), `sierra` (saw), `triangulo` (tri), `cuadrada` (square).

Para usarlos escribes: `sonido("bd")` o su forma corta `s("bd")`.

---

## 3. Nota (note)

Una **nota** es una altura musical: aguda o grave. Puedes escribirla con su
nombre (`do`, `re`, `mi`, `fa`, `sol`, `la`, `si`) y su octava, o con un número.

Ejemplos:

```
nota("do3")        # do en la octava 3
nota("re#4")       # re sostenido en la octava 4
nota(60)           # el número 60 también es una nota (do4 en el sistema MIDI)
```

No necesitas entender los números: puedes usar siempre los nombres (`do`, `re`…).
Cuantos más alto el número de octava (3, 4, 5…), más aguda suena.

---

## 4. Parámetro (parameter)

Un **parámetro** es un "mando" que cambia cómo suena el patrón. Son como los
potenciómetros de una consola de mezcla:

- **volumen** (`gain`): qué tan fuerte suena.
- **paneo** (`pan`): si suena más a la izquierda o a la derecha.
- **filtro** (`cutoff`): si suena graves (oscuro) o agudos (brillante).
- **reverberación** (`room`): si suena en un cuarto pequeño o en una catedral.

Para aplicar un parámetro usas un punto (`.`) y su nombre:

```
sonido("bd").volumen(0.8)
```

Se lee: "sonido bd, con volumen 0.8".

---

## Cómo se juntan

La receta siempre es:

```
[un patrón] . [un parámetro] (opcional) .out()
```

Por ejemplo:

```
mi_beat = sonido("bd sd hh cp").volumen(0.7)
mi_beat.out()
```

Eso es Codic en una frase: **creas un patrón, le das forma con parámetros, y lo
haces sonar con `.out()`.**

---

## El punto (`.`) es tu amigo

El punto `.` significa "y además aplícale esto". Puedes encadenar varios:

```
sonido("bd sd").volumen(0.8).paneo(-0.5).filtro(1200).room(0.3)
```

Cada `.` agrega una capa de sonido. No hay orden obligatorio; ponlos en el
orden que quieras.

# Mini-notación (el atajo de texto)

Dentro de las comillas `sonido("…")` y `nota("…")` puedes escribir ritmos
rapidísimo usando símbolos cortos. A esto lo llamamos **mini-notación**. Es como
taquigrafía musical.

---

## El espacio = en orden

Cada palabra separada por espacio suena en orden, una tras otra, dentro del ciclo:

```
sonido("bd sd hh cp")     # kick, luego snare, luego hihat, luego clap
nota("do re mi sol")      # do, luego re, luego mi, luego sol
```

---

## La coma (`,`) = al mismo tiempo

Las palabras separadas por coma suenan **a la vez** (apiladas):

```
sonido("bd, hh")          # kick y hihat sonando juntos
```

---

## Los corchetes `[ ]` = grupo

Agrupan varias cosas para tratarlas como una sola unidad. Aquí dentro la coma
también significa "a la vez":

```
sonido("[bd hh] sd")      # (kick+hihat) y luego snare
```

---

## Las llaves `{ }` = polímetro

Toca varias secuencias a la vez pero con compases distintos, de modo que se
cruzan de forma interesante:

```
sonido("{bd sd, hh hh hh}")   # dos ritmos que chocan entre sí
```

Puedes añadir `%` para fijar cuántas subdivisiones:

```
sonido("{bd sd, hh*3}%4")
```

---

## Los signos `< >` = alternar

Cada ciclo (vuelta) suena una cosa distinta, rotando:

```
sonido("<bd sd hh cp>")   # ciclo 1: bd, ciclo 2: sd, ciclo 3: hh, ciclo 4: cp, y vuelve a bd
```

---

## El asterisco `*` = repetir (más rápido)

```
sonido("bd*4")            # el kick suena 4 veces en el ciclo (más rápido)
sonido("hh*8")            # hihat 8 veces
```

---

## La barra `/` = alargar (más lento)

```
sonido("bd/2")            # el kick suena la mitad de rápido (más espaciado)
```

---

## El signo `?` = azar

Pone una probabilidad de 50% de que suene o no:

```
sonido("bd? sd? hh?")     # cada golpe puede aparecer o no, al azar
```

---

## El signo `:` = número de nota

Para sonidos de batería puedes fijar un "tono" con dos puntos y un número:

```
sonido("bd:3")            # el sonido bd suena afinado a la nota 3
```

---

## Paréntesis `(n, m)` = ritmo euclídeo

Crea patrones rítmicos repartidos de forma pareja (como el taladro de un neumático):

```
sonido("bd(3,8)")         # 3 golpes repartidos en 8, tipo "10010010"
sonido("sd(5,8)")         # 5 golpes repartidos en 8
```

Puedes añadir un tercer número para rotar el patrón: `bd(3,8,1)`.

---

## El símbolo `~` = silencio

Deja un hueco en blanco:

```
sonido("bd ~ sd")         # kick, silencio, snare
```

---

## Ejemplo combinado

```
ritmo = sonido("bd*4 [sd hh] hh*2 ~ cp?")
ritmo.out()
```

Tómate un momento en leerlo símbolo por símbolo. Esa sola línea dice: "kick 4
veces, luego snare+hihat juntos, luego hihat 2 veces, un silencio, y un clap
que a veces aparece". En un compás. Se repite para siempre.

# Cómo escribir en Codang

Codang es un lenguaje pequeño y amable. Solo necesitas recordar tres cosas:
**asignar un nombre**, **encadenar con punto**, y **hacer sonar con `.out()`**.

---

## Asignar un nombre (=)

Cuando quieres guardar una idea musical para usarla después, le pones un nombre
con el signo igual (`=`). No escribes `:=` ni otras cosas: solo `=`.

```
mi_beat = sonido("bd sd hh cp")
```

Ahora `mi_beat` es esa idea. Puedes usarla más abajo.

> Nota: el nombre puede tener letras, números y guion bajo (`_`). Mejor sin
> espacios. Ejemplos buenos: `bateria`, `melodia_1`, `bajo_feo`.

---

## Encadenar con punto (.)

Para cambiar cómo suena, escribes un punto y el nombre del mando:

```
mi_beat = sonido("bd sd hh cp").volumen(0.7).room(0.2)
```

Cada punto agrega un cambio. Es como apilar capas de pintura.

---

## Hacer sonar (.out())

Para que una idea suene de verdad, le pones `.out()` al final:

```
sonido("bd sd hh cp").out()
```

Si no escribes `.out()`, Codic solo "recuerda" la idea pero no la reproduce.
(Puedes tener muchas ideas guardadas y hacer sonar solo algunas.)

---

## Funciones (func …)

A veces quieres una "receta" que repites con distintos valores. En Codang eso es
una **función**. Se escribe con `func`, el nombre, los ingredientes entre
paréntesis, y dos puntos `:`.

```
func mas_rapido(x):
    x.rapido(2)
```

Se lee: "la función `mas_rapido` recibe algo llamado `x` y lo devuelve el doble
de rápido". El `x.rapido(2)` es lo que hace.

Para usarla:

```
bateria = sonido("bd sd hh cp")
bateria.mas_rapido().out()
```

Las funciones sirven para no repetirte. También las usa Codic para los efectos
aleatorios (ver `06-combinar-y-vivir.md`).

---

## Comentarios

Puedes escribir notas para ti mismo que Codic ignora, usando `#`:

```
# Este es mi beat de prueba
sonido("bd sd").volumen(0.8)   # fuerte pero no tanto
```

---

## Un programa completo

Un "programa" en Codic es solo una lista de frases, de arriba hacia abajo.
Ejemplo de un archivo `cancion.cdc`:

```
# Mi primera canción
bateria = sonido("bd sd hh cp").volumen(0.7)
bajo    = nota("do2 re2 mi2 sol2").volumen(0.5)
melodia = nota("do5 mi5 sol5").volumen(0.4).sintetizador("sierra")

# Toco todo junto
pila(bateria, bajo, melodia).out()
```

Eso es todo lo que necesitas saber de "sintaxis". El resto de la guía es qué
palabras (sonidos, notas, mandos) puedes usar.

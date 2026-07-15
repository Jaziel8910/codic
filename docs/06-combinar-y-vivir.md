# Combinar patrones y darles vida

Aquí aprendes a juntar varias ideas para que suenen a la vez, y a hacer que la
música "respire" y cambie sola.

---

## Tocar varias cosas a la vez

### pila (stack)

La forma más simple: pones varios patrones juntos y suenan al mismo tiempo.

```
bateria = sonido("bd sd hh cp")
bajo    = nota("do2 re2 mi2 sol2")
melodia = nota("do5 mi5 sol5")

pila(bateria, bajo, melodia).out()
```

Todo suena junto. Cada patrón conserva su propio ritmo.

### secuencia (cat / fastcat)

Toca unos después de otros, dentro del ciclo:

```
secuencia(sonido("bd"), sonido("sd"), sonido("hh"))
```

### secuenciaLenta (slowcat)

Cada patrón ocupa un ciclo entero, uno tras otro:

```
secuenciaLenta(sonido("bd*4"), sonido("hh*8"))
```

### polímetro / polirritmia / timecat

Para cruzar ritmos de distinta longitud (avanzado). Ejemplo:

```
polirritmia(sonido("bd*3"), sonido("hh*4"))
```

---

## Separar canales (jux)

`jux` (de "juxtapose") manda una copia a la izquierda y otra a la derecha, y
puedes transformar una de ellas:

```
sonido("bd*4 hh*4").jux(func(x): x.rapido(2))
```

Esto pone el ritmo normal a la izquierda y el doble de rápido a la derecha.

Variantes: `juxBy(pan, f)` (cantidad de separación), `juxFlip(f)`, `juxFlipBy(pan, f)`.

---

## Capas y superposición

Ya viste `capas` y `superponer` en el doc anterior. Recordatorio rápido:

```
func oct(x): x.transponer(12)
sonido("do3 mi3 sol3").capas("oct")        # melodía + 1 octava arriba
sonido("bd").superponer("oct")             # bd + bd agudo
```

---

## Hacer que cambie solo (aleatorio)

Esto es lo que hace que Codic suene "vivo", como un músico improvisando:

```
func Doble(x): x.rapido(2)

base = sonido("bd*4 hh*8")
base.a veces("Doble").out()        # a veces el beat se acelera
```

Otras palabras mágicas: `a menudo`, `rara vez`, `casi siempre`, `casi nunca`,
`cada(n, f)`, `algunosCiclos(f)`.

Todas reciben el nombre de una función **entre comillas**.

---

## inside / dentro / fuera

Aplicas una transformación solo a una porción del tiempo:

```
func eco(x): x.delay(0.3)
sonido("bd*4").dentro(0, 0.5, "eco")   # eco solo en la primera mitad
```

---

## Ejemplo "vivo" completo

```
# Batería que a veces se acelera
func Rapido(x): x.rapido(2)
bateria = sonido("bd sd hh cp").volumen(0.7)
bateria.a menudo("Rapido").out()

# Bajo que cambia de nota al azar
bajo = nota(elegir("do2", "re2", "mi2", "sol2")).volumen(0.5)
bajo.out()

# Melodía con paneo automático
melodia = nota("do5 re5 mi5 sol5").volumen(0.35).lfo("pan", 1, 0.5)
melodia.out()
```

Cada `.out()` hace sonar esa idea. Como son patrones distintos, suenan juntos,
pero cada uno tiene su propia vida.

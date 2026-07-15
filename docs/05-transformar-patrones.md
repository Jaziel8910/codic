# Transformar patrones

Estos mandos cambian la **estructura** de tu patrón: su velocidad, su orden, su
repetición. Se aplican con punto, igual que los parámetros.

---

## Velocidad y tiempo

| Mando | Qué hace |
|-------|----------|
| `rapido(n)` / `fast` | suena n veces más rápido |
| `lento(n)` / `slow` | suena n veces más lento |
| `adelantar(n)` / `early` | corre el patrón hacia el inicio |
| `atrasar(n)` / `late` | corre el patrón hacia el final |
| `iter(n)` | desplaza el patrón ciclo a ciclo |
| `iterBack(n)` | desplaza hacia atrás |

```
sonido("bd sd").rapido(2)     # el beat va el doble de rápido
sonido("bd sd").lento(2)      # la mitad de rápido
```

---

## Orden

| Mando | Qué hace |
|-------|----------|
| `revertir()` / `rev` | toca al revés |
| `palindromo()` / `palindrome` | ida y vuelta (abc → abcba) |
| `brak()` | efecto "breakbeat" (ralentiza y acelera) |
| `mezclar()` / `shuffle` / `scramble` | desordena al azar |

```
sonido("bd sd hh cp").revertir()
```

---

## Recorte y zoom

| Mando | Qué hace |
|-------|----------|
| `zoom(a, b)` | usa solo la parte del ciclo entre a y b (0–1) |
| `estirar(n)` / `stretch` | estira el patrón |
| `contraer(n)` / `contract` | lo aprieta |
| `expandir(n)` / `expand` | lo expande |
| `encajar(n)` / `fit` | hace que quepa en n ciclos |
| `doblar(n)` / `fold` | repite dentro del mismo espacio |
| `encoger(n)` / `shrink` | reduce longitud |

```
sonido("bd sd hh cp").zoom(0, 0.5)   # solo la primera mitad del ciclo
```

---

## Dentro de una parte

| Mando | Qué hace |
|-------|----------|
| `dentro(a, b, f)` / `within` | aplica una función solo entre a y b |
| `fuera(a, b, f)` / `outside` | aplica una función fuera de ese rango |
| `adentro(f)` / `inside` | aplica f a cada sub-parte |
| `hacia(f)` / `into` | mete el patrón dentro de f |
| `ritmo(f)` / `pace` | cambia la velocidad localmente |

```
func doble(x): x.rapido(2)
sonido("bd*4").dentro(0, 0.5, "doble")   # la primera mitad va doble rápido
```

---

## Repetición y capas

| Mando | Qué hace |
|-------|----------|
| `repetirCiclos(n)` / `repeatCycles` | repite el patrón n ciclos |
| `capas(f1, f2, …)` / `layer` | aplica varias funciones y las apila |
| `superponer(f1, f2, …)` / `superimpose` | original + versiones encima |
| `golpear(n)` / `ply` | n golpes por nota |
| `golpearCon(n, f)` / `plyWith` | n golpes procesados por f |
| `prensar(n)` / `press` | comprime n en uno |

```
func agudo(x): x.transponer(12)
sonido("bd").capas("agudo")         # el bd original + una versión 1 octava arriba
```

---

## Swing y groove

| Mando | Qué hace |
|-------|----------|
| `swing(n)` / `swing` | desfase ritmico (groove) |
| `swingBy(n, f)` | swing solo donde f diga |

```
sonido("bd*8").swing(0.25)
```

---

## Trocear (chunk)

Divide el ciclo en n pedazos y aplica una función a cada uno:

| Mando | Qué hace |
|-------|----------|
| `trozo(n, f)` / `chunk` | aplica f al trozo n-ésimo |
| `trozoLento(n, f)` / `slowChunk` | igual pero lento |
| `trozoRapido(n, f)` / `fastChunk` | igual pero rápido |
| `trozoEn(n, f)` / `chunkInto` | reparte el ciclo en n y mete f |

---

## Ancla y recorrido

| Mando | Qué hace |
|-------|----------|
| `ancla(n, f)` / `anchor` | ancla el inicio y aplica f |
| `fancla` / `fanchor` | ancla y espeja |
| `pancla` / `panchor` | ancla y panea |
| `recorrido(f)` / `tour` | recorre varias transformaciones |

---

## Espacio y fuerza

| Mando | Qué hace |
|-------|----------|
| `crecer(n)` / `grow` | crece progresivamente |
| `apurar(n)` / `hurry` | acelera |
| `mantener(n)` / `keep` | mantiene n eventos |
| `soltar(n)` / `drop` | quita n eventos |
| `linger` | alarga el final |
| `tomar(n)` / `take` | toma n eventos |
| `esparcir(n)` / `spread` | esparce en n |
| `exprimir(f)` / `squeeze` | exprime según otro patrón |
| `como(n)` / `as` | convierte el valor a otro tipo |
| `comoNumero()` / `asNumber` | fuerza a número |

---

## Ritmos (Euclídeos y más)

| Mando | Qué hace |
|-------|----------|
| `euclid(pasos, golpes, rot)` | reparte golpes en pasos |
| `euclidRot(…)` | con rotación |
| `euclidLegato(…)` | con notas sostenidas |
| `euclidish(grupos, golpes)` | grupos de distinto tamaño |

```
sonido("bd").euclid(8, 3, 0)    # 3 golpes en 8: 10010010
sonido("sd").euclid(8, 5, 0)    # 5 golpes en 8
```

También como función suelta: `euclid(8,3)`, `euclidish([3,2], 5)`.

---

## Azar y vida

| Mando | Qué hace |
|-------|----------|
| `degradar()` / `degrade` | quita notas al azar (50%) |
| `degradarPor(p)` / `degradeBy` | quita con probabilidad p |
| `a veces(f)` / `sometimes` | a veces aplica f |
| `a menudo(f)` / `often` | frecuentemente |
| `rara vez(f)` / `rarely` | pocas veces |
| `casi siempre(f)` / `almostAlways` | casi siempre |
| `casi nunca(f)` / `almostNever` | casi nunca |
| `siempre(f)` / `always` | siempre |
| `nunca(f)` / `never` | nunca |
| `algunosCiclos(f)` / `someCycles` | f en algunos ciclos |
| `cada(n, f)` / `every` | f cada n ciclos |

```
func doble(x): x.rapido(2)
sonido("bd*4").a veces("doble")   # a veces va doble rápido
sonido("bd*8").cada(4, "doble")   # cada 4 ciclos va doble
```

> Las funciones se pasan **por su nombre** entre comillas: `"doble"`.

---

## Selección aleatoria

| Función suelta | Qué hace |
|----------------|----------|
| `elegir(a, b, c)` / `choose` | elige uno al azar |
| `elegirCiclos(…)` / `chooseCycles` | elige uno por ciclo |
| `elegirPesos([p, a], [q, b])` / `wchoose` | elige con pesos |
| `corrido(n)` / `run` | cuenta del 0 a n−1 |
| `irand(n)` | entero al azar 0..n−1 |
| `rand2()` | número al azar −1..1 |
| `randL()` | al azar 0..1 (logarítmico) |
| `rangoX(a, b)` / `rangex` | rango exponencial |

```
sonido(elegir("bd", "sd", "hh"))      # cada ciclo un sonido distinto
nota(corrido(8)).escala("mayor")      # escala ascendente
```

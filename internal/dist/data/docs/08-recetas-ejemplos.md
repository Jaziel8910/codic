# Recetas y ejemplos listos

Copia, pega y cambia lo que quieras. Cada bloque es un archivo `.cdc` completo.

---

## Beat de batería básico

```
sonido("bd sd hh cp").volumen(0.8).out()
```

## Beat con groove (swing)

```
sonido("bd*4 [sd hh] hh*2").swing(0.25).volumen(0.8).out()
```

## Hi-hats rápidos con filtro

```
sonido("hh*16").filtro(9000).resonancia(0.3).volumen(0.4).out()
```

## Bombo con ritmo euclídeo

```
sonido("bd").euclid(8, 3, 0).volumen(0.9).out()
sonido("sd").euclid(8, 5, 0).volumen(0.7).out()
```

## Melodía suave con sintetizador

```
nota("do5 re5 mi5 sol5 do6").sintetizador("sierra").volumen(0.4).out()
```

## Escala ascendente

```
nota(corrido(8)).escala("mayor").volumen(0.5).out()
```

## Bajo con paneo automático

```
nota("do2 re2 mi2 sol2").volumen(0.6).lfo("pan", 0.5, 0.8).out()
```

## Dos cosas a la vez (stack)

```
pila(
  sonido("bd sd hh cp").volumen(0.7),
  nota("do3 mi3 sol3").volumen(0.4).sintetizador("triangulo")
).out()
```

## Canción que cambia sola

```
func Doble(x): x.rapido(2)
func Eco(x): x.delay(0.25)

pila(
  sonido("bd*4 hh*8").a menudo("Doble"),
  nota(elegir("do3","re3","mi3","sol3")).cada(2, "Eco")
).out()
```

## Ambiente con mucho eco

```
sonido("hh*4").room(0.8).delay(0.4).delayfeedback(0.5).volumen(0.3).out()
```

## Voz/palmas con efecto vocal

```
sonido("cp*4").vocal("a").volumen(0.6).out()
```

## Beat con distorsión

```
sonido("bd sd").distorsion(0.3).volumen(0.8).out()
```

## Canción completa para exportar

```
album("Mi Cancion")
bpm(120)

stem("bateria", "stems/bateria.wav", 0.9)
track("bajo", nota("do2 re2 mi2 sol2").volumen(0.6))
track("lead", nota("do5 mi5 sol5").volumen(0.4))

export("mi_cancion.dawproject")
```

---

## Consejos para empezar

1. Empieza solo con `sonido("bd sd hh cp").out()` y escúchalo.
2. Agrega un `.volumen(0.8)` y luego un `.room(0.3)`.
3. Cambia los nombres: `bd`, `sd`, `hh`, `cp`, `oh`, `cy`…
4. Prueba `rapido(2)`, `revertir()`, `swing(0.25)`.
5. Cuando suene bien, envuélvelo en `pila(...)` con una melodía.
6. Por último, haz un `album(...)` + `export(...)` para llevártelo a tu DAW.

No hay forma de "romper" nada: si suena raro, cambia un número y vuelve a
intentar.

# Bienvenida a Codic

**Codic** es un estudio de música que vive dentro de la terminal (esa pantalla
negra con letras). No necesitas saber tocar un instrumento, ni saber qué es la
programación. Solo escribes frases sencillas y Codic las convierte en música que
suena en bucle, para siempre, hasta que lo detengas.

El lenguaje en el que escribes se llama **Codang** (se pronuncia "códang").
Está diseñado para que cualquier persona, sin experiencia técnica, pueda
componer beats, melodías y canciones completas.

---

## ¿Qué puedes hacer con Codic?

- Crear ritmos de batería (kick, snare, hi-hat…) que suenan en bucle.
- Hacer melodías con notas musicales (`do`, `re`, `mi`… o números).
- Mezclar varios ritmos y melodías a la vez.
- Alterar la música al azar, para que nunca suene igual dos veces.
- Armar un **álbum** con varias pistas y exportarlo a un programa de música
  profesional (un "DAW" como Bitwig, Reaper o Ableton).

---

## Tu primera canción

Escribe esto en un archivo llamado `miprimera.cdc`:

```
pista = sonido("bd sd hh cp")
pista.out()
```

Eso es todo. Significa: "toca, en orden, un kick (`bd`), un snare (`sd`),
un hi-hat (`hh`) y un clap (`cp`), y que suene". Cuando lo reproduces,
escucharás un beat de batería básico que se repite.

Para escucharlo, desde la terminal:

```
codic file miprimera.cdc
```

---

## Cómo leer esta documentación

Está pensada para alguien que **nunca ha programado**. Por eso explicamos cada
idea desde cero, con ejemplos que puedes copiar y pegar.

Archivos de esta guía:

1. `01-ideas-basicas.md` — las ideas que necesitas (patrón, sonido, nota, parámetro).
2. `02-escribir-codang.md` — cómo se escribe el lenguaje.
3. `03-mini-notacion.md` — el atajo de texto para escribir ritmos rápido.
4. `04-sonido-parametros.md` — cómo cambiar el volumen, el paneo, el filtro, etc.
5. `05-transformar-patrones.md` — acelerar, invertir, repetir, al azar…
6. `06-combinar-y-vivir.md` — juntar varios patrones y darles vida.
7. `07-albumes-stems-exportar.md` — álbumes, pistas de terceros y exportar.
8. `08-recetas-ejemplos.md` — recetas listas para usar.
9. `09-referencia-completa.md` — la lista de todo lo que existe.

---

## Una nota sobre "bucle" (loop)

En Codic casi todo sucede dentro de un **ciclo** (loop). Imagina un ciclo como
una vuelta de carrusel: cuando termina, vuelve a empezar, una y otra vez.
Por defecto un ciclo dura un compás (4 golpes). Tú decides qué pasa en esa
vuelta, y Codic la repite para siempre.

No tienes que pensar en "detener" la música; Codic la mantiene sonando en bucle
hasta que cierras el programa.

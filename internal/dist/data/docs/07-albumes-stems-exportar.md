# Álbumes, stems de terceros y exportar

Hasta ahora componías una idea y la hacías sonar. Codic también puede armar un
**álbum** (una canción con varias pistas) y **exportarlo** a un programa de
música profesional mediante el formato abierto **DAWproject**.

Esto sirve para: seguir retocando tu canción en Bitwig, Reaper, Ableton, etc.;
mezclar con stems (pistas de audio) que bajaste de internet; o entregar tu
trabajo a otro músico.

---

## El álbum

Empiezas nombrando tu álbum/canción:

```
album("Mi Primer Album")
```

Eso es el título que llevará el archivo exportado.

---

## Stems de terceros

Un **stem** es un archivo de audio (`.wav`, `.mp3`) que ya tienes: una batería
que descargaste, una voz grabada, un loop de alguien más. Lo agregas así:

```
stem("bateria", "stems/bateria.wav")            # ruta al archivo
stem("voz", "stems/voz.wav", 0.9)               # con volumen 0.9
stem("bajo", "stems/bajo.wav", 0.8, -0.3)       # volumen 0.8 y paneo a la izquierda
```

El primer texto es el **nombre** de la pista (para ti). El segundo es la
**ruta** al archivo en tu computadora. Los números opcionales son volumen y
paneo.

> Consejo: pon tus archivos de audio en una carpeta `stems/` junto a tu
> archivo `.cdc` para no perderte.

---

## Pistas desde tus patrones

Además de stems, puedes convertir cualquier patrón de Codic en una pista del
álbum:

```
track("bajo", nota("do2 re2 mi2 sol2").volumen(0.5))
track("lead", sonido("bd*4").nota(36), 0.7)     # volumen 0.7
```

El primer texto es el nombre; el segundo es tu patrón (lo que ya sabes hacer).
Opcionalmente, volumen y paneo.

Al exportar, los patrones se convierten en **notas MIDI** (si son notas) o se
guardan como la información necesaria para que el DAW las reproduzca.

---

## Exportar a DAWproject

Cuando ya tienes tu álbum armado, lo escribes a un archivo:

```
export("mi_album.dawproject")
```

o

```
export_dawproject("mi_album.dawproject")
```

Eso crea un archivo `.dawproject` (formato abierto, compatible con Bitwig,
Reaper y otros). Ábrelo en tu DAW y ahí tendrás tus pistas: los stems como
audio, y tus patrones como clips de notas.

Si no pones nombre de archivo, Codic usa el nombre del álbum:

```
album("Mi Album Cool")
# ... pistas ...
export()                  # crea "mi_album_cool.dawproject"
```

---

## Desde la terminal (sin abrir el script)

También puedes exportar un archivo `.cdc` directamente:

```
codic export cancion.cdc
codic export cancion.cdc salida.dawproject
```

Codic lee el archivo, arma el álbum y escribe el `.dawproject`. No suena nada;
solo guarda el archivo.

---

## Ejemplo completo de álbum

Archivo `album.cdc`:

```
album("Cancion de Prueba")
bpm(110)

# Pista de batería que bajé de internet
stem("bateria", "stems/bateria.wav", 0.9)

# Mi bajo en Codang
track("bajo", nota("do2 re2 mi2 sol2").volumen(0.6))

# Mi melodía
func Oct(x): x.transponer(12)
track("lead", nota("do5 mi5 sol5").volumen(0.4).capas("Oct"))

# Exportar
export("cancion.dawproject")
```

Al ejecutar `codic export album.cdc` obtienes `cancion.dawproject` listo para
abrir en tu DAW.

---

## Resumen de las funciones nuevas

| Función | Qué hace |
|---------|----------|
| `album("nombre")` | título del álbum/proyecto |
| `bpm(n)` | tempo (latidos por minuto) |
| `stem("nombre", "ruta", volumen?, pan?)` | pista de audio de terceros |
| `track("nombre", patron, volumen?, pan?)` | pista desde un patrón de Codic |
| `export("archivo.dawproject")` | exporta el álbum |
| `export_dawproject("archivo.dawproject")` | igual que `export` |

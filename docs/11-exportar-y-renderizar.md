# Exportar y renderizar tu canción

Codic puede convertir tus canciones de Codang en archivos que cualquiera puede escuchar, sin necesidad de abrir el programa.

## Renderizar a WAV (audio)

El comando `render` toma tu archivo `.cdc`, lo toca en "cámara lenta" internamente y guarda el sonido resultante en un archivo `.wav`. Es ideal para compartir tu tema por WhatsApp, subirlo a SoundCloud o editarlo en otro programa.

```bash
codic render micancion.cdc salida.wav 8
```

- `micancion.cdc` — tu archivo de Codang.
- `salida.wav` — nombre del archivo de audio que se creará (puedes ponerle otra ruta).
- `8` — segundos que durará el audio (opcional, por defecto 8).

El audio se genera a 44100 Hz en estéreo, con todos los sonidos, efectos, samples y notas que hayas usado. El tempo (`cps`) se respeta tal como lo configuraste en el código.

Ejemplo de `micancion.cdc`:

```python
cps(0.5)
bateria = sound("808bd*2 909sd 808hc")
melodia = note("c e g b").s("sine").cutoff(800)
bateria.out()
melodia.out()
```

Después de ejecutar el comando tendrás `salida.wav` listo para reproducir en cualquier reproductor.

> Consejo: si tu canción es más larga, aumenta los segundos. El render recorre la canción ciclo a ciclo, así que todo lo que suene dentro de ese tiempo aparecerá en el archivo.

## Exportar a DAWProject (para ABL Live / Bitwig)

Si quieres abrir tu canción en un programa de edición profesional (como ABL Live o Bitwig Studio), usa el comando `export`:

```bash
codic export micancion.cdc album.dawproject
```

Esto crea un proyecto con una pista por cada sección de tu canción, conservando los sonidos y efectos como clips. Ábrelo en tu DAW para mezclar, masterizar o añadir más instrumentos.

## Diferencias rápidas

| Comando | Resultado | Para qué sirve |
| --- | --- | --- |
| `render` | archivo `.wav` | escuchar y compartir el audio terminado |
| `export` | archivo `.dawproject` | seguir editando en un programa profesional |

Ambos usan el mismo motor de sonido de Codic, así que lo que oyes en vivo es lo que se guarda.

# Codic standard library — scales
#
# Scales are selected by name in Codang via .constrain("minor")
# and .scale("dorian"). This file documents the built-in scale
# vocabulary. Import support (Fase 5.3) will let you `import "stdlib/scales"`.
#
# Available scale names:
#   major / ionian     0 2 4 5 7 9 11
#   minor / aeolian   0 2 3 5 7 8 10
#   dorian             0 2 3 5 7 9 10
#   phrygian          0 1 3 5 7 8 10
#   lydian            0 2 4 6 7 9 11
#   mixolydian        0 2 4 5 7 9 10
#   locrian           0 1 3 5 6 8 10
#   pentatonic        0 2 4 7 9
#   minorpenta        0 3 5 7 10
#   blues            0 3 5 6 7 10
#   chromatic        0 1 2 3 4 5 6 7 8 9 10 11
#
# Example:
#   note("c3 e3 g3").constrain("minor", "nearest").out()

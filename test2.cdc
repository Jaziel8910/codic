# @title Mi Cancion
# @bpm: 120

beat = s("bd sd hh cp").fast(2)
melody = note("c3 d3 e3 g3").s("sawtooth").gain(0.3)
melody = melody.cutoff(sine.slow(4).range(200, 2000))

beat.out()
melody.out()
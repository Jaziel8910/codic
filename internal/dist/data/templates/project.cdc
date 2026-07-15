# @bpm 120
# @key "c minor"
# Template for a new Codic project track.
# Add it to a project with:  codic project add my_track.cdc

melody = note("c3 d3 e3 g3").s("sawtooth").cutoff(1200).gain(0.5).out()
beat   = s("bd sd hh cp").fast(2).gain(0.8).out()

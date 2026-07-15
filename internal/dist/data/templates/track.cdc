# @bpm 120
# Template for a new Codic track.
# Render with:  codic render my_track.cdc out.wav -d 8
# Play with:     codic play my_track.cdc

kick = s("bd").euclid(8, 4, 0).out()
hat  = s("hh").euclid(8, 3, 1).speed(2).gain(0.4).out()
bass = note("c2 e2 g2 a2").s("sawtooth").cutoff(600).gain(0.3).out()

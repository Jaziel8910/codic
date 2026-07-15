# @bpm 128
# A small example song shipped with Codic.

kick = s("bd").euclid(16, 4, 0).gain(0.9).out()
hat  = s("hh").euclid(16, 6, 2).speed(2).gain(0.35).out()
bass = note("c2 c2 g2 bb2").s("sawtooth").cutoff(500).gain(0.4).out()
lead = note("c4 eb4 g4").s("square").cutoff(2000).gain(0.25).out()

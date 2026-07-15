# Codic standard library — drum kit reference
#
# Default embedded sample kit (auto-installed by `codic install`).
# Trigger with s("name"):
#   bd   kick drum        sd   snare
#   hh   closed hat       ho   open hat
#   cp   clap            cb   cowbell
#   lt   low tom         mt   mid tom
#   808  TR-808 sub       rs   rimshot
#   ma   maraca          cl   clave
#
# Example:
#   stack(s("bd").euclid(8,4,0), s("hh*2"), s("sd").every(4, "rev")).out()

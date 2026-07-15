# Codic standard library — rhythms
#
# Reusable Euclidean / mini-notation rhythm snippets.
# Copy these into your own .cdc files.

# Four-on-the-floor kick
kick = s("bd").euclid(8, 4, 0)

# Offbeat hats
hats = s("hh").euclid(8, 3, 1).speed(2)

# Broken beat
broken = s("bd sd ~ hh sd ~ hh")

# A simple 16-step techno groove
techno = cat(
  s("bd ~ ~ ~ bd ~ ~ ~"),
  s("~ ~ hh ~ ~ ~ hh ~"),
  s("~ sd ~ ~ ~ sd ~ ~"),
)

# Arpeggiated minor chord
arp = note("c3 e3 g3 c4").arpeggiate("up").slow(2)

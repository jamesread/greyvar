# world-generator

Generates a Tiled world (`<name>.world` plus one layered `row.col.tmj` per map)
from Perlin noise. Tiles are chosen from a corner wangset in a `.tsx` tileset,
so no texture names are hard-coded: terrains are wangset color names.

```bash
make
./world-generator -name myIsle -world-rows 2 -world-cols 3 \
  -entity pressureButton:sand:3 -entity pot:grass:2
GREYVAR_WORLD=myIsle_tiled ../server/greyvar-server
```

How it works:

- `-terrains water,sand,grass` stacks terrains bottom to top. The first fills the
  base layer; each later one gets its own layer drawn with the wangset's
  transparent ("no terrain") transition tiles, like the hand-made maps.
- Noise is sampled per tile corner across the whole world, so adjacent maps
  share edges and scroll transitions line up.
- Corner patterns the tileset has no tile for (usually the two diagonals) are
  filled in automatically; tiles that are blank in the tileset image are
  skipped and reported.
- Flowers are not terrain variants. Tiles marked `decor=true` in
  `-decor-tileset` (default: the generated `flora.tsx`) are scattered on a
  `decor` layer over the spawn terrain (or `-decor-terrain`), each free cell
  with probability `-decor-density` (default 0.15). They animate and never collide.
- A `spawn` marker is placed on the centre map, on `-spawn-terrain` (default:
  the top terrain).

Run `./world-generator -h` for all flags.

## Puzzles

`world-generator puzzle` writes a one-map puzzle world: two to four stone
rooms in a chain, the player starting on an entry teleporter and rift coins
waiting by the exit teleporter, each door (a stone barrier) locked by levers, a key hidden in a pot for a
chest, push blocks on pressure plates, or a timed switch (alone or combined).
The layout comes from `pkg/puzzlegen`, which the server can also call.

```sh
./world-generator puzzle -level 4 -seed 7 -preview     # print it and its solution
./world-generator puzzle -level 9 -seed 3               # write ../server/dat/worlds/puzzle_9_3_tiled
./world-generator puzzle -level 6 -mechanics lever,keyChest
```

- `-level` follows a difficulty curve: levels 1-8 are hand-tuned; 9 and up
  pair mechanics per door from the seed, with longer lever combinations, two
  blocks, keys hidden in earlier rooms, more decoys and tighter timers.
- Every layout is checked by a solver before it is written. It must have no
  shortcut (no door open at the start, no way round a door), be solvable,
  and need at least one action per door; otherwise the next layout is tried.
- The solver works on whole moves, not footsteps: it sets each door's levers
  in one go, breaks only pots that hide keys, and treats a timed switch as a
  run that must reach the far side of its door within the timer. Layouts are
  verified by pushing alone (a smaller search); `Solve` also plans pulls
  (the use button), e.g. to free a block pushed into a corner.
- Push blocks (`pushBlock`), plates held by blocks and timed switches
  (`reset_after`) are described in `server/dat/worlds/PUZZLES.md`.
- Load a written world in game with the console: `world-load puzzle_9_3_tiled`
  (and `world-load isleOfStarting_dev_tiled` to go back). In game you rarely
  need to: the server opens puzzles itself as rifts (the gold rift portal, or
  the console `rift` command); see `server/dat/worlds/PUZZLES.md`.

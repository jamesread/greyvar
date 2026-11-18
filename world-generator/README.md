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
- A `spawn` marker is placed on the centre map, on `-spawn-terrain` (default:
  the top terrain).

Run `./world-generator -h` for all flags.

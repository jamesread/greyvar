# House interiors

`res/img/textures/tilesets/house_interior.{png,tsx}` and the isleOfStarting
house interiors (`0.0.tailorShop`, `0.0.professorHouse`, `1.2.salon`,
`1.2.redHouse`, `0.1.hermanHouse`) are built by the scripts here.

- `furniture/<n>.png`: 64 pieces made with PixelLab in one call
  (`create_1_direction_object`, size 32, one description per piece): beds
  0-1, table 2, chair 3, bookshelf 4, fireplace 5, plant 6, drawers 7, rugs
  8-9, lamp 10, barrel 11, stove 12, kitchen counter 13, clock 14, painting
  15, window 16, study desk 17, globe 18, chalkboard 19, mannequin 20, fabric
  rolls 21, sewing table 22, barber chair 23, vanity 24, basin 25, teller
  counter 26, vault 27, coin table 28, ledger desk 29, side table 30, crate
  31, armchair 32, grandfather clock 33, coat stand 34, sofa 35; 36-63 are
  extra pieces in the same style. Packed onto the 16px grid; rugs (8, 9) and
  wall hangings (14, 15, 16, 47, 50) have no collision.
- `build_tileset.py`: draws the room pieces (plank floors, 2-tile back wall
  in five wallpapers, beam walls, doormat) and packs the furniture.
  `python3 build_tileset.py <outdir>` writes the png, tsx and
  `house_index.json`.
- `build_maps.py <worlds/isleOfStarting_dev_tiled> <house_index.json>`
  rewrites each house's layers (room, rugs, wall decor, furniture) from its
  room box and furniture list, keeping the map's objects.

## Scale: small pieces (`furniture-small/`)

The player is one tile (16px); most original pieces are 2x2. Pieces that
should be smaller were redrawn by PixelLab (`create_image_pro`, the original
piece as a reference image, 64 candidates each) at 1x1, 1x2 or 2x1, and the
candidate closest to the original kept. They are named (`chair.png`,
`bed.png`, ...) and packed after the originals; `build_maps.py` uses names
for them and numbers for the 2x2 showpieces it keeps (fireplace, wardrobe,
chalkboard, teller counter, vault, window, painting, pans, rugs). The
originals all stay in the tileset.

  | piece | redrawn from | job | candidate |
  |---|---|---|---|
  | armchair | 32 | 5c9576af | 0 |
  | barber | 23 | 535be7a9 | 23 |
  | barrel | redrawn (art rules) | a67ce4c5 | 1 |
  | basin | 25 | 006874a4 | 30 |
  | bed | 0 | d501c97d | 0 |
  | bench | redrawn (art rules) | 66f037d0 | 12 |
  | bookshelf | 4 | ca09d2f5 | 32 |
  | chair | redrawn (art rules) | 29dc244f | 2 |
  | clock | 14 | 85c9112e | 24 |
  | coatstand | redrawn (art rules) | 8b1f08c7 | 5 |
  | cointable | 28 | f273d9ae | 0 |
  | counter | redrawn (art rules) | df068392 | 8 |
  | crate | redrawn (art rules) | f3c0d75f | 53 |
  | desk | redrawn (art rules) | 4a1a4656 | 8 |
  | drawers | redrawn (art rules) | a4997730 | 18 |
  | fabricshelf | 21 | b611dd10 | 0 |
  | globe | 18 | e36256c1 | 4 |
  | grandfather | redrawn (art rules) | 4741828e | 4 |
  | lamp | redrawn (art rules) | 6b773687 | 4 |
  | ledger | redrawn (art rules) | 01a2181e | 2 |
  | mannequin | 20 | b7d1cdc9 | 33 |
  | mirror | 51 | 6285a9ad | 48 |
  | plant | 6 | 196ac55e | 1 |
  | sewing | redrawn (art rules) | 2a9336ef | 1 (outline darkened) |
  | sidetable | 30 | e5ea59ad | 8 |
  | sofa | 35 | 68c608eb | 2 |
  | stool | redrawn (art rules) | 4a92af30 | 1 |
  | stove | 12 | 1cbe8705 | 19 |
  | table | redrawn (art rules) | ae528d4e | 0 |
  | vanity | redrawn (art rules) | 1f93e5a6 | 3 |

Pieces marked "redrawn (art rules)" were redrawn again to follow
[../../ART_RULES.md](../../ART_RULES.md): a flat front view (no rotation), dark
walnut instead of mid-brown wood so they stand out from the floor, and a
fixed 15-colour palette. Candidates were kept only if they passed
`res/tools/artcheck.py` on the house floor.

For wide/tall pieces the prompt must say the object spans the FULL width
(or height) of the canvas, or PixelLab draws a 1-tile object in the middle.

The prison (cellOfEscape) still uses the old `interior.tsx`.
Rebuilding the maps overwrites hand edits to those layers: add furniture to
`build_maps.py` instead, or stop using the script for a house you edit in
Tiled.

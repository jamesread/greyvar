# Art rules

Every new or redrawn asset follows these rules, so the game reads as one
consistent world. The measurable ones are checked by
`res/tools/artcheck.py` (see [Checking](#checking)); perspective is checked
by eye.

Reference assets that follow the rules: the player, the rift portals, shop
signs, the chest, pots, the fireplace, beds and bookshelves, and nearly all
terrain tiles. Examples that break them are named in each section.

## 1. Grid and scale

- One tile is **16×16 px**. One art pixel is one game pixel: never scale art
  up or down, and never mix pixel sizes in one asset.
- **A character is one tile**: a 16×16 frame with a body about 10 px wide
  and 16 px tall (the child size). **Adults** are about 1.5 tiles tall: a
  33×33 frame with a figure 17 px wide (front and back; 13 px from the side)
  and 27 px tall, standing on the same one-tile footprint.
- **Characters are an odd number of pixels wide**, on an odd frame, so they
  have a true centre column to be symmetric about (the adult frame is 33,
  centred on column 16). New character sizes follow this. Everything is sized against that, as it would be in real
  life (a chair seat is below a character's waist, a table about half its
  height, a grandfather clock taller than it).
- Footprints, in tiles (the cells an object stands on and blocks):

| Footprint | For | Examples |
|---|---|---|
| 1×1 | small things | chair, stool, plant, lamp, barrel, crate, pot, chest, sign, lever, basin, small table |
| 1×2 (tall) | upright things one tile wide | bookshelf, grandfather clock, mannequin, coat stand, mirror, vanity, single bed |
| 2×1 (wide) | long, low things | dining table, desk, bench, sofa, counter, sewing table |
| 2×2 (largest) | showpieces only | fireplace, double bed, wardrobe, vault, teller counter, chalkboard |

- Nothing is bigger than 2×2 except buildings (a house exterior is 4×4),
  trees and terrain.
- Low furniture (chairs, stools, tables, benches) is clearly shorter than a
  character: at most about 12 px of a 16 px cell.
- Doors and doorways are one tile wide.
- Wall hangings (windows, paintings, clocks) fit the 2-tile back wall of an
  interior: up to 2×2.

## 2. Canvas and aspect ratio

- Canvas width and height are **multiples of 16**, and the footprint is one
  of 1×1, 1×2, 2×1 or 2×2.
- **Entities may instead be one pixel short of the grid: 15 px per cell**
  (15, or 31 for two cells). The odd size gives the sprite a true centre
  pixel, so symmetric entities (portals, buttons, switches, chests) have a
  middle column and row to centre on. Use 16 when the entity fills its cell
  or tiles against neighbours (push blocks, barriers, doors), and 15 when it
  is centred and symmetric. Terrain and wall tiles are always 16.
- The object sits **bottom-aligned and horizontally centred** in its canvas.
  It should use the canvas: a 2×1 piece spans at least 24 px wide and a 1×2
  piece at least 24 px tall.
- A tall object's footprint is its **bottom** cell; its upper cell stands in
  front of the wall or air behind it.
- Animation frames share one canvas size; frames sit side by side in the
  sheet.

## 3. Perspective: 2D top-down

The whole game uses one view, the flat 16-bit RPG view:

- **Ground** (terrain, floors, rugs, paths) is seen straight down: flat, no
  perspective, no outline, tiling seamlessly.
- **Anything standing** (characters, furniture, signs, props, buildings) is
  seen from above and in front, and what shows depends on its shape:
  - **Tall, upright things** (characters, bookshelves, wardrobes, clocks,
    signs, mannequins) show mostly their **front face**, with at most a thin
    strip of their top above it.
  - **Low, flat-topped things** (tables, desks, counters, beds, benches'
    seats, chests' lids) show mostly their **top surface**: it fills most of
    the sprite, with a short front edge and the legs below it. A table drawn
    side-on (all front, no top) is wrong even if it is square to the viewer.
- **Edges are horizontal and vertical.** No rotation, no three-quarter
  turn, no isometric or oblique angles, no diagonal sides, no vanishing
  points.
- A symmetric object (chair, bed, barrel, clock) is drawn symmetric.
- Characters have front, back and side views (four directions). Objects
  have one view, the front.
- Wall hangings are drawn flat, straight on.
- Light comes from the top left: highlights on top and left edges, shade on
  bottom and right edges. No cast shadows.

Fits: the player, portals, signs, chest, pot, fireplace, bed, bookshelf.
Breaks it: the **rift lever** (base and handle drawn at an angle), the
**wooden chair** (turned three-quarters), and a **table drawn side-on**
(only its front edge and legs, no table top).

## 4. Pixels and colour

- **Binary alpha**: every pixel is fully opaque or fully transparent. No
  anti-aliasing against transparency, no blur, no soft gradients.
- **A closed 1 px dark outline** around every object and character, darker
  than both the object and the ground beneath it. Ground tiles have no
  outline.
- **Keep the palette small**: aim for 16 colours or fewer per sprite (the
  chest uses 6, the player 13, the fireplace 12). Many near-identical shades
  usually mean soft, muddy shading; prefer a few clearly separated steps.
  This is guidance (artcheck warns), not a hard rule: some good assets, such
  as the shop signs, use more.
- Build from the project's colours: the bright base palette (grass
  `#b5e61d`, hill `#22b14c`, sand `#efe4b0`, water `#00a2e8`, castle
  `#c3c3c3`) and the shading ramps already in the tilesets. A new material
  gets a ramp of 3 or 4 steps (outline, shade, base, highlight).

## 5. Contrast against the ground

Everything placed on the ground must stand out from **every ground it is
placed on** (grass, sand, hill, castle stone, house floor...).

| Rule | Minimum |
|---|---|
| Outline: mean colour difference from the ground | ΔE ≥ 35 |
| Body: share of body pixels clearly unlike the ground (ΔE > 20) | ≥ 70% |
| Body: mean colour difference from the ground | ΔE ≥ 30 |

ΔE is the CIELAB colour difference (CIE76): about 2 is barely noticeable,
20 clearly different, 50 or more strongly contrasting. The thresholds were
set from these examples:

| Example | Verdict | Body clearly unlike ground | Body mean ΔE |
|---|---|---|---|
| fireplace on the house floor | good | 87% | 35 |
| chest on grass | good | 100% | 69 |
| flower on grass | good | 100% | 72 |
| bookshelf on the house floor | good | 74% | 33 |
| wooden table on the house floor | bad | 4% | 17 |
| wooden chair on the house floor | bad | 53% | 27 |
| Herman's ledger desk on the house floor | bad | 43% | 21 |

Ground tiles keep their own detail quiet (speckles and grain within about
ΔE 15 of the base colour), so objects carry the contrast.

When an object fails, change the **object** first, not the ground: a
different material (dark walnut, painted wood, stone, metal), cloth or
cushions in a contrasting colour, or a darker outline. Wooden furniture is
the common case: on the honey-brown house floor (`#9b6937`, L* 49) it must be
clearly darker (L* about 30 or less) or lighter, or carry colour.

Grounds for reference:

| Ground | Mean colour | L* (lightness) | Chroma |
|---|---|---|---|
| grass | `#b7e61c` | 85 | 89 |
| sand | `#efe4aa` | 90 | 30 |
| hill | `#26b24a` | 64 | 72 |
| castle stone | `#8f8f97` | 60 | 5 |
| mud | `#a26926` | 50 | 48 |
| house floor | `#9b6937` | 49 | 38 |

## 6. Making art with PixelLab

- Ask for the view explicitly: *"flat 2D top-down RPG sprite, front view,
  facing the viewer, straight on, not rotated, symmetrical"*. For tables
  and other flat-topped furniture add *"seen from above: the top surface
  fills most of the sprite, with a thin front edge and short legs"*. Never ask for
  a "3/4 view", "isometric" or "angled" view: that is where the angled
  lever and chair came from.
- **Characters** (`create_character`): use `view: "side"` (eye level). It
  gives a straight-on front and back and a true side profile. The
  `low top-down` and `high top-down` views draw the figure turned and tilted,
  almost isometric. Lock `palette` to the rig's key colours so the parts
  split into layers exactly, then animate with `mode: "skeleton-v3"`: the
  template modes redraw every frame and the figure wobbles. `spritegen
  import-pixellab` also lines each animation's frames up on the head.
- Generate at the native canvas (16×16, 16×32, 32×16, 32×32). For a wide or
  tall piece, say it *"spans the FULL width (height) of the canvas"*, or a
  one-tile object comes back in the middle.
- Pass a `palette` (the project's colours) or reduce colours afterwards to
  16 or fewer.
- To redraw an existing piece, give it as a reference image and say what to
  keep (design, colours).
- Before accepting a result, run `artcheck` against every ground it will
  stand on, and check the perspective by eye.

## Checking

```bash
# one sprite (or one frame of a sheet) on a ground
python3 res/tools/artcheck.py res/img/textures/entities/lever.png --frame 16 16 0 --ground castle
# every interior piece in res/sprites/interiors/furniture-small on the house floor
python3 res/tools/artcheck.py --interiors
```

It fails on canvas size and footprint, binary alpha, and outline and body
contrast, and warns on colour count; `-v` also lists what passes. Grounds: `grass`, `sand`,
`hill`, `castle`, `mud`, `house_floor`. Perspective, outline closure and
scale against the character are reviewed by eye.

# Layered sprite rigs

Characters are drawn as stacked **layers** (pants, body, shirt, head, hair,
plus a held item). The client recolours each layer at runtime and every layer
shows the same **pose**. The rig directory is the only source of truth:

```
res/sprites/player/
  rig.yml     layers + variants, colour slots (key ramps), palettes, defaults, output paths
  poses.yml   ordered poses: part per layer, per-layer offsets, hand anchor, itemZ
  clips.yml   clips as {pose, ms} steps, idle behaviour, entity state -> clip
  parts/<layer>/<variant>/<part>.png   16x16 parts, rig colours only
```

`make sprites` (`go run ./sprite-generator build`) generates everything else.
**Never edit these generated files by hand:**

- `res/img/textures/entities/player/<layer>_<variant>.png` holds one sheet per
  layer variant.
- `res/img/textures/entities/player/player.rig.json` is the manifest the client
  and server load.
- `res/img/textures/entities/player/composite.png` and
  `server/dat/entdefs/player.tsx` show the default look for the server and Tiled.

`make sprites-check` fails if the rig is invalid or a generated file is stale.
`make test` runs it.

## Rules that keep edits safe

- **Never read or paint the generated sheets.** Work with parts and YAML.
- **Use symbolic colours.** A part may only use slot key colours
  (`skin.0`..`skin.3`, `hair.0`..`hair.2`, `shirt.*`, `pants.*`, where 0 is
  darkest), `fixed.*` (outline and eyes) or transparency. The validator rejects
  anything else.
- **Reuse before you draw.** A new frame is usually an existing part with an
  `offsets` entry in `poses.yml`.
- **Always look at the result**: `spritegen preview -clip <clip> -out /tmp/x.png`
  renders an 8x strip you can view.

## Commands

All commands run from the repository root. The rig directory defaults to
`res/sprites/player`.

```bash
go run ./sprite-generator check                      # validate + stale check
go run ./sprite-generator build                      # regenerate outputs
go run ./sprite-generator preview -out /tmp/p.png [-clip idle] [-set hair=blonde,hair=long,shirt=green] [-scale 8]
go run ./sprite-generator part show head/default/idle_0          # coordinate grid + colour legend
go run ./sprite-generator part derive -layer head -from idle_0 -to wink -edit 9,4=skin.0 [-edit x1-x2,y=clear]
go run ./sprite-generator part longhair -from short -to long     # derive a whole hair variant
go run ./sprite-generator part prune [-n]                        # delete parts no pose uses
```

Diagonal walks reuse the horizontal walk parts, as classic Zelda does. They are
still separate poses (`walkUpLeft_0`...), so real diagonal art only needs
new parts and a `parts:` change. Walk-down uses the idle upper body, so the head
stays identical to idle, over an authored leg cycle (`pants/walkDown_step*`).
When repointing poses, run `part prune` afterwards and re-run
`part longhair -force` so the long-hair variant matches the short one.

`part derive` copies a part (or starts blank without `-from`) and applies
symbolic pixel edits. By default it writes the part for **every** variant of
the layer, because each variant must provide every part its poses use. It
refuses to overwrite an existing part unless you pass `-force`.

## Recipes

### Add an idle variant (or any one-shot animation)

1. Run `part show` on the parts you need. Use `part derive` to make new ones,
   for example a changed head for an expression.
2. Add poses to `poses.yml`. Keep the head steady: idle variants should change
   one thing at a time.
   ```yaml
   - name: down_wink
     parts: {pants: idle_0, body: idle_0, shirt: idle_0, head: down_wink, hair: idle_0}
     hand: [12, 11]
   ```
3. Add the clip and weight it in `clips.yml`:
   ```yaml
   clips:
     idle_wink:
       - {pose: down_wink, ms: 200}
       - {pose: down_stand, ms: 200}
   idle:
     variants:
       - {clip: idle_wink, weight: 2}
   ```
4. Run `check`, `build` and `preview -clip idle_wink`. The client picks it up on
   a browser refresh.

The idle loop plays `idle.base` and, every `intervalMs` (a random value in the
range), one weighted random variant once.

### Add a colour option

Add a palette to the slot in `rig.yml`. It needs the same number of colours as
the slot's key ramp, ordered dark to light:

```yaml
palettes:
  shirt:
    purple: ['#24123a', '#4a2480', '#6030a8', '#8850d0']
```

Then run `build`. The server offers it in default looks and accepts it in
`appearance shirt=purple`. You don't need to restart the server for the client
side, but the server reads `player.rig.json` at startup.

### Add a hairstyle (or any layer variant)

1. Create `parts/hair/<style>/` with one PNG for every hair part that
   `poses.yml` uses. Start from `part longhair`, or from
   `part derive -variant <style> ...` per part.
2. Add the style to the layer's `variants` in `rig.yml`.
3. Run `check`, `build` and `preview -set hair=<style>`.

### Add a held item

1. Draw a small PNG (around 8x5 pixels) under `res/img/textures/items/held/`.
2. In the itemdef (`server/dat/itemdefs/<item>.yml`) set:
   ```yaml
   held:
     sprite: img/textures/items/held/<item>.png
     grip: [1, 2]   # pixel that sits on the hand
   ```
3. Each pose's `hand: [x, y]` sets where it goes. `itemZ: behind` draws it
   under the body, which the up-facing poses use. A pose without `hand` hides
   the item.
4. In game, run `appearance held=<item>`. `held=-` clears it.

## Use (action key)

Pressing E plays `use<Facing>` from `clips.yml` once, then the facing idle:
`useLeft`/`useRight` swing the near arm out of the sleeve (body parts
`useLeft_half`/`_reach`, `useRight_*`, shirt `useLeft`/`useRight`), `useDown`
brings both hands up in front of the torso (body and shirt `useDown`; the
hands are skin pixels on the shirt layer), and `useUp` tucks the forearms
behind the back (body `useUp`). Diagonals use the side clip. The local player
plays it on key down; the server sends `use<Facing>` to everyone else as a
one-off `EntityStateChange` (the entity's state does not change). Only
movement cuts it short. To restyle it, edit those parts or poses and keep each
clip short (a test keeps it under 600ms).

## Transient states (walk*, bump*)

The client plays `walk<Dir>` while moves arrive. The server sends
`bump<Dir>` (an `EntityStateChange`) for every move it fully blocks, so the
player plays the matching `bump<Dir>` clip from `clips.yml`. Both fall back to
idle about 250ms after the events stop. Bump poses are walk parts leaning 1px
into the wall horizontally; vertical pushes dip the head and hair 1px instead,
because the frame has no headroom. To restyle a bump, edit its
`bump<Dir>_lean*` poses or its clip timings.

## Facing idles

The sprite remembers its last facing from `walk*`/`bump*` states. When it
stops, it plays `idle<Facing>` from `clips.yml`. Diagonals fall back to
`idleLeft`/`idleRight`, the same side view their walks use. Facing down plays
the front `idle`, which has the random blink, glance and shrug variants. To
give another facing its own variants, add more steps to its `idle<Facing>` clip.

## Runtime pieces

- **Client:** `webclient/js/sprites/`
  - `PlayerRig.js` loads the manifest and makes recoloured textures.
  - `LayeredSprite.js` is the Phaser container.
  - `AnimationController.js` and `IdleScheduler.js` are pure and unit tested.
- **Server:** `datlib/rigs` provides validation and username-seeded default
  looks. `server/pkg/greyvarserver/frame_appearance.go` handles
  `AppearanceRequest`.
- **Protocol:** `Appearance` on `EntitySpawn`, plus `EntityAppearanceChange` and
  `AppearanceRequest`.

# Flora tileset (animated decor tiles)

Flowers are animated **tiles**, not entities: no collision, no interaction,
and they cost nothing on the server. The source lives in `res/sprites/flora/`:

```
flora.yml          sway pattern (offsets px + ms per frame), one entry per flower
parts/<name>.png   16x16 transparent flower, upright (frame 0)
```

`make sprites` (`spritegen flora build`) writes
`res/img/textures/tilesets/flora.png` and `flora.tsx`. Each flower is one
animated tile with type `flora` and property `flora=<name>`; place that tile
in Tiled. The client plays the animation, starting each map cell on a
different frame.

```bash
go run ./sprite-generator flora preview -out /tmp/flora.png   # rows = flowers, columns = frames
go run ./sprite-generator flora check                         # validate + stale check
```

Flora tiles carry the property `decor=true`. Decor tiles draw on top, but
collision, teleports and other tile rules use the terrain beneath them
(`Grid.CollisionTileAt`). Put them on a `decor` layer above the terrain.

### Move flowers out of other tilesets

Old maps had flowers painted into grass tiles. Each flower's `sources` in
`flora.yml` lists those tiles, and `migrate.replacement` names the plain tile
drawn in their place. To migrate:

```bash
go run ./sprite-generator flora migrate -n server/dat/worlds/*/*.tmj   # dry run
go run ./sprite-generator flora migrate server/dat/worlds/*/*.tmj
```

The tool edits each map's JSON text in place, so Tiled's formatting is kept.
It re-parses and checks the result before writing, and skips maps that
already use the flora tileset.

### Add a flower

1. Draw `parts/<name>.png` (16x16, transparent background, no semi-transparent pixels).
2. Add an entry to `flora.yml` and list its `heads`. These are inclusive
   `[x0, y0, x1, y1]` boxes that sway. Leave stems outside them so they stay
   planted. For clusters, give each head a box and use `phases` so the heads
   don't move together.
3. Run `flora check`, `make sprites` and `flora preview`.

# Crops tileset and crop definitions

Crops use the same generator as flora: source in `res/sprites/crops/`
(`flora.yml` with `decor: false`, plus parts). It builds
`res/img/textures/tilesets/crops.png` and `crops.tsx`. Each stage tile carries
the properties `crop=<id>` and `stage=<n>`. Crop rules live in
`server/dat/cropdefs/<id>.yml`:

```yaml
id: hay
seed: hay_seeds        # holding this and interacting with an empty walkable cell plants it
starterSeeds: 5        # given to new players
stages: [{growMs: 30000}, {growMs: 45000}, {}]   # the last stage is ripe
harvest: {item: hay, count: 2, seeds: 1}
```

At runtime the server owns crop cells. It grows them by timestamp, so they
keep growing while the server is off, and saves them to
`server/state/crops/<world>.json` (gitignored; override with
`GREYVAR_STATE_DIR`). Clients see crops as a `crops` grid layer and get live
`TileChange` messages. The interact key (E) targets the cell the player faces
when no nearby entity has anything to do.

### Add a crop

1. Draw or generate one part per stage in `res/sprites/crops/parts/`.
2. Add one entry per stage to `res/sprites/crops/flora.yml`, with
   `props: {crop: <id>, stage: "<n>"}`. Run `make sprites`.
3. Add `server/dat/itemdefs/<seed>.yml` and the harvest item, then
   `server/dat/cropdefs/<id>.yml`.
4. Restart the server. It logs and disables any crop whose items or stage
   tiles are missing.

# Animated water: the tide on sand shores (PixelLab)

The sand shorelines of `grass_terrains` play a gentle tide: a foam line washes
up the sand and slides back, leaving the wet sand darker. Open water and plain
sand stay still. The art comes from PixelLab, not from code.
`res/sprites/water/` holds the source:

- `tide_base.png` and `island_base.png` are small maps composed from this
  tileset's own tiles. The lake (water in sand) gives the opaque water/sand
  tiles of hand-made maps. The island (sand in water) gives the see-through
  sand tiles the world generator layers over water.
- `*_styled.png` are PixelLab img2img passes (`create_image_pixflux`) that
  add the foam line and wet sand.
- `frames/*.png` are PixelLab animations (`animate_image`) of those, with the
  first frame pinned as the last so they loop.
- `tileanim.yml` holds the layouts, the colour map and the prompts, seeds and
  job ids used.

`make sprites` (`spritegen tiles animate`) cuts each shore tile out of the
frames. Only pixels within `band` of the tile's own waterline change; the
rest stay the original art, so shore tiles still meet plain neighbours
seamlessly. PixelLab colours are mapped onto the tileset palette. Frames are
appended below the 20 hand-drawn rows of `grass_terrains.png`, and each shore
tile gets an `<animation>` plus `synced=true` in `grass_terrains.tsx`. The
hand-drawn rows and the rest of the `.tsx` are never changed, and runs are
idempotent. After re-exporting the PNG from GIMP, run `make sprites` again;
`make sprites-check` reports it until you do.

Some see-through sand edge tiles are used both over water (generated maps)
and over grass (hand-made maps). `make sprites` marks them
`needs_water_below`, and marks all-water tiles `water`. When the server sends
a map, it checks each such cell: if the first ground tile underneath (flowers
and other decor are skipped) is not water, the cell is drawn still. So foam
appears only where sand meets water, never between sand and grass.

Synced animations run from one shared clock on the client, so the whole
shoreline breathes together. Flowers and crops start on a different frame per
cell.

To change the look, regenerate the PixelLab steps listed in `tileanim.yml`
(e.g. another seed or action text), replace `frames/*.png`, and run
`make sprites`.

## Taller cliff faces (`res/sprites/faces`)

South-facing hill edges get a face 12px taller than the hand-drawn tiles,
without changing the tilesets or the maps' terrain:

- `faces.yml` lists the source tilesets (hill tiles are found from the
  `terrains` wangset's corners; steps are listed by id), the face colours,
  the rows that repeat (thin rock strata) and the maps to apply to.
- `make sprites` writes `res/img/textures/tilesets/cliff_faces.{png,tsx}`
  and two locked layers in each listed map: `face caps (generated)` (over
  the face tile, replacing its base outline) and `faces (generated)` (the
  extension into the cell below, with collision).
- Generated tiles have `overlay=true`: the server adds their collision to
  the tile beneath instead of replacing it (`Grid.CollisionTileAt`).

Rules: edit the terrain in Tiled as usual, then `make sprites`; never edit
the generated layers or `cliff_faces.*`. `make sprites-check` fails if a
listed map changed since. To roll out to another map, add it under `maps:`.
Leave room for the taller face: the row under a south cliff is now mostly
cliff (only its bottom 4px are walkable).

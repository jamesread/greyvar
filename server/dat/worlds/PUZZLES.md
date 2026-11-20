# Puzzle wiring (Tiled object properties)

Puzzles are built from map objects and their custom properties; no code is
needed per map. `cellOfEscape_dev_tiled` is the worked example.

| Property      | On                       | Meaning |
|---------------|--------------------------|---------|
| `id`          | any object               | Name other objects refer to (e.g. `sw1`). |
| `loot`        | lootable (pot), chest    | `item[:count], ...` given when broken/opened; overrides the entdef's loot. |
| `requires`    | entdefs with `interact.open` (chest) | Item needed (and used up) to open. |
| `msg`         | chest, signpost          | Text shown when opened / touched. |
| `locked_msg`  | chest                    | Text shown when `requires` is missing (default "It's locked."). |
| `open_when`   | doors (`cellDoor`)       | `id=state, id=state`: open while all hold, closed otherwise. Re-checked whenever any state changes. |
| `reset_after` | toggles (`lever`, `wallSwitch`) | Milliseconds: once flipped away from its initial state the lever springs back (a timed switch). The flipping player sees "Tick, tock... hurry!". Flipping it back by hand stops the countdown. |
| `teleport_world`, `teleport_destination` | `teleportPad` | Where the pad warps to. |

Entity definitions (`server/dat/entdefs/*.yml`) control behaviour:

- `interact: {toggle: [a, b]}`: the interact key (E) cycles through these states (e.g. `lever`, `wallSwitch`).
- `interact: {open: open}`: E opens it once, using `requires`/`loot`/`msg` above.
- `interact: {open: open, close: shut}` (e.g. `chest`): as above, then E closes it (`shut`) and opens it again; `requires`/`loot`/`msg` only apply to the first opening.
- `passableStates: [open]`: a `solid` entity stops blocking in these states (e.g. `cellDoor`).
- `pushable: true` (with `solid`; e.g. `pushBlock`): lean on it (keep walking
  into it, square on, for about a quarter of a second) and it slides one tile,
  if that tile is free floor: no wall tile, solid entity, door (open or
  closed), teleporter or player. The use button (E) pulls it instead: square
  on beside it, you step back a tile and it follows, if there is room behind
  you. So a block pushed against a wall or into a corner can be recovered.

Pressure plates (`pressureButton`, any `collision_type: pressure` entity that
is not a teleporter) are `pressed` while any player stands on them or a push
block sits on them, and pop back up when the last one leaves. A door with
`open_when = plate1=pressed` therefore stays open while a block holds the
plate, or while a friend stands on it.

All state changes are sent to every player on the grid, so puzzles are shared
in multiplayer. Objects go in the cell Tiled shows them in: a tile object
(placed from a tileset, so it has a `gid`) is anchored at its bottom-left
corner, any other object at its top-left, and an object dragged slightly off
the grid goes to the nearest cell (`objectCell` in `datlib/gridfiles/tmj.go`).

## cellOfEscape walkthrough

1. A signpost hints that something rattles in a pot. The pot at row 9, col 2
   has `loot = rusty_key`.
2. The chest (row 10, col 8) has `requires = rusty_key` and `loot = note`. Its
   `msg` gives the lever order.
3. Levers `sw1`, `sw2` and `sw3` (row 3, cols 10/11/13). The cell door
   (row 8, col 13) has
   `open_when = sw1=flickedRight, sw2=flickedLeft, sw3=flickedRight`.
4. The escape pad (row 10, col 13) warps to `isleOfStarting_dev_tiled` / `1.1.tmj`.

`server/pkg/greyvarserver/escape_test.go` plays this with the real data, so a
map edit that makes the puzzle unsolvable fails `go test`.

## Doors and warps

Step on an entity with a `teleport_destination` to go to that grid:

| Property               | Meaning |
|------------------------|---------|
| `teleport_destination` | Grid file, e.g. `1.2.redHouse.tmj`. |
| `teleport_x`, `teleport_y` | Arrival cell (tiles). Default: the destination's spawn point. Put it next to the way back, not on it. |
| `teleport_world`       | Another world. Default: the current one (teleport pads must set it). |

`houseDoor` (a house's front door) and `stoneDoorway` (the way out of an
interior) are in `entdefs/doors.tsx`. Houses are built from
`res/img/textures/tilesets/house.tsx` (solid; leave the door cell empty and
place a `houseDoor` there). Blue-roofed houses are 4 tiles tall (roof, roof
bottom, windows, door row); red and purple ones 3 (their roof tiles include
the wall). Mirror the left column (flip horizontally). `houseWindowsTop` is the
wall above the door (its lintel), `houseWindows` the window: for a door on the
right, swap them (see Professor Ivan's house on 0.0). Example: the red
house on isleOfStarting 1.2 and its inside, `1.2.redHouse.tmj`.

## Caves

`caveDoor` (`entdefs/caveDoor.tsx`, art made with PixelLab) is a cave mouth for
the brown hill walls: place it on the hill's edge row (the row above the cliff
face; the sprite is 16x32 and its arch sits in the face below). Walk up into
the face to go in; it takes `teleport_destination`, `teleport_x` and
`teleport_y` like `houseDoor`. Inside, `caveExit` (daylight at the mouth) leads
back out; set its `teleport_x`/`teleport_y` to the cell below the door. Cave
maps use `res/img/textures/tilesets/cave.tsx` (wangset `cave`: floor and rock;
rock corners collide). Example: the debug cave, `1.1.debugCave.tmj`, entered
from the top left of isleOfStarting 1.1, which holds the teleporter to the
debug island. Keep a cave's edges solid rock: its file name gives it 1.1's
position, so an open edge would lead onto the neighbouring map.

## Shop signs

Hang a sign over a shop's door: an object of type `signHair`, `signTailor`,
`signScience` or `signFarm` (`entdefs/shopsigns.tsx`, art in
`res/sprites/shopsigns`) one cell above the door, with `msg` set to the
shop's name. It does not block, and its name shows while the player is
within `touch_range` (40px, set on the tileset tile) - e.g. at the door.
`touch_range` works for any touch entity without collision shapes.

## Rifts (generated puzzles)

`world-generator puzzle -level N -seed S` writes whole puzzle worlds from these
pieces (levers, key pots and chests, push blocks on plates, timed switches),
each checked solvable before it is written. See `world-generator/README.md`.
Doorways between their stone rooms are closed by stone barriers:
`barrierV` (upright, across a left-right passage) and `barrierH` (across an
up-down passage), which sink into the ground when open; `cellDoor` remains the
prison-cell door for hand-made maps.

In game the server opens them on demand as **rifts**
(`server/pkg/greyvarserver/rifts.go`):

- A rift is a world `rift_<n>` written to `$GREYVAR_STATE_DIR/rifts/rift_<n>/`,
  loaded and mounted like any other. Generating takes a few milliseconds (the
  hardest levels about 0.2s).
- You arrive on the rift's **entry teleporter** (`teleportPad`,
  `rift_exit = leave`). Step off it and back on to flee: you go back where you
  came from and the rift is not cleared (the rest of the party stays in).
- The **exit teleporter** (`rift_exit = complete`) clears the level for
  everyone in the rift and sends each back where they came from ("Rift level N
  cleared!").
- **Rift coins** (`riftCoin` pickups, item `rift_coin`) lie in the last room
  around the exit, more at higher levels: the reward for getting there. They
  pile up in the inventory; nothing takes them yet.
- Players entering with the same key (from the same portal) share one rift
  while it is open: a party. Each remembers where they came from.
- When the last player leaves (exit, entry teleporter, `world-load`, teleport
  or disconnect) the rift is unmounted and its directory deleted. Leftover
  rift directories are removed when the server starts.
- `world-reload` in a rift regenerates it from its spec (a reset).
- Progress, each player's highest level cleared, is kept in
  `$GREYVAR_STATE_DIR/rifts.json`; a player who asks for no particular level
  gets one more than their best.

Ways in (`rift_entry.go`):

- **Rift portal** (`riftPortal`, gold pad; one on isleOfStarting 1.1 by the
  spawn): step on it for a rift one level above your best. Anyone who steps on
  the same portal while that rift is open joins the party. Optional map
  properties: `rift_level` (always this level) and `rift_mechanics` (e.g.
  `lever,keyChest`). After the rift you are back on the portal; step off and
  on again for the next level.
- **Console**: `rift` (your next level), `rift 5`, `rift 5 seed=7`,
  `rift mechanics=lever,keyChest`, and `rift leave` to flee. A console rift is
  private (nobody joins it).

`rifts_test.go` plays the solver's solution for a range of levels and seeds
through the real server (moves, interactions, pushes and pulls, timers on a
simulated clock, the exit), so a rule change that breaks generated puzzles
fails `go test`.

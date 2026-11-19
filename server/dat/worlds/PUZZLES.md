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
| `teleport_world`, `teleport_destination` | `teleportPad` | Where the pad warps to. |

Entity definitions (`server/dat/entdefs/*.yml`) control behaviour:

- `interact: {toggle: [a, b]}`: the interact key (E) cycles through these states (e.g. `wallSwitch`).
- `interact: {open: open}`: E opens it once, using `requires`/`loot`/`msg` above.
- `interact: {open: open, close: shut}` (e.g. `chest`): as above, then E closes it (`shut`) and opens it again; `requires`/`loot`/`msg` only apply to the first opening.
- `passableStates: [open]`: a `solid` entity stops blocking in these states (e.g. `cellDoor`).

All state changes are sent to every player on the grid, so puzzles are shared
in multiplayer. Objects are placed at `row = floor(y / 16)`.

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

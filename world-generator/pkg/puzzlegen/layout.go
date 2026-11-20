package puzzlegen

import (
	"fmt"
	"math"
	"math/rand"
	"sort"
	"strings"

	"github.com/jamesread/greyvar/world-generator/pkg/gridgen"
)

// Map geometry: one 16x16 map. Terrain level 0 is the grass outside, level 1
// the stone ("wall" wangset colour) rooms. The wall wangset draws a solid
// wall line along the boundary of every stone region, so a room is a stone
// rectangle; two rooms are kept one corner apart (two walls back to back)
// and a door punches a two-corner gap through both.
const (
	mapSize      = 16
	levelOutside = 0
	levelRoom    = 1
	keyItem      = "rusty_key"
	decoyItem    = "shard"
	maxCoins     = 8
)

// Cell is a tile position.
type Cell struct{ R, C int }

func (c Cell) add(d Cell) Cell { return Cell{c.R + d.R, c.C + d.C} }

var dirs = []Cell{{-1, 0}, {1, 0}, {0, -1}, {0, 1}}

// Room is a stone rectangle in corner coordinates (inclusive). Its interior
// (fully walkable) cells are rows R0..R1-1, cols C0..C1-1.
type Room struct{ R0, R1, C0, C1 int }

func (r Room) contains(c Cell) bool {
	return c.R >= r.R0 && c.R < r.R1 && c.C >= r.C0 && c.C < r.C1
}

func (r Room) cells() []Cell {
	var out []Cell
	for row := r.R0; row < r.R1; row++ {
		for col := r.C0; col < r.C1; col++ {
			out = append(out, Cell{row, col})
		}
	}
	return out
}

// Room slots for 2..4 rooms, in chain order (each joins the next).
var slots = map[int][]Room{
	2: {{3, 13, 1, 7}, {3, 13, 9, 15}},
	3: {{1, 7, 1, 7}, {1, 7, 9, 15}, {9, 15, 9, 15}},
	4: {{1, 7, 1, 7}, {1, 7, 9, 15}, {9, 15, 9, 15}, {9, 15, 1, 7}},
}

// Kind of a puzzle entity; Type() is the entity definition it maps to.
type Kind int

const (
	KLever Kind = iota
	KTimedLever
	KPot
	KChest
	KBlock
	KPlate
	KDoor
	KSign
	KExit
	KSpawn
	KEntry // the teleporter the player arrives on; back on it = abandon
	KCoin  // a rift coin (pickup) near the exit
)

// Type is the server entity definition (map object type) for a kind. Doors
// depend on their doorway; see Ent.Type.
func (k Kind) Type() string {
	return [...]string{"lever", "lever", "pot", "chest", "pushBlock", "pressureButton", "barrierV", "signpost", "teleportPad", "spawn", "teleportPad", "riftCoin"}[k]
}

// Barrier entity definitions: a doorway between rooms side by side is a
// left-right passage closed by an upright (vertical) barrier; one between
// rooms above each other is closed by a horizontal one.
const (
	barrierVertical   = "barrierV"
	barrierHorizontal = "barrierH"
)

// Type is the entity definition this entity is placed as.
func (e *Ent) Type() string {
	if e.Def != "" {
		return e.Def
	}
	return e.Kind.Type()
}

// Cond is one condition of a door's open_when: entity ID in state.
type Cond struct{ ID, State string }

// Ent is a placed puzzle entity.
type Ent struct {
	ID       string
	Kind     Kind
	Cell     Cell
	Room     int
	Loot     string // pot/chest loot
	Requires string // chest: item needed (used up) to open
	Msg      string // chest/sign text
	ResetMs  int    // timed lever
	Conds    []Cond // door
	Door     int    // timed lever: index of the door it opens
	Def      string // entity definition, when not the kind's default (doors)
	// door geometry: the two corridor cells, and the approach cell inside
	// each neighbouring room (side[0] in room Rooms[0], side[1] in Rooms[1]).
	Corridor [2]Cell
	Side     [2]Cell
	Rooms    [2]int
	budget   int // timed lever: steps allowed from the lever to the far side
}

// Puzzle is a laid-out (not yet verified) puzzle.
type Puzzle struct {
	Spec  Spec
	Rooms []Room
	Walk  map[Cell]bool // interiors + corridors
	Ents  []*Ent
	Doors []*Ent
	Field *gridgen.CornerField
	Spawn Cell
	Exit  Cell
}

type builder struct {
	p        *Puzzle
	rng      *rand.Rand
	reserved map[Cell]bool
	used     map[Cell]bool
	n        map[string]int
}

func (b *builder) id(prefix string) string {
	b.n[prefix]++
	return fmt.Sprintf("%s%d", prefix, b.n[prefix])
}

// layout builds the rooms, doors and every mechanic's pieces.
func layout(spec Spec, rng *rand.Rand) (*Puzzle, error) {
	rooms := slots[spec.Rooms]
	p := &Puzzle{Spec: spec, Rooms: rooms, Walk: map[Cell]bool{}, Field: gridgen.NewCornerField(mapSize, mapSize)}
	b := &builder{p: p, rng: rng, reserved: map[Cell]bool{}, used: map[Cell]bool{}, n: map[string]int{}}

	for _, r := range rooms {
		for row := r.R0; row <= r.R1; row++ {
			for col := r.C0; col <= r.C1; col++ {
				p.Field.Set(row, col, levelRoom)
			}
		}
		for _, c := range r.cells() {
			p.Walk[c] = true
		}
	}
	for i := 0; i+1 < len(rooms); i++ {
		d, err := b.door(i, i+1)
		if err != nil {
			return nil, err
		}
		p.Doors = append(p.Doors, d)
		p.Ents = append(p.Ents, d)
	}

	// Spawn on the entry teleporter in a corner of the first room (out of
	// the way: walking back onto it abandons the rift), the exit as far as
	// possible into the last.
	r0 := rooms[0]
	corners := []Cell{{r0.R0, r0.C0}, {r0.R0, r0.C1 - 1}, {r0.R1 - 1, r0.C0}, {r0.R1 - 1, r0.C1 - 1}}
	rng.Shuffle(len(corners), func(i, j int) { corners[i], corners[j] = corners[j], corners[i] })
	spawn, ok := b.free(0, corners)
	if !ok {
		return nil, fmt.Errorf("no room for the spawn")
	}
	p.Spawn = spawn
	b.place(&Ent{ID: "spawn", Kind: KSpawn, Cell: spawn, Room: 0})
	b.place(&Ent{ID: "entry", Kind: KEntry, Cell: spawn, Room: 0})
	b.reserved[spawn] = true
	for _, d := range dirs {
		b.reserved[spawn.add(d)] = true // room to step off it
	}
	last := len(rooms) - 1
	exit := b.farthest(last, p.Doors[last-1].Side[1])
	p.Exit = exit
	b.place(&Ent{ID: "exit", Kind: KExit, Cell: exit, Room: last})
	b.reserved[exit] = true

	for i, lock := range spec.Locks {
		if err := b.lock(i, lock); err != nil {
			return nil, err
		}
	}
	b.coins(spec, last, exit)
	for k := 0; k < spec.RedHerrings; k++ {
		room := rng.Intn(len(rooms))
		if rng.Intn(2) == 0 {
			if c, ok := b.free(room, b.topRowFirst(room)); ok {
				b.place(&Ent{ID: b.id("decoyLever"), Kind: KLever, Cell: c, Room: room})
			}
		} else if c, ok := b.free(room, nil); ok {
			b.place(&Ent{ID: b.id("decoyPot"), Kind: KPot, Cell: c, Room: room, Loot: decoyItem})
		}
	}
	return p, nil
}

// coins scatters rift coins over the free cells of the last room nearest
// the exit: more at higher levels.
func (b *builder) coins(spec Spec, room int, exit Cell) {
	n := min(2+spec.Level/2, maxCoins)
	dist := bfs(b.p.Walk, exit, func(c Cell) bool { return !b.p.Rooms[room].contains(c) })
	cells := b.p.Rooms[room].cells()
	sort.SliceStable(cells, func(i, j int) bool {
		di, dj := dist[cells[i]], dist[cells[j]]
		if di != dj {
			return di < dj
		}
		return cells[i].R*100+cells[i].C < cells[j].R*100+cells[j].C
	})
	for _, c := range cells {
		if n == 0 {
			break
		}
		if b.used[c] || b.reserved[c] {
			continue
		}
		b.place(&Ent{ID: b.id("coin"), Kind: KCoin, Cell: c, Room: room})
		n--
	}
}

// door joins rooms a and b through the corner gap between them.
func (b *builder) door(ai, bi int) (*Ent, error) {
	a, c := b.p.Rooms[ai], b.p.Rooms[bi]
	d := &Ent{ID: b.id("door"), Kind: KDoor, Rooms: [2]int{ai, bi}}
	switch {
	case a.C1+2 == c.C0 || c.C1+2 == a.C0: // side by side
		left, right := a, c
		if c.C1+2 == a.C0 {
			left, right = c, a
		}
		g := left.C1 + 1
		lo, hi := max(left.R0, right.R0)+1, min(left.R1, right.R1)-2
		if hi < lo {
			return nil, fmt.Errorf("rooms %d and %d do not overlap", ai, bi)
		}
		row := lo + b.rng.Intn(hi-lo+1)
		d.Def = barrierVertical
		b.p.Field.Set(row, g, levelRoom)
		b.p.Field.Set(row+1, g, levelRoom)
		d.Corridor = [2]Cell{{row, g - 1}, {row, g}}
		leftSide, rightSide := Cell{row, g - 2}, Cell{row, g + 1}
		if left == a {
			d.Side = [2]Cell{leftSide, rightSide}
		} else {
			d.Side = [2]Cell{rightSide, leftSide}
		}
	case a.R1+2 == c.R0 || c.R1+2 == a.R0: // one above the other
		top, bottom := a, c
		if c.R1+2 == a.R0 {
			top, bottom = c, a
		}
		g := top.R1 + 1
		lo, hi := max(top.C0, bottom.C0)+1, min(top.C1, bottom.C1)-2
		if hi < lo {
			return nil, fmt.Errorf("rooms %d and %d do not overlap", ai, bi)
		}
		col := lo + b.rng.Intn(hi-lo+1)
		d.Def = barrierHorizontal
		b.p.Field.Set(g, col, levelRoom)
		b.p.Field.Set(g, col+1, levelRoom)
		d.Corridor = [2]Cell{{g - 1, col}, {g, col}}
		topSide, bottomSide := Cell{g - 2, col}, Cell{g + 1, col}
		if top == a {
			d.Side = [2]Cell{topSide, bottomSide}
		} else {
			d.Side = [2]Cell{bottomSide, topSide}
		}
	default:
		return nil, fmt.Errorf("rooms %d and %d are not neighbours", ai, bi)
	}
	d.Cell = d.Corridor[0]
	d.Room = ai
	for _, cc := range d.Corridor {
		b.p.Walk[cc] = true
		b.reserved[cc] = true
	}
	// Keep the doorway approaches clear on both sides.
	for _, s := range d.Side {
		b.reserved[s] = true
		for _, dd := range dirs {
			b.reserved[s.add(dd)] = true
		}
	}
	b.used[d.Cell] = true
	return d, nil
}

func (b *builder) place(e *Ent) {
	b.p.Ents = append(b.p.Ents, e)
	b.used[e.Cell] = true
}

// free picks an unused, unreserved interior cell of room, trying prefer first.
func (b *builder) free(room int, prefer []Cell) (Cell, bool) {
	ok := func(c Cell) bool { return b.p.Rooms[room].contains(c) && !b.used[c] && !b.reserved[c] }
	for _, c := range prefer {
		if ok(c) {
			return c, true
		}
	}
	cells := b.p.Rooms[room].cells()
	b.rng.Shuffle(len(cells), func(i, j int) { cells[i], cells[j] = cells[j], cells[i] })
	for _, c := range cells {
		if ok(c) {
			return c, true
		}
	}
	return Cell{}, false
}

// inner prefers cells away from the room's walls (blocks need space to push).
func (b *builder) inner(room int) []Cell {
	r := b.p.Rooms[room]
	var out []Cell
	for _, c := range r.cells() {
		if c.R > r.R0 && c.R < r.R1-1 && c.C > r.C0 && c.C < r.C1-1 {
			out = append(out, c)
		}
	}
	b.rng.Shuffle(len(out), func(i, j int) { out[i], out[j] = out[j], out[i] })
	return out
}

// topRowFirst lists the room's top interior row (levers stand along the north
// wall), shuffled.
func (b *builder) topRowFirst(room int) []Cell {
	r := b.p.Rooms[room]
	var out []Cell
	for col := r.C0; col < r.C1; col++ {
		out = append(out, Cell{r.R0, col})
	}
	b.rng.Shuffle(len(out), func(i, j int) { out[i], out[j] = out[j], out[i] })
	return out
}

// farthest is the free interior cell of room farthest (by steps) from start.
func (b *builder) farthest(room int, start Cell) Cell {
	dist := bfs(b.p.Walk, start, func(c Cell) bool { return !b.p.Rooms[room].contains(c) && c != start })
	best, bestD := start, -1
	cells := b.p.Rooms[room].cells()
	sort.Slice(cells, func(i, j int) bool { return cells[i].R*100+cells[i].C < cells[j].R*100+cells[j].C })
	for _, c := range cells {
		if d, ok := dist[c]; ok && d > bestD && !b.used[c] && !b.reserved[c] {
			best, bestD = c, d
		}
	}
	return best
}

func (b *builder) lock(i int, lock Lock) error {
	door := b.p.Doors[i]
	for _, m := range lock.Mechanics {
		switch m {
		case Lever:
			n := max(lock.Levers, 1)
			var levers []*Ent
			for k := 0; k < n; k++ {
				c, ok := b.free(i, b.topRowFirst(i))
				if !ok {
					return fmt.Errorf("no space for lever %d in room %d", k, i)
				}
				e := &Ent{ID: b.id("lever"), Kind: KLever, Cell: c, Room: i}
				b.place(e)
				levers = append(levers, e)
			}
			// Required positions: random, at least one away from the start.
			want := make([]bool, n)
			for k := range want {
				want[k] = b.rng.Intn(2) == 0
			}
			want[b.rng.Intn(n)] = true
			sort.Slice(levers, func(x, y int) bool {
				if levers[x].Cell.R != levers[y].Cell.R {
					return levers[x].Cell.R < levers[y].Cell.R
				}
				return levers[x].Cell.C < levers[y].Cell.C
			})
			var words []string
			for k, e := range levers {
				st, w := "flickedLeft", "left"
				if want[k] {
					st, w = "flickedRight", "right"
				}
				door.Conds = append(door.Conds, Cond{e.ID, st})
				words = append(words, w)
			}
			if n > 1 {
				hint := "Levers, top row first, left to right: " + strings.Join(words, ", ") + "."
				if lock.HintChest {
					if err := b.keyAndChest(i, false, "", "A note: "+hint); err != nil {
						return err
					}
				} else if c, ok := b.free(i, b.topRowFirst(i)); ok {
					b.place(&Ent{ID: b.id("sign"), Kind: KSign, Cell: c, Room: i, Msg: hint})
				}
			}
		case KeyChest:
			keyRoom := i
			if lock.KeyEarly && i > 0 {
				keyRoom = b.rng.Intn(i)
			}
			chest, err := b.keyAndChestIn(i, keyRoom, "Click! Somewhere a door unlocks.")
			if err != nil {
				return err
			}
			door.Conds = append(door.Conds, Cond{chest.ID, "open"})
		case PushBlock:
			n := max(lock.Blocks, 1)
			for k := 0; k < n; k++ {
				plateCell, ok := b.free(i, b.inner(i))
				if !ok {
					return fmt.Errorf("no space for a plate in room %d", i)
				}
				plate := &Ent{ID: b.id("plate"), Kind: KPlate, Cell: plateCell, Room: i}
				b.place(plate)
				blockCell, ok := b.free(i, b.inner(i))
				if !ok {
					return fmt.Errorf("no space for a block in room %d", i)
				}
				b.place(&Ent{ID: b.id("block"), Kind: KBlock, Cell: blockCell, Room: i})
				door.Conds = append(door.Conds, Cond{plate.ID, "pressed"})
			}
			if c, ok := b.free(i, b.topRowFirst(i)); ok {
				b.place(&Ent{ID: b.id("sign"), Kind: KSign, Cell: c, Room: i, Msg: "Heavy blocks keep pressure plates down."})
			}
		case TimedSwitch:
			c := b.farthest(i, door.Side[0])
			if b.used[c] || b.reserved[c] {
				return fmt.Errorf("no space for a timed lever in room %d", i)
			}
			e := &Ent{ID: b.id("timer"), Kind: KTimedLever, Cell: c, Room: i, Door: i}
			b.place(e)
			door.Conds = append(door.Conds, Cond{e.ID, "flickedRight"})
			if c, ok := b.free(i, b.topRowFirst(i)); ok {
				b.place(&Ent{ID: b.id("sign"), Kind: KSign, Cell: c, Room: i, Msg: "Quick! That door won't stay open for long."})
			}
		}
	}
	return nil
}

// keyAndChest places a pot hiding a key (in keyRoom) and a locked chest in room.
func (b *builder) keyAndChest(room int, early bool, loot, msg string) error {
	keyRoom := room
	if early && room > 0 {
		keyRoom = b.rng.Intn(room)
	}
	chest, err := b.keyAndChestIn(room, keyRoom, msg)
	if chest != nil {
		chest.Loot = loot
	}
	return err
}

func (b *builder) keyAndChestIn(room, keyRoom int, msg string) (*Ent, error) {
	potCell, ok := b.free(keyRoom, nil)
	if !ok {
		return nil, fmt.Errorf("no space for a pot in room %d", keyRoom)
	}
	b.place(&Ent{ID: b.id("pot"), Kind: KPot, Cell: potCell, Room: keyRoom, Loot: keyItem})
	chestCell, ok := b.free(room, nil)
	if !ok {
		return nil, fmt.Errorf("no space for a chest in room %d", room)
	}
	chest := &Ent{ID: b.id("chest"), Kind: KChest, Cell: chestCell, Room: room, Requires: keyItem, Msg: msg}
	b.place(chest)
	return chest, nil
}

// bfs returns step distances from start over walk, never entering cells for
// which blocked is true.
func bfs(walk map[Cell]bool, start Cell, blocked func(Cell) bool) map[Cell]int {
	dist := map[Cell]int{start: 0}
	queue := []Cell{start}
	for len(queue) > 0 {
		c := queue[0]
		queue = queue[1:]
		for _, d := range dirs {
			n := c.add(d)
			if _, seen := dist[n]; seen || !walk[n] || blocked(n) {
				continue
			}
			dist[n] = dist[c] + 1
			queue = append(queue, n)
		}
	}
	return dist
}

// Player speed on the server: 6 px per move at 15 moves a second.
const playerPxPerSecond = 6 * 15

// timerMs is the window for a run of steps tiles, with slack and a moment to react.
func timerMs(steps int, slack float64) int {
	return int(math.Ceil(float64(steps)*16/playerPxPerSecond*1000*slack)) + 600
}

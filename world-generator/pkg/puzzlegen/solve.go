package puzzlegen

import (
	"fmt"
)

// Action is one step of a solution: stand on Stand, then do something.
type Action struct {
	Kind   string // "toggle", "break", "open", "push", "pull", "timed", "exit"
	Entity string // entity ID acted on
	Stand  Cell   // where the player stands to do it
	Dir    Cell   // the way the block moves one tile (push or pull)
	To     Cell   // where the player ends up (push: the block's old cell; pull: a tile back; timed: the far side)
}

func (a Action) String() string {
	switch a.Kind {
	case "push":
		return fmt.Sprintf("push %s %s from (%d,%d)", a.Entity, dirName(a.Dir), a.Stand.R, a.Stand.C)
	case "pull":
		return fmt.Sprintf("pull %s %s from (%d,%d)", a.Entity, dirName(a.Dir), a.Stand.R, a.Stand.C)
	case "timed":
		return fmt.Sprintf("flip %s at (%d,%d) and run to (%d,%d)", a.Entity, a.Stand.R, a.Stand.C, a.To.R, a.To.C)
	case "exit":
		return fmt.Sprintf("walk to the exit at (%d,%d)", a.Stand.R, a.Stand.C)
	}
	return fmt.Sprintf("%s %s from (%d,%d)", a.Kind, a.Entity, a.Stand.R, a.Stand.C)
}

func dirName(d Cell) string {
	switch d {
	case Cell{-1, 0}:
		return "up"
	case Cell{1, 0}:
		return "down"
	case Cell{0, -1}:
		return "left"
	}
	return "right"
}

// Solution is the solver's verdict.
type Solution struct {
	Solvable bool
	Actions  []Action
	States   int // states explored
}

const (
	maxStates = 400000
	nCells    = mapSize * mapSize
	maxBlocks = 8
	noCell    = 0xff
)

func idx(c Cell) int    { return c.R*mapSize + c.C }
func cellOf(i int) Cell { return Cell{i / mapSize, i % mapSize} }

func inMap(c Cell) bool { return c.R >= 0 && c.R < mapSize && c.C >= 0 && c.C < mapSize }

// state is everything that changes while solving. It is comparable, so it
// is its own map key.
type state struct {
	player uint8  // canonical: the smallest reachable cell index
	levers uint32 // bit set = flickedRight
	broken uint32 // key pots broken
	opened uint32 // chests opened
	keys   int8
	blocks [maxBlocks]uint8
}

// condKind says what a door condition looks at.
type condKind uint8

const (
	condLever condKind = iota
	condChest
	condPlate
)

type cond struct {
	kind condKind
	i    int   // lever / chest index
	cell uint8 // plate cell
	want bool  // flickedRight / open / pressed
}

// leverGroup is the ordinary levers one door depends on. Doors only ever
// need opening, so the solver sets a whole group to its combination in one
// move rather than searching every combination (and never touches decoys).
type leverGroup struct {
	door int
	mask uint32
	want uint32
	ids  []int
}

type solver struct {
	p       *Puzzle
	walk    [nCells]bool
	room    [nCells]int8 // room index, -1 outside rooms (corridors)
	static  [nCells]bool // never passable: off the walk, levers, chests, signs, decoy pots
	levers  []*Ent       // ordinary and timed levers (bit index = position)
	pots    []*Ent       // pots that hide keys
	chests  []*Ent
	blocks  []*Ent
	doors   [][]cond
	groups  []leverGroup
	corr    [][2]uint8 // door corridor cells
	exit    int
	scratch [2][nCells]int16
	queue   []uint8
}

func newSolver(p *Puzzle) *solver {
	s := &solver{p: p, exit: idx(p.Exit), queue: make([]uint8, 0, nCells)}
	for i := range s.room {
		s.room[i] = -1
	}
	for c, ok := range p.Walk {
		if ok && inMap(c) {
			s.walk[idx(c)] = true
		}
	}
	for ri, r := range p.Rooms {
		for _, c := range r.cells() {
			s.room[idx(c)] = int8(ri)
		}
	}
	for i := range s.static {
		s.static[i] = !s.walk[i]
	}
	byID := map[string]cond{}
	for _, e := range p.Ents {
		at := idx(e.Cell)
		switch e.Kind {
		case KLever, KTimedLever:
			byID[e.ID] = cond{kind: condLever, i: len(s.levers)}
			s.levers = append(s.levers, e)
			s.static[at] = true
		case KPot:
			if e.Loot == keyItem {
				s.pots = append(s.pots, e)
			} else {
				s.static[at] = true // breaking a decoy never helps
			}
		case KChest:
			byID[e.ID] = cond{kind: condChest, i: len(s.chests)}
			s.chests = append(s.chests, e)
			s.static[at] = true
		case KBlock:
			s.blocks = append(s.blocks, e)
		case KSign:
			s.static[at] = true
		case KEntry:
			// Stepping back onto it abandons the rift: never walk over it
			// (the player starts there; the flood still leaves from it).
			s.static[at] = true
		case KPlate:
			byID[e.ID] = cond{kind: condPlate, cell: uint8(at)}
		}
	}
	for di, d := range p.Doors {
		var cs []cond
		g := leverGroup{door: di}
		for _, c := range d.Conds {
			k, ok := byID[c.ID]
			if !ok {
				// Unknown entity: the door can never open.
				cs = append(cs, cond{kind: condLever, i: -1})
				continue
			}
			switch k.kind {
			case condLever:
				k.want = c.State == "flickedRight"
				if s.levers[k.i].Kind == KLever {
					g.mask |= 1 << k.i
					if k.want {
						g.want |= 1 << k.i
					}
					g.ids = append(g.ids, k.i)
				}
			case condChest:
				k.want = c.State == "open"
			case condPlate:
				k.want = c.State == "pressed"
			}
			cs = append(cs, k)
		}
		s.doors = append(s.doors, cs)
		if g.mask != 0 {
			s.groups = append(s.groups, g)
		}
		s.corr = append(s.corr, [2]uint8{uint8(idx(d.Corridor[0])), uint8(idx(d.Corridor[1]))})
	}
	return s
}

func (s *solver) doorOpen(di int, st *state) bool {
	for _, c := range s.doors[di] {
		var holds bool
		switch c.kind {
		case condLever:
			if c.i < 0 {
				return false
			}
			holds = st.levers&(1<<c.i) != 0
		case condChest:
			holds = st.opened&(1<<c.i) != 0
		case condPlate:
			for bi := range s.blocks {
				if st.blocks[bi] == c.cell {
					holds = true
				}
			}
		}
		if holds != c.want {
			return false
		}
	}
	return true
}

// solidMap marks every cell the player (or a block) cannot enter in st.
func (s *solver) solidMap(st *state, out *[nCells]bool) {
	*out = s.static
	for i, pot := range s.pots {
		if st.broken&(1<<i) == 0 {
			out[idx(pot.Cell)] = true
		}
	}
	for bi := range s.blocks {
		out[st.blocks[bi]] = true
	}
	for di := range s.doors {
		if !s.doorOpen(di, st) {
			out[s.corr[di][0]] = true
			out[s.corr[di][1]] = true
		}
	}
}

// flood fills dist (-1 = unreachable) with step distances from start.
func (s *solver) flood(start int, solid *[nCells]bool, dist *[nCells]int16) {
	for i := range dist {
		dist[i] = -1
	}
	dist[start] = 0
	q := append(s.queue[:0], uint8(start))
	for h := 0; h < len(q); h++ {
		c := int(q[h])
		r, col := c/mapSize, c%mapSize
		for _, d := range dirs {
			nr, nc := r+d.R, col+d.C
			if nr < 0 || nr >= mapSize || nc < 0 || nc >= mapSize {
				continue
			}
			n := nr*mapSize + nc
			if dist[n] >= 0 || solid[n] {
				continue
			}
			dist[n] = dist[c] + 1
			q = append(q, uint8(n))
		}
	}
	s.queue = q
}

// canonical sets st.player to the smallest cell reachable from from.
func (s *solver) canonical(st *state, from int) {
	var solid [nCells]bool
	s.solidMap(st, &solid)
	dist := &s.scratch[1]
	s.flood(from, &solid, dist)
	for i := 0; i < nCells; i++ {
		if dist[i] >= 0 {
			st.player = uint8(i)
			return
		}
	}
	st.player = uint8(from)
}

// stand is the reachable neighbour of target nearest the player.
func stand(dist *[nCells]int16, target Cell) (Cell, bool) {
	best, found := Cell{}, false
	for _, d := range dirs {
		n := target.add(d)
		if !inMap(n) || dist[idx(n)] < 0 {
			continue
		}
		if !found || dist[idx(n)] < dist[idx(best)] {
			best, found = n, true
		}
	}
	return best, found
}

type node struct {
	st     state
	parent int32
	acts   []Action
}

// Solve searches for a way from the spawn to the exit, breadth-first over
// moves (a lever group set in one go counts as one move), so the solution
// is short; its Actions spell out every lever flip. It may pull blocks (the
// use button), e.g. out of a corner they were pushed into.
func Solve(p *Puzzle) Solution { return solve(p, true) }

// solvePushOnly is Solve without pulls: a smaller search, used to verify
// generated layouts (one solvable by pushing is solvable with pulls too).
func solvePushOnly(p *Puzzle) Solution { return solve(p, false) }

func solve(p *Puzzle, pulls bool) Solution {
	s := newSolver(p)
	if len(s.blocks) > maxBlocks || len(s.levers) > 32 || len(s.pots) > 32 || len(s.chests) > 32 {
		return Solution{}
	}
	var start state
	for i := range start.blocks {
		start.blocks[i] = noCell
	}
	for bi, b := range s.blocks {
		start.blocks[bi] = uint8(idx(b.Cell))
	}
	s.canonical(&start, idx(p.Spawn))

	nodes := []node{{st: start, parent: -1}}
	seen := map[state]bool{start: true}
	var solid [nCells]bool
	dist := &s.scratch[0]
	for i := 0; i < len(nodes); i++ {
		if len(nodes) > maxStates {
			return Solution{States: len(nodes)}
		}
		cur := nodes[i].st
		s.solidMap(&cur, &solid)
		s.flood(int(cur.player), &solid, dist)
		if dist[s.exit] >= 0 {
			acts := []Action{{Kind: "exit", Entity: "exit", Stand: p.Exit, To: p.Exit}}
			for j := i; nodes[j].parent >= 0; j = int(nodes[j].parent) {
				acts = append(append([]Action(nil), nodes[j].acts...), acts...)
			}
			return Solution{Solvable: true, Actions: acts, States: len(nodes)}
		}
		push := func(next state, from Cell, acts ...Action) {
			s.canonical(&next, idx(from))
			if seen[next] {
				return
			}
			seen[next] = true
			nodes = append(nodes, node{st: next, parent: int32(i), acts: acts})
		}

		for _, g := range s.groups {
			if cur.levers&g.mask == g.want {
				continue
			}
			var acts []Action
			var at Cell
			ok := true
			for _, li := range g.ids {
				if (cur.levers>>li)&1 == (g.want>>li)&1 {
					continue
				}
				c, found := stand(dist, s.levers[li].Cell)
				if !found {
					ok = false
					break
				}
				acts = append(acts, Action{Kind: "toggle", Entity: s.levers[li].ID, Stand: c, To: c})
				at = c
			}
			if !ok {
				continue
			}
			next := cur
			next.levers = cur.levers&^g.mask | g.want
			push(next, at, acts...)
		}
		for li, lv := range s.levers {
			if lv.Kind != KTimedLever {
				continue
			}
			if c, ok := stand(dist, lv.Cell); ok {
				s.timed(&cur, lv, li, c, push)
			}
		}
		for pi, pot := range s.pots {
			if cur.broken&(1<<pi) != 0 {
				continue
			}
			if c, ok := stand(dist, pot.Cell); ok {
				next := cur
				next.broken |= 1 << pi
				next.keys++
				push(next, c, Action{Kind: "break", Entity: pot.ID, Stand: c, To: c})
			}
		}
		for ci, ch := range s.chests {
			if cur.opened&(1<<ci) != 0 || (ch.Requires != "" && cur.keys == 0) {
				continue
			}
			if c, ok := stand(dist, ch.Cell); ok {
				next := cur
				next.opened |= 1 << ci
				if ch.Requires != "" {
					next.keys--
				}
				push(next, c, Action{Kind: "open", Entity: ch.ID, Stand: c, To: c})
			}
		}
		for bi := range s.blocks {
			bc := cellOf(int(cur.blocks[bi]))
			for _, d := range dirs {
				from := Cell{bc.R - d.R, bc.C - d.C}
				to := bc.add(d)
				if !inMap(from) || !inMap(to) || dist[idx(from)] < 0 {
					continue
				}
				// Blocks stay inside rooms (never in a doorway).
				if s.room[idx(to)] < 0 || solid[idx(to)] {
					continue
				}
				next := cur
				next.blocks[bi] = uint8(idx(to))
				push(next, bc, Action{Kind: "push", Entity: s.blocks[bi].ID, Stand: from, Dir: d, To: bc})
			}
			// Pull (the use button): standing at bc+d, the player steps
			// back to bc+2d and the block follows into bc+d.
			for _, d := range dirs {
				if !pulls {
					break
				}
				at, back := bc.add(d), bc.add(d).add(d)
				if !inMap(at) || !inMap(back) || dist[idx(at)] < 0 || solid[idx(back)] {
					continue
				}
				if s.room[idx(at)] < 0 {
					continue
				}
				next := cur
				next.blocks[bi] = uint8(idx(at))
				push(next, back, Action{Kind: "pull", Entity: s.blocks[bi].ID, Stand: at, Dir: d, To: back})
			}
		}
	}
	return Solution{States: len(nodes)}
}

// timed: flipping a spring lever opens its door for a while; the run counts
// only if the far side is reached within the lever's step budget. The lever
// has sprung back by the time anything else is done.
func (s *solver) timed(cur *state, lv *Ent, li int, at Cell, push func(state, Cell, ...Action)) {
	flipped := *cur
	flipped.levers |= 1 << li
	if !s.doorOpen(lv.Door, &flipped) {
		return
	}
	var solid [nCells]bool
	s.solidMap(&flipped, &solid)
	dist := &s.scratch[1]
	s.flood(idx(at), &solid, dist)
	door := s.p.Doors[lv.Door]
	// Only a run that crosses the doorway is useful.
	var reach []Cell
	for _, far := range door.Side {
		d := dist[idx(far)]
		if d < 0 || int(d) > lv.budget || s.room[idx(far)] == s.room[idx(at)] {
			continue
		}
		reach = append(reach, far)
	}
	for _, far := range reach {
		push(*cur, far, Action{Kind: "timed", Entity: lv.ID, Stand: at, To: far})
	}
}

// Shortcut reports a way round the intended chain: a door that is already
// open at the start, or a door the exit can be reached without (with every
// other door open, every pot broken and no blocks in the way).
func Shortcut(p *Puzzle) error {
	s := newSolver(p)
	var start state
	for i := range start.blocks {
		start.blocks[i] = noCell
	}
	for bi, b := range s.blocks {
		start.blocks[bi] = uint8(idx(b.Cell))
	}
	for di, d := range p.Doors {
		if s.doorOpen(di, &start) {
			return fmt.Errorf("%s is open at the start", d.ID)
		}
	}
	var dist [nCells]int16
	for di, d := range p.Doors {
		solid := s.static
		for _, pot := range s.pots {
			solid[idx(pot.Cell)] = false
		}
		solid[s.corr[di][0]] = true
		solid[s.corr[di][1]] = true
		s.flood(idx(p.Spawn), &solid, &dist)
		if dist[s.exit] >= 0 {
			return fmt.Errorf("the exit can be reached without passing %s", d.ID)
		}
	}
	return nil
}

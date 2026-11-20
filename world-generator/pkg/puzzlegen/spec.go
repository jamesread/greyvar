// Package puzzlegen generates small puzzle worlds: rooms joined by doors,
// each door locked by one or more mechanics whose solution lies in the rooms
// before it. Every generated puzzle is checked by a solver before it is
// written, so a puzzle that cannot be finished is never handed out.
package puzzlegen

import (
	"fmt"
	"math/rand"
)

// Mechanic is one kind of lock a door can have.
type Mechanic string

const (
	// Lever: wall levers that must be set (a combination when more than one).
	Lever Mechanic = "lever"
	// KeyChest: a pot hides a key that opens a chest; the door opens with the chest.
	KeyChest Mechanic = "keyChest"
	// PushBlock: push blocks onto pressure plates to hold the door open.
	PushBlock Mechanic = "pushBlock"
	// TimedSwitch: a lever that springs back; reach the door before it does.
	TimedSwitch Mechanic = "timedSwitch"
)

// AllMechanics in order of introduction.
var AllMechanics = []Mechanic{Lever, KeyChest, PushBlock, TimedSwitch}

// Lock is what one door needs; all of its mechanics must hold (AND).
type Lock struct {
	Mechanics []Mechanic
	Levers    int  // levers in a Lever combination (default 1)
	Blocks    int  // blocks/plates for PushBlock (default 1)
	KeyEarly  bool // KeyChest: hide the key in an earlier room (backtracking)
	HintChest bool // Lever: the combination is written on a note in a locked chest
}

// Spec describes the puzzle to generate.
type Spec struct {
	Level       int
	Seed        int64
	Rooms       int    // 2..4
	Locks       []Lock // one per door: len(Locks) == Rooms-1
	RedHerrings int    // decoy pots and levers that do nothing useful
	// TimerSlack scales the timed-switch window over the shortest run
	// (1.6 is relaxed, 1.2 is tight).
	TimerSlack float64
}

// Validate checks the spec is something the generator can lay out.
func (s Spec) Validate() error {
	if s.Rooms < 2 || s.Rooms > 4 {
		return fmt.Errorf("rooms must be 2..4, got %d", s.Rooms)
	}
	if len(s.Locks) != s.Rooms-1 {
		return fmt.Errorf("need %d locks for %d rooms, got %d", s.Rooms-1, s.Rooms, len(s.Locks))
	}
	for i, l := range s.Locks {
		if len(l.Mechanics) == 0 {
			return fmt.Errorf("lock %d has no mechanics", i)
		}
		for _, m := range l.Mechanics {
			if !known(m) {
				return fmt.Errorf("lock %d: unknown mechanic %q", i, m)
			}
		}
		if l.Levers > 4 || l.Blocks > 2 {
			return fmt.Errorf("lock %d: at most 4 levers and 2 blocks", i)
		}
	}
	return nil
}

func known(m Mechanic) bool {
	for _, k := range AllMechanics {
		if k == m {
			return true
		}
	}
	return false
}

// ParseMechanics parses "lever,pushBlock" (empty means all).
func ParseMechanics(list []string) ([]Mechanic, error) {
	if len(list) == 0 {
		return AllMechanics, nil
	}
	var out []Mechanic
	for _, s := range list {
		m := Mechanic(s)
		if !known(m) {
			return nil, fmt.Errorf("unknown mechanic %q (have %v)", s, AllMechanics)
		}
		out = append(out, m)
	}
	return out, nil
}

// Curve is the default difficulty ladder: each level adds rooms, longer
// lever combinations, backtracking, more blocks, tighter timers and decoys.
// allowed restricts the mechanics used (nil = all); a lock whose mechanic is
// not allowed falls back to the first allowed one.
func Curve(level int, seed int64, allowed []Mechanic) Spec {
	if level < 1 {
		level = 1
	}
	s := Spec{Level: level, Seed: seed, TimerSlack: 1.6}
	switch level {
	case 1:
		s.Rooms = 2
		s.Locks = []Lock{{Mechanics: []Mechanic{Lever}, Levers: 1}}
	case 2:
		s.Rooms = 2
		s.Locks = []Lock{{Mechanics: []Mechanic{KeyChest}}}
	case 3:
		s.Rooms = 2
		s.Locks = []Lock{{Mechanics: []Mechanic{PushBlock}, Blocks: 1}}
	case 4:
		s.Rooms = 3
		s.Locks = []Lock{
			{Mechanics: []Mechanic{Lever}, Levers: 3},
			{Mechanics: []Mechanic{KeyChest}},
		}
	case 5:
		s.Rooms = 3
		s.Locks = []Lock{
			{Mechanics: []Mechanic{PushBlock}, Blocks: 1},
			{Mechanics: []Mechanic{TimedSwitch}},
		}
		s.RedHerrings = 1
	case 6:
		s.Rooms = 4
		s.Locks = []Lock{
			{Mechanics: []Mechanic{Lever}, Levers: 3, HintChest: true},
			{Mechanics: []Mechanic{PushBlock}, Blocks: 1},
			{Mechanics: []Mechanic{KeyChest}, KeyEarly: true},
		}
		s.RedHerrings = 1
		s.TimerSlack = 1.5
	case 7:
		s.Rooms = 4
		s.Locks = []Lock{
			{Mechanics: []Mechanic{KeyChest, Lever}, Levers: 2},
			{Mechanics: []Mechanic{PushBlock}, Blocks: 2},
			{Mechanics: []Mechanic{TimedSwitch}},
		}
		s.RedHerrings = 2
		s.TimerSlack = 1.4
	case 8:
		s.Rooms = 4
		s.Locks = []Lock{
			{Mechanics: []Mechanic{Lever, KeyChest}, Levers: 3, HintChest: true},
			{Mechanics: []Mechanic{PushBlock, Lever}, Blocks: 2, Levers: 2},
			{Mechanics: []Mechanic{TimedSwitch, KeyChest}, KeyEarly: true},
		}
		s.RedHerrings = 2
		s.TimerSlack = 1.35
	default:
		s = endless(level, seed)
	}
	if allowed != nil {
		for i := range s.Locks {
			s.Locks[i].Mechanics = restrict(s.Locks[i].Mechanics, allowed)
			if s.Locks[i].Levers == 0 {
				s.Locks[i].Levers = 1
			}
			if s.Locks[i].Blocks == 0 {
				s.Locks[i].Blocks = 1
			}
		}
	}
	return s
}

// endless is the curve past the hand-made levels: four rooms whose locks are
// drawn from the seed, each pairing two mechanics, with longer combinations,
// more blocks, more backtracking, more decoys and tighter timers as the
// level rises. One lock (at most) gets two blocks, which keeps the solver's
// search small.
func endless(level int, seed int64) Spec {
	rng := rand.New(rand.NewSource(seed*104729 + int64(level)))
	over := level - 8
	s := Spec{Level: level, Seed: seed, Rooms: 4}
	s.RedHerrings = min(2+over/2, 4)
	s.TimerSlack = max(1.2, 1.35-0.03*float64(over))
	levers := min(2+over/2, 4)
	twoBlocks := level >= 10
	// Every puzzle uses all four mechanics somewhere; which door gets which
	// pairing is up to the seed.
	pairs := [][]Mechanic{
		{Lever, KeyChest},
		{PushBlock, Lever},
		{TimedSwitch, KeyChest},
	}
	rng.Shuffle(len(pairs), func(i, j int) { pairs[i], pairs[j] = pairs[j], pairs[i] })
	for i, ms := range pairs {
		l := Lock{Mechanics: ms, Levers: levers, Blocks: 1}
		for _, m := range ms {
			switch m {
			case Lever:
				l.HintChest = rng.Intn(2) == 0
			case KeyChest:
				l.KeyEarly = i > 0 && rng.Intn(3) > 0
			case PushBlock:
				if twoBlocks {
					l.Blocks = 2
				}
			}
		}
		s.Locks = append(s.Locks, l)
	}
	return s
}

func restrict(ms []Mechanic, allowed []Mechanic) []Mechanic {
	ok := map[Mechanic]bool{}
	for _, a := range allowed {
		ok[a] = true
	}
	var out []Mechanic
	seen := map[Mechanic]bool{}
	for _, m := range ms {
		if ok[m] && !seen[m] {
			out = append(out, m)
			seen[m] = true
		}
	}
	if len(out) == 0 && len(allowed) > 0 {
		out = []Mechanic{allowed[0]}
	}
	return out
}

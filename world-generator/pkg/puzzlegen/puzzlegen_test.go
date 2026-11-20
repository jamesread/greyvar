package puzzlegen

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/jamesread/greyvar/datlib/worldfiles"
)

const (
	repoTilesets = "../../../res/img/textures/tilesets"
	repoEntdefs  = "../../../server/dat/entdefs"
)

func seedsFor(t *testing.T) int64 {
	if testing.Short() {
		return 5
	}
	return 25
}

// Every level of the curve generates, for many seeds, quickly enough for the
// server to do it on demand, and the verified solution reaches the exit.
func TestCurveIsSolvable(t *testing.T) {
	seeds := seedsFor(t)
	for level := 1; level <= 16; level++ {
		var slowest time.Duration
		for seed := int64(1); seed <= seeds; seed++ {
			t0 := time.Now()
			res, err := Build(Curve(level, seed, nil), Options{})
			if d := time.Since(t0); d > slowest {
				slowest = d
			}
			if err != nil {
				t.Fatalf("level %d seed %d: %v", level, seed, err)
			}
			acts := res.Solution.Actions
			if !res.Solution.Solvable || len(acts) == 0 || acts[len(acts)-1].Kind != "exit" {
				t.Fatalf("level %d seed %d: bad solution %v", level, seed, acts)
			}
		}
		if slowest > 2*time.Second {
			t.Errorf("level %d: slowest build took %v", level, slowest)
		}
	}
}

// Puzzles get longer as the level rises (on average over seeds).
func TestDifficultyRises(t *testing.T) {
	seeds := seedsFor(t)
	avg := func(level int) float64 {
		total := 0
		for seed := int64(1); seed <= seeds; seed++ {
			res, err := Build(Curve(level, seed, nil), Options{})
			if err != nil {
				t.Fatalf("level %d seed %d: %v", level, seed, err)
			}
			total += len(res.Solution.Actions) - 1
		}
		return float64(total) / float64(seeds)
	}
	bands := [][]int{{1, 2}, {4, 5}, {6, 7}, {9, 12}, {14, 16}}
	prev := 0.0
	for _, band := range bands {
		sum := 0.0
		for _, l := range band {
			sum += avg(l)
		}
		mean := sum / float64(len(band))
		t.Logf("levels %v: %.1f actions on average", band, mean)
		if mean <= prev {
			t.Errorf("levels %v average %.1f actions, not more than the band before (%.1f)", band, mean, prev)
		}
		prev = mean
	}
}

func TestBuildIsDeterministic(t *testing.T) {
	for _, level := range []int{1, 5, 8, 11} {
		a, err := Build(Curve(level, 42, nil), Options{})
		if err != nil {
			t.Fatal(err)
		}
		b, err := Build(Curve(level, 42, nil), Options{})
		if err != nil {
			t.Fatal(err)
		}
		if Describe(a) != Describe(b) {
			t.Fatalf("level %d: two builds with the same seed differ", level)
		}
	}
	a, _ := Build(Curve(8, 1, nil), Options{})
	b, _ := Build(Curve(8, 2, nil), Options{})
	if Describe(a) == Describe(b) {
		t.Error("different seeds should give different puzzles")
	}
}

// Past the hand-made levels the curve keeps varying with level and seed.
func TestEndlessLevelsVary(t *testing.T) {
	seen := map[string]bool{}
	for level := 9; level <= 14; level++ {
		res, err := Build(Curve(level, 7, nil), Options{})
		if err != nil {
			t.Fatal(err)
		}
		seen[Describe(res)[strings.Index(Describe(res), "\n"):]] = true
	}
	if len(seen) < 5 {
		t.Errorf("levels 9-14 with one seed gave only %d distinct puzzles", len(seen))
	}
}

func TestMechanicsCanBeRestricted(t *testing.T) {
	for level := 1; level <= 12; level++ {
		spec := Curve(level, 3, []Mechanic{Lever})
		res, err := Build(spec, Options{})
		if err != nil {
			t.Fatalf("level %d levers only: %v", level, err)
		}
		for _, e := range res.Puzzle.Ents {
			switch e.Kind {
			case KBlock, KPlate, KTimedLever:
				t.Fatalf("level %d levers only: placed a %s", level, e.Kind.Type())
			case KChest:
				if e.Msg == "" || !strings.HasPrefix(e.Msg, "A note") {
					t.Fatalf("level %d levers only: placed a key chest", level)
				}
			}
		}
	}
}

func TestSpecValidate(t *testing.T) {
	bad := []Spec{
		{Rooms: 1},
		{Rooms: 5},
		{Rooms: 2},
		{Rooms: 2, Locks: []Lock{{}}},
		{Rooms: 2, Locks: []Lock{{Mechanics: []Mechanic{"teleport"}}}},
		{Rooms: 2, Locks: []Lock{{Mechanics: []Mechanic{Lever}, Levers: 5}}},
	}
	for i, s := range bad {
		if s.Validate() == nil {
			t.Errorf("spec %d should be rejected: %+v", i, s)
		}
	}
	if _, err := ParseMechanics([]string{"lever", "nope"}); err == nil {
		t.Error("unknown mechanic should be rejected")
	}
}

func build(t *testing.T, level int, seed int64) *Puzzle {
	t.Helper()
	res, err := Build(Curve(level, seed, nil), Options{})
	if err != nil {
		t.Fatal(err)
	}
	return res.Puzzle
}

func find(p *Puzzle, k Kind) *Ent {
	for _, e := range p.Ents {
		if e.Kind == k {
			return e
		}
	}
	return nil
}

// A door that depends on something that does not exist never opens.
func TestSolverRejectsMissingLever(t *testing.T) {
	p := build(t, 1, 1)
	p.Doors[0].Conds = []Cond{{"leverThatIsNotThere", "flickedRight"}}
	if Solve(p).Solvable {
		t.Fatal("solvable with a door waiting on a missing lever")
	}
}

// A timed door whose window is too short cannot be beaten.
func TestSolverRespectsTimerBudget(t *testing.T) {
	p := build(t, 5, 1)
	timer := find(p, KTimedLever)
	if !Solve(p).Solvable {
		t.Fatal("level 5 should be solvable as generated")
	}
	timer.budget = 1
	if Solve(p).Solvable {
		t.Fatal("solvable with a 1-step timer window")
	}
}

// A key chest needs its key; take the key out of the pot and it is stuck.
func TestSolverNeedsTheKey(t *testing.T) {
	p := build(t, 2, 1)
	find(p, KPot).Loot = decoyItem
	if Solve(p).Solvable {
		t.Fatal("solvable without a key for the chest")
	}
}

func TestShortcutIsDetected(t *testing.T) {
	p := build(t, 1, 1)
	if err := Shortcut(p); err != nil {
		t.Fatalf("generated puzzle has a shortcut: %v", err)
	}
	// Want the lever where it already is: the door starts open.
	p.Doors[0].Conds[0].State = "flickedLeft"
	if Shortcut(p) == nil {
		t.Error("a door that starts open should count as a shortcut")
	}
	// Knock a hole in the wall between the rooms.
	p = build(t, 1, 1)
	d := p.Doors[0]
	for r := 0; r < mapSize; r++ {
		hole := Cell{r, d.Corridor[0].C}
		if r != d.Corridor[0].R && p.Walk[Cell{r, d.Side[0].C}] && p.Walk[Cell{r, d.Side[1].C}] {
			p.Walk[hole] = true
			p.Walk[Cell{r, d.Corridor[1].C}] = true
			break
		}
	}
	if Shortcut(p) == nil {
		t.Error("a second gap between the rooms should count as a shortcut")
	}
}

// Generate writes a world the server's loader reads, with every puzzle
// entity (and its wiring) intact.
func TestGenerateWritesLoadableWorld(t *testing.T) {
	for _, level := range []int{1, 4, 8, 12} {
		dir := t.TempDir()
		res, err := Generate(Curve(level, 9, nil), Options{
			WorldName:        "puzzle_test",
			OutDir:           dir,
			TilesetPath:      filepath.Join(repoTilesets, "grass_terrains.tsx"),
			DecorTilesetPath: filepath.Join(repoTilesets, "flora.tsx"),
			EntdefsDir:       repoEntdefs,
		})
		if err != nil {
			t.Fatalf("level %d: %v", level, err)
		}
		world, err := worldfiles.LoadWorldFromDir("puzzle_test", dir)
		if err != nil {
			t.Fatalf("level %d: world does not load: %v", level, err)
		}
		g := world.Grids[res.SpawnGrid]
		if g == nil {
			t.Fatalf("level %d: no spawn grid %q", level, res.SpawnGrid)
		}
		if len(g.SpawnPoints) != 1 || g.SpawnPoints[0].Row != uint32(res.Puzzle.Spawn.R) || g.SpawnPoints[0].Col != uint32(res.Puzzle.Spawn.C) {
			t.Errorf("level %d: spawn points %+v, want one at %v", level, g.SpawnPoints, res.Puzzle.Spawn)
		}
		want := map[string]*Ent{}
		for _, e := range res.Puzzle.Ents {
			if e.Kind != KSpawn {
				want[e.ID] = e
			}
		}
		for _, ge := range g.Entities {
			e := want[ge.Properties["id"]]
			if e == nil {
				t.Errorf("level %d: unexpected entity %s %+v", level, ge.Definition, ge.Properties)
				continue
			}
			delete(want, e.ID)
			if ge.Definition != e.Type() || ge.Row != uint32(e.Cell.R) || ge.Col != uint32(e.Cell.C) {
				t.Errorf("level %d: %s loaded as %s at %d,%d", level, e.ID, ge.Definition, ge.Row, ge.Col)
			}
			if e.Kind == KDoor && ge.Properties["open_when"] == "" {
				t.Errorf("level %d: %s has no open_when", level, e.ID)
			}
		}
		for id := range want {
			t.Errorf("level %d: %s is missing after loading (covered or dropped)", level, id)
		}
		raw, err := os.ReadFile(res.WorldPath)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(raw), `"puzzle_level"`) || !strings.Contains(string(raw), fmt.Sprintf(`"value": "%d"`, level)) {
			t.Errorf("level %d: world file lacks the puzzle_level property", level)
		}
	}
}

// Doorways get a barrier matching their passage: an upright one between
// rooms side by side, a horizontal one between rooms above each other.
func TestBarrierOrientationFollowsTheDoorway(t *testing.T) {
	seen := map[string]bool{}
	for seed := int64(1); seed <= 10; seed++ {
		p := build(t, 8, seed)
		for _, d := range p.Doors {
			sideBySide := d.Corridor[0].R == d.Corridor[1].R
			want := barrierHorizontal
			if sideBySide {
				want = barrierVertical
			}
			if d.Type() != want {
				t.Fatalf("seed %d: %s across corridor %v is %s, want %s", seed, d.ID, d.Corridor, d.Type(), want)
			}
			seen[d.Type()] = true
		}
	}
	if !seen[barrierVertical] || !seen[barrierHorizontal] {
		t.Fatalf("expected both orientations, saw %v", seen)
	}
}

// A block in a corner can be pulled out, so the puzzle stays solvable.
func TestSolverPullsBlocksOutOfCorners(t *testing.T) {
	p := build(t, 3, 1)
	block := find(p, KBlock)
	r := p.Rooms[block.Room]
	corner := Cell{r.R0, r.C0}
	for _, e := range p.Ents {
		if e != block && e.Cell == corner {
			e.Cell = block.Cell
		}
	}
	block.Cell = corner
	sol := Solve(p)
	if !sol.Solvable {
		t.Fatal("a cornered block can be pulled out; the puzzle should be solvable")
	}
	pulled := false
	for _, a := range sol.Actions {
		pulled = pulled || a.Kind == "pull"
	}
	if !pulled {
		t.Fatalf("the solution should pull the block: %v", sol.Actions)
	}
}

// The player starts on the entry teleporter (in a corner of the first
// room), solutions never walk back over it, and rift coins lie in the last
// room around the exit, more at higher levels.
func TestRiftEntryAndCoins(t *testing.T) {
	prevCoins := 0
	for _, level := range []int{1, 4, 8, 14} {
		res, err := Build(Curve(level, 3, nil), Options{})
		if err != nil {
			t.Fatal(err)
		}
		p := res.Puzzle
		entry := find(p, KEntry)
		if entry == nil || entry.Cell != p.Spawn || entry.Type() != "teleportPad" || toObject(entry).Props["rift_exit"] != "leave" {
			t.Fatalf("level %d: entry teleporter %+v at spawn %v", level, entry, p.Spawn)
		}
		if toObject(find(p, KExit)).Props["rift_exit"] != "complete" {
			t.Fatalf("level %d: the exit should complete the rift", level)
		}
		for _, a := range res.Solution.Actions {
			if a.Stand == p.Spawn {
				t.Fatalf("level %d: the solution stands on the entry teleporter: %s", level, a)
			}
		}
		last := len(p.Rooms) - 1
		coins := 0
		for _, e := range p.Ents {
			if e.Kind == KCoin {
				coins++
				if e.Room != last || !p.Rooms[last].contains(e.Cell) || e.Type() != "riftCoin" {
					t.Fatalf("level %d: coin %s at %v is not in the last room", level, e.ID, e.Cell)
				}
			}
		}
		if coins < 2 || coins < prevCoins {
			t.Fatalf("level %d: %d coins (previous level band %d)", level, coins, prevCoins)
		}
		prevCoins = coins
	}
}

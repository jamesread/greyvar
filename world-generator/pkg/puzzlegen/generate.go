package puzzlegen

import (
	"fmt"
	"math"
	"math/rand"
	"strconv"
	"strings"

	"github.com/jamesread/greyvar/world-generator/pkg/gridgen"
)

// Options says where and with which art a puzzle world is written.
type Options struct {
	WorldName        string
	OutDir           string
	TilesetPath      string // grass_terrains.tsx (wangset "terrains", colours grass + wall)
	DecorTilesetPath string // flora.tsx; empty for none
	EntdefsDir       string
	MaxAttempts      int // layouts tried before giving up (default 40)
	// MinActions rejects puzzles solvable in fewer actions (default: rooms-1).
	MinActions int
}

// Result describes a generated, verified puzzle.
type Result struct {
	Spec      Spec
	Puzzle    *Puzzle
	Solution  Solution
	Attempts  int
	WorldPath string
	SpawnGrid string
}

// Build lays out and verifies a puzzle without writing anything.
func Build(spec Spec, opts Options) (*Result, error) {
	if err := spec.Validate(); err != nil {
		return nil, err
	}
	attempts := opts.MaxAttempts
	if attempts <= 0 {
		attempts = 40
	}
	minActions := opts.MinActions
	if minActions <= 0 {
		minActions = spec.Rooms - 1
	}
	var lastErr error
	for a := 0; a < attempts; a++ {
		rng := rand.New(rand.NewSource(spec.Seed*7919 + int64(a)))
		p, err := layout(spec, rng)
		if err != nil {
			lastErr = err
			continue
		}
		setTimers(p, spec)
		if err := Shortcut(p); err != nil {
			lastErr = fmt.Errorf("layout %d has a shortcut: %v", a, err)
			continue
		}
		sol := solvePushOnly(p)
		if !sol.Solvable {
			lastErr = fmt.Errorf("layout %d is not solvable (%d states)", a, sol.States)
			continue
		}
		if len(sol.Actions)-1 < minActions {
			lastErr = fmt.Errorf("layout %d is too easy (%d actions)", a, len(sol.Actions)-1)
			continue
		}
		return &Result{Spec: spec, Puzzle: p, Solution: sol, Attempts: a + 1, SpawnGrid: "0.0.tmj"}, nil
	}
	return nil, fmt.Errorf("no solvable layout in %d attempts: %v", attempts, lastErr)
}

// Generate builds a verified puzzle and writes it as a Tiled world.
func Generate(spec Spec, opts Options) (*Result, error) {
	res, err := Build(spec, opts)
	if err != nil {
		return nil, err
	}
	p := res.Puzzle
	objects := []gridgen.Object{}
	for _, e := range p.Ents {
		objects = append(objects, toObject(e))
	}
	cfg := &gridgen.GenerationConfig{
		WorldName:        opts.WorldName,
		Title:            fmt.Sprintf("Puzzle level %d", spec.Level),
		Author:           "puzzlegen",
		OutDir:           opts.OutDir,
		TilesetPath:      opts.TilesetPath,
		WangSet:          "terrains",
		EntdefsDir:       opts.EntdefsDir,
		Terrains:         []string{"grass", "wall"},
		MapRows:          mapSize,
		MapCols:          mapSize,
		WorldRows:        1,
		WorldCols:        1,
		Seed:             spec.Seed,
		SpawnTerrain:     "wall",
		DecorTilesetPath: opts.DecorTilesetPath,
		DecorTerrain:     "grass",
		DecorDensity:     0.12,
		Field:            p.Field,
		Objects:          map[string][]gridgen.Object{"0.0.tmj": objects},
		NoAutoSpawn:      true,
		WorldProperties: map[string]string{
			"puzzle_level":   strconv.Itoa(spec.Level),
			"puzzle_seed":    strconv.FormatInt(spec.Seed, 10),
			"puzzle_actions": strconv.Itoa(len(res.Solution.Actions) - 1),
		},
	}
	if opts.DecorTilesetPath == "" {
		cfg.DecorDensity = 0
	}
	gres, err := gridgen.Generate(cfg)
	if err != nil {
		return nil, err
	}
	if gres.CornersRaised > 0 {
		// The painted walls would no longer match the layout the solver checked.
		return nil, fmt.Errorf("tileset could not draw the layout as is (%d corners raised)", gres.CornersRaised)
	}
	res.WorldPath = gres.WorldPath
	res.SpawnGrid = gres.SpawnGrid
	return res, nil
}

// setTimers sizes each spring lever's window from the shortest run between
// it and the far side of its door (static obstacles only), with the spec's
// slack.
func setTimers(p *Puzzle, spec Spec) {
	statics := map[Cell]bool{}
	for _, e := range p.Ents {
		switch e.Kind {
		case KLever, KTimedLever, KPot, KChest, KSign, KBlock:
			statics[e.Cell] = true
		}
	}
	slack := spec.TimerSlack
	if slack <= 0 {
		slack = 1.5
	}
	for _, e := range p.Ents {
		if e.Kind != KTimedLever {
			continue
		}
		door := p.Doors[e.Door]
		best := math.MaxInt32
		for _, d := range dirs {
			stand := e.Cell.add(d)
			if !p.Walk[stand] || statics[stand] {
				continue
			}
			dist := bfs(p.Walk, stand, func(c Cell) bool { return statics[c] })
			if v, ok := dist[door.Side[1]]; ok && v < best {
				best = v
			}
		}
		if best == math.MaxInt32 {
			best = 2 * mapSize
		}
		e.budget = int(math.Ceil(float64(best) * slack))
		e.ResetMs = timerMs(e.budget, 1)
	}
}

func toObject(e *Ent) gridgen.Object {
	props := map[string]string{"id": e.ID}
	switch e.Kind {
	case KPot:
		props["loot"] = e.Loot
		if e.Loot == "" {
			props["loot"] = decoyItem
		}
	case KChest:
		if e.Requires != "" {
			props["requires"] = e.Requires
			props["locked_msg"] = "It's locked. Maybe a key is hidden nearby."
		}
		if e.Loot != "" {
			props["loot"] = e.Loot
		}
		if e.Msg != "" {
			props["msg"] = e.Msg
		}
	case KSign:
		props["msg"] = e.Msg
		props["collision_type"] = "touch"
	case KTimedLever:
		props["reset_after"] = strconv.Itoa(e.ResetMs)
	case KDoor:
		var conds []string
		for _, c := range e.Conds {
			conds = append(conds, c.ID+"="+c.State)
		}
		props["open_when"] = strings.Join(conds, ", ")
	case KExit:
		props["rift_exit"] = "complete"
	case KEntry:
		props["rift_exit"] = "leave"
	case KSpawn:
		props = nil
	}
	return gridgen.Object{Type: e.Type(), Name: e.ID, Row: e.Cell.R, Col: e.Cell.C, Props: props}
}

// Describe is a text picture of the puzzle and its solution, for previews.
func Describe(res *Result) string {
	p := res.Puzzle
	grid := make([][]byte, mapSize)
	for r := range grid {
		grid[r] = []byte(strings.Repeat(" ", mapSize))
		for c := 0; c < mapSize; c++ {
			if p.Walk[Cell{r, c}] {
				grid[r][c] = '.'
			} else {
				grid[r][c] = '#'
			}
		}
	}
	glyph := map[Kind]byte{KLever: 'L', KTimedLever: 'T', KPot: 'o', KChest: 'C', KBlock: 'B', KPlate: '_', KDoor: 'D', KSign: 'S', KExit: 'X', KSpawn: '@', KEntry: '@', KCoin: '$'}
	for _, e := range p.Ents {
		grid[e.Cell.R][e.Cell.C] = glyph[e.Kind]
	}
	var sb strings.Builder
	fmt.Fprintf(&sb, "level %d, seed %d: %d rooms, %d actions (layout attempt %d, %d states)\n",
		res.Spec.Level, res.Spec.Seed, res.Spec.Rooms, len(res.Solution.Actions)-1, res.Attempts, res.Solution.States)
	for _, row := range grid {
		sb.WriteString("  " + string(row) + "\n")
	}
	for _, d := range p.Doors {
		var conds []string
		for _, c := range d.Conds {
			conds = append(conds, c.ID+"="+c.State)
		}
		fmt.Fprintf(&sb, "  %s opens when %s\n", d.ID, strings.Join(conds, ", "))
	}
	for i, a := range res.Solution.Actions {
		fmt.Fprintf(&sb, "  %2d. %s\n", i+1, a)
	}
	return sb.String()
}

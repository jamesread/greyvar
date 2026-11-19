package rig

import (
	"encoding/json"
	"fmt"
	"image"
	"path/filepath"
	"sort"
	"strings"
)

// Output is one generated file.
type Output struct {
	Path string // absolute or rig-relative-resolved path
	Data []byte
}

// RigJSON is the client-facing manifest written next to the layer sheets.
type RigJSON struct {
	Name         string                         `json:"name"`
	FrameSize    int                            `json:"frameSize"`
	Columns      int                            `json:"columns"`
	FrameCount   int                            `json:"frameCount"`
	Layers       []LayerJSON                    `json:"layers"`
	SlotOrder    []string                       `json:"slotOrder"`
	Slots        map[string][]string            `json:"slots"`
	Palettes     map[string]map[string][]string `json:"palettes"`
	Defaults     Defaults                       `json:"defaults"`
	Poses        []PoseJSON                     `json:"poses"`
	Clips        map[string][]StepJSON          `json:"clips"`
	Idle         IdleSpec                       `json:"idle"`
	States       map[string]string              `json:"states"`
	InitialState string                         `json:"initialState"`
	Hitbox       [4]int                         `json:"hitbox"`
}

// LayerJSON lists a layer's sheets (web paths under /res) and which colour
// slots appear in it, so the client only recolours what matters.
type LayerJSON struct {
	Name     string            `json:"name"`
	Default  string            `json:"default"`
	Variants []string          `json:"variants"`
	Sheets   map[string]string `json:"sheets"`
	Slots    []string          `json:"slots"`
}

// PoseJSON is a pose's frame index and item anchor.
type PoseJSON struct {
	Name  string  `json:"name"`
	Frame int     `json:"frame"`
	Hand  *[2]int `json:"hand"`
	ItemZ string  `json:"itemZ"`
}

// StepJSON is a clip step resolved to a frame index.
type StepJSON struct {
	Pose  string `json:"pose"`
	Frame int    `json:"frame"`
	Ms    int    `json:"ms"`
}

func (r *Rig) resolve(rel string) string {
	if filepath.IsAbs(rel) {
		return filepath.Clean(rel)
	}
	return filepath.Clean(filepath.Join(r.Dir, rel))
}

func (r *Rig) webPath(abs string) (string, error) {
	root := r.resolve(r.Spec.Outputs.WebRoot)
	rel, err := filepath.Rel(root, abs)
	if err != nil || strings.HasPrefix(rel, "..") {
		return "", fmt.Errorf("%s is not under webRoot %s", abs, root)
	}
	return filepath.ToSlash(rel), nil
}

// sheetPath is where a layer variant's sheet is written.
func (r *Rig) sheetPath(layer, variant string) string {
	return filepath.Join(r.resolve(r.Spec.Outputs.SheetsDir), layer+"_"+variant+".png")
}

func (r *Rig) manifestPath() string {
	return filepath.Join(r.resolve(r.Spec.Outputs.SheetsDir), r.Spec.Name+".rig.json")
}

// partCache loads each part PNG once.
type partCache map[string]*image.NRGBA

func (pc partCache) get(r *Rig, layer, variant, part string) (*image.NRGBA, error) {
	p := r.PartPath(layer, variant, part)
	if img, ok := pc[p]; ok {
		return img, nil
	}
	img, err := ReadPNG(p)
	if err != nil {
		return nil, err
	}
	pc[p] = img
	return img, nil
}

func (r *Rig) sheetSize() (int, int) {
	fs, cols := r.Spec.FrameSize, r.Spec.Columns
	rows := (len(r.Poses) + cols - 1) / cols
	if rows == 0 {
		rows = 1
	}
	return cols * fs, rows * fs
}

// renderLayerSheet draws every pose's part for one layer variant.
func (r *Rig) renderLayerSheet(pc partCache, layer, variant string) (*image.NRGBA, error) {
	w, h := r.sheetSize()
	sheet := image.NewNRGBA(image.Rect(0, 0, w, h))
	fs, cols := r.Spec.FrameSize, r.Spec.Columns
	for i, pose := range r.Poses {
		part, ok := pose.Parts[layer]
		if !ok || part == "" {
			continue
		}
		img, err := pc.get(r, layer, variant, part)
		if err != nil {
			return nil, fmt.Errorf("pose %s layer %s variant %s: %w", pose.Name, layer, variant, err)
		}
		off := pose.Offsets[layer]
		cx, cy := (i%cols)*fs, (i/cols)*fs
		for y := 0; y < img.Bounds().Dy(); y++ {
			for x := 0; x < img.Bounds().Dx(); x++ {
				c := img.NRGBAAt(x, y)
				if c.A == 0 {
					continue
				}
				tx, ty := x+off[0], y+off[1]
				if tx < 0 || ty < 0 || tx >= fs || ty >= fs {
					continue // validate reports this
				}
				sheet.SetNRGBA(cx+tx, cy+ty, c)
			}
		}
	}
	return sheet, nil
}

// RecolourMap maps key colours to the chosen palette for each slot.
func (r *Rig) RecolourMap(choice map[string]string) (map[RGB]RGB, error) {
	m := map[RGB]RGB{}
	for _, slot := range r.Spec.SlotOrder {
		name := choice[slot]
		if name == "" {
			continue
		}
		ramp, ok := r.Spec.Palettes[slot][name]
		if !ok {
			return nil, fmt.Errorf("unknown palette %s.%s", slot, name)
		}
		for i, k := range r.Spec.Slots[slot] {
			from, err := ParseHex(k)
			if err != nil {
				return nil, err
			}
			to, err := ParseHex(ramp[i])
			if err != nil {
				return nil, err
			}
			m[from] = to
		}
	}
	return m, nil
}

func recolour(img *image.NRGBA, m map[RGB]RGB) {
	if len(m) == 0 {
		return
	}
	b := img.Bounds()
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			c, ok := rgbAt(img, x, y)
			if !ok {
				continue
			}
			if to, ok := m[c]; ok {
				setRGB(img, x, y, to)
			}
		}
	}
}

// Compose renders frame-strip images for a look (parts + palettes).
// It returns a full sheet with every pose.
func (r *Rig) Compose(parts, palettes map[string]string) (*image.NRGBA, error) {
	pc := partCache{}
	w, h := r.sheetSize()
	out := image.NewNRGBA(image.Rect(0, 0, w, h))
	m, err := r.RecolourMap(palettes)
	if err != nil {
		return nil, err
	}
	for _, l := range r.Spec.Layers {
		variant := r.DefaultVariant(l)
		if v, ok := parts[l.Name]; ok && v != "" {
			variant = v
		}
		sheet, err := r.renderLayerSheet(pc, l.Name, variant)
		if err != nil {
			return nil, err
		}
		recolour(sheet, m)
		overlay(out, sheet)
	}
	return out, nil
}

func overlay(dst, src *image.NRGBA) {
	b := src.Bounds()
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			c := src.NRGBAAt(x, y)
			if c.A != 0 {
				dst.SetNRGBA(x, y, c)
			}
		}
	}
}

func (r *Rig) slotsIn(img *image.NRGBA) map[string]bool {
	keyToSlot := map[RGB]string{}
	for slot, ramp := range r.Spec.Slots {
		for _, k := range ramp {
			if c, err := ParseHex(k); err == nil {
				keyToSlot[c] = slot
			}
		}
	}
	used := map[string]bool{}
	b := img.Bounds()
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			if c, ok := rgbAt(img, x, y); ok {
				if s, ok := keyToSlot[c]; ok {
					used[s] = true
				}
			}
		}
	}
	return used
}

// Build generates every output. It does not write anything; callers decide
// whether to write (build) or compare (check).
func (r *Rig) Build() ([]Output, error) {
	if errs := r.Validate(); len(errs) > 0 {
		return nil, fmt.Errorf("rig invalid:\n  %s", strings.Join(errs, "\n  "))
	}
	pc := partCache{}
	var outs []Output
	idx := r.PoseIndex()

	man := RigJSON{
		Name:         r.Spec.Name,
		FrameSize:    r.Spec.FrameSize,
		Columns:      r.Spec.Columns,
		FrameCount:   len(r.Poses),
		SlotOrder:    r.Spec.SlotOrder,
		Slots:        r.Spec.Slots,
		Palettes:     r.Spec.Palettes,
		Defaults:     r.Spec.Defaults,
		Clips:        map[string][]StepJSON{},
		Idle:         r.Clips.Idle,
		States:       r.Clips.States,
		InitialState: r.Clips.InitialState,
		Hitbox:       r.Spec.HitboxOrDefault(),
	}

	for _, l := range r.Spec.Layers {
		lj := LayerJSON{Name: l.Name, Default: r.DefaultVariant(l), Variants: LayerVariants(l), Sheets: map[string]string{}}
		slots := map[string]bool{}
		for _, v := range LayerVariants(l) {
			sheet, err := r.renderLayerSheet(pc, l.Name, v)
			if err != nil {
				return nil, err
			}
			for s := range r.slotsIn(sheet) {
				slots[s] = true
			}
			p := r.sheetPath(l.Name, v)
			web, err := r.webPath(p)
			if err != nil {
				return nil, err
			}
			lj.Sheets[v] = web
			outs = append(outs, Output{Path: p, Data: EncodePNG(sheet)})
		}
		for _, s := range r.Spec.SlotOrder {
			if slots[s] {
				lj.Slots = append(lj.Slots, s)
			}
		}
		if lj.Slots == nil {
			lj.Slots = []string{}
		}
		man.Layers = append(man.Layers, lj)
	}

	for i, p := range r.Poses {
		z := p.ItemZ
		if z == "" {
			z = "front"
		}
		man.Poses = append(man.Poses, PoseJSON{Name: p.Name, Frame: i, Hand: p.Hand, ItemZ: z})
	}
	for name, steps := range r.Clips.Clips {
		for _, s := range steps {
			man.Clips[name] = append(man.Clips[name], StepJSON{Pose: s.Pose, Frame: idx[s.Pose], Ms: s.Ms})
		}
	}

	js, err := json.MarshalIndent(man, "", "  ")
	if err != nil {
		return nil, err
	}
	outs = append(outs, Output{Path: r.manifestPath(), Data: append(js, '\n')})

	composite, err := r.Compose(r.Spec.Defaults.Parts, r.Spec.Defaults.Palettes)
	if err != nil {
		return nil, err
	}
	compPath := r.resolve(r.Spec.Outputs.Composite)
	outs = append(outs, Output{Path: compPath, Data: EncodePNG(composite)})

	if r.Spec.Outputs.Tsx != "" {
		tsx, err := r.renderTsx(compPath, composite.Bounds().Dx(), composite.Bounds().Dy())
		if err != nil {
			return nil, err
		}
		outs = append(outs, Output{Path: r.resolve(r.Spec.Outputs.Tsx), Data: tsx})
	}

	sort.Slice(outs, func(i, j int) bool { return outs[i].Path < outs[j].Path })
	return outs, nil
}

func (r *Rig) renderTsx(compPath string, w, h int) ([]byte, error) {
	tsxPath := r.resolve(r.Spec.Outputs.Tsx)
	rel, err := filepath.Rel(filepath.Dir(tsxPath), compPath)
	if err != nil {
		return nil, err
	}
	idx := r.PoseIndex()
	type tile struct {
		id    int
		typ   string
		steps []Step
	}
	var tiles []tile
	seen := map[int]string{}
	states := make([]string, 0, len(r.Clips.States))
	for s := range r.Clips.States {
		states = append(states, s)
	}
	sort.Strings(states)
	for _, state := range states {
		steps := r.Clips.Clips[r.Clips.States[state]]
		id := idx[steps[0].Pose]
		if other, dup := seen[id]; dup {
			return nil, fmt.Errorf("states %s and %s both start on pose %s; Tiled needs a distinct first frame per state", other, state, steps[0].Pose)
		}
		seen[id] = state
		typ := r.Spec.Name + "_" + state
		if state == r.Clips.InitialState {
			typ = r.Spec.Name
		}
		tiles = append(tiles, tile{id, typ, steps})
	}
	sort.Slice(tiles, func(i, j int) bool { return tiles[i].id < tiles[j].id })

	var b strings.Builder
	fs := r.Spec.FrameSize
	b.WriteString(`<?xml version="1.0" encoding="UTF-8"?>` + "\n")
	fmt.Fprintf(&b, "<!-- GENERATED by sprite-generator from res/sprites/%s. Do not edit; run `make sprites`. -->\n", r.Spec.Name)
	fmt.Fprintf(&b, `<tileset version="1.10" tiledversion="1.12.2" name="%s" tilewidth="%d" tileheight="%d" tilecount="%d" columns="%d">`+"\n",
		r.Spec.Name, fs, fs, len(r.Poses), r.Spec.Columns)
	fmt.Fprintf(&b, ` <image source="%s" width="%d" height="%d"/>`+"\n", filepath.ToSlash(rel), w, h)
	for _, t := range tiles {
		fmt.Fprintf(&b, ` <tile id="%d" type="%s">`+"\n  <animation>\n", t.id, t.typ)
		for _, s := range t.steps {
			fmt.Fprintf(&b, `   <frame tileid="%d" duration="%d"/>`+"\n", idx[s.Pose], s.Ms)
		}
		b.WriteString("  </animation>\n </tile>\n")
	}
	b.WriteString("</tileset>\n")
	return []byte(b.String()), nil
}

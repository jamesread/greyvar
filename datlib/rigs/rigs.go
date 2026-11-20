// Package rigs reads the client-facing <name>.rig.json written by
// sprite-generator, so the server can validate and default appearances
// without depending on the generator itself.
package rigs

import (
	"encoding/json"
	"fmt"
	"hash/fnv"
	"os"
	"path/filepath"
	"sort"

	"github.com/jamesread/greyvar/datlib/common"
)

// Rig is the subset of rig.json the server needs.
type Rig struct {
	Name      string                         `json:"name"`
	Layers    []Layer                        `json:"layers"`
	// Hitbox is the collision box [x, y, w, h] within the 16px frame.
	Hitbox [4]int32 `json:"hitbox"`
	SlotOrder []string                       `json:"slotOrder"`
	Palettes  map[string]map[string][]string `json:"palettes"`
	// Size is "child" or "adult" (rig.yml size); Anchor is the frame point
	// on the bottom centre of the character's cell (its feet).
	Size   string `json:"size"`
	Anchor [2]int `json:"anchor"`
	// Hidden[slot]: palettes only data uses (a mannequin's grey skin), never
	// picked for a random look.
	Hidden map[string][]string `json:"hidden"`
	Defaults  struct {
		Palettes map[string]string `json:"palettes"`
		Parts    map[string]string `json:"parts"`
	} `json:"defaults"`
}

// Layer is one draw layer and its selectable variants.
type Layer struct {
	Name     string   `json:"name"`
	Default  string   `json:"default"`
	Variants []string `json:"variants"`
}

// Appearance is a validated look. Rig is the size it is drawn at
// ("child" or "adult"); empty means the default size.
type Appearance struct {
	Rig      string
	Parts    map[string]string
	Palettes map[string]string
	HeldItem string
}

// PlayerRigPath is where the child rig manifest lives under ResDir (the
// original player rig).
func PlayerRigPath() string {
	return RigPath("child")
}

// RigPath is where a size's rig manifest lives under ResDir.
func RigPath(size string) string {
	return RigPathIn(common.ResDir(), size)
}

// RigPathIn is where a size's rig manifest lives under the res directory.
func RigPathIn(res, size string) string {
	if res == "" {
		return ""
	}
	if size == "child" {
		return filepath.Join(res, "img", "textures", "entities", "player", "player.rig.json")
	}
	return filepath.Join(res, "img", "textures", "entities", size, size+".rig.json")
}

// Sizes are the character sizes, default first.
var Sizes = []string{"adult", "child"}

// Set is every loaded rig by size. New looks use Default (the first of Sizes
// that is loaded).
type Set struct {
	BySize  map[string]*Rig
	Default string
}

// LoadSet loads the rig for each size that has a manifest.
func LoadSet() (*Set, []error) {
	return LoadSetFrom(common.ResDir())
}

// LoadSetFrom loads the rig for each size from a res directory.
func LoadSetFrom(res string) (*Set, []error) {
	set := &Set{BySize: map[string]*Rig{}}
	var errs []error
	for _, size := range Sizes {
		path := RigPathIn(res, size)
		if path == "" {
			continue
		}
		r, err := Load(path)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		if r.Size == "" {
			r.Size = size
		}
		set.BySize[size] = r
		if set.Default == "" {
			set.Default = size
		}
	}
	return set, errs
}

// NewSet wraps already loaded rigs (tests).
func NewSet(rs ...*Rig) *Set {
	set := &Set{BySize: map[string]*Rig{}}
	for _, r := range rs {
		set.BySize[r.Size] = r
	}
	for _, size := range Sizes {
		if _, ok := set.BySize[size]; ok {
			set.Default = size
			break
		}
	}
	return set
}

// Size is the rig for a size, or the default rig when that size is missing.
func (s *Set) Size(size string) *Rig {
	if s == nil {
		return nil
	}
	if r, ok := s.BySize[size]; ok {
		return r
	}
	return s.BySize[s.Default]
}

// For is the rig an appearance is drawn with.
func (s *Set) For(a *Appearance) *Rig {
	if a == nil {
		return s.Size("")
	}
	return s.Size(a.Rig)
}

// DefaultLook is a stable default look at the default size.
func (s *Set) DefaultLook(seed string) Appearance {
	return s.Size("").Default(seed)
}

// Normalize fits a look onto this rig: parts and palettes the rig lacks fall
// back to its defaults (so a child look, or an old look with no size, still
// loads as an adult), and Rig is set to this rig's size.
func (r *Rig) Normalize(a Appearance) Appearance {
	out := Appearance{Rig: r.Size, Parts: map[string]string{}, Palettes: map[string]string{}, HeldItem: a.HeldItem}
	for _, l := range r.Layers {
		if l.Default != "" {
			out.Parts[l.Name] = l.Default
		}
	}
	for k, v := range r.Defaults.Palettes {
		out.Palettes[k] = v
	}
	for k, v := range a.Parts {
		if l, ok := r.layer(k); ok && contains(l.Variants, v) {
			out.Parts[k] = v
		}
	}
	for k, v := range a.Palettes {
		if _, ok := r.Palettes[k][v]; ok {
			out.Palettes[k] = v
		}
	}
	return out
}

// Load reads a rig.json.
func Load(path string) (*Rig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var r Rig
	if err := json.Unmarshal(data, &r); err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	return &r, nil
}

func (r *Rig) layer(name string) (Layer, bool) {
	for _, l := range r.Layers {
		if l.Name == name {
			return l, true
		}
	}
	return Layer{}, false
}

// PaletteNames returns a slot's palettes, sorted.
func (r *Rig) PaletteNames(slot string) []string {
	names := make([]string, 0, len(r.Palettes[slot]))
	for n := range r.Palettes[slot] {
		names = append(names, n)
	}
	sort.Strings(names)
	return names
}

func pick(seed, key string, options []string) string {
	h := fnv.New64a()
	_, _ = h.Write([]byte(seed + "/" + key))
	// FNV's low bits barely change between seeds like player9/player10;
	// finalise (splitmix64) so consecutive players get varied looks.
	x := h.Sum64()
	x ^= x >> 30
	x *= 0xbf58476d1ce4e5b9
	x ^= x >> 27
	x *= 0x94d049bb133111eb
	x ^= x >> 31
	return options[x%uint64(len(options))]
}

// Default derives a stable look from seed (e.g. a username): the same seed
// always gets the same look, different seeds usually differ.
func (r *Rig) Default(seed string) Appearance {
	a := Appearance{Rig: r.Size, Parts: map[string]string{}, Palettes: map[string]string{}}
	for _, l := range r.Layers {
		switch {
		case len(l.Variants) > 0 && l.Variants[0] == "none":
			// Optional layers (outfits) start empty: new players wear their
			// shirt and trousers, not a random dress.
			a.Parts[l.Name] = "none"
		case len(l.Variants) > 1:
			a.Parts[l.Name] = pick(seed, "part:"+l.Name, l.Variants)
		}
	}
	for _, slot := range r.SlotOrder {
		var names []string
		for _, n := range r.PaletteNames(slot) {
			if !contains(r.Hidden[slot], n) {
				names = append(names, n)
			}
		}
		if len(names) > 0 {
			a.Palettes[slot] = pick(seed, "palette:"+slot, names)
		}
	}
	return a
}

// Apply returns base with the requested changes, or an error naming the
// first unknown layer/variant/slot/palette. base is not modified.
func (r *Rig) Apply(base Appearance, parts, palettes map[string]string) (Appearance, error) {
	out := Appearance{Rig: r.Size, Parts: map[string]string{}, Palettes: map[string]string{}, HeldItem: base.HeldItem}
	for k, v := range base.Parts {
		out.Parts[k] = v
	}
	for k, v := range base.Palettes {
		out.Palettes[k] = v
	}
	for _, k := range sortedKeys(parts) {
		l, ok := r.layer(k)
		if !ok {
			return base, fmt.Errorf("unknown layer %q", k)
		}
		if !contains(l.Variants, parts[k]) {
			return base, fmt.Errorf("unknown %s variant %q (have %v)", k, parts[k], l.Variants)
		}
		out.Parts[k] = parts[k]
	}
	for _, k := range sortedKeys(palettes) {
		if _, ok := r.Palettes[k]; !ok {
			return base, fmt.Errorf("unknown colour slot %q", k)
		}
		if _, ok := r.Palettes[k][palettes[k]]; !ok {
			return base, fmt.Errorf("unknown %s palette %q (have %v)", k, palettes[k], r.PaletteNames(k))
		}
		out.Palettes[k] = palettes[k]
	}
	return out, nil
}

// CheckWear reports whether a dialog `wear` entry means something on this
// rig: a layer and one of its variants (hair: bun), or a colour slot and one
// of its palettes (hair: blonde). `hair` is both a layer and a slot.
func (r *Rig) CheckWear(key, value string) error {
	l, isLayer := r.layer(key)
	if isLayer && contains(l.Variants, value) {
		return nil
	}
	_, isSlot := r.Palettes[key]
	if isSlot {
		if _, ok := r.Palettes[key][value]; ok {
			return nil
		}
	}
	switch {
	case isLayer && isSlot:
		return fmt.Errorf("%s %q is neither a %s style %v nor a colour %v", key, value, key, l.Variants, r.PaletteNames(key))
	case isLayer:
		return fmt.Errorf("unknown %s variant %q (have %v)", key, value, l.Variants)
	case isSlot:
		return fmt.Errorf("unknown %s palette %q (have %v)", key, value, r.PaletteNames(key))
	}
	return fmt.Errorf("unknown layer or colour slot %q", key)
}

func contains(xs []string, x string) bool {
	for _, v := range xs {
		if v == x {
			return true
		}
	}
	return false
}

func sortedKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

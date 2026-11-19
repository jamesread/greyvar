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

// Appearance is a validated look.
type Appearance struct {
	Parts    map[string]string
	Palettes map[string]string
	HeldItem string
}

// PlayerRigPath is where the player rig manifest lives under ResDir.
func PlayerRigPath() string {
	res := common.ResDir()
	if res == "" {
		return ""
	}
	return filepath.Join(res, "img", "textures", "entities", "player", "player.rig.json")
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
	a := Appearance{Parts: map[string]string{}, Palettes: map[string]string{}}
	for _, l := range r.Layers {
		if len(l.Variants) > 1 {
			a.Parts[l.Name] = pick(seed, "part:"+l.Name, l.Variants)
		}
	}
	for _, slot := range r.SlotOrder {
		if names := r.PaletteNames(slot); len(names) > 0 {
			a.Palettes[slot] = pick(seed, "palette:"+slot, names)
		}
	}
	return a
}

// Apply returns base with the requested changes, or an error naming the
// first unknown layer/variant/slot/palette. base is not modified.
func (r *Rig) Apply(base Appearance, parts, palettes map[string]string) (Appearance, error) {
	out := Appearance{Parts: map[string]string{}, Palettes: map[string]string{}, HeldItem: base.HeldItem}
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

// Package cropdefs loads crop definitions (dat/cropdefs/*.yml) and resolves
// each growth stage to a tile in the crops tileset by its properties.
package cropdefs

import (
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"gopkg.in/yaml.v2"
)

// CropDefinition is one crop type.
type CropDefinition struct {
	ID           string  `yaml:"id"`
	Title        string  `yaml:"title"`
	Seed         string  `yaml:"seed"`
	StarterSeeds int32   `yaml:"starterSeeds"`
	Stages       []Stage `yaml:"stages"`
	Harvest      Harvest `yaml:"harvest"`
}

// Stage is one growth stage; the last stage is ripe and never advances.
type Stage struct {
	GrowMs int64 `yaml:"growMs"`
}

// Harvest is what a ripe crop yields.
type Harvest struct {
	Item  string `yaml:"item"`
	Count int32  `yaml:"count"`
	Seeds int32  `yaml:"seeds"`
}

// RipeStage is the index of the harvestable stage.
func (c *CropDefinition) RipeStage() int { return len(c.Stages) - 1 }

// Validate checks the definition against the known items.
func (c *CropDefinition) Validate(itemExists func(string) bool) []string {
	var errs []string
	if c.ID == "" {
		errs = append(errs, "id is required")
	}
	if len(c.Stages) < 2 {
		errs = append(errs, "needs at least 2 stages (growing + ripe)")
	}
	for i, st := range c.Stages[:max(0, len(c.Stages)-1)] {
		if st.GrowMs <= 0 {
			errs = append(errs, fmt.Sprintf("stage %d: growMs must be > 0", i))
		}
	}
	for _, it := range []string{c.Seed, c.Harvest.Item} {
		if it == "" || !itemExists(it) {
			errs = append(errs, fmt.Sprintf("unknown item %q", it))
		}
	}
	if c.Harvest.Count <= 0 {
		errs = append(errs, "harvest.count must be > 0")
	}
	return errs
}

// LoadAll reads every *.yml in dir, keyed by id.
func LoadAll(dir string) (map[string]*CropDefinition, error) {
	files, err := filepath.Glob(filepath.Join(dir, "*.yml"))
	if err != nil {
		return nil, err
	}
	out := map[string]*CropDefinition{}
	for _, f := range files {
		data, err := os.ReadFile(f)
		if err != nil {
			return nil, err
		}
		var def CropDefinition
		if err := yaml.UnmarshalStrict(data, &def); err != nil {
			return nil, fmt.Errorf("%s: %w", f, err)
		}
		out[def.ID] = &def
	}
	return out, nil
}

// StageTiles maps crop id -> stage index -> local tile id in the crops
// tileset, from tile properties crop=<id> and stage=<n>.
func StageTiles(tsxPath string) (map[string]map[int]int, error) {
	data, err := os.ReadFile(tsxPath)
	if err != nil {
		return nil, err
	}
	var ts struct {
		Tiles []struct {
			ID    int `xml:"id,attr"`
			Props []struct {
				Name  string `xml:"name,attr"`
				Value string `xml:"value,attr"`
			} `xml:"properties>property"`
		} `xml:"tile"`
	}
	if err := xml.Unmarshal(data, &ts); err != nil {
		return nil, fmt.Errorf("%s: %w", tsxPath, err)
	}
	out := map[string]map[int]int{}
	for _, t := range ts.Tiles {
		crop, stage := "", -1
		for _, p := range t.Props {
			switch strings.ToLower(p.Name) {
			case "crop":
				crop = p.Value
			case "stage":
				if n, err := strconv.Atoi(p.Value); err == nil {
					stage = n
				}
			}
		}
		if crop == "" || stage < 0 {
			continue
		}
		if out[crop] == nil {
			out[crop] = map[int]int{}
		}
		out[crop][stage] = t.ID
	}
	return out, nil
}

// MissingStageTiles lists stages of def with no tile.
func MissingStageTiles(def *CropDefinition, tiles map[string]map[int]int) []int {
	var missing []int
	for i := range def.Stages {
		if _, ok := tiles[def.ID][i]; !ok {
			missing = append(missing, i)
		}
	}
	sort.Ints(missing)
	return missing
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

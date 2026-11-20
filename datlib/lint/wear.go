package lint

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jamesread/greyvar/datlib/rigs"
)

// LintWearProperties checks every map object's `wear` property (a
// mannequin's outfit: "outfit=dress,garment=blue") against the default
// (adult) character rig, like dialog `wear`.
func LintWearProperties(datDir, resDir string) []Check {
	set, _ := rigs.LoadSetFrom(resDir)
	rig := set.Size("")
	if rig == nil {
		return nil // LintDialogs already warns that the rigs are missing
	}
	var out []Check
	maps, _ := filepath.Glob(filepath.Join(datDir, "worlds", "*", "*.tmj"))
	for _, m := range maps {
		raw, err := os.ReadFile(m)
		if err != nil {
			continue
		}
		var doc tmjWearDoc
		if json.Unmarshal(raw, &doc) != nil {
			continue // LintTMJMaps reports unreadable maps
		}
		check := Check{Type: "tmj", Filename: m}
		var walk func([]tmjWearLayer)
		walk = func(layers []tmjWearLayer) {
			for _, l := range layers {
				walk(l.Layers)
				for _, o := range l.Objects {
					for _, p := range o.Properties {
						if p.Name != "wear" {
							continue
						}
						v, _ := p.Value.(string)
						where := fmt.Sprintf("object %d (%s) at (%d,%d)", o.ID, o.Type, int(o.X)/16, int(o.Y)/16)
						entries := 0
						for _, kv := range strings.Split(v, ",") {
							k, val, ok := strings.Cut(strings.TrimSpace(kv), "=")
							k, val = strings.TrimSpace(k), strings.TrimSpace(val)
							if !ok || k == "" || val == "" {
								check.AddError(fmt.Sprintf("%s: wear %q: want key=value pairs", where, v))
								continue
							}
							entries++
							if err := rig.CheckWear(k, val); err != nil {
								check.AddError(fmt.Sprintf("%s: wear: %v (%s rig)", where, err, rig.Size))
							}
						}
						if entries == 0 {
							check.AddError(fmt.Sprintf("%s: empty wear", where))
						}
					}
				}
			}
		}
		walk(doc.Layers)
		if len(check.Issues) > 0 {
			out = append(out, check)
		}
	}
	return out
}

type tmjWearDoc struct {
	Layers []tmjWearLayer `json:"layers"`
}

type tmjWearLayer struct {
	Layers  []tmjWearLayer `json:"layers"`
	Objects []struct {
		ID         int     `json:"id"`
		Type       string  `json:"type"`
		X          float64 `json:"x"`
		Y          float64 `json:"y"`
		Properties []struct {
			Name  string `json:"name"`
			Value any    `json:"value"`
		} `json:"properties"`
	} `json:"objects"`
}

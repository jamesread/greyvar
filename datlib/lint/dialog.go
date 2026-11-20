package lint

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/jamesread/greyvar/datlib/dialogs"
	"github.com/jamesread/greyvar/datlib/entdefs"
	"github.com/jamesread/greyvar/datlib/rigs"
)

// LintDialogs validates dat/dialogs/*.yml (start node, gotos, options, that
// deposit/{balance:} items exist and that every `wear` entry is a style or
// colour of the default (adult) character rig), that every entdef `dialog:`
// names an existing dialog, and that every NPC look fits its rig.
func LintDialogs(datDir, resDir string) []Check {
	var out []Check
	set, rigErrs := rigs.LoadSetFrom(resDir)
	rig := set.Size("")
	if rig == nil {
		check := Check{Type: "dialog", Filename: filepath.Join(datDir, "dialogs")}
		check.AddWarning(fmt.Sprintf("no character rigs under %q (%v): wear and NPC looks not checked", resDir, rigErrs))
		out = append(out, check)
	}
	matches, _ := filepath.Glob(filepath.Join(datDir, "dialogs", "*.yml"))
	for _, m := range matches {
		check := Check{Type: "dialog", Filename: m}
		d, err := dialogs.ReadFile(m)
		if err != nil {
			check.AddError(err.Error())
		} else {
			for _, item := range d.Items() {
				if _, err := os.Stat(filepath.Join(datDir, "itemdefs", item+".yml")); err != nil {
					check.AddError(fmt.Sprintf("deposit/balance item %q: no itemdefs/%s.yml", item, item))
				}
			}
			if rig != nil {
				for _, name := range sortedNodeNames(d) {
					if d.Nodes[name] == nil {
						continue
					}
					for _, o := range d.Nodes[name].Options {
						for _, k := range sortedWearKeys(o.Wear) {
							if err := rig.CheckWear(k, o.Wear[k]); err != nil {
								check.AddError(fmt.Sprintf("node %q option %q: wear: %v (%s rig)", name, o.Label, err, rig.Size))
							}
						}
					}
				}
			}
		}
		out = append(out, check)
	}
	defs, _ := filepath.Glob(filepath.Join(datDir, "entdefs", "*.yml"))
	for _, m := range defs {
		var def *entdefs.EntityDefinition
		withQuietLogs(func() { def, _ = entdefs.ReadEntdefFile(m) })
		if def == nil {
			continue
		}
		if look := def.Appearance; look != nil && rig != nil {
			r := set.Size(look.Rig)
			if look.Rig != "" && set.BySize[look.Rig] == nil {
				check := Check{Type: "entdef", Filename: m}
				check.AddError(fmt.Sprintf("appearance rig %q: no such character size (have %v)", look.Rig, rigs.Sizes))
				out = append(out, check)
			} else if _, err := r.Apply(r.Default(def.Title), look.Parts, look.Palettes); err != nil {
				check := Check{Type: "entdef", Filename: m}
				check.AddError(fmt.Sprintf("appearance: %v (%s rig)", err, r.Size))
				out = append(out, check)
			}
		}
		if def.Dialog == "" {
			continue
		}
		if _, err := os.Stat(filepath.Join(datDir, "dialogs", def.Dialog+".yml")); err != nil {
			check := Check{Type: "entdef", Filename: m}
			check.AddError(fmt.Sprintf("dialog %q: no dialogs/%s.yml", def.Dialog, def.Dialog))
			out = append(out, check)
		}
	}
	return out
}

func sortedNodeNames(d *dialogs.Dialog) []string {
	names := make([]string, 0, len(d.Nodes))
	for n := range d.Nodes {
		names = append(names, n)
	}
	sort.Strings(names)
	return names
}

func sortedWearKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

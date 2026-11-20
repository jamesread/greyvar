package lint

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/jamesread/greyvar/datlib/itemdefs"
)

// LintItemdefs checks dat/itemdefs/*.yml load, and that their icon and held
// sprite files exist under resDir (when resDir is known).
func LintItemdefs(datDir, resDir string) []Check {
	var out []Check
	matches, _ := filepath.Glob(filepath.Join(datDir, "itemdefs", "*.yml"))
	for _, m := range matches {
		check := Check{Type: "itemdef", Filename: m}
		var def *itemdefs.ItemDefinition
		var err error
		withQuietLogs(func() { def, err = itemdefs.ReadItemdefFile(m) })
		if err != nil {
			check.AddError(err.Error())
			out = append(out, check)
			continue
		}
		if def.Icon == "" {
			check.AddWarning("no icon (shown blank in the HUD hotbar)")
		}
		if resDir != "" {
			paths := map[string]string{"icon": def.Icon}
			if def.Held != nil {
				paths["held sprite"] = def.Held.Sprite
			}
			for what, rel := range paths {
				if rel == "" {
					continue
				}
				if _, statErr := os.Stat(filepath.Join(resDir, filepath.FromSlash(rel))); statErr != nil {
					check.AddError(fmt.Sprintf("%s %q not found under res/", what, rel))
				}
			}
		}
		out = append(out, check)
	}
	return out
}

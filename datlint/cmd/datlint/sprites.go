package main

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jamesread/greyvar/datlib/lint"
	"github.com/jamesread/greyvar/sprite-generator/pkg/flora"
	"github.com/jamesread/greyvar/sprite-generator/pkg/rig"
)

// spriteChecks validates every sprite source under res/sprites (player rigs
// and the flora tileset) and reports generated files that are out of date,
// the same as `make sprites-check`.
func spriteChecks(resDir string) []lint.Check {
	if resDir == "" {
		return nil
	}
	dirs, _ := filepath.Glob(filepath.Join(resDir, "sprites", "*"))
	sort.Strings(dirs)
	var checks []lint.Check
	for _, dir := range dirs {
		switch {
		case exists(filepath.Join(dir, "rig.yml")):
			check := lint.Check{Type: "rig", Filename: dir}
			if r, err := rig.Load(dir); err != nil {
				check.AddError(err.Error())
			} else {
				addBuildIssues(&check, r.Build)
			}
			checks = append(checks, check)
		case exists(filepath.Join(dir, "flora.yml")):
			check := lint.Check{Type: "flora", Filename: dir}
			if f, err := flora.Load(dir); err != nil {
				check.AddError(err.Error())
			} else {
				addBuildIssues(&check, f.Build)
			}
			checks = append(checks, check)
		}
	}
	return checks
}

func addBuildIssues(check *lint.Check, build func() ([]rig.Output, error)) {
	outs, err := build()
	if err != nil {
		// Validation errors come back as "<kind> invalid:\n  issue\n  issue".
		lines := strings.Split(err.Error(), "\n")
		if len(lines) > 1 {
			lines = lines[1:]
		}
		for _, l := range lines {
			check.AddError(strings.TrimSpace(l))
		}
		return
	}
	for _, path := range rig.Stale(outs) {
		check.AddError("generated file out of date (run `make sprites`): " + path)
	}
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

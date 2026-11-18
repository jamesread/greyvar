package common

import (
	"os"
	"path/filepath"
)

func DatDir() string {
	if env := os.Getenv("GREYVAR_DAT_DIR"); env != "" {
		return env
	}

	if _, err := os.Stat("dat"); err == nil {
		absPath, _ := os.Getwd()
		return absPath + "/dat"
	}

	return ""
}

// ResDir is the shared asset root (repo res/). GREYVAR_RES_DIR (or the
// webclient's GREYVAR_RES) overrides; otherwise it is found next to DatDir
// (server/dat -> ../../res). Returns "" when it cannot be found.
func ResDir() string {
	for _, env := range []string{"GREYVAR_RES_DIR", "GREYVAR_RES"} {
		if v := os.Getenv(env); v != "" {
			return v
		}
	}
	dat := DatDir()
	if dat == "" {
		return ""
	}
	candidate := filepath.Clean(filepath.Join(dat, "..", "..", "res"))
	if _, err := os.Stat(candidate); err == nil {
		return candidate
	}
	return ""
}

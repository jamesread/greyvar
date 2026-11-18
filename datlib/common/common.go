package common

import (
	"os"
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

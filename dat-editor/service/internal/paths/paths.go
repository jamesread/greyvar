package paths

import (
	"os"
	"path/filepath"

	"github.com/jamesread/golure/pkg/dirs"
	datlib "github.com/jamesread/greyvar/datlib/common"
)

func ApplyDatDir(explicit string) {
	if datDir := ResolveDatDir(explicit); datDir != "" {
		os.Setenv("GREYVAR_DAT_DIR", datDir)
	}
}

func ResolveDatDir(explicit string) string {
	if explicit != "" {
		return explicit
	}

	if env := os.Getenv("GREYVAR_DAT_DIR"); env != "" {
		return env
	}

	if d := datlib.DatDir(); d != "" && isValidDatDir(d) {
		return d
	}

	dat, err := dirs.GetFirstExistingDirectory("dat", []string{
		"../../server/dat",
		"../server/dat",
		"dat",
	})
	if err == nil && isValidDatDir(dat) {
		return dat
	}

	return ""
}

func ResolveResDir(explicit string) string {
	if explicit != "" {
		return absPath(explicit)
	}

	if env := os.Getenv("GREYVAR_RES"); env != "" {
		return absPath(env)
	}

	res, err := dirs.GetFirstExistingDirectory("res", []string{
		"../../res",
		"../../../res",
		"../res",
	})
	if err == nil {
		return absPath(res)
	}

	if datDir := os.Getenv("GREYVAR_DAT_DIR"); datDir != "" {
		if derived := resFromDatDir(datDir); isValidResDir(derived) {
			return derived
		}
	}

	return ""
}

func DatDirLabel() string {
	if env := os.Getenv("GREYVAR_DAT_DIR"); env != "" {
		return env
	}
	return datlib.DatDir()
}

func resFromDatDir(datDir string) string {
	// server/dat -> greyvar/res
	return absPath(filepath.Join(filepath.Dir(filepath.Dir(datDir)), "res"))
}

func absPath(path string) string {
	if path == "" {
		return ""
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return path
	}
	return abs
}

func isValidResDir(resDir string) bool {
	path := filepath.Join(resDir, "img", "textures", "tiles")
	st, err := os.Stat(path)
	return err == nil && st.IsDir()
}

func isValidDatDir(datDir string) bool {
	for _, name := range []string{"entdefs", "worlds"} {
		path := filepath.Join(datDir, name)
		st, err := os.Stat(path)
		if err != nil || !st.IsDir() {
			return false
		}
	}
	return true
}

package paths

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestResolveDatDir_fromServiceDir(t *testing.T) {
	t.Chdir(filepath.Join("..", ".."))

	t.Setenv("GREYVAR_DAT_DIR", "")

	datDir := ResolveDatDir("")
	require.NotEmpty(t, datDir)
	require.True(t, isValidDatDir(datDir))
}

func TestApplyDatDir_setsEnv(t *testing.T) {
	t.Chdir(filepath.Join("..", ".."))

	t.Setenv("GREYVAR_DAT_DIR", "")

	ApplyDatDir("")
	require.NotEmpty(t, os.Getenv("GREYVAR_DAT_DIR"))
	require.True(t, isValidDatDir(os.Getenv("GREYVAR_DAT_DIR")))
}

func TestResolveResDir_fromDatDir(t *testing.T) {
	t.Chdir(filepath.Join("..", ".."))

	t.Setenv("GREYVAR_RES", "")
	datDir, err := filepath.Abs(filepath.Join("..", "..", "server", "dat"))
	require.NoError(t, err)
	t.Setenv("GREYVAR_DAT_DIR", datDir)

	resDir := ResolveResDir("")
	require.NotEmpty(t, resDir)
	require.True(t, filepath.IsAbs(resDir))
	require.True(t, isValidResDir(resDir))
}

func TestDatDirLabel_prefersEnv(t *testing.T) {
	t.Setenv("GREYVAR_DAT_DIR", "/configured/dat")
	require.Equal(t, "/configured/dat", DatDirLabel())
}

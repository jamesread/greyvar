package itemdefs

import (
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

// Every itemdef's held sprite must exist and its grip must be inside it.
func TestHeldSpritesExist(t *testing.T) {
	repo := filepath.Join("..", "..")
	files, err := filepath.Glob(filepath.Join(repo, "server", "dat", "itemdefs", "*.yml"))
	if err != nil || len(files) == 0 {
		t.Fatalf("no itemdefs found: %v", err)
	}
	checked := 0
	for _, f := range files {
		def, err := ReadItemdefFile(f)
		if err != nil {
			t.Fatalf("%s: %v", f, err)
		}
		if def.Held == nil {
			continue
		}
		checked++
		path := filepath.Join(repo, "res", filepath.FromSlash(def.Held.Sprite))
		fh, err := os.Open(path)
		if err != nil {
			t.Errorf("%s: held sprite: %v", f, err)
			continue
		}
		cfg, err := png.DecodeConfig(fh)
		fh.Close()
		if err != nil {
			t.Errorf("%s: %v", path, err)
			continue
		}
		g := def.Held.Grip
		if g[0] < 0 || g[1] < 0 || int(g[0]) >= cfg.Width || int(g[1]) >= cfg.Height {
			t.Errorf("%s: grip %v outside %dx%d sprite", f, g, cfg.Width, cfg.Height)
		}
	}
	if checked == 0 {
		t.Error("expected at least one itemdef with a held sprite")
	}
}

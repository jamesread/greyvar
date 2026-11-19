package flora

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Tiled-style formatting: the migration must keep it.
const fixtureMap = `{ "compressionlevel":-1,
 "height":2,
 "layers":[
        {
         "data":[11, 6,
            2147483654, 11],
         "height":2,
         "id":1,
         "name":"ground",
         "opacity":1,
         "type":"tilelayer",
         "visible":true,
         "width":2,
         "x":0,
         "y":0
        },
        {
         "draworder":"topdown",
         "id":2,
         "name":"objects",
         "objects":[],
         "opacity":1,
         "type":"objectgroup",
         "visible":true,
         "x":0,
         "y":0
        }],
 "nextlayerid":3,
 "nextobjectid":1,
 "tilesets":[
        {
         "firstgid":1,
         "source":"grass.tsx"
        }],
 "width":2
}`

func TestMigrateMap(t *testing.T) {
	f := writeFlora(t, testYML+"migrate:\n  replacement: {grass.tsx: 10}\n")
	mapDir := filepath.Join(f.Dir, "maps")
	_ = os.MkdirAll(mapDir, 0o755)
	_ = os.WriteFile(filepath.Join(mapDir, "grass.tsx"), []byte(`<tileset name="grass" tilecount="20" columns="5"></tileset>`), 0o644)
	mapPath := filepath.Join(mapDir, "0.0.tmj")
	_ = os.WriteFile(mapPath, []byte(fixtureMap), 0o644)
	if _, err := f.Build(); err != nil {
		t.Fatal(err)
	}

	res, err := f.Migrate(mapPath, true)
	if err != nil {
		t.Fatal(err)
	}
	if res.Moved != 2 || len(res.Layers) != 1 || res.Layers[0] != "decor" {
		t.Fatalf("unexpected result %+v", res)
	}
	out, _ := os.ReadFile(mapPath)
	var doc mapDoc
	if err := json.Unmarshal(out, &doc); err != nil {
		t.Fatal(err)
	}
	// grass.tsx#5 (gid 6) is the flower; plain tile 10 is gid 11; flora starts at gid 21.
	ground, decor := doc.Layers[0], doc.Layers[1]
	if want := []uint32{11, 11, 2147483659, 11}; !equal(ground.Data, want) {
		t.Errorf("ground = %v, want %v (flip flags kept)", ground.Data, want)
	}
	if want := []uint32{0, 21, 2147483669, 0}; decor.Name != "decor" || decor.ID != 3 || !equal(decor.Data, want) {
		t.Errorf("decor = %s id %d %v, want %v", decor.Name, decor.ID, decor.Data, want)
	}
	if doc.NextLayerID != 4 || doc.Layers[2].Name != "objects" {
		t.Errorf("layer order/nextlayerid wrong: %d %v", doc.NextLayerID, doc.Layers[2].Name)
	}
	if last := doc.Tilesets[1]; last.FirstGID != 21 || !strings.HasSuffix(last.Source, "out/flora.tsx") {
		t.Errorf("flora tileset entry = %+v", last)
	}
	// Formatting kept: untouched lines are byte-identical.
	for _, line := range []string{`{ "compressionlevel":-1,`, `         "draworder":"topdown",`, `         "source":"grass.tsx"`} {
		if !strings.Contains(string(out), line+"\n") {
			t.Errorf("lost formatting for %q", line)
		}
	}
	// Re-running is a no-op.
	again, err := f.Migrate(mapPath, true)
	if err != nil || again.Skipped == "" {
		t.Fatalf("second run should skip, got %+v %v", again, err)
	}
}

func equal(a, b []uint32) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

const embeddedFixture = `{ "height":1,
 "layers":[
        {
         "data":[1, 2, 2147483649],
         "height":1,
         "id":1,
         "name":"biome",
         "type":"tilelayer",
         "width":3
        }],
 "nextlayerid":2,
 "tilesets":[
        {
         "columns":0,
         "firstgid":1,
         "name":"export",
         "tilecount":2,
         "tiles":[
                {
                 "id":0,
                 "image":"..\/tiles\/water.png"
                },
                {
                 "id":1,
                 "image":"..\/tiles\/sand.png"
                }]
        }],
 "width":3
}`

func TestRemapEmbeddedTile(t *testing.T) {
	dir := t.TempDir()
	tsx := filepath.Join(dir, "water.tsx")
	_ = os.WriteFile(tsx, []byte(`<tileset name="water" tilecount="3" columns="3"></tileset>`), 0o644)
	mapPath := filepath.Join(dir, "maps", "0.0.tmj")
	_ = os.MkdirAll(filepath.Dir(mapPath), 0o755)
	_ = os.WriteFile(mapPath, []byte(embeddedFixture), 0o644)

	res, err := RemapEmbeddedTile(mapPath, RemapOptions{FromImage: "water.png", ToTsx: tsx, ToTile: 0}, true)
	if err != nil {
		t.Fatal(err)
	}
	if res.Moved != 2 {
		t.Fatalf("want 2 water cells remapped, got %+v", res)
	}
	out, _ := os.ReadFile(mapPath)
	var doc struct {
		Layers []struct {
			Data []uint32 `json:"data"`
		} `json:"layers"`
		Tilesets []struct {
			FirstGID uint32 `json:"firstgid"`
			Source   string `json:"source"`
		} `json:"tilesets"`
	}
	if err := json.Unmarshal(out, &doc); err != nil {
		t.Fatal(err)
	}
	// Embedded set covers gids 1-2, so water.tsx starts at 3; flip flag kept; sand untouched.
	if want := []uint32{3, 2, 2147483651}; !equal(doc.Layers[0].Data, want) {
		t.Fatalf("data = %v, want %v", doc.Layers[0].Data, want)
	}
	if ts := doc.Tilesets[1]; ts.FirstGID != 3 || ts.Source != "../water.tsx" {
		t.Fatalf("tileset entry = %+v", ts)
	}
	if again, _ := RemapEmbeddedTile(mapPath, RemapOptions{FromImage: "water.png", ToTsx: tsx}, true); again.Skipped == "" {
		t.Fatal("second run should skip")
	}
}

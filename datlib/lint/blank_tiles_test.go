package lint

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// writeBlankTileFixture writes a 2-tile tileset whose tile 1 is fully
// transparent but still carries a wang tag, plus a map that draws tile 1.
func writeBlankTileFixture(t *testing.T) (tsxPath, tmjPath string) {
	t.Helper()
	dir := t.TempDir()

	img := image.NewNRGBA(image.Rect(0, 0, 32, 16))
	for y := 0; y < 16; y++ {
		for x := 0; x < 16; x++ {
			img.Set(x, y, color.NRGBA{40, 160, 40, 255})
		}
	}
	f, err := os.Create(filepath.Join(dir, "t.png"))
	if err != nil {
		t.Fatal(err)
	}
	if err := png.Encode(f, img); err != nil {
		t.Fatal(err)
	}
	f.Close()

	tsxPath = filepath.Join(dir, "t.tsx")
	tsx := `<?xml version="1.0" encoding="UTF-8"?>
<tileset version="1.10" name="t" tilewidth="16" tileheight="16" tilecount="2" columns="2">
 <image source="t.png" width="32" height="16"/>
 <wangsets>
  <wangset name="terrains" type="corner" tile="-1">
   <wangcolor name="grass" color="#24ff20" tile="-1" probability="1"/>
   <wangtile tileid="0" wangid="0,1,0,1,0,1,0,1"/>
   <wangtile tileid="1" wangid="0,1,0,1,0,1,0,1"/>
  </wangset>
 </wangsets>
</tileset>`
	if err := os.WriteFile(tsxPath, []byte(tsx), 0o644); err != nil {
		t.Fatal(err)
	}

	tmjPath = filepath.Join(dir, "0.0.tmj")
	tmj := `{"width":2,"height":1,"tilewidth":16,"tileheight":16,
 "tilesets":[{"firstgid":1,"source":"t.tsx"}],
 "layers":[{"type":"tilelayer","name":"ground","width":2,"height":1,"data":[1,2],"visible":true}]}`
	if err := os.WriteFile(tmjPath, []byte(tmj), 0o644); err != nil {
		t.Fatal(err)
	}
	return tsxPath, tmjPath
}

func hasIssue(c Check, substr string) bool {
	for _, issue := range c.Issues {
		if strings.Contains(issue.Issue, substr) {
			return true
		}
	}
	return false
}

func TestCheckTilesetWarnsOnStaleWangTags(t *testing.T) {
	tsxPath, _ := writeBlankTileFixture(t)
	check := CheckTileset(tsxPath)
	if !hasIssue(check, `wangset "terrains" tags tiles that are blank in the image: [1]`) {
		t.Fatalf("expected stale wang warning, got %+v", check.Issues)
	}
}

func TestCheckTMJWarnsOnBlankTileUse(t *testing.T) {
	_, tmjPath := writeBlankTileFixture(t)
	check := CheckTMJFile(tmjPath, map[string]struct{}{})
	if !hasIssue(check, `layer "ground" draws 1 cells with tiles that are blank in t.tsx`) {
		t.Fatalf("expected blank tile use warning, got %+v", check.Issues)
	}
}

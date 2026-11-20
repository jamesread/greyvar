package lint

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jamesread/greyvar/datlib/tiled"
)

func TestValidateDatDirEmpty(t *testing.T) {
	if err := ValidateDatDir(""); err == nil {
		t.Fatal("expected error for empty dat dir")
	}
}

func TestValidateDatDirIncomplete(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "worlds"), 0o755); err != nil {
		t.Fatal(err)
	}
	err := ValidateDatDir(root)
	if err == nil {
		t.Fatal("expected error when entdefs/ is missing")
	}
	if !strings.Contains(err.Error(), "entdefs/") {
		t.Fatalf("error = %q, want missing entdefs/", err)
	}
}

func TestValidateDatDirOK(t *testing.T) {
	root := t.TempDir()
	for _, name := range []string{"worlds", "entdefs"} {
		if err := os.MkdirAll(filepath.Join(root, name), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if err := ValidateDatDir(root); err != nil {
		t.Fatalf("ValidateDatDir: %v", err)
	}
}

func TestDiscoverDatDirDiscoversServerDat(t *testing.T) {
	root := t.TempDir()
	serverDat := filepath.Join(root, "server", "dat")
	for _, name := range []string{"worlds", "entdefs"} {
		if err := os.MkdirAll(filepath.Join(serverDat, name), 0o755); err != nil {
			t.Fatal(err)
		}
	}

	toolDir := filepath.Join(root, "datlint")
	if err := os.MkdirAll(toolDir, 0o755); err != nil {
		t.Fatal(err)
	}
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(cwd) })
	if err := os.Chdir(toolDir); err != nil {
		t.Fatal(err)
	}
	t.Setenv("GREYVAR_DAT_DIR", "")

	got, err := DiscoverDatDir()
	if err != nil {
		t.Fatalf("DiscoverDatDir: %v", err)
	}
	want, err := filepath.EvalSymlinks(serverDat)
	if err != nil {
		want = filepath.Clean(serverDat)
	}
	gotEval, err := filepath.EvalSymlinks(got)
	if err != nil {
		gotEval = filepath.Clean(got)
	}
	if gotEval != want {
		t.Fatalf("DiscoverDatDir = %q, want %q", gotEval, want)
	}
}

func TestCheckEntdefOK(t *testing.T) {
	dir := t.TempDir()
	tsx := `<?xml version="1.0"?><tileset name="t" tilewidth="16" tileheight="16" tilecount="1" columns="1">
 <image source="bob.png" width="16" height="16"/>
 <tile id="0" type="player"/>
</tileset>`
	if err := os.WriteFile(filepath.Join(dir, "bob.tsx"), []byte(tsx), 0o644); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "player.yml")
	body := "title: player\ninitialState: idle\nstates:\n  idle: {}\n"
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}

	visuals, err := tiled.LoadEntityVisualCatalog(dir)
	if err != nil {
		t.Fatal(err)
	}
	check := CheckEntdef(path, visuals)
	if len(check.Issues) != 0 {
		t.Fatalf("issues = %+v, want none", check.Issues)
	}
}

func TestCheckEntdefMissingTilesetType(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "ghost.yml")
	body := "title: ghost\ninitialState: idle\nstates:\n  idle: {}\n"
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}

	check := CheckEntdef(path, &tiled.EntityVisualCatalog{})
	if !check.HasErrors() {
		t.Fatal("expected missing tileset type error")
	}
}

func TestCheckTiledWorldMissingMap(t *testing.T) {
	dir := t.TempDir()
	world := filepath.Join(dir, "demo.world")
	body := `{"type":"world","maps":[{"fileName":"missing.tmj","x":0,"y":0,"width":16,"height":16}]}`
	if err := os.WriteFile(world, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}

	check := CheckTiledWorld(world, dir)
	if !check.HasErrors() {
		t.Fatal("expected missing map error")
	}
}

func TestCheckTMJMissingTileset(t *testing.T) {
	dir := t.TempDir()
	tmj := filepath.Join(dir, "map.tmj")
	body := `{"tilesets":[{"firstgid":1,"source":"grass.tsx"}],"layers":[]}`
	if err := os.WriteFile(tmj, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}

	check := CheckTMJFile(tmj, nil)
	if !check.HasErrors() {
		t.Fatal("expected error for missing tileset on tmj")
	}
}

func TestCheckTMJUnknownEntdef(t *testing.T) {
	dir := t.TempDir()
	tmj := filepath.Join(dir, "map.tmj")
	body := `{"tilesets":[],"layers":[{"type":"objectgroup","objects":[{"type":"noSuchEntity","x":0,"y":0}]}]}`
	if err := os.WriteFile(tmj, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}

	check := CheckTMJFile(tmj, map[string]struct{}{"chest": {}})
	if !check.HasErrors() {
		t.Fatal("expected unknown entdef error")
	}
}

func TestCheckTexdefMissingTexture(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "rock.yaml")
	if err := os.WriteFile(path, []byte("---\ntraversable: false\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	check := CheckTexdef(path, t.TempDir())
	if !check.HasErrors() {
		t.Fatal("expected missing texture error")
	}
}

func TestFindFilesByExt(t *testing.T) {
	root := t.TempDir()
	world := filepath.Join(root, "worlds", "demo")
	if err := os.MkdirAll(world, 0o755); err != nil {
		t.Fatal(err)
	}
	tmj := filepath.Join(world, "0.0.tmj")
	if err := os.WriteFile(tmj, []byte(`{}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(world, "old.grid"), []byte(`{}`), 0o644); err != nil {
		t.Fatal(err)
	}

	tmjs, err := FindFilesByExt(root, ".tmj")
	if err != nil {
		t.Fatalf("FindFilesByExt: %v", err)
	}
	if len(tmjs) != 1 || tmjs[0] != tmj {
		t.Fatalf("tmj paths = %v, want [%q]", tmjs, tmj)
	}
}

func TestCount(t *testing.T) {
	check := Check{Type: "entdef", Filename: "x.yml"}
	check.AddWarning("w")
	check.AddError("e")
	errors, warnings := Count([]Check{check})
	if errors != 1 || warnings != 1 {
		t.Fatalf("Count = %d,%d", errors, warnings)
	}
}

func TestCheckEntdefPushableMustBeSolid(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "block.yml")
	body := "title: block\npushable: true\ninitialState: still\nstates:\n  still: {}\n"
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	check := CheckEntdef(path, &tiled.EntityVisualCatalog{})
	found := false
	for _, issue := range check.Issues {
		if strings.Contains(issue.Issue, "solid") {
			found = true
		}
	}
	if !found {
		t.Fatalf("issues = %+v, want a pushable-must-be-solid error", check.Issues)
	}
}

func TestCheckTMJResetAfter(t *testing.T) {
	for value, wantErr := range map[string]bool{`"1500"`: false, `1500`: false, `"0"`: true, `"soon"`: true, `"-5"`: true, `"2.5"`: true} {
		dir := t.TempDir()
		tmj := filepath.Join(dir, "map.tmj")
		body := `{"tilesets":[],"layers":[{"type":"objectgroup","objects":[{"type":"wallSwitch","x":0,"y":0,
			"properties":[{"name":"reset_after","type":"string","value":` + value + `}]}]}]}`
		if err := os.WriteFile(tmj, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
		check := CheckTMJFile(tmj, map[string]struct{}{"wallSwitch": {}})
		got := false
		for _, issue := range check.Issues {
			if strings.Contains(issue.Issue, "reset_after") {
				got = true
			}
		}
		if got != wantErr {
			t.Errorf("reset_after %s: error %v, want %v (%+v)", value, got, wantErr, check.Issues)
		}
	}
}

// Dialog `wear` entries and NPC looks are checked against the character
// rigs in res/ (adult by default).
func TestLintDialogsChecksWearAndLooksAgainstTheRig(t *testing.T) {
	res := filepath.Join("..", "..", "res")
	dat := t.TempDir()
	for _, d := range []string{"dialogs", "entdefs"} {
		if err := os.MkdirAll(filepath.Join(dat, d), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	write := func(rel, body string) {
		if err := os.WriteFile(filepath.Join(dat, rel), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("dialogs/salon.yml", `speaker: Test
start: hello
nodes:
  hello:
    text: "Hi"
    options:
      - {label: "Bun", wear: {hair: bun}}
      - {label: "Blonde", wear: {hair: blonde}}
      - {label: "Beard", wear: {beard: beard}}
      - {label: "Mohawk", wear: {hair: mohawk}}
      - {label: "Cape", wear: {cape: red}}
`)
	write("entdefs/good.yml", "title: good\nstates: {idle: {}}\nappearance:\n  parts: {hair: bob, beard: beard}\n  palettes: {hair: grey}\n")
	write("entdefs/bad.yml", "title: bad\nstates: {idle: {}}\nappearance:\n  parts: {hair: mohawk}\n")
	write("entdefs/tiny.yml", "title: tiny\nstates: {idle: {}}\nappearance:\n  rig: giant\n")

	var errs []string
	for _, c := range LintDialogs(dat, res) {
		for _, e := range c.Issues {
			if e.Severity == SeverityError {
				errs = append(errs, filepath.Base(c.Filename)+": "+e.Issue)
			}
		}
	}
	joined := strings.Join(errs, "\n")
	for _, want := range []string{`option "Mohawk"`, `option "Cape"`, "bad.yml: appearance", `tiny.yml: appearance rig "giant"`} {
		if !strings.Contains(joined, want) {
			t.Errorf("want an error mentioning %s; got:\n%s", want, joined)
		}
	}
	for _, unwanted := range []string{`option "Bun"`, `option "Blonde"`, `option "Beard"`, "good.yml"} {
		if strings.Contains(joined, unwanted) {
			t.Errorf("unexpected error for %s:\n%s", unwanted, joined)
		}
	}
}

// Map objects' `wear` (mannequin outfits) are checked against the rig too.
func TestLintWearPropertiesChecksMapOutfits(t *testing.T) {
	res := filepath.Join("..", "..", "res")
	dat := t.TempDir()
	w := filepath.Join(dat, "worlds", "w")
	if err := os.MkdirAll(w, 0o755); err != nil {
		t.Fatal(err)
	}
	doc := `{"layers":[{"type":"objectgroup","objects":[
 {"id":1,"type":"mannequin","x":16,"y":32,"properties":[{"name":"wear","type":"string","value":"outfit=dress,garment=blue,trim=white"}]},
 {"id":2,"type":"mannequin","x":32,"y":32,"properties":[{"name":"wear","type":"string","value":"outfit=gown,garment=plaid"}]},
 {"id":3,"type":"mannequin","x":48,"y":32,"properties":[{"name":"wear","type":"string","value":"just a dress"}]}]}]}`
	if err := os.WriteFile(filepath.Join(w, "0.tmj"), []byte(doc), 0o644); err != nil {
		t.Fatal(err)
	}
	var errs []string
	for _, c := range LintWearProperties(dat, res) {
		for _, e := range c.Issues {
			errs = append(errs, e.Issue)
		}
	}
	joined := strings.Join(errs, "\n")
	if strings.Contains(joined, "object 1 ") {
		t.Errorf("valid outfit flagged:\n%s", joined)
	}
	for _, want := range []string{`object 2 (mannequin) at (2,2): wear: unknown outfit variant "gown"`, `garment palette "plaid"`, "object 3 (mannequin) at (3,2): wear \"just a dress\": want key=value"} {
		if !strings.Contains(joined, want) {
			t.Errorf("want %s in:\n%s", want, joined)
		}
	}
}

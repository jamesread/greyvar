package lint

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	grids "github.com/jamesread/greyvar/datlib/gridfiles"
	"github.com/jamesread/greyvar/datlib/tiled"
)

type tmjDoc struct {
	Tilesets []tmjTilesetRef `json:"tilesets"`
	Layers   []tmjLayer      `json:"layers"`
}

type tmjTilesetRef struct {
	FirstGID uint32 `json:"firstgid"`
	Source   string `json:"source"`
}

type tmjLayer struct {
	Type    string         `json:"type"`
	Name    string         `json:"name"`
	Data    []uint32       `json:"data"`
	Objects []tmjMapObject `json:"objects"`
}

type tmjMapObject struct {
	Template string `json:"template"`
	Type     string `json:"type"`
	Class    string `json:"class"`
	Name     string `json:"name"`
}

// LintTMJMaps loads referenced tilesets then every worlds/**/*.tmj map.
func LintTMJMaps(datDir string, entdefNames map[string]struct{}) []Check {
	worldsRoot := filepath.Join(datDir, "worlds")
	tmjPaths, err := FindFilesByExt(worldsRoot, ".tmj")
	if err != nil {
		check := Check{Type: "map", Filename: worldsRoot}
		check.AddError(fmt.Sprintf("scan tmj maps: %v", err))
		return []Check{check}
	}

	if len(tmjPaths) == 0 {
		return nil
	}

	var out []Check
	tilesets, collectChecks := collectTilesetPaths(tmjPaths)
	out = append(out, collectChecks...)

	for _, path := range tilesets {
		if _, err := os.Stat(path); err != nil {
			continue
		}
		out = append(out, CheckTileset(path))
	}

	for _, path := range tmjPaths {
		out = append(out, CheckTMJFile(path, entdefNames))
	}
	return out
}

// LintTMXMaps warns on legacy worlds/**/*.tmx maps.
func LintTMXMaps(datDir string) []Check {
	worldsRoot := filepath.Join(datDir, "worlds")
	tmxPaths, err := FindFilesByExt(worldsRoot, ".tmx")
	if err != nil {
		check := Check{Type: "map", Filename: worldsRoot}
		check.AddError(fmt.Sprintf("scan tmx maps: %v", err))
		return []Check{check}
	}

	out := make([]Check, 0, len(tmxPaths))
	for _, path := range tmxPaths {
		check := Check{Type: "map", Filename: path}
		check.AddWarning("legacy .tmx map; server loads .tmj/.grid only")
		out = append(out, check)
	}
	return out
}

func collectTilesetPaths(tmjPaths []string) ([]string, []Check) {
	seen := map[string]struct{}{}
	var out []string
	var checks []Check

	for _, tmjPath := range tmjPaths {
		data, err := os.ReadFile(tmjPath)
		if err != nil {
			check := Check{Type: "map", Filename: tmjPath}
			check.AddError(fmt.Sprintf("read map while collecting tilesets: %v", err))
			checks = append(checks, check)
			continue
		}

		var doc tmjDoc
		if err := json.Unmarshal(data, &doc); err != nil {
			check := Check{Type: "map", Filename: tmjPath}
			check.AddError(fmt.Sprintf("parse map while collecting tilesets: %v", err))
			checks = append(checks, check)
			continue
		}

		mapDir := filepath.Dir(tmjPath)
		for _, ref := range doc.Tilesets {
			source := strings.TrimSpace(ref.Source)
			if source == "" {
				continue
			}
			resolved := filepath.Clean(filepath.Join(mapDir, filepath.FromSlash(source)))
			if _, ok := seen[resolved]; ok {
				continue
			}
			seen[resolved] = struct{}{}
			out = append(out, resolved)
		}
	}

	sort.Strings(out)
	return out, checks
}

// CheckTileset loads one .tsx tileset file.
func CheckTileset(path string) Check {
	check := Check{Type: "tileset", Filename: path}

	var warnings []string
	var err error
	withQuietLogs(func() {
		_, warnings, err = tiled.LoadTilesetFile(path)
	})
	for _, warning := range warnings {
		check.AddWarning(warning)
	}
	if err != nil {
		check.AddError(err.Error())
		return check
	}

	addStaleWangIssues(&check, path)

	return check
}

// addStaleWangIssues warns about terrain (wang) tags on tiles that are fully
// transparent in the tileset image, usually left behind by an image edit.
func addStaleWangIssues(check *Check, path string) {
	blank, err := blankTilesCached(path)
	if err != nil || len(blank) == 0 {
		return
	}

	sets, err := tiled.LoadWangSets(path)
	if err != nil {
		check.AddWarning(fmt.Sprintf("read wangsets: %v", err))
		return
	}

	for _, ws := range sets {
		var stale []int
		for _, t := range ws.Tiles {
			if blank[t.TileID] {
				stale = append(stale, t.TileID)
			}
		}
		if len(stale) > 0 {
			sort.Ints(stale)
			check.AddWarning(fmt.Sprintf("wangset %q tags tiles that are blank in the image: %v", ws.Name, stale))
		}
	}
}

var blankTilesByTileset = map[string]map[int]bool{}

// blankTilesCached returns tiled.BlankTiles for a .tsx, caching per path.
// Tilesets without a single atlas image yield no blank tiles.
func blankTilesCached(path string) (map[int]bool, error) {
	if blank, ok := blankTilesByTileset[path]; ok {
		return blank, nil
	}
	blank, err := tiled.BlankTiles(path)
	if err != nil {
		blank = nil
	}
	blankTilesByTileset[path] = blank
	return blank, err
}

// addBlankTileUseIssues warns when a map draws tiles that are fully
// transparent in their tileset image (they render as nothing).
func addBlankTileUseIssues(check *Check, mapDir string, doc tmjDoc) {
	type tilesetRange struct {
		first  uint32
		source string
		blank  map[int]bool
	}

	var ranges []tilesetRange
	for _, ref := range doc.Tilesets {
		source := strings.TrimSpace(ref.Source)
		if source == "" || ref.FirstGID == 0 {
			continue
		}
		resolved := filepath.Clean(filepath.Join(mapDir, filepath.FromSlash(source)))
		blank, _ := blankTilesCached(resolved)
		ranges = append(ranges, tilesetRange{first: ref.FirstGID, source: source, blank: blank})
	}
	if len(ranges) == 0 {
		return
	}
	sort.Slice(ranges, func(i, j int) bool { return ranges[i].first < ranges[j].first })

	for _, layer := range doc.Layers {
		if layer.Type != "tilelayer" {
			continue
		}
		counts := map[string]int{}
		for _, raw := range layer.Data {
			gid := tiled.DecodeGID(raw).ID
			if gid == 0 {
				continue
			}
			for i := len(ranges) - 1; i >= 0; i-- {
				if gid >= ranges[i].first {
					if ranges[i].blank[int(gid-ranges[i].first)] {
						counts[ranges[i].source]++
					}
					break
				}
			}
		}
		for source, n := range counts {
			check.AddWarning(fmt.Sprintf("layer %q draws %d cells with tiles that are blank in %s", layer.Name, n, filepath.Base(source)))
		}
	}
}

// CheckTMJFile loads one .tmj map and validates refs / entdefs.
func CheckTMJFile(path string, entdefNames map[string]struct{}) Check {
	check := Check{Type: "map", Filename: path}

	data, err := os.ReadFile(path)
	if err != nil {
		check.AddError(err.Error())
		return check
	}

	var doc tmjDoc
	if err := json.Unmarshal(data, &doc); err != nil {
		check.AddError(fmt.Sprintf("parse tmj: %v", err))
		return check
	}

	mapDir := filepath.Dir(path)
	for _, ref := range doc.Tilesets {
		source := strings.TrimSpace(ref.Source)
		if source == "" {
			continue
		}
		resolved := filepath.Clean(filepath.Join(mapDir, filepath.FromSlash(source)))
		if _, err := os.Stat(resolved); err != nil {
			check.AddError(fmt.Sprintf("missing tileset %q", source))
		}
	}

	addBlankTileUseIssues(&check, mapDir, doc)

	for _, layer := range doc.Layers {
		if layer.Type != "objectgroup" {
			continue
		}
		for _, obj := range layer.Objects {
			template := strings.TrimSpace(obj.Template)
			if template != "" {
				if issue := templateRefIssue(mapDir, template); issue != "" {
					check.AddError(issue)
				}
			}
			addObjectEntdefIssue(&check, mapDir, obj, entdefNames)
		}
	}

	withQuietLogs(func() {
		_, err = grids.ReadGrid(path)
	})
	if err != nil {
		check.AddError(err.Error())
	}

	return check
}

func addObjectEntdefIssue(check *Check, mapDir string, obj tmjMapObject, entdefNames map[string]struct{}) {
	name := firstNonEmpty(obj.Type, obj.Class, obj.Name)

	if template := strings.TrimSpace(obj.Template); template != "" {
		if tplPath := resolveTemplatePath(mapDir, template); tplPath != "" {
			var tpl *tiled.Template
			withQuietLogs(func() {
				tpl, _ = tiled.LoadTemplate(tplPath)
			})
			if tpl != nil && strings.TrimSpace(tpl.EntityName) != "" && name == "" {
				name = strings.TrimSpace(tpl.EntityName)
			}
		}
	}

	if name == "" || strings.EqualFold(name, "spawn") {
		return
	}

	if _, ok := entdefNames[name]; !ok {
		check.AddError(fmt.Sprintf("unknown entdef %q", name))
	}
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if s := strings.TrimSpace(v); s != "" {
			return s
		}
	}
	return ""
}

func resolveTemplatePath(mapDir, template string) string {
	path := filepath.Clean(filepath.Join(mapDir, filepath.FromSlash(template)))
	candidates := []string{path}
	if filepath.Ext(path) == "" {
		candidates = append(candidates, path+".tx")
	}
	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}
	return ""
}

func templateRefIssue(mapDir, template string) string {
	if resolveTemplatePath(mapDir, template) != "" {
		return ""
	}

	path := filepath.Clean(filepath.Join(mapDir, filepath.FromSlash(template)))
	candidates := []string{path}
	if filepath.Ext(path) == "" {
		candidates = append(candidates, path+".tx")
	}
	return fmt.Sprintf("missing object template %q (tried %s)", template, strings.Join(candidates, ", "))
}

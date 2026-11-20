package faces

// Tall wall faces: unlike the hills, the wall tiles themselves are drawn as
// the top of a taller face (open at the bottom), and the tileset holds the
// lower half. A tile with the property face_below (and face_below_short)
// continues its face into the cell below; this pass places those lower
// tiles on a locked layer "<layer> faces (generated)" directly under the
// layer holding the wall, so a nearer wall in that cell draws over it.
//
// The short variant (collision over its top 6px only) is used when the cell
// below starts the same terrain again under every half the face covers
// (two walls back to back); the mirror of tiled.FaceBelowShort in datlib.

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

// WallSuffix ends the name of every generated wall-face layer.
const WallSuffix = " faces (generated)"

var (
	reFaceBelow      = regexp.MustCompile(`<property name="face_below"[^>]*value="(\d+)"`)
	reFaceBelowShort = regexp.MustCompile(`<property name="face_below_short"[^>]*value="(\d+)"`)
	reFirstGID       = regexp.MustCompile(`"firstgid":\s*(\d+)`)
	reDataArray      = regexp.MustCompile(`"data":\s*\[([^\]]*)\]`)
	reObjectGID      = regexp.MustCompile(`"gid":\s*(\d+)`)
	reNumber         = regexp.MustCompile(`\d+`)
)

// faceBelows reads face_below / face_below_short per tile ({full, short}).
func faceBelows(tsx string) map[int][2]int {
	out := map[int][2]int{}
	for _, b := range reTileBlock.FindAllStringSubmatch(tsx, -1) {
		m := reFaceBelow.FindStringSubmatch(b[2])
		if m == nil {
			continue
		}
		id, _ := strconv.Atoi(b[1])
		full, _ := strconv.Atoi(m[1])
		short := full
		if s := reFaceBelowShort.FindStringSubmatch(b[2]); s != nil {
			short, _ = strconv.Atoi(s[1])
		}
		out[id] = [2]int{full, short}
	}
	return out
}

func isGenerated(name string) bool {
	return name == CapLayer || name == FaceLayer || strings.HasSuffix(name, WallSuffix)
}

// faceShort: corners are TL, TR, BR, BL (as readTsx stores them).
func faceShort(face, below [4]int, terrain int) bool {
	left := face[0] == terrain && face[3] != terrain
	right := face[1] == terrain && face[2] != terrain
	if !left && !right {
		return false
	}
	return (!left || below[3] == terrain) && (!right || below[2] == terrain)
}

type wallSet struct {
	first int
	path  string
	info  *tsxInfo
}

func mapSets(m tmjMap, dir string, skip string) ([]wallSet, error) {
	var sets []wallSet
	for _, t := range m.Tilesets {
		if t.Source == "" {
			sets = append(sets, wallSet{first: t.FirstGID})
			continue
		}
		p := filepath.Clean(filepath.Join(dir, filepath.FromSlash(t.Source)))
		s := wallSet{first: t.FirstGID, path: p}
		if p != skip {
			info, err := readTsx(p)
			if err != nil {
				return nil, err
			}
			s.info = info
		}
		sets = append(sets, s)
	}
	return sets, nil
}

func setOf(sets []wallSet, gid int64) (int, int) {
	g := int(gid & gidMask)
	best := -1
	for i, s := range sets {
		if g >= s.first && (best < 0 || s.first > sets[best].first) {
			best = i
		}
	}
	if g == 0 || best < 0 {
		return -1, 0
	}
	return best, g - sets[best].first
}

// applyWallFaces writes the wall-face layers into a map's text. It returns
// the new text and how many lower faces were placed.
func applyWallFaces(text, dir, skip string) (string, int, error) {
	var m tmjMap
	if err := json.Unmarshal([]byte(text), &m); err != nil {
		return "", 0, err
	}
	sets, err := mapSets(m, dir, skip)
	if err != nil {
		return "", 0, err
	}
	W, H := m.Width, m.Height
	type placement struct{ set, id int }
	byLayer := map[string][]placement{} // source layer -> lower faces per cell (set -1 = none)
	var order []string
	existing := map[string]bool{}
	need := map[int]int{} // set -> tile count it must have room for
	placed := 0
	for _, l := range m.Layers {
		if l.Type != "tilelayer" {
			continue
		}
		if isGenerated(l.Name) {
			if strings.HasSuffix(l.Name, WallSuffix) {
				existing[strings.TrimSuffix(l.Name, WallSuffix)] = true
			}
			continue
		}
		var cells []placement
		for r := 0; r+1 < H; r++ {
			for c := 0; c < W; c++ {
				cell := r*W + c
				if cell >= len(l.Data) {
					continue
				}
				si, local := setOf(sets, l.Data[cell])
				if si < 0 || sets[si].info == nil {
					continue
				}
				fb, ok := sets[si].info.faceBelows[local]
				if !ok {
					continue
				}
				id := fb[0]
				for _, ws := range sets[si].info.wangsets {
					fc, ok := ws.corners[local]
					if !ok {
						continue
					}
					terrain := fc[0]
					if terrain == 0 {
						terrain = fc[1]
					}
					var bc [4]int
					if bsi, blocal := setOf(sets, l.Data[cell+W]); bsi == si {
						bc = ws.corners[blocal]
					}
					if faceShort(fc, bc, terrain) {
						id = fb[1]
					}
					break
				}
				if cells == nil {
					cells = make([]placement, W*H)
					for i := range cells {
						cells[i].set = -1
					}
				}
				cells[cell+W] = placement{si, id}
				if id+1 > need[si] {
					need[si] = id + 1
				}
				placed++
			}
		}
		if cells != nil {
			byLayer[l.Name] = cells
			order = append(order, l.Name)
		}
	}
	for name := range existing {
		if _, ok := byLayer[name]; !ok {
			order = append(order, name) // clear a stale layer
		}
	}
	if len(order) == 0 {
		return text, 0, nil
	}

	// Make room for the lower-face tiles in each tileset's gid range, the
	// way Tiled renumbers firstgids when a tileset has grown.
	for si, n := range need {
		if text, err = ensureRoom(text, sets[si].first, max(n, sets[si].info.count)); err != nil {
			return "", 0, err
		}
	}
	if err := json.Unmarshal([]byte(text), &m); err != nil {
		return "", 0, err
	}
	if sets, err = mapSets(m, dir, skip); err != nil {
		return "", 0, err
	}
	for _, name := range order {
		data := make([]int64, W*H)
		for i, p := range byLayer[name] {
			if p.set >= 0 {
				data[i] = int64(sets[p.set].first + p.id)
			}
		}
		if text, err = upsertLayerBefore(text, name+WallSuffix, data, W, H, name); err != nil {
			return "", 0, err
		}
	}
	return text, placed, nil
}

// ensureRoom shifts every tileset after the one at first (and every gid
// into them) so that it has room for count tiles.
func ensureRoom(text string, first, count int) (string, error) {
	elems, _, err := arrayElements(text, `"tilesets":`)
	if err != nil {
		return "", err
	}
	next := 0
	for _, e := range elems {
		if m := reFirstGID.FindStringSubmatch(text[e[0]:e[1]]); m != nil {
			if g, _ := strconv.Atoi(m[1]); g > first && (next == 0 || g < next) {
				next = g
			}
		}
	}
	if next == 0 || first+count <= next {
		return text, nil
	}
	delta := first + count - next
	shift := func(s string) string {
		v, _ := strconv.ParseInt(s, 10, 64)
		if g := v & gidMask; g >= int64(next) {
			v = (v &^ gidMask) | (g + int64(delta))
		}
		return strconv.FormatInt(v, 10)
	}
	// Edit back to front so earlier offsets stay valid.
	for i := len(elems) - 1; i >= 0; i-- {
		e := elems[i]
		seg := reFirstGID.ReplaceAllStringFunc(text[e[0]:e[1]], func(m string) string {
			n := reNumber.FindString(m)
			return strings.Replace(m, n, shift(n), 1)
		})
		text = text[:e[0]] + seg + text[e[1]:]
	}
	layers, _, err := arrayElements(text, `"layers":`)
	if err != nil {
		return "", err
	}
	for i := len(layers) - 1; i >= 0; i-- {
		e := layers[i]
		seg := reDataArray.ReplaceAllStringFunc(text[e[0]:e[1]], func(m string) string {
			open := strings.Index(m, "[")
			return m[:open] + reNumber.ReplaceAllStringFunc(m[open:], shift)
		})
		seg = reObjectGID.ReplaceAllStringFunc(seg, func(m string) string {
			n := reNumber.FindString(m)
			return strings.Replace(m, n, shift(n), 1)
		})
		text = text[:e[0]] + seg + text[e[1]:]
	}
	return text, nil
}

// upsertLayerBefore replaces the named layer's data, or inserts a new locked
// tile layer just before the layer called before.
func upsertLayerBefore(text, name string, data []int64, w, h int, before string) (string, error) {
	elems, _, err := arrayElements(text, `"layers":`)
	if err != nil {
		return "", err
	}
	at := -1
	for i, e := range elems {
		var l tmjLayer
		if err := json.Unmarshal([]byte(text[e[0]:e[1]]), &l); err != nil {
			return "", err
		}
		if l.Name == name {
			return text[:e[0]] + strings.TrimLeft(layerText(name, l.ID, data, w, h), " ") + text[e[1]:], nil
		}
		if l.Name == before && at < 0 {
			at = i
		}
	}
	if at < 0 {
		return "", fmt.Errorf("no layer %q", before)
	}
	nm := reNextLayer.FindStringSubmatchIndex(text)
	if nm == nil {
		return "", fmt.Errorf("missing nextlayerid")
	}
	id, _ := strconv.Atoi(text[nm[2]:nm[3]])
	text = text[:nm[2]] + strconv.Itoa(id+1) + text[nm[3]:]
	elems, _, _ = arrayElements(text, `"layers":`)
	start := elems[at][0]
	return text[:start] + strings.TrimLeft(layerText(name, id, data, w, h), " ") + ", \n" + strings.Repeat(" ", 8) + text[start:], nil
}

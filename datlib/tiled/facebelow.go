package tiled

import (
	"encoding/xml"
	"fmt"
	"os"
	"strconv"
)

// FaceBelow names the lower half of a tall south face: a tile whose face
// continues into the cell below has the property face_below (and
// face_below_short) set to the tile drawn there. Both lower tiles carry the
// same art; Full collides over its whole face, Short only over its top 6px,
// for when the cell below starts another wall (which then blocks the rest).
type FaceBelow struct {
	Full  int
	Short int
}

type tsxFaceFile struct {
	Tiles []struct {
		ID         int `xml:"id,attr"`
		Properties []struct {
			Name  string `xml:"name,attr"`
			Value string `xml:"value,attr"`
		} `xml:"properties>property"`
	} `xml:"tile"`
}

// FacesBelow reads every tile's face_below / face_below_short properties.
// Short falls back to Full when only face_below is set.
func FacesBelow(tsxPath string) (map[int]FaceBelow, error) {
	raw, err := os.ReadFile(tsxPath)
	if err != nil {
		return nil, err
	}
	var f tsxFaceFile
	if err := xml.Unmarshal(raw, &f); err != nil {
		return nil, fmt.Errorf("%s: %w", tsxPath, err)
	}
	out := map[int]FaceBelow{}
	for _, t := range f.Tiles {
		fb := FaceBelow{Full: -1, Short: -1}
		for _, p := range t.Properties {
			n, err := strconv.Atoi(p.Value)
			switch p.Name {
			case "face_below":
				if err != nil {
					return nil, fmt.Errorf("%s: tile %d: face_below %q is not a tile id", tsxPath, t.ID, p.Value)
				}
				fb.Full = n
			case "face_below_short":
				if err != nil {
					return nil, fmt.Errorf("%s: tile %d: face_below_short %q is not a tile id", tsxPath, t.ID, p.Value)
				}
				fb.Short = n
			}
		}
		if fb.Full < 0 {
			continue
		}
		if fb.Short < 0 {
			fb.Short = fb.Full
		}
		out[t.ID] = fb
	}
	return out, nil
}

// FaceSides reports which halves of a face tile (corners of one wangset
// colour, TL, TR, BR, BL as in Corners) drop to lower ground: a side faces
// south when its top corner is the terrain and its bottom corner is not.
func FaceSides(c Corners, terrain int) (left, right bool) {
	return c[0] == terrain && c[3] != terrain, c[1] == terrain && c[2] != terrain
}

// FaceBelowShort reports whether the lower face under a face tile should be
// the short one: the cell below starts the same terrain again (another wall)
// under every half the face covers, so that wall's own collision takes over.
func FaceBelowShort(face, below Corners, terrain int) bool {
	left, right := FaceSides(face, terrain)
	if !left && !right {
		return false
	}
	return (!left || below[3] == terrain) && (!right || below[2] == terrain)
}

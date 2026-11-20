package tiled

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFacesBelow(t *testing.T) {
	path := filepath.Join(t.TempDir(), "t.tsx")
	os.WriteFile(path, []byte(`<?xml version="1.0" encoding="UTF-8"?>
<tileset name="t" tilewidth="16" tileheight="16" tilecount="4" columns="4">
 <tile id="1">
  <properties>
   <property name="face_below" type="int" value="2"/>
   <property name="face_below_short" type="int" value="3"/>
  </properties>
 </tile>
 <tile id="0"><properties><property name="face_below" type="int" value="2"/></properties></tile>
</tileset>`), 0o644)
	got, err := FacesBelow(path)
	if err != nil {
		t.Fatal(err)
	}
	if got[1] != (FaceBelow{2, 3}) || got[0] != (FaceBelow{2, 2}) || len(got) != 2 {
		t.Fatalf("got %+v", got)
	}
}

func TestFaceBelowShort(t *testing.T) {
	south := Corners{1, 1, 0, 0} // stone on top, open below: a full-width face
	wall := Corners{0, 0, 1, 1}  // another wall starting in the cell below
	open := Corners{0, 0, 0, 0}  // open ground below
	half := Corners{0, 0, 1, 0}  // a wall starting under the right half only
	endR := Corners{0, 1, 0, 0}  // face under the right half only
	if !FaceBelowShort(south, wall, 1) || FaceBelowShort(south, open, 1) || FaceBelowShort(south, half, 1) {
		t.Fatal("a full-width face is short only over a full-width wall")
	}
	if !FaceBelowShort(endR, half, 1) || FaceBelowShort(endR, Corners{0, 0, 0, 1}, 1) {
		t.Fatal("a right-half face only needs the wall under its right half")
	}
}

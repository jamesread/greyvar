package greyvarserver

import (
	"testing"

	"github.com/jamesread/greyvar/datlib/tiled"
)

func TestTileAnimationsToPB(t *testing.T) {
	got := tileAnimationsToPB(map[int][]tiled.AnimFrame{
		9: {{TileID: 9, DurationMs: 100}, {TileID: 10, DurationMs: 150}},
		2: {{TileID: 2, DurationMs: 200}, {TileID: 3, DurationMs: 200}},
	}, map[int]bool{9: true})
	if len(got) != 2 || got[0].TileId != 2 || got[1].TileId != 9 {
		t.Fatalf("want sorted by tile id, got %+v", got)
	}
	if got[1].Frames[1] != 10 || got[1].DurationsMs[1] != 150 {
		t.Fatalf("frames/durations wrong: %+v", got[1])
	}
	if !got[1].Synced || got[0].Synced {
		t.Fatal("synced flag should follow the tile")
	}
	if len(tileAnimationsToPB(nil, nil)) != 0 {
		t.Fatal("nil map should give no animations")
	}
}

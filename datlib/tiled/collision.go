package tiled

import "math"

// CollisionRect is a blocking rectangle in tile-local pixel coordinates.
type CollisionRect struct {
	X float64
	Y float64
	W float64
	H float64
}

func collisionRectsFromObjects(objects []tsxObject) []CollisionRect {
	if len(objects) == 0 {
		return nil
	}

	out := make([]CollisionRect, 0, len(objects))
	for _, obj := range objects {
		if obj.Width <= 0 || obj.Height <= 0 {
			continue
		}
		out = append(out, snapCollisionRect(obj.X, obj.Y, obj.Width, obj.Height))
	}

	return out
}

// snapCollisionRect rounds a rect's edges to whole pixels. Shapes drawn by
// hand in Tiled land on fractions (5.909, 9.727...); movement is in whole
// pixels, so unsnapped edges stop the player up to 1px short of the art.
// A shape never snaps away entirely: it keeps at least 1px per axis.
func snapCollisionRect(x, y, w, h float64) CollisionRect {
	x0, y0 := math.Round(x), math.Round(y)
	x1, y1 := math.Round(x+w), math.Round(y+h)
	if x1 <= x0 {
		x1 = x0 + 1
	}
	if y1 <= y0 {
		y1 = y0 + 1
	}
	return CollisionRect{X: x0, Y: y0, W: x1 - x0, H: y1 - y0}
}

func collisionRectsFromTSXTile(tile tsxTile) []CollisionRect {
	return collisionRectsFromObjects(tile.ObjectGroup.Objects)
}

// TransformCollisionRect applies tile flip flags to a tile-local collision rect.
func TransformCollisionRect(r CollisionRect, tileWidth, tileHeight float64, flipH, flipV bool) CollisionRect {
	out := r
	if flipH {
		out.X = tileWidth - r.X - r.W
	}
	if flipV {
		out.Y = tileHeight - r.Y - r.H
	}
	return out
}

package rig

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
	"strconv"
	"strings"
)

// RGB is an opaque colour.
type RGB struct{ R, G, B uint8 }

// ParseHex parses #rrggbb.
func ParseHex(s string) (RGB, error) {
	s = strings.TrimPrefix(strings.TrimSpace(s), "#")
	if len(s) != 6 {
		return RGB{}, fmt.Errorf("colour %q: want #rrggbb", s)
	}
	v, err := strconv.ParseUint(s, 16, 32)
	if err != nil {
		return RGB{}, fmt.Errorf("colour %q: %w", s, err)
	}
	return RGB{uint8(v >> 16), uint8(v >> 8), uint8(v)}, nil
}

// Hex formats as #rrggbb.
func (c RGB) Hex() string { return fmt.Sprintf("#%02x%02x%02x", c.R, c.G, c.B) }

// Luma is perceptual brightness 0..255.
func (c RGB) Luma() float64 {
	return 0.299*float64(c.R) + 0.587*float64(c.G) + 0.114*float64(c.B)
}

// HSL returns hue (0..360), saturation and lightness (0..1).
func (c RGB) HSL() (h, s, l float64) {
	r, g, b := float64(c.R)/255, float64(c.G)/255, float64(c.B)/255
	mx := max(r, g, b)
	mn := min(r, g, b)
	l = (mx + mn) / 2
	if mx == mn {
		return 0, 0, l
	}
	d := mx - mn
	if l > 0.5 {
		s = d / (2 - mx - mn)
	} else {
		s = d / (mx + mn)
	}
	switch mx {
	case r:
		h = (g - b) / d
		if g < b {
			h += 6
		}
	case g:
		h = (b-r)/d + 2
	default:
		h = (r-g)/d + 4
	}
	return h * 60, s, l
}

func rgbAt(img *image.NRGBA, x, y int) (RGB, bool) {
	c := img.NRGBAAt(x, y)
	if c.A < 128 {
		return RGB{}, false
	}
	return RGB{c.R, c.G, c.B}, true
}

func setRGB(img *image.NRGBA, x, y int, c RGB) {
	img.SetNRGBA(x, y, color.NRGBA{c.R, c.G, c.B, 255})
}

// ReadPNG loads a PNG as NRGBA.
func ReadPNG(path string) (*image.NRGBA, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	src, err := png.Decode(f)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	b := src.Bounds()
	out := image.NewNRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
	draw.Draw(out, out.Bounds(), src, b.Min, draw.Src)
	return out, nil
}

// EncodePNG encodes deterministically (Go's encoder has no timestamps).
func EncodePNG(img image.Image) []byte {
	var buf bytes.Buffer
	enc := png.Encoder{CompressionLevel: png.BestCompression}
	if err := enc.Encode(&buf, img); err != nil {
		panic(err) // only fails on writer errors; bytes.Buffer never errors
	}
	return buf.Bytes()
}

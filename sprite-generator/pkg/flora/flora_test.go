package flora

import (
	"bytes"
	"image"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jamesread/greyvar/sprite-generator/pkg/rig"
)

const testYML = `name: testflora
tileSize: 16
columns: 4
outputs: {image: out/flora.png, tsx: out/flora.tsx}
sway: {offsets: [0, 1, 0, -1], ms: [100, 200, 100, 200]}
flora:
  - name: dot
    sources: [grass.tsx#5]
    heads: [[4, 2, 6, 4]]
`

func writeFlora(t *testing.T, yml string) *Flora {
	t.Helper()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "flora.yml"), []byte(yml), 0o644); err != nil {
		t.Fatal(err)
	}
	img := image.NewNRGBA(image.Rect(0, 0, 16, 16))
	head, _ := rig.ParseHex("#ff0000")
	stem, _ := rig.ParseHex("#00ff00")
	img.Pix[img.PixOffset(5, 3)+0], img.Pix[img.PixOffset(5, 3)+3] = head.R, 255
	for y := 5; y < 10; y++ {
		o := img.PixOffset(5, y)
		img.Pix[o+1], img.Pix[o+3] = stem.G, 255
	}
	_ = os.MkdirAll(filepath.Join(dir, "parts"), 0o755)
	if err := os.WriteFile(filepath.Join(dir, "parts", "dot.png"), rig.EncodePNG(img), 0o644); err != nil {
		t.Fatal(err)
	}
	f, err := Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	return f
}

func TestFrameSwaysHeadOnly(t *testing.T) {
	f := writeFlora(t, testYML)
	part, _ := rig.ReadPNG(f.partPath(f.Spec.Flora[0]))
	for fr, wantX := range []int{5, 6, 5, 4} {
		img := f.frame(f.Spec.Flora[0], part, fr)
		if img.NRGBAAt(wantX, 3).A == 0 {
			t.Errorf("frame %d: head should be at x=%d", fr, wantX)
		}
		if img.NRGBAAt(5, 7).A == 0 {
			t.Errorf("frame %d: stem must not move", fr)
		}
	}
}

func TestBuildTsxAndDeterminism(t *testing.T) {
	f := writeFlora(t, testYML)
	a, err := f.Build()
	if err != nil {
		t.Fatal(err)
	}
	b, _ := f.Build()
	for i := range a {
		if !bytes.Equal(a[i].Data, b[i].Data) {
			t.Fatalf("%s not deterministic", a[i].Path)
		}
	}
	tsx := string(a[1].Data)
	for _, want := range []string{`tilecount="4"`, `type="testflora"`, `name="decor" type="bool" value="true"`, `name="flora" value="dot"`, `name="sources" value="grass.tsx#5"`, `<frame tileid="3" duration="200"/>`, `source="flora.png"`} {
		if !strings.Contains(tsx, want) {
			t.Errorf("tsx missing %s:\n%s", want, tsx)
		}
	}
}

func TestValidateCatchesMistakes(t *testing.T) {
	cases := map[string]string{
		"out of the tile":     strings.Replace(testYML, "[[4, 2, 6, 4]]", "[[0, 2, 15, 4]]\n    phases: [0]", 1),
		"same length":         strings.Replace(testYML, "ms: [100, 200, 100, 200]", "ms: [100]", 1),
		"one entry per head":  strings.Replace(testYML, "heads: [[4, 2, 6, 4]]", "heads: [[4, 2, 6, 4]]\n    phases: [0, 1]", 1),
		"at least one head":   strings.Replace(testYML, "heads: [[4, 2, 6, 4]]", "heads: []", 1),
	}
	// Make "out of the tile" actually push a pixel off: move the head pixel to x=0.
	for want, yml := range cases {
		f := writeFlora(t, yml)
		if want == "out of the tile" {
			img := image.NewNRGBA(image.Rect(0, 0, 16, 16))
			img.Pix[img.PixOffset(0, 3)+3] = 255
			_ = os.WriteFile(f.partPath(f.Spec.Flora[0]), rig.EncodePNG(img), 0o644)
		}
		errs := strings.Join(f.Validate(), "\n")
		if !strings.Contains(errs, want) {
			t.Errorf("want %q in:\n%s", want, errs)
		}
	}
}

func TestPropsAndNonDecor(t *testing.T) {
	yml := strings.Replace(testYML, "name: testflora", "name: testcrops\ndecor: false", 1)
	yml = strings.Replace(yml, "    heads: [[4, 2, 6, 4]]", "    heads: [[4, 2, 6, 4]]\n    props: {crop: hay, stage: \"2\"}", 1)
	f := writeFlora(t, yml)
	outs, err := f.Build()
	if err != nil {
		t.Fatal(err)
	}
	tsx := string(outs[1].Data)
	if strings.Contains(tsx, `name="decor"`) {
		t.Error("decor: false must not mark tiles decor")
	}
	for _, want := range []string{`type="testcrops"`, `<property name="crop" value="hay"/>`, `<property name="stage" value="2"/>`, "res/sprites/testcrops"} {
		if !strings.Contains(tsx, want) {
			t.Errorf("tsx missing %s", want)
		}
	}
}

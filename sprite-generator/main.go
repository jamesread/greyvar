// spritegen builds layered sprite rigs (see pkg/rig and res/sprites/README.md).
//
//	spritegen check   [rig-dir]   validate the rig and fail if outputs are stale
//	spritegen build   [rig-dir]   regenerate sheets, rig.json, composite and .tsx
//	spritegen preview [rig-dir] -out file.png [-clip idle] [-set hair=blonde,style=long] [-scale 8]
//	spritegen part show head/default/idle_0 [rig-dir]
//	spritegen part derive -layer head -from idle_0 -to blink -edit 6,4=skin.0 [-edit ...] [rig-dir]
//	spritegen part prune [-n]   (delete parts no pose references)
//	spritegen part longhair -from short -to long [-rows 3]   (derive a whole hair variant)
//	spritegen part skirt [-from default -to skirt -rows 2]   (derive skirts from trousers)
//	spritegen import-pixellab -src <frames> [-poses] [-variant hair=long] [-layers hair]   (PixelLab character -> rig layers)
//	spritegen part outfit -to dress [-style dress] [-poses] [-rows 3]   (one-piece outfits over shirt + trousers)
//	spritegen flora build|check [flora-dir]        (animated flora tileset)
//	spritegen flora preview -out file.png [flora-dir]
//	spritegen flora migrate [-n] map.tmj...         (baked flower tiles -> decor layer)
//	spritegen tiles animate|animate-check [dir]      (animate tiles of one colour, e.g. water)
//	spritegen tiles remap [-n] -from-image water.png -to res/.../water.tsx#0 map.tmj...
//	spritegen import  -sheet ss.png -tsx player.tsx -out res/sprites/player   (one-time migration)
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/jamesread/greyvar/sprite-generator/pkg/faces"
	"github.com/jamesread/greyvar/sprite-generator/pkg/flora"
	"github.com/jamesread/greyvar/sprite-generator/pkg/rig"
	"github.com/jamesread/greyvar/sprite-generator/pkg/tileanim"
)

const defaultRig = "res/sprites/player"

func main() {
	if len(os.Args) < 2 {
		usage()
	}
	cmd, args := os.Args[1], os.Args[2:]
	var err error
	switch cmd {
	case "check":
		err = runCheck(args)
	case "build":
		err = runBuild(args)
	case "preview":
		err = runPreview(args)
	case "part":
		err = runPart(args)
	case "flora":
		err = runFlora(args)
	case "tiles":
		err = runTiles(args)
	case "import":
		err = runImport(args)
	case "import-pixellab":
		err = runImportPixelLab(args)
	default:
		usage()
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "spritegen:", err)
		os.Exit(1)
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, "usage: spritegen check|build|preview|part|flora|tiles|import|import-pixellab [flags] [rig-dir]")
	os.Exit(2)
}

func rigDir(fs *flag.FlagSet) string {
	if fs.NArg() > 0 {
		return fs.Arg(0)
	}
	return defaultRig
}

func runCheck(args []string) error {
	fs := flag.NewFlagSet("check", flag.ExitOnError)
	_ = fs.Parse(args)
	r, err := rig.Load(rigDir(fs))
	if err != nil {
		return err
	}
	outs, err := r.Build()
	if err != nil {
		return err
	}
	if stale := rig.Stale(outs); len(stale) > 0 {
		return fmt.Errorf("generated files are out of date; run `make sprites`:\n  %s", strings.Join(stale, "\n  "))
	}
	fmt.Printf("ok: %s (%d poses, %d outputs up to date)\n", r.Spec.Name, len(r.Poses), len(outs))
	return nil
}

func runBuild(args []string) error {
	fs := flag.NewFlagSet("build", flag.ExitOnError)
	_ = fs.Parse(args)
	r, err := rig.Load(rigDir(fs))
	if err != nil {
		return err
	}
	outs, err := r.Build()
	if err != nil {
		return err
	}
	stale := map[string]bool{}
	for _, p := range rig.Stale(outs) {
		stale[p] = true
	}
	for _, o := range outs {
		if !stale[o.Path] {
			continue
		}
		if err := os.MkdirAll(filepath.Dir(o.Path), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(o.Path, o.Data, 0o644); err != nil {
			return err
		}
		fmt.Println("wrote", o.Path)
	}
	fmt.Printf("built %s: %d poses, %d outputs (%d changed)\n", r.Spec.Name, len(r.Poses), len(outs), len(stale))
	return nil
}

func runPreview(args []string) error {
	fs := flag.NewFlagSet("preview", flag.ExitOnError)
	out := fs.String("out", "", "output PNG (required)")
	clip := fs.String("clip", "", "clip to render (default: every pose)")
	set := fs.String("set", "", "look overrides: <slot>=<palette> or <layer>=<variant>, comma separated")
	scale := fs.Int("scale", 8, "pixel scale")
	_ = fs.Parse(reorder(args))
	if *out == "" {
		return fmt.Errorf("preview: -out is required")
	}
	r, err := rig.Load(rigDir(fs))
	if err != nil {
		return err
	}
	opts := rig.PreviewOptions{Clip: *clip, Scale: *scale, Parts: map[string]string{}, Palettes: map[string]string{}}
	if *set != "" {
		for _, kv := range strings.Split(*set, ",") {
			k, v, ok := strings.Cut(kv, "=")
			if !ok {
				return fmt.Errorf("preview: bad -set entry %q", kv)
			}
			// Slot and layer names overlap (hair, shirt, pants): a value that
			// names a palette is a palette, anything else is a part variant.
			if _, isPalette := r.Spec.Palettes[k][v]; isPalette {
				opts.Palettes[k] = v
			} else if _, isLayer := r.Layer(k); isLayer {
				opts.Parts[k] = v
			} else {
				return fmt.Errorf("preview: %q is neither a %s palette nor a layer variant", v, k)
			}
		}
	}
	img, err := r.Preview(opts)
	if err != nil {
		return err
	}
	return os.WriteFile(*out, rig.EncodePNG(img), 0o644)
}

func runImport(args []string) error {
	fs := flag.NewFlagSet("import", flag.ExitOnError)
	sheet := fs.String("sheet", "res/img/textures/entities/ss_playerBob.png", "baked sheet")
	tsx := fs.String("tsx", "server/dat/entdefs/player.tsx", "tileset with clip animations")
	out := fs.String("out", defaultRig, "rig directory to create")
	name := fs.String("name", "player", "entity name")
	force := fs.Bool("force", false, "overwrite a non-empty rig directory")
	_ = fs.Parse(args)
	return rig.Import(rig.ImportOptions{
		Sheet: *sheet, Tsx: *tsx, OutDir: *out, Name: *name, Force: *force,
		Outputs: rig.Outputs{
			SheetsDir: "../../img/textures/entities/" + *name,
			Composite: "../../img/textures/entities/" + *name + "/composite.png",
			Tsx:       "../../../server/dat/entdefs/" + *name + ".tsx",
			WebRoot:   "../..",
		},
	})
}

// reorder moves a leading positional rig-dir after the flags so both
// `preview dir -out x` and `preview -out x dir` work.
func reorder(args []string) []string {
	if len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		return append(args[1:], args[0])
	}
	return args
}

type multiFlag []string

func (m *multiFlag) String() string     { return strings.Join(*m, " ") }
func (m *multiFlag) Set(v string) error { *m = append(*m, v); return nil }

func runPart(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("part: want show or derive")
	}
	switch args[0] {
	case "show":
		fs := flag.NewFlagSet("part show", flag.ExitOnError)
		_ = fs.Parse(args[1:])
		if fs.NArg() < 1 {
			return fmt.Errorf("part show: want <layer>/<variant>/<part> [rig-dir]")
		}
		ref := strings.Split(fs.Arg(0), "/")
		if len(ref) != 3 {
			return fmt.Errorf("part show: want <layer>/<variant>/<part>, got %q", fs.Arg(0))
		}
		dir := defaultRig
		if fs.NArg() > 1 {
			dir = fs.Arg(1)
		}
		r, err := rig.Load(dir)
		if err != nil {
			return err
		}
		out, err := r.ShowPart(ref[0], ref[1], ref[2])
		if err != nil {
			return err
		}
		fmt.Print(out)
		return nil
	case "derive":
		fs := flag.NewFlagSet("part derive", flag.ExitOnError)
		layer := fs.String("layer", "", "layer (required)")
		variant := fs.String("variant", "all", "variant, or 'all' to derive in every variant of the layer")
		from := fs.String("from", "", "existing part to copy (empty = blank)")
		to := fs.String("to", "", "new part name (required)")
		force := fs.Bool("force", false, "overwrite an existing part")
		var edits multiFlag
		fs.Var(&edits, "edit", "x,y=colour or x1-x2,y=colour (colour: <slot>.<n>, fixed.<n>, clear); repeatable")
		_ = fs.Parse(args[1:])
		if *layer == "" || *to == "" {
			return fmt.Errorf("part derive: -layer and -to are required")
		}
		r, err := rig.Load(rigDir(fs))
		if err != nil {
			return err
		}
		l, ok := r.Layer(*layer)
		if !ok {
			return fmt.Errorf("part derive: unknown layer %q", *layer)
		}
		var parsed []rig.PixelEdit
		for _, e := range edits {
			pe, err := rig.ParsePixelEdit(e)
			if err != nil {
				return err
			}
			parsed = append(parsed, pe)
		}
		variants := []string{*variant}
		if *variant == "all" {
			variants = rig.LayerVariants(l)
		}
		for _, v := range variants {
			path, err := r.DerivePart(*layer, v, *from, *to, parsed, *force)
			if err != nil {
				return err
			}
			fmt.Println("wrote", path)
		}
		return nil
	case "prune":
		fs := flag.NewFlagSet("part prune", flag.ExitOnError)
		dry := fs.Bool("n", false, "only list unused parts")
		_ = fs.Parse(args[1:])
		r, err := rig.Load(rigDir(fs))
		if err != nil {
			return err
		}
		unused, err := r.UnusedParts()
		if err != nil {
			return err
		}
		for _, f := range unused {
			if *dry {
				fmt.Println("unused", f)
				continue
			}
			if err := os.Remove(f); err != nil {
				return err
			}
			fmt.Println("removed", f)
		}
		fmt.Printf("%d unused parts\n", len(unused))
		return nil
	case "longhair":
		fs := flag.NewFlagSet("part longhair", flag.ExitOnError)
		layer := fs.String("layer", "hair", "hair layer")
		from := fs.String("from", "short", "source variant")
		to := fs.String("to", "long", "new variant")
		rows := fs.Int("rows", 3, "rows to extend below the hairline")
		force := fs.Bool("force", false, "overwrite existing parts")
		_ = fs.Parse(args[1:])
		r, err := rig.Load(rigDir(fs))
		if err != nil {
			return err
		}
		paths, err := r.DeriveLongHair(rig.LongHairOptions{Layer: *layer, From: *from, To: *to, Rows: *rows, Force: *force})
		if err != nil {
			return err
		}
		fmt.Printf("wrote %d parts to %s/%s; add %q to the layer's variants in rig.yml\n", len(paths), *layer, *to, *to)
		return nil
	case "skirt":
		fs := flag.NewFlagSet("part skirt", flag.ExitOnError)
		layer := fs.String("layer", "pants", "legwear layer")
		from := fs.String("from", "default", "source variant (trousers)")
		to := fs.String("to", "skirt", "new variant")
		rows := fs.Int("rows", 2, "skirt height from the top of the legs")
		force := fs.Bool("force", false, "overwrite existing parts")
		_ = fs.Parse(args[1:])
		r, err := rig.Load(rigDir(fs))
		if err != nil {
			return err
		}
		paths, err := r.DeriveSkirt(rig.SkirtOptions{Layer: *layer, From: *from, To: *to, Rows: *rows, Force: *force})
		if err != nil {
			return err
		}
		fmt.Printf("wrote %d parts to %s/%s; add %q to the layer's variants in rig.yml\n", len(paths), *layer, *to, *to)
		return nil
	case "outfit":
		fs := flag.NewFlagSet("part outfit", flag.ExitOnError)
		to := fs.String("to", "", "variant to write (e.g. none, dress)")
		style := fs.String("style", "", "none or dress (default: the variant name)")
		poses := fs.Bool("poses", false, "first point every pose's outfit part at its shirt/trousers (rewrites poses.yml)")
		rows := fs.Int("rows", 3, "skirt height below the waist, including the hem")
		force := fs.Bool("force", false, "overwrite existing parts")
		_ = fs.Parse(args[1:])
		dir := rigDir(fs)
		r, err := rig.Load(dir)
		if err != nil {
			return err
		}
		o := rig.OutfitOptions{To: *to, Style: *style, SkirtRows: *rows, Force: *force}
		if *poses {
			n, err := r.SetOutfitPoses(o)
			if err != nil {
				return err
			}
			fmt.Printf("poses.yml: %d poses now name their outfit part\n", n)
			if r, err = rig.Load(dir); err != nil {
				return err
			}
		}
		if *to == "" {
			return nil
		}
		paths, err := r.DeriveOutfit(o)
		if err != nil {
			return err
		}
		fmt.Printf("wrote %d parts to outfit/%s; add %q to the outfit layer's variants in rig.yml\n", len(paths), *to, *to)
		return nil
	default:
		return fmt.Errorf("part: unknown subcommand %q", args[0])
	}
}

const defaultFlora = "res/sprites/flora"

func runFlora(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("flora: want build, check or preview")
	}
	fs := flag.NewFlagSet("flora "+args[0], flag.ExitOnError)
	out := fs.String("out", "", "preview output PNG")
	scale := fs.Int("scale", 8, "preview pixel scale")
	dry := fs.Bool("n", false, "migrate: report only, do not write")
	if args[0] == "migrate" {
		_ = fs.Parse(args[1:])
	} else {
		_ = fs.Parse(reorder(args[1:]))
	}
	dir := defaultFlora
	if fs.NArg() > 0 && args[0] != "migrate" {
		dir = fs.Arg(0)
	}
	f, err := flora.Load(dir)
	if err != nil {
		return err
	}
	switch args[0] {
	case "migrate":
		// flora migrate [-n] <map.tmj>... (flora dir is the default)
		if fs.NArg() == 0 {
			return fmt.Errorf("flora migrate: give one or more .tmj maps")
		}
		f, err = flora.Load(defaultFlora)
		if err != nil {
			return err
		}
		for _, m := range fs.Args() {
			res, err := f.Migrate(m, !*dry)
			if err != nil {
				return err
			}
			if res.Skipped != "" {
				fmt.Printf("skip  %s (%s)\n", m, res.Skipped)
				continue
			}
			verb := "moved"
			if *dry {
				verb = "would move"
			}
			fmt.Printf("%s %d flowers in %s -> %s\n", verb, res.Moved, m, strings.Join(res.Layers, ", "))
		}
		return nil
	case "preview":
		if *out == "" {
			return fmt.Errorf("flora preview: -out is required")
		}
		img, err := f.Preview(*scale)
		if err != nil {
			return err
		}
		return os.WriteFile(*out, rig.EncodePNG(img), 0o644)
	case "check", "build":
		outs, err := f.Build()
		if err != nil {
			return err
		}
		stale := rig.Stale(outs)
		if args[0] == "check" {
			if len(stale) > 0 {
				return fmt.Errorf("generated flora files are out of date; run `make sprites`:\n  %s", strings.Join(stale, "\n  "))
			}
			fmt.Printf("ok: %s (%d flora, %d outputs up to date)\n", f.Spec.Name, len(f.Spec.Flora), len(outs))
			return nil
		}
		for _, p := range stale {
			for _, o := range outs {
				if o.Path == p {
					if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
						return err
					}
					if err := os.WriteFile(p, o.Data, 0o644); err != nil {
						return err
					}
					fmt.Println("wrote", p)
				}
			}
		}
		fmt.Printf("built %s: %d flora (%d changed)\n", f.Spec.Name, len(f.Spec.Flora), len(stale))
		return nil
	default:
		return fmt.Errorf("flora: unknown subcommand %q", args[0])
	}
}

func runTiles(args []string) error {
	if len(args) > 0 && (args[0] == "animate" || args[0] == "animate-check") {
		dir := "res/sprites/water"
		if len(args) > 1 {
			dir = args[1]
		}
		a, err := tileanim.Load(dir)
		if err != nil {
			return err
		}
		outs, err := a.Build()
		if err != nil {
			return err
		}
		stale := rig.Stale(outs)
		if args[0] == "animate-check" {
			if len(stale) > 0 {
				return fmt.Errorf("generated tile animation is out of date; run `make sprites`:\n  %s", strings.Join(stale, "\n  "))
			}
			fmt.Printf("ok: %s (%d animated tiles, up to date)\n", dir, tileanim.AnimatedCount(outs))
			return nil
		}
		for _, o := range outs {
			for _, p := range stale {
				if p == o.Path {
					if err := os.WriteFile(o.Path, o.Data, 0o644); err != nil {
						return err
					}
					fmt.Println("wrote", o.Path)
				}
			}
		}
		fmt.Printf("animated %s: %d tiles (%d files changed)\n", dir, tileanim.AnimatedCount(outs), len(stale))
		return nil
	}
	if len(args) > 0 && (args[0] == "faces" || args[0] == "faces-check") {
		return runFaces(args[0] == "faces-check", args[1:])
	}
	if len(args) == 0 || args[0] != "remap" {
		return fmt.Errorf("tiles: want remap, animate, animate-check, faces or faces-check")
	}
	fs := flag.NewFlagSet("tiles remap", flag.ExitOnError)
	fromImage := fs.String("from-image", "", "embedded tile image basename to replace (e.g. water.png)")
	to := fs.String("to", "", "replacement as path/to/tileset.tsx#tileid")
	dry := fs.Bool("n", false, "report only, do not write")
	_ = fs.Parse(args[1:])
	tsx, idStr, ok := strings.Cut(*to, "#")
	id, err := strconv.Atoi(idStr)
	if *fromImage == "" || !ok || err != nil || fs.NArg() == 0 {
		return fmt.Errorf("tiles remap: need -from-image, -to tileset.tsx#id and maps")
	}
	abs, err := filepath.Abs(tsx)
	if err != nil {
		return err
	}
	for _, m := range fs.Args() {
		res, err := flora.RemapEmbeddedTile(m, flora.RemapOptions{FromImage: *fromImage, ToTsx: abs, ToTile: id}, !*dry)
		if err != nil {
			return err
		}
		if res.Skipped != "" {
			fmt.Printf("skip  %s (%s)\n", m, res.Skipped)
			continue
		}
		fmt.Printf("remapped %d tiles in %s (%s)\n", res.Moved, m, strings.Join(res.Layers, ", "))
	}
	return nil
}

// runFaces builds (or checks) the taller cliff faces: res/sprites/faces.
func runFaces(check bool, args []string) error {
	dir := "res/sprites/faces"
	if len(args) > 0 {
		dir = args[0]
	}
	f, err := faces.Load(dir)
	if err != nil {
		return err
	}
	res, err := f.Build()
	if err != nil {
		return err
	}
	placed := 0
	for _, n := range res.Placed {
		placed += n
	}
	summary := fmt.Sprintf("%s (%d face tiles, %d placed on %d maps", dir, len(res.Tiles), placed, len(res.Placed))
	stale := rig.Stale(res.Outputs)
	if check {
		if len(stale) > 0 {
			return fmt.Errorf("generated cliff faces are out of date; run `make sprites`:\n  %s", strings.Join(stale, "\n  "))
		}
		fmt.Printf("ok: %s, up to date)\n", summary)
		return nil
	}
	for _, p := range stale {
		for _, o := range res.Outputs {
			if o.Path == p {
				if err := os.WriteFile(o.Path, o.Data, 0o644); err != nil {
					return err
				}
				fmt.Println("wrote", o.Path)
			}
		}
	}
	fmt.Printf("faces: %s, %d files changed)\n", summary, len(stale))
	return nil
}

// runImportPixelLab splits a palette-locked PixelLab character into rig
// layers: spritegen import-pixellab -src res/sprites/adult/source/base -poses res/sprites/adult
func runImportPixelLab(args []string) error {
	fs := flag.NewFlagSet("import-pixellab", flag.ExitOnError)
	src := fs.String("src", "", "directory of <anim>_<dir>_<i>.png frames")
	poses := fs.Bool("poses", false, "also write poses.yml and clips.yml (the base character)")
	layers := fs.String("layers", "", "comma-separated layers to write (default pants,body,shirt,hair)")
	variants := fs.String("variant", "", "layer=variant pairs the parts go to, e.g. hair=long (default: default)")
	force := fs.Bool("force", false, "overwrite existing parts")
	noStab := fs.Bool("no-stabilize", false, "keep PixelLab's frame positions (default: line each animation up on its first frame's head)")
	_ = fs.Parse(args)
	r, err := rig.Load(rigDir(fs))
	if err != nil {
		return err
	}
	o := rig.PixelLabImportOptions{Source: *src, Poses: *poses, Force: *force, NoStabilize: *noStab, Variant: map[string]string{}}
	if *layers != "" {
		o.Layers = strings.Split(*layers, ",")
	}
	for _, kv := range strings.Split(*variants, ",") {
		if k, v, ok := strings.Cut(kv, "="); ok {
			o.Variant[k] = v
		}
	}
	paths, err := r.ImportPixelLab(o)
	if err != nil {
		return err
	}
	fmt.Printf("wrote %d parts\n", len(paths))
	return nil
}

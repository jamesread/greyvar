// spritegen builds layered sprite rigs (see pkg/rig and res/sprites/README.md).
//
//	spritegen check   [rig-dir]   validate the rig and fail if outputs are stale
//	spritegen build   [rig-dir]   regenerate sheets, rig.json, composite and .tsx
//	spritegen preview [rig-dir] -out file.png [-clip idle] [-set hair=blonde,style=long] [-scale 8]
//	spritegen part show head/default/idle_0 [rig-dir]
//	spritegen part derive -layer head -from idle_0 -to blink -edit 6,4=skin.0 [-edit ...] [rig-dir]
//	spritegen part prune [-n]   (delete parts no pose references)
//	spritegen part longhair -from short -to long [-rows 3]   (derive a whole hair variant)
//	spritegen import  -sheet ss.png -tsx player.tsx -out res/sprites/player   (one-time migration)
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jamesread/greyvar/sprite-generator/pkg/rig"
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
	case "import":
		err = runImport(args)
	default:
		usage()
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "spritegen:", err)
		os.Exit(1)
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, "usage: spritegen check|build|preview|part|import [flags] [rig-dir]")
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
	default:
		return fmt.Errorf("part: unknown subcommand %q", args[0])
	}
}

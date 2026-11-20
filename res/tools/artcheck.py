#!/usr/bin/env python3
"""Check art against res/ART_RULES.md (the measurable rules).

  python3 res/tools/artcheck.py <sprite.png> [--ground NAME] [--frame W H I]
  python3 res/tools/artcheck.py --interiors      # every furniture-small piece on the house floor

Checks: canvas of 16 px per cell (or 15/31 for entities centred on a middle
pixel) and an allowed footprint (1x1, 1x2, 2x1, 2x2),
binary alpha, outline contrast and body contrast against the ground it
stands on (failures), and colour count (a warning). Perspective can't be
measured: review it by eye. Exit status 1 when anything fails.
"""
import argparse, json, math, os, sys
from PIL import Image

ROOT = os.path.normpath(os.path.join(os.path.dirname(os.path.abspath(__file__)), '..', '..'))
TS = os.path.join(ROOT, 'res', 'img', 'textures', 'tilesets')

# Ground tiles, by name: (tileset png, columns, tile id). house_floor is
# looked up from the interiors builder (first plank floor).
GROUNDS = {
    'grass': ('grass_terrains.png', 20, 67), 'sand': ('grass_terrains.png', 20, 64),
    'hill': ('grass_terrains.png', 20, 70), 'castle': ('grass_terrains.png', 20, 73),
    'mud': ('grass_terrains.png', 20, 61), 'house_floor': ('house_interior.png', 16, 1),
}

# Thresholds (ART_RULES.md, "Contrast"); calibrated on the good/bad examples.
BODY_FAR_DE = 20      # a body pixel "clearly differs" from the ground above this ΔE
BODY_FAR_SHARE = 0.70 # share of body pixels that must clearly differ
BODY_MEAN_DE = 30     # mean ΔE of the body against the ground
OUTLINE_DE = 35       # mean ΔE of the outline ring against the ground
MAX_COLOURS = 16
FOOTPRINTS = {(1, 1), (1, 2), (2, 1), (2, 2)}


def lin(c):
    c /= 255
    return c / 12.92 if c <= 0.04045 else ((c + 0.055) / 1.055) ** 2.4


def lab(p):
    r, g, b = (lin(v) for v in p[:3])
    x = (0.4124 * r + 0.3576 * g + 0.1805 * b) / 0.95047
    y = 0.2126 * r + 0.7152 * g + 0.0722 * b
    z = (0.0193 * r + 0.1192 * g + 0.9505 * b) / 1.08883
    f = lambda t: t ** (1 / 3) if t > 0.008856 else 7.787 * t + 16 / 116
    return (116 * f(y) - 16, 500 * (f(x) - f(y)), 200 * (f(y) - f(z)))


def de(a, b):
    return math.dist(lab(a), lab(b))


def ground_colour(name):
    png, cols, tid = GROUNDS[name]
    im = Image.open(os.path.join(TS, png)).convert('RGBA')
    t = im.crop((tid % cols * 16, tid // cols * 16, tid % cols * 16 + 16, tid // cols * 16 + 16))
    px = [p for p in t.getdata() if p[3] == 255]
    return tuple(sum(p[i] for p in px) / len(px) for i in range(3))


def check(im, ground=None, canvas=True):
    """Returns a list of (rule, ok, detail)."""
    im = im.convert('RGBA')
    W, H = im.size
    out = []
    if canvas:
        # 16 px per cell, or 15 (31) for entities centred on a middle pixel.
        cells = lambda n: n // 16 if n % 16 == 0 else (n + 1) // 16 if n in (15, 31) else None
        ok = cells(W) is not None and cells(H) is not None and (cells(W), cells(H)) in FOOTPRINTS
        out.append(('canvas', ok, f'{W}x{H}px (16 or 15 px per cell)'))
    data = list(im.getdata())
    partial = sum(1 for p in data if 0 < p[3] < 255)
    out.append(('binary alpha', partial == 0, f'{partial} partly transparent px'))
    colours = len({p[:3] for p in data if p[3]})
    # Guidance, not a failure: some good assets (shop signs) use more.
    out.append(('colours', True if colours <= MAX_COLOURS else 'warn', f'{colours} (aim for {MAX_COLOURS} or fewer)'))
    if ground:
        g = ground_colour(ground)
        P = im.load()
        op = {(x, y) for y in range(H) for x in range(W) if P[x, y][3] == 255}
        ring = {(x, y) for x, y in op if any((x + dx, y + dy) not in op for dx, dy in ((1, 0), (-1, 0), (0, 1), (0, -1)))}
        body = (op - ring) or ring
        o = sum(de(P[q], g) for q in ring) / max(len(ring), 1)
        b = [de(P[q], g) for q in body]
        far = sum(v > BODY_FAR_DE for v in b) / len(b)
        mean = sum(b) / len(b)
        out.append(('outline contrast', o >= OUTLINE_DE, f'ΔE {o:.0f} vs {ground} (min {OUTLINE_DE})'))
        out.append(('body contrast', far >= BODY_FAR_SHARE and mean >= BODY_MEAN_DE,
                    f'{far:.0%} of body clearly differs (min {BODY_FAR_SHARE:.0%}), mean ΔE {mean:.0f} (min {BODY_MEAN_DE})'))
    return out


def report(name, results):
    bad = [r for r in results if r[1] is False]
    warn = [r for r in results if r[1] == 'warn']
    print(f"{'FAIL' if bad else 'warn' if warn else 'ok  '} {name}")
    for rule, ok, detail in results:
        if ok is not True or '-v' in sys.argv:
            mark = 'x' if ok is False else '!' if ok == 'warn' else ' '
            print(f"     {mark} {rule}: {detail}")
    return not bad


def main():
    ap = argparse.ArgumentParser(description=__doc__, formatter_class=argparse.RawDescriptionHelpFormatter)
    ap.add_argument('png', nargs='?')
    ap.add_argument('--ground', choices=sorted(GROUNDS))
    ap.add_argument('--frame', nargs=3, type=int, metavar=('W', 'H', 'I'), help='check one frame of a sheet')
    ap.add_argument('--interiors', action='store_true', help='every res/sprites/interiors/furniture-small piece on house_floor')
    ap.add_argument('-v', action='store_true', help='show passing rules too')
    a = ap.parse_args()
    ok = True
    if a.interiors:
        d = os.path.join(ROOT, 'res', 'sprites', 'interiors', 'furniture-small')
        for f in sorted(os.listdir(d)):
            ground = None if f == 'clock.png' else 'house_floor'   # wall-hung
            ok &= report(f, check(Image.open(os.path.join(d, f)), ground))
    elif a.png:
        im = Image.open(a.png)
        canvas = True
        if a.frame:
            w, h, i = a.frame
            cols = im.width // w
            im = im.crop((i % cols * w, i // cols * h, i % cols * w + w, i // cols * h + h))
        ok = report(a.png, check(im, a.ground, canvas))
    else:
        ap.print_help(); return 2
    return 0 if ok else 1


if __name__ == '__main__':
    sys.exit(main())

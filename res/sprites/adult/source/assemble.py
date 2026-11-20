#!/usr/bin/env python3
"""Assemble the adult rig's parts from the PixelLab sources in this folder.

Every folder here holds one PixelLab character state, palette-locked to the
rig's key colours (see download.cfg in each for where the frames came from):

    base_v2/   the base character: standing pose (idle_<dir>_0) and the
               skeleton walk (walk_<dir>_<i>) in four directions
    bald/      the base, bald            -> the head of the body layer
    long/ ponytail/ bun/ curly/ bob/     -> hair/<style>
    beard/     the base with a beard     -> beard/beard
    dress/     the base in a dress, with its own skeleton walk and breathing
               -> outfit/dress

Hands: the idle has them hanging naturally at the sides; a second stance,
hands in pockets (pockets_<dir>_0..3, clips pockets*, state `pockets`),
tucks them over the hips (not in a dress). See tuck_hands.

The breathing idle (breathe_<dir>_0..3) is made here from the standing pose:
two frames as drawn, then two with the head and shoulders 1 px lower (the
out-breath), on every layer. (PixelLab's skeleton breathing-idle barely moves
at this size and its side views do not match the standing pose.)
The form layer (mannequin) is derived: the finished body in the form colours,
eyes filled in.

States other than base_v2 and dress have only the standing pose: hair and
heads do not change shape while walking, so

  * every frame's head (face and hair) is the standing pose's head for that
    direction, moved with the frame (PixelLab repaints the head's shading on
    every walk frame and it flickers);
  * the body layer's head is the bald state's, so any hairstyle (or none)
    has a whole head under it;
  * each hairstyle is lifted from its state, lined up with the base on the
    unchanged clothes.

Run from the repository root after changing the sources:

    python3 res/sprites/adult/source/assemble.py
    make sprites

It runs `spritegen import-pixellab` for the colour split (base_v2 straight
into the rig, which also writes poses.yml and clips.yml; the others into a
scratch copy), then rewrites the parts it owns.
"""
import os
import shutil
import subprocess
import sys
import tempfile

import yaml
from PIL import Image

HERE = os.path.dirname(os.path.abspath(__file__))
RIG = os.path.dirname(HERE)
ROOT = os.path.abspath(os.path.join(RIG, '..', '..', '..'))
PARTS = os.path.join(RIG, 'parts')
with open(os.path.join(RIG, 'rig.yml')) as _fh:
    _SPEC = yaml.safe_load(_fh)
FS = _SPEC['frameSize']
ANCHOR_Y = _SPEC.get('anchor', [FS // 2, FS])[1]
DIRS = ['south', 'east', 'north', 'west']
STYLES = ['long', 'ponytail', 'bun', 'curly', 'bob']
LAYERS = ['pants', 'body', 'shirt', 'outfit', 'hair']


def spritegen(*args):
    subprocess.run(['go', 'run', './sprite-generator', *args], cwd=ROOT, check=True,
                   stdout=subprocess.DEVNULL)


def split(name, scratch):
    """Colour-split one source state into layers; returns {layer: {frame: px}}."""
    rig = os.path.join(scratch, name)
    os.makedirs(rig, exist_ok=True)
    shutil.copy(os.path.join(RIG, 'rig.yml'), rig)
    for f in ('poses.yml', 'clips.yml'):
        with open(os.path.join(rig, f), 'w') as fh:
            fh.write('poses: []\n' if f == 'poses.yml' else 'clips: {}\n')
    spritegen('import-pixellab', '-src', os.path.join(HERE, name), '-force',
              '-layers', ','.join(LAYERS), rig)
    out = {}
    for layer in LAYERS:
        d = os.path.join(rig, 'parts', layer, 'default')
        out[layer] = {f[:-4]: load(os.path.join(d, f)) for f in os.listdir(d)} if os.path.isdir(d) else {}
    return out


def load(path):
    im = Image.open(path).convert('RGBA')
    return {(x, y): im.getpixel((x, y))[:3] for y in range(im.height) for x in range(im.width)
            if im.getpixel((x, y))[3]}


# Centring on the frame's centre column (CX): PixelLab draws the front and
# back views 16 px wide, symmetric about the line between two columns. The
# middle column is doubled so they are 17 px wide and symmetric about CX; the
# side views (13 px, odd already) are moved onto CX. Set per direction from
# the base's standing pose (see centre_shifts) and applied to every part.
CX = _SPEC.get('anchor', [FS // 2, FS])[0]
WIDEN = ('south', 'north')
SHIFT = {}


def direction(frame):
    return frame.split('_')[1]


BREATHE_FRAMES = 4
CHEST = {}  # direction -> last row that sinks on the out-breath


def breathe(px, frame):
    """The out-breath (breathe frames 2 and 3): rows down to the chest move
    1 px down, over the row below; the rest stays."""
    if not frame.startswith(('breathe_', 'pockets_')) or int(frame.rsplit('_', 1)[1]) < 2:
        return px
    chest = CHEST[direction(frame)]
    out = {p: c for p, c in px.items() if p[1] > chest + 1}
    out.update({(x, y + 1): c for (x, y), c in px.items() if y <= chest})
    out.update({p: c for p, c in px.items() if p[1] == chest + 1 and (p[0], p[1]) not in out})
    return out


def centre(px, frame):
    d = direction(frame)
    if d in WIDEN:
        out = {(x + 1 if x >= CX else x, y): c for (x, y), c in px.items()}
        out.update({(CX, y): c for (x, y), c in px.items() if x == CX - 1})
        return out
    return shift(px, SHIFT.get(d, 0), 0)


def centre_shifts(base):
    """How far each side view moves so its standing figure is centred, and
    each direction's chest row (two rows below the collar: the shoulders)."""
    for d in DIRS:
        CHEST[d] = min(y for (x, y) in base['shirt'][f'idle_{d}_0']) + 1
    for d in ('east', 'west'):
        f = f'idle_{d}_0'
        xs = [x for layer in ('body', 'shirt', 'pants', 'hair') for (x, y) in base[layer][f]]
        SHIFT[d] = CX - (min(xs) + max(xs)) // 2


def save(px, layer, variant, frame):
    px = centre(breathe(px, frame), frame)
    d = os.path.join(PARTS, layer, variant)
    os.makedirs(d, exist_ok=True)
    im = Image.new('RGBA', (FS, FS))
    for (x, y), c in px.items():
        if 0 <= x < FS and 0 <= y < FS:
            im.putpixel((x, y), c + (255,))
    im.save(os.path.join(d, frame + '.png'))


def shift(px, dx, dy):
    return {(x + dx, y + dy): c for (x, y), c in px.items()}


def best_offset(ref, cur, r=3):
    """The (dx, dy) that moves `ref` onto `cur` with the fewest differences."""
    best = None
    for dy in range(-r, r + 1):
        for dx in range(-r, r + 1):
            diff = len(set(shift(ref, dx, dy).items()) ^ set(cur.items()))
            if best is None or diff < best[0]:
                best = (diff, dx, dy)
    return best[1], best[2]


# PixelLab draws the shoes in the darkest hair colours, so the colour split
# files them under hair. Below this row hair-coloured pixels are shoes: they
# go to the pants layer in the fixed dark colours (so they neither follow the
# head nor turn blonde with the hair).
SHOE_ROW = ANCHOR_Y - 5  # the soles are the row above the anchor
HAIR_KEYS = [(0x3e, 0x28, 0x2b), (0x4e, 0x2c, 0x29), (0x61, 0x3a, 0x41)]
SHOE = {HAIR_KEYS[0]: (0x15, 0x12, 0x15), HAIR_KEYS[1]: (0x1f, 0x1a, 0x1f), HAIR_KEYS[2]: (0x1f, 0x1a, 0x1f)}


def shoes(px):
    return {p: SHOE.get(c, c) for p, c in px.items() if p[1] >= SHOE_ROW}


def above_shoes(px):
    return {p: c for p, c in px.items() if p[1] < SHOE_ROW}


def no_hair_colours(px):
    """Pixels with any shoe-coloured hair keys turned to the shoe colours."""
    return {p: (SHOE.get(c, c) if p[1] >= SHOE_ROW else c) for p, c in px.items()}


def cover(outfit, under):
    """Grow `outfit` over every pixel of `under` it leaves showing (the dress
    and the base differ by a pixel at the collar and between the feet), each
    filled pixel taking the commonest colour of its outfit neighbours."""
    out = dict(outfit)
    todo = [p for p in under if p not in out]
    while todo:
        left = []
        for (x, y) in todo:
            near = [out[(x + dx, y + dy)] for dy in (-1, 0, 1) for dx in (-1, 0, 1)
                    if (dx or dy) and (x + dx, y + dy) in out]
            if near:
                out[(x, y)] = max(set(near), key=near.count)
            else:
                left.append((x, y))
        if len(left) == len(todo):
            break
        todo = left
    return out


OUTLINE = (0x15, 0x12, 0x15)
EYE = (0x1f, 0x1a, 0x1f)


def hexrgb(h):
    h = h.lstrip('#')
    return tuple(int(h[i:i + 2], 16) for i in (0, 2, 4))


SKIN_KEYS = [hexrgb(h) for h in _SPEC['slots']['skin']]
FORM_KEYS = [hexrgb(h) for h in _SPEC['slots']['form']]


def mannequin(body):
    """The body's silhouette as a faceless mannequin: skin keys -> form keys
    (same ramp step), eyes -> the form's base colour, outline kept."""
    out = {}
    for p, c in body.items():
        if c in SKIN_KEYS:
            out[p] = FORM_KEYS[SKIN_KEYS.index(c)]
        elif c == EYE:
            out[p] = FORM_KEYS[2]
        else:
            out[p] = c
    return out


def close_stance(px):
    """The front standing pose, legs as the back one has them: the gap
    between the legs dark (outline) down to the shoes instead of see-through,
    and each leg full width to the ankle (PixelLab tapers the inner edge
    to one pixel there), so they read as close together."""
    out = dict(px)
    ys = sorted({y for (x, y) in px})
    bottom = ys[-1]
    for y in ys:
        row = sorted(x for (x, yy) in px if yy == y)
        for x in range(row[0] + 1, row[-1]):
            if (x, y) not in out and y < bottom:
                out[(x, y)] = OUTLINE  # close the gap
    ankle = bottom - 2  # the row above the shoes
    for x in range(min(x for (x, y) in px), max(x for (x, y) in px) + 1):
        up = out.get((x, ankle - 1))
        if out.get((x, ankle)) == OUTLINE and up not in (None, OUTLINE):
            out[(x, ankle)] = up  # keep the leg full width
    return out


TUCK = 1  # px the hands move in towards the hips (hands in pockets)
FRONT_BACK_STANDING = tuple(f'{a}_{d}_' for a in ('idle', 'breathe', 'pockets') for d in ('south', 'north'))
# Softer hands: round each hand's bottom outer corner and outline the hand's
# bottom in the darkest skin shade instead of black (it follows skin tone).
SOFT_EDGE = True


def lengthen_sleeves(st, frame):
    """Sleeves 1 px longer: for each arm (skin below the collar, under a
    sleeve), the hem moves one row down across the arm's whole width. The new
    pixels take the sleeve colour above them (on the shirt layer; the body
    keeps its pixel). Hands lower down are not touched."""
    shirt, body = st['shirt'][frame], st['body'][frame]
    cut = min(y for (x, y) in shirt)
    band = {p for p, c in body.items() if c in SKIN_KEYS and cut + 1 < p[1] <= cut + 6 and p not in shirt}
    seen = set()
    for start in sorted(band):
        if start in seen:
            continue
        arm, todo = set(), [start]
        while todo:  # one arm: 4-connected skin in the band
            p = todo.pop()
            if p in arm or p not in band:
                continue
            arm.add(p)
            x, y = p
            todo += [(x + 1, y), (x - 1, y), (x, y + 1), (x, y - 1)]
        seen |= arm
        # the sleeve's lowest row over this arm: shirt right above its top
        bottoms = []
        for x in {x for (x, y) in arm}:
            top = min(y for (xx, y) in arm if xx == x)
            if shirt.get((x, top - 1)) not in (None, OUTLINE):
                bottoms.append(top - 1)
        if not bottoms:
            continue  # not under a sleeve (a hand across the body)
        hem = max(bottoms) + 1
        for (x, y) in sorted(arm, key=lambda p: p[1]):
            if y > hem:
                continue
            # colour: the sleeve pixel above, else the nearest sleeve pixel in the row above
            above = shirt.get((x, y - 1))
            if above in (None, OUTLINE):
                near = [shirt[(x + d, y - 1)] for d in (1, -1, 2, -2) if shirt.get((x + d, y - 1)) not in (None, OUTLINE)]
                near += [shirt[(x + d, y)] for d in (1, -1) if shirt.get((x + d, y)) not in (None, OUTLINE)]
                if not near:
                    continue
                above = near[0]
            shirt[(x, y)] = above


def tuck_hands(st, frame, layers=('pants', 'body', 'shirt'), cut=None, pockets=False):
    """Front and back standing: PixelLab's hands flare 1 px outwards below
    the arm. Two stances fix that:

    * natural (default): the hand narrows to the arm's width and hangs
      straight at the side, its fingertips rounded and edged in the darkest
      skin shade;
    * pockets: the hand rows move 1 px in, over the hip (hands in pockets),
      with the same soft edge.

    Works on `layers` of state `st` together (st[layer][frame] = pixels, in
    draw order); a moved pixel replaces whatever was there. `cut` is the
    collar row (default: the shirt's top)."""
    top = {}  # position -> layer drawn there (draw order: last wins)
    for l in layers:
        for p in st[l][frame]:
            top[p] = l
    if cut is None:
        cut = min(y for (x, y) in st['shirt'][frame])
    rows = sorted({y for (x, y) in top if y > cut + 3})
    for side in (-1, 1):  # -1: the figure's left edge, 1: its right edge
        edge = {y: (min if side < 0 else max)(x for (x, yy) in top if yy == y) for y in rows}
        arm = [y for y in rows if top.get((edge[y], y)) is not None]
        outer = (min if side < 0 else max)(edge[y] for y in arm)
        hand = [y for y in arm if edge[y] == outer]
        if not hand or all(edge[y] == outer for y in arm[:len(arm) // 2]):
            continue  # no flare
        def segment(y):
            """The arm at row y: outer edge to the first outline inside it,
            as [(x, layer, colour)], and its skin width."""
            seg, x, skin = [], edge.get(y, outer), 0
            while (x, y) in top:
                c = st[top[(x, y)]][frame][(x, y)]
                seg.append((x, top[(x, y)], c))
                if c in SKIN_KEYS:
                    skin += 1
                elif skin and c == OUTLINE:
                    break
                x -= side
            return seg, skin

        if not pockets:
            # Natural: line the hand up under the forearm. A row wider than
            # the forearm loses its outer column (its outer skin pixel
            # becomes the outline); a row as wide but sticking out (the back
            # view) moves 1 px in.
            # the forearm's width: the narrowest bare-arm row here (the
            # sleeve may cover the row above the hand)
            ref = min(w for w in (segment(y)[1] for y in [hand[0] - 1] + hand) if w > 0)
            for y in hand:
                seg, skin = segment(y)
                if skin > ref:
                    for l in layers:
                        st[l][frame].pop((outer, y), None)
                    owner = top.get((outer - side, y), layers[0])
                    st[owner][frame][(outer - side, y)] = OUTLINE
                else:
                    for x, l, c in seg:
                        del st[l][frame][(x, y)]
                    for x, l, c in seg:
                        for other in layers:
                            st[other][frame].pop((x - side, y), None)
                        st[l][frame][(x - side, y)] = c
            last = hand[-1]
            ox = outer - side  # the hand's outer outline now
            # round the fingertips: no outline under the outer corner, the
            # rest of the edge under the hand in the darkest skin shade
            for l in layers:
                st[l][frame].pop((ox, last + 1), None)
                st[l][frame].pop((outer, last + 1), None)
            x = ox - side
            while True:
                owner = next((l for l in reversed(layers) if (x, last) in st[l][frame]), None)
                if owner is None or st[owner][frame][(x, last)] not in SKIN_KEYS:
                    break
                below = next((l for l in reversed(layers) if (x, last + 1) in st[l][frame]), None)
                if below is not None and st[below][frame][(x, last + 1)] == OUTLINE:
                    st[below][frame][(x, last + 1)] = SKIN_KEYS[0]
                elif below is None:
                    st[owner][frame][(x, last + 1)] = SKIN_KEYS[0]
                x -= side
            top = {}
            for l in layers:
                for p in st[l][frame]:
                    top[p] = l
            continue
        for y in hand:
            # the arm segment: outer edge to the first outline inside it
            seg, x, seen_skin = [], edge[y], False
            while (x, y) in top:
                c = st[top[(x, y)]][frame][(x, y)]
                seg.append((x, top[(x, y)], c))
                if c in SKIN_KEYS:
                    seen_skin = True
                elif seen_skin and c == OUTLINE:
                    break
                x -= side
            for x, l, c in seg:
                del st[l][frame][(x, y)]
            for x, l, c in seg:
                for other in layers:
                    st[other][frame].pop((x - side * TUCK, y), None)
                st[l][frame][(x - side * TUCK, y)] = c
        # close the outline under the moved hand
        last = hand[-1]
        soften = []
        for x in range(outer - side, outer + (TUCK + 6) * -side, -side):
            owner = top.get((x, last))
            if owner and st[owner][frame].get((x, last)) in SKIN_KEYS and not any((x, last + 1) in st[l][frame] for l in layers):
                st[owner][frame][(x, last + 1)] = OUTLINE
        # Soften: the hand's outer bottom corner goes (outline moves in one),
        # and its bottom edge is drawn in the darkest skin shade.
        ox = outer - side * TUCK  # the hand's outer outline, after the tuck
        cx = ox - side            # the skin pixel at its outer bottom corner
        for l in layers:
            st[l][frame].pop((ox, last), None)
            st[l][frame].pop((ox, last + 1), None)
        corner_owner = top.get((cx, last), layers[0])
        st[corner_owner][frame][(cx, last)] = OUTLINE
        if SOFT_EDGE:
            x = cx - side
            while True:
                owner = next((l for l in reversed(layers) if (x, last) in st[l][frame]), None)
                if owner is None or st[owner][frame][(x, last)] not in SKIN_KEYS:
                    break
                below = next((l for l in reversed(layers) if (x, last + 1) in st[l][frame]), None)
                if below is not None and st[below][frame][(x, last + 1)] == OUTLINE:
                    st[below][frame][(x, last + 1)] = SKIN_KEYS[0]
                x -= side
        top = {}
        for l in layers:
            for p in st[l][frame]:
                top[p] = l


def merge(*layers):
    out = {}
    for px in layers:
        out.update(px)
    return out


def main():
    scratch = tempfile.mkdtemp(prefix='adult-assemble-')
    try:
        # The base goes straight into the rig: parts for every frame plus
        # poses.yml / clips.yml (which list every layer of rig.yml).
        for layer in os.listdir(PARTS) if os.path.isdir(PARTS) else []:
            shutil.rmtree(os.path.join(PARTS, layer))
        # The base's frames plus the breathing idle: copies of the standing
        # pose (breathe() moves the out-breath frames when they are saved).
        src = os.path.join(scratch, 'base_src')
        shutil.copytree(os.path.join(HERE, 'base_v2'), src)
        for d in DIRS:
            for i in range(BREATHE_FRAMES):
                for anim in ('breathe', 'pockets'):
                    shutil.copy(os.path.join(src, f'idle_{d}_0.png'), os.path.join(src, f'{anim}_{d}_{i}.png'))
        spritegen('import-pixellab', '-src', src, '-poses', '-force', '-variant', 'hair=short', RIG)
        base = {l: {} for l in LAYERS}
        for layer, variant in (('pants', 'default'), ('body', 'default'), ('shirt', 'default'), ('hair', 'short')):
            d = os.path.join(PARTS, layer, variant)
            base[layer] = {f[:-4]: load(os.path.join(d, f)) for f in os.listdir(d)}
        frames = sorted(base['body'])
        states = {n: split(n, scratch) for n in ['bald', 'beard', 'dress', *STYLES]}
        centre_shifts(base)
        for st in [base, *states.values()]:
            for f in list(st['hair']):
                st['pants'][f] = merge(st['pants'].get(f, {}), shoes(st['hair'][f]))
                st['hair'][f] = above_shoes(st['hair'][f])
        for f in base['pants']:
            if f == 'idle_south_0' or f.startswith(('breathe_south_', 'pockets_south_')):
                base['pants'][f] = close_stance(base['pants'][f])
            lengthen_sleeves(base, f)
            if f.startswith(FRONT_BACK_STANDING):
                tuck_hands(base, f, pockets=f.startswith('pockets_'))
    finally:
        shutil.rmtree(scratch, ignore_errors=True)

    def stand(d):
        return f'idle_{d}_0'

    def clothes(st, frame, cut):
        px = merge(st['pants'].get(frame, {}), st['shirt'].get(frame, {}))
        return {p: c for p, c in px.items() if p[1] >= cut}

    for d in DIRS:
        s = stand(d)
        ys = [y for (x, y) in base['shirt'][s]]
        cut = min(ys)  # the collar: rows above it are the head and neck
        head = lambda px: {p: c for p, c in px.items() if p[1] < cut}
        base_head = merge(head(base['body'][s]), base['hair'][s])

        # Where each state's standing pose sits against the base's.
        align = {n: best_offset(clothes(st, s, cut), clothes(base, s, cut))
                 for n, st in states.items() if n != 'dress'}
        bald_head = shift(head(states['bald']['body'][s]), *align['bald'])
        # The bald state's head can end a row above the base's collar: keep
        # the base's neck below it so head and shirt join.
        neck_top = max(y for (x, y) in bald_head) + 1
        bald_head.update({p: c for p, c in base['body'][s].items() if neck_top <= p[1] < cut})
        # From behind the base's nape is hair: carry the neck down instead.
        for y in range(neck_top, cut):
            if not any(py == y for (px, py) in bald_head):
                bald_head.update({(x, y): c for (x, py), c in list(bald_head.items()) if py == y - 1})
        hair = {'short': base['hair'][s]}
        for style in STYLES:
            hair[style] = shift(states[style]['hair'][s], *align[style])
        hair['bald'] = {}

        # Beard: hair-coloured pixels on the lower face of the beard state
        # that are not the base's (short) hair. None from behind.
        beard = {}
        if d != 'north':
            eyes = [y for (x, y), c in base['body'][s].items() if c == (0x1f, 0x1a, 0x1f)]
            low = (max(eyes) + 1) if eyes else cut - 4
            bh = shift(states['beard']['hair'][s], *align['beard'])
            beard = {p: c for p, c in bh.items() if p[1] >= low and p not in base['hair'][s]}

        for f in [fr for fr in frames if f'_{d}_' in fr]:
            # The frame's own head moved against the standing pose's.
            off = (0, 0) if f == s else best_offset(base_head, merge(head(base['body'][f]), base['hair'][f]))
            body = {p: c for p, c in base['body'][f].items() if p[1] >= cut + off[1]}
            body.update(shift(bald_head, *off))
            save(body, 'body', 'default', f)
            save(mannequin(body), 'form', 'mannequin', f)
            save({}, 'form', 'none', f)
            save(base['pants'][f], 'pants', 'default', f)
            save(base['shirt'][f], 'shirt', 'default', f)
            for style, px in hair.items():
                save(shift(px, *off), 'hair', style, f)
            save({}, 'beard', 'none', f)
            save(shift(beard, *off), 'beard', 'beard', f)
            # Dress: the dress state's whole figure below the collar, from
            # its own walk (the skirt moves with the legs).
            dress = states['dress']
            src = f if f in dress['outfit'] else s
            below = merge(*(dress[l].get(src, {}) for l in ('pants', 'body', 'shirt', 'outfit')))
            dress_px = no_hair_colours({p: c for p, c in below.items() if p[1] >= cut + off[1]})
            if f.startswith(FRONT_BACK_STANDING):
                # a dress has no pockets: natural hands in both stances
                one = {'all': {f: dress_px}}
                tuck_hands(one, f, layers=('all',), cut=cut + off[1])
                dress_px = one['all'][f]
            dress_px = cover(dress_px, merge(base['shirt'][f], base['pants'][f]))
            save(dress_px, 'outfit', 'dress', f)
            save({}, 'outfit', 'none', f)
    print(f'assembled {len(frames)} frames: hair {sorted(hair)}, beard, dress')


if __name__ == '__main__':
    sys.exit(main())

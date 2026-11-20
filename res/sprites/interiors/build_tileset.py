"""House interior tileset builder (res/img/textures/tilesets/house_interior.*).

Room pieces are drawn here (16px, flat top-down with a 2-tile back wall);
furniture is the PixelLab set (furn/<i>.png, 32x32 each) packed onto the
16px grid. Writes the PNG, the TSX and a JSON index used by hmaps.py.
"""
import json, os, random, sys
HERE = os.path.dirname(os.path.abspath(__file__))
from PIL import Image

COLS = 16
T = 16
OUT = sys.argv[1] if len(sys.argv) > 1 else '.'

def rgba(h, a=255):
    h = h.lstrip('#'); return (int(h[0:2], 16), int(h[2:4], 16), int(h[4:6], 16), a)

def sh(c, f):
    return tuple(max(0, min(255, int(v * f))) for v in c[:3]) + (255,)

VOID = rgba('#141018')
BEAM = rgba('#5a3a1c'); BEAM_HI = rgba('#8a5c2e'); BEAM_LO = rgba('#3a2410'); BEAM_D = rgba('#24160a')
PLANK = [rgba('#b07a42'), rgba('#a8723c'), rgba('#b8844c'), rgba('#a06a36')]
GAP = rgba('#6e4622')
WAINSCOT = rgba('#6a4424'); WAIN_HI = rgba('#8c5e34'); WAIN_LO = rgba('#4a2e16')
SKIRT = rgba('#3a2410')
RAIL = rgba('#9a6a3a')

WALLPAPERS = {  # base, pattern, top shade
    'cream': (rgba('#efe4c4'), rgba('#dccfa8')),
    'sage': (rgba('#cfdcb4'), rgba('#b4c896')),
    'rose': (rgba('#f0cccc'), rgba('#dcaeb4')),
    'sky': (rgba('#cce0f0'), rgba('#aac8e0')),
    'bank': (rgba('#3c6a44'), rgba('#2e5636')),   # deep green with gold dots
}
GOLD = rgba('#d8b44a')

tiles = []   # (name, image, props)
def add(name, im, collision=None, props=None):
    tiles.append((name, im, collision, props or {}))
    return len(tiles) - 1

def blank(c=None):
    return Image.new('RGBA', (T, T), c or (0, 0, 0, 0))

# ---- floor ----
def floor(seed):
    rng = random.Random(seed)
    im = blank()
    for row in range(4):                       # 4 planks per tile, 4px each
        y0 = row * 4
        off = (row * 5 + (seed % 3) * 3) % 16 if row % 2 else (row * 7) % 16
        joints = {off % 16, (off + 8) % 16} if rng.random() < .5 else {off % 16}
        tone = PLANK[(row + seed) % 4]
        for y in range(y0, y0 + 4):
            for x in range(16):
                if y == y0 + 3:
                    c = GAP
                elif x in joints:
                    c = GAP
                else:
                    c = tone
                    if y == y0:
                        c = sh(tone, 1.08)
                    elif y == y0 + 2 and rng.random() < .25:
                        c = sh(tone, .94)
                im.putpixel((x, y), c)
        if rng.random() < .25:                  # a knot
            kx = rng.randint(2, 13)
            im.putpixel((kx, y0 + 1), sh(tone, .75)); im.putpixel((kx + 1, y0 + 1), sh(tone, .85))
        if rng.random() < .2:                   # nail heads beside a joint
            for j in joints:
                if 1 <= j <= 14:
                    im.putpixel((j + 1, y0 + 1), rgba('#3a2a1a'))
    return im

# ---- walls ----
def beam_v(side):
    """Side wall: a 6px beam on the room side of the cell, void outside."""
    im = blank(VOID)
    xs = range(10, 16) if side == 'L' else range(0, 6)
    for y in range(16):
        for x in xs:
            i = x - xs[0]
            c = BEAM
            if (side == 'L' and i == 0) or (side == 'R' and i == 5):
                c = BEAM_D
            elif (side == 'L' and i == 1) or (side == 'R' and i == 4):
                c = BEAM_HI
            elif (side == 'L' and i == 5) or (side == 'R' and i == 0):
                c = BEAM_LO
            if y % 8 == 7 and c == BEAM:
                c = BEAM_LO                           # board joints
            im.putpixel((x, y), c)
    return im

def beam_h(top=True, cap=None):
    """Bottom wall: a 6px beam along the top of the cell, void below."""
    im = blank(VOID)
    for y in range(6):
        for x in range(16):
            c = BEAM
            if y == 0: c = BEAM_HI
            elif y == 5: c = BEAM_D
            elif y == 4: c = BEAM_LO
            if x % 8 == 7 and 1 <= y <= 3: c = BEAM_LO
            im.putpixel((x, y), c)
    return im

def wall_upper(paper, pat):
    """Back wall, upper tile: the wall top (beam) then wallpaper."""
    im = blank()
    for y in range(16):
        for x in range(16):
            if y < 4:
                c = [BEAM_D, BEAM_HI, BEAM, BEAM_LO][y]
            elif y == 4:
                c = sh(paper, .8)                    # shadow under the beam
            else:
                c = paper
                if (x % 4 == 1 and y % 4 == 2) or (x % 4 == 3 and y % 4 == 0):
                    c = pat                          # small diamond motif
            im.putpixel((x, y), c)
    return im

def wall_lower(paper, pat, gold=False):
    """Back wall, lower tile: wallpaper, chair rail, wainscot, skirting."""
    im = blank()
    for y in range(16):
        for x in range(16):
            if y < 5:
                c = paper
                if (x % 4 == 1 and y % 4 == 2) or (x % 4 == 3 and y % 4 == 0):
                    c = GOLD if gold else pat
            elif y == 5:
                c = RAIL
            elif y == 6:
                c = sh(RAIL, .7)
            elif y < 14:
                c = WAINSCOT                         # panelled wainscot
                px = x % 8
                if px == 0: c = WAIN_LO
                elif px == 1 or y == 7: c = WAIN_HI
                elif y == 13: c = WAIN_LO
            elif y == 14:
                c = SKIRT
            else:
                c = BEAM_D
            im.putpixel((x, y), c)
    return im

def corner_upper(side):
    im = beam_v(side)
    for y in range(4):                               # the wall top turns the corner
        xs = range(10, 16) if side == 'L' else range(0, 6)
        for x in xs:
            im.putpixel((x, y), [BEAM_D, BEAM_HI, BEAM, BEAM_LO][y])
    return im

def doormat():
    im = floor(7)
    for y in range(4, 13):
        for x in range(2, 14):
            edge = x in (2, 13) or y in (4, 12)
            c = rgba('#7a3a2a') if edge else (rgba('#a85a3a') if (x + y) % 3 else rgba('#c07040'))
            im.putpixel((x, y), c)
    return im

def door_cap(side):
    """Bottom beam ending at the doorway (a post)."""
    im = beam_h()
    xs = range(12, 16) if side == 'L' else range(0, 4)
    for y in range(0, 8):
        for x in xs:
            c = BEAM_HI if y == 0 else (BEAM_D if y == 7 or x in (12, 3) else BEAM)
            im.putpixel((x, y), c)
    return im

full = [(0, 0, 16, 16)]
idx = {}
idx['void'] = add('void', blank(VOID), full)
idx['floor'] = [add(f'floor{i}', floor(i), None, {'floor': True}) for i in range(8)]
idx['doormat'] = add('doormat', doormat())
idx['side_L'] = add('side_L', beam_v('L'), full)
idx['side_R'] = add('side_R', beam_v('R'), full)
idx['corner_L'] = add('corner_L', corner_upper('L'), full)
idx['corner_R'] = add('corner_R', corner_upper('R'), full)
idx['bottom'] = add('bottom', beam_h(), full)
idx['door_L'] = add('door_L', door_cap('L'), full)
idx['door_R'] = add('door_R', door_cap('R'), full)
idx['bottom_L'] = add('bottom_L', Image.alpha_composite(beam_v('L'), beam_h()), full)
idx['bottom_R'] = add('bottom_R', Image.alpha_composite(beam_v('R'), beam_h()), full)
idx['walls'] = {}
for name, (paper, pat) in WALLPAPERS.items():
    idx['walls'][name] = {
        'upper': add(f'{name}_upper', wall_upper(paper, pat), full),
        'lower': add(f'{name}_lower', wall_lower(paper, pat, gold=(name == 'bank')), full),
    }
# pad the structure block to a whole row
while len(tiles) % COLS:
    add('pad', blank())

# ---- furniture ----
# index: (cells wide, cells tall, kind) ; kind: floor (collides), rug (no
# collision, drawn under), wall (hung on the back wall, no collision)
WALL = {14, 15, 16, 47, 50, 'clock'}
RUG = {8, 9}
# Pieces redrawn at the character's scale (1x1, 1x2, 2x1), named; see
# README. Packed after the original 64.
SMALL = sorted(f[:-4] for f in os.listdir(os.path.join(HERE, 'furniture-small')) if f.endswith('.png'))
FURN = {}
furn_start = len(tiles)          # whole rows from here are first-fit packed
occupied = set()

def fits(n, cw, ch):
    col = n % COLS
    if col + cw > COLS:
        return False
    return all(n + r * COLS + c not in occupied for r in range(ch) for c in range(cw))

for i in list(range(64)) + SMALL:
    path = os.path.join(HERE, 'furniture', f'{i}.png') if isinstance(i, int) else os.path.join(HERE, 'furniture-small', f'{i}.png')
    src = Image.open(path).convert('RGBA')
    piece = src.crop(src.getbbox())
    w, h = piece.size
    cw, ch = (w + T - 1) // T, (h + T - 1) // T
    block = Image.new('RGBA', (cw * T, ch * T))
    block.paste(piece, ((cw * T - w) // 2, ch * T - h))
    kind = 'wall' if i in WALL else 'rug' if i in RUG else 'floor'
    base = furn_start
    while not fits(base, cw, ch):
        base += 1
    ids = []
    for r in range(ch):
        row = []
        for c in range(cw):
            tid = base + r * COLS + c
            occupied.add(tid)
            while len(tiles) <= tid:
                add('pad', blank())
            cell = block.crop((c * T, r * T, c * T + T, r * T + T))
            cb = cell.getbbox()
            coll = None
            if kind == 'floor' and cb:
                x0, y0, x1, y1 = cb
                coll = [(x0, y0, x1 - x0, y1 - y0)]
            tiles[tid] = (f'furn{i}_{r}_{c}', cell, coll, {})
            row.append(tid if cb else None)   # empty corner of the block: not placed
        ids.append(row)
    FURN[i] = {'cells': ids, 'kind': kind, 'w': cw, 'h': ch}
idx['furn'] = FURN
while len(tiles) % COLS:
    add('pad', blank())

rows = len(tiles) // COLS
sheet = Image.new('RGBA', (COLS * T, rows * T))
for n, (_, im, _, _) in enumerate(tiles):
    sheet.paste(im, ((n % COLS) * T, (n // COLS) * T))
sheet.save(f'{OUT}/house_interior.png')

xml = ['<?xml version="1.0" encoding="UTF-8"?>',
       '<!-- House interiors (isleOfStarting houses). Room pieces: a 2-tile back',
       '     wall (wallpaper: cream, sage, rose, sky, bank), beam side and front',
       '     walls, plank floors, doormat. Furniture made with PixelLab',
       '     (create_1_direction_object), packed onto the 16px grid; rugs and',
       '     wall hangings have no collision. Built by a script; see README. -->',
       f'<tileset version="1.10" tiledversion="1.12.2" name="house_interior" tilewidth="16" tileheight="16" tilecount="{len(tiles)}" columns="{COLS}">',
       f' <image source="house_interior.png" width="{COLS * T}" height="{rows * T}"/>']
for n, (name, im, coll, props) in enumerate(tiles):
    if not coll and not props:
        continue
    xml.append(f' <tile id="{n}">')
    if props:
        xml.append('  <properties>')
        for k, v in props.items():
            xml.append(f'   <property name="{k}" type="bool" value="{str(v).lower()}"/>')
        xml.append('  </properties>')
    if coll:
        xml.append('  <objectgroup draworder="index" id="2">')
        for j, (x, y, w, h) in enumerate(coll):
            xml.append(f'   <object id="{j + 1}" x="{x}" y="{y}" width="{w}" height="{h}"/>')
        xml.append('  </objectgroup>')
    xml.append(' </tile>')
xml.append('</tileset>')
open(f'{OUT}/house_interior.tsx', 'w').write('\n'.join(xml) + '\n')
json.dump(idx, open(f'{OUT}/house_index.json', 'w'), indent=1)
print('tiles', len(tiles), 'rows', rows)

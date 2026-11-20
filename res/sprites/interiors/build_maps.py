"""Write the isleOfStarting house interiors with house_interior.tsx.

Each house keeps its room box (so door teleports and NPC spots stay valid):
R0 = top wall row (now the lower half of a 2-tile back wall; R0-1 its upper
half), R1 = bottom wall row with the doorway at DOOR, C0/C1 side walls.
Furniture: (piece, col, row) = top-left cell of the piece's block.
"""
import json, random, sys, copy

WD = sys.argv[1]           # worlds/isleOfStarting_dev_tiled
IDX = json.load(open(sys.argv[2]))
TSX_REL = '../../../../res/img/textures/tilesets/house_interior.tsx'

F = {(int(k) if k.isdigit() else k): v for k, v in IDX['furn'].items()}

HOUSES = {
    # Furniture at the character's scale: named pieces are 1x1, 1x2 or 2x1
    # (furniture-small/); numbered ones are the 2x2 originals kept as
    # showpieces (fireplace, wardrobe, chalkboard, teller counter, vault,
    # window, painting, pans, rugs).
    '1.2.redHouse.tmj': dict(  # Bill the farmer
        box=(4, 10, 3, 12), door=7, paper='cream', seed=1,
        wall=[(16, 4, 3), (47, 9, 3)],
        rugs=[(9, 6, 8)],
        furn=[(5, 6, 4), ('bed', 11, 4), ('plant', 4, 5), ('stove', 8, 5), ('counter', 9, 5),
              ('armchair', 10, 7), ('lamp', 11, 7), ('table', 4, 8), ('chair', 4, 9), ('chair', 5, 9),
              ('barrel', 11, 8), ('crate', 11, 9)]),
    '0.0.tailorShop.tmj': dict(  # Geraldine
        box=(2, 11, 1, 11), door=6, paper='rose', seed=2,
        wall=[(16, 5, 1), (15, 8, 1)],
        rugs=[(8, 5, 7)],
        # The mannequins are `mannequin` entities in the object layer (kept
        # on rebuild): they wear real outfits you can try on (E).
        furn=[(60, 2, 2), ('drawers', 4, 3), ('fabricshelf', 8, 3),
              ('sewing', 2, 7), ('stool', 2, 8),
              ('coatstand', 2, 9), ('sofa', 3, 10), ('mirror', 10, 9), ('plant', 9, 10)]),
    '0.0.professorHouse.tmj': dict(  # Ivan
        box=(2, 11, 1, 11), door=6, paper='sage', seed=3,
        wall=[(15, 5, 1)],
        rugs=[(9, 5, 8)],
        furn=[('bookshelf', 2, 2), ('bookshelf', 3, 2), ('bookshelf', 4, 2), ('grandfather', 7, 2),
              (19, 9, 2), ('desk', 2, 7), ('chair', 2, 8), ('globe', 10, 6), ('armchair', 2, 10),
              ('lamp', 3, 10), ('sidetable', 10, 10), ('plant', 10, 3), ('stool', 3, 8),
              ('sofa', 7, 10)]),
    '1.2.salon.tmj': dict(  # Claire
        box=(2, 11, 1, 11), door=6, paper='sky', seed=4,
        wall=[(16, 5, 1)],
        rugs=[(9, 5, 8)],
        furn=[('vanity', 2, 2), ('vanity', 4, 2), ('vanity', 8, 2), ('vanity', 10, 2),
              ('barber', 2, 4), ('barber', 4, 4), ('barber', 8, 4), ('barber', 10, 4),
              ('basin', 2, 8), ('basin', 3, 8), ('sidetable', 10, 7), ('mirror', 10, 9),
              ('plant', 2, 10), ('bench', 7, 10), ('drawers', 7, 3), ('stool', 2, 9)]),
    '0.1.hermanHouse.tmj': dict(  # Herman's Rift Bank
        box=(2, 11, 1, 11), door=6, paper='bank', seed=5,
        wall=[(15, 7, 1), ('clock', 5, 1)],
        rugs=[(9, 5, 8)],
        furn=[('cointable', 2, 3), ('cointable', 3, 3), ('ledger', 4, 3), (27, 9, 2),
              (26, 3, 5), (26, 5, 5), (26, 7, 5), ('grandfather', 2, 5), ('plant', 9, 5),
              ('plant', 10, 5), ('bench', 2, 10), ('armchair', 4, 10), ('plant', 10, 10),
              ('lamp', 10, 8), ('sidetable', 5, 10), ('plant', 2, 8)]),
}


def build(name, h, base_doc):
    R0, R1, C0, C1 = h['box']
    W = H = 16
    rng = random.Random(h['seed'])
    room = [IDX['void'] + 1] * (W * H)
    rugs = [0] * (W * H)
    deco = [0] * (W * H)
    furn = [0] * (W * H)

    def put(layer, r, c, tid):
        if 0 <= r < H and 0 <= c < W:
            layer[r * W + c] = tid + 1

    wp = IDX['walls'][h['paper']]
    for c in range(C0, C1 + 1):
        if c == C0:
            put(room, R0 - 1, c, IDX['corner_L']); put(room, R0, c, IDX['side_L'])
        elif c == C1:
            put(room, R0 - 1, c, IDX['corner_R']); put(room, R0, c, IDX['side_R'])
        else:
            put(room, R0 - 1, c, wp['upper']); put(room, R0, c, wp['lower'])
    for r in range(R0 + 1, R1):
        put(room, r, C0, IDX['side_L']); put(room, r, C1, IDX['side_R'])
        for c in range(C0 + 1, C1):
            put(room, r, c, rng.choice(IDX['floor']))
    for c in range(C0, C1 + 1):
        t = IDX['bottom']
        if c == C0: t = IDX['bottom_L']
        elif c == C1: t = IDX['bottom_R']
        elif c == h['door'] - 1: t = IDX['door_L']
        elif c == h['door'] + 1: t = IDX['door_R']
        elif c == h['door']: t = rng.choice(IDX['floor'])
        put(room, R1, c, t)

    def place(layer, items):
        for piece, c, r in items:
            for dr, row in enumerate(F[piece]['cells']):
                for dc, tid in enumerate(row):
                    if tid is not None:
                        put(layer, r + dr, c + dc, tid)
    place(rugs, h['rugs']); place(deco, h['wall']); place(furn, h['furn'])

    doc = copy.deepcopy(base_doc)
    objects = None
    for l in doc.get('layers', []):
        if l['type'] == 'objectgroup':
            objects = l
    if objects is None:
        objects = {'draworder': 'topdown', 'id': 5, 'name': 'Object Layer 1', 'objects': [],
                   'opacity': 1, 'type': 'objectgroup', 'visible': True, 'x': 0, 'y': 0}
    for o in objects['objects']:
        if o.get('type') == 'stoneDoorway':
            o['type'] = 'houseExit'

    def layer(i, nm, data):
        return {'data': data, 'height': H, 'id': i, 'name': nm, 'opacity': 1,
                'type': 'tilelayer', 'visible': True, 'width': W, 'x': 0, 'y': 0}
    objects['id'] = 5
    doc['layers'] = [layer(1, 'room', room), layer(2, 'rugs', rugs), layer(3, 'wall decor', deco),
                     layer(4, 'furniture', furn), objects]
    doc['tilesets'] = [{'firstgid': 1, 'source': TSX_REL}]
    doc['nextlayerid'] = 6
    doc['width'], doc['height'] = W, H
    return doc


if __name__ == '__main__':
    import os
    for name, h in HOUSES.items():
        path = os.path.join(WD, name)
        if os.path.exists(path):
            base = json.load(open(path))
        else:
            base = json.load(open(os.path.join(WD, '0.0.tailorShop.tmj')))
            for l in base['layers']:
                if l['type'] == 'objectgroup':
                    l['objects'] = []
        doc = build(name, h, base)
        if name == '0.1.hermanHouse.tmj':
            objs = [l for l in doc['layers'] if l['type'] == 'objectgroup'][0]
            objs['objects'] = [
                {'height': 16, 'id': 1, 'name': 'Herman', 'rotation': 0, 'type': 'herman',
                 'visible': True, 'width': 16, 'x': 96, 'y': 64},
                {'height': 16, 'id': 2, 'name': 'way out', 'rotation': 0, 'type': 'houseExit',
                 'visible': True, 'width': 16, 'x': 96, 'y': 176,
                 'properties': [{'name': 'teleport_destination', 'type': 'string', 'value': '0.1.tmj'},
                                {'name': 'teleport_x', 'type': 'int', 'value': 7},
                                {'name': 'teleport_y', 'type': 'int', 'value': 10}]},
            ]
            doc['nextobjectid'] = 3
        json.dump(doc, open(path, 'w'), indent=1)
        print('wrote', name)

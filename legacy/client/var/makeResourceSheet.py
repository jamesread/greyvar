#!/usr/bin/env python3

import sys

from PIL import Image

import argparse

parser = argparse.ArgumentParser();
parser.add_argument("--paper_width", default = 2480)
parser.add_argument("--paper_height", default = 3508)
parser.add_argument("--tile_size", default = 256)
parser.add_argument("--tileMargin", default = 40)
parser.add_argument("--pageMargin", default = 50)
parser.add_argument("--tiles", nargs = "*", default = [])
parser.add_argument("--entities", nargs = "*", default = [])
args = parser.parse_args();

list_images = []

for tileRef in args.tiles:
    list_images.append("../../../res/img/textures/tiles/" + tileRef + ".png")

for entRef in args.entities:
    list_images.append("../../../res/img/textures/entities/" + entRef + ".png")

new_image = Image.new("RGBA", (args.paper_width, args.paper_height), (255, 0, 0, 0))

x = args.pageMargin
y = args.pageMargin

for resource_filename in list_images:
    if x + args.tile_size + (args.pageMargin * 2) > args.paper_width:
        x = args.pageMargin
        y += args.tile_size + args.tileMargin

    open_image = Image.open(resource_filename)
    open_image = open_image.resize((args.tile_size, args.tile_size))

    box = (x, y)

    print(box)

    new_image.paste(open_image, box)

    x += args.tile_size + args.tileMargin

new_image.save("out.png")

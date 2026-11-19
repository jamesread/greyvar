# Steps tiles

`res/img/textures/tilesets/steps.png` and `steps.tsx`: stairs up a hill cliff,
with no collision, so the player can climb from the ground onto the plateau.

How they were made:

1. `context_sand.png` and `context_grass.png`: 3x3-tile scenes from
   `grass_terrains`: plateau (70) on top, a cliff row (170 over sand, or 90
   over grass) in the middle, and ground (64 sand, or 67 grass) below.
2. `pixellab_*.png`: PixelLab `create_image_pixflux` img2img of each
   (strength 110, seed 500; jobs f51fdfdb / 35f94d72) with "a narrow stone
   staircase of three steps cut into the middle of the brown cliff face ...".
3. The staircase columns of the grass result's centre tile (rows 6-15,
   x 4-13) are pasted into the original cliff tile (170 over sand for tile
   0, 90 over grass for tile 1), so the plateau above and the cliff either
   side are untouched and meet their neighbours seamlessly. Both tiles use
   the grass staircase: the sand one PixelLab drew was only 6px wide, too
   narrow for the 8px player hitbox once the taller faces block beside it.
4. `res/sprites/faces` grows the steps with the rest of the cliff (the
   stairs repeat downwards and stay walkable).

Use tile 0 where a cliff meets sand, and tile 1 where it meets grass. It
replaces the cliff tile on the terrain layer.

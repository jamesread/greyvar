<?xml version="1.0" encoding="UTF-8"?>
<!-- Stone barriers for doorways in the grey stone (wall terrain) rooms of
     generated puzzles. barrierV closes a left-right passage, barrierH an
     up-down one; _open is the barrier sunk into the ground (walkable).
     Art made with PixelLab on the wall tiles' palette. -->
<tileset version="1.10" tiledversion="1.12.2" name="barrier" tilewidth="16" tileheight="16" tilecount="4" columns="4">
 <image source="../../../res/img/textures/entities/barrier.png" width="64" height="16"/>
 <tile id="0" type="barrierV"/>
 <tile id="1" type="barrierV_open"/>
 <tile id="2" type="barrierH"/>
 <tile id="3" type="barrierH_open"/>
</tileset>

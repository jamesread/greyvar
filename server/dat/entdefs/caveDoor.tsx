<?xml version="1.0" encoding="UTF-8"?>
<!-- caveDoor: a cave mouth (art by PixelLab, painted into the hill's cliff
     face). 16x32: place it on a hill's edge row so the arch sits in the face
     below. Its trigger reaches below the face, where a player walking up
     into it stands (but not the whole cell, so arriving there is safe). -->
<tileset version="1.10" tiledversion="1.12.2" name="caveDoor" tilewidth="16" tileheight="32" tilecount="1" columns="1">
 <image source="../../../res/img/textures/entities/caveDoor.png" width="16" height="32"/>
 <tile id="0" type="caveDoor">
  <properties>
   <property name="collision_type" value="pressure"/>
  </properties>
  <objectgroup draworder="index" id="2">
   <object id="1" x="0" y="28" width="16" height="12"/>
  </objectgroup>
 </tile>
</tileset>

<?xml version="1.0" encoding="UTF-8"?>
<!-- Steps up a hill cliff: walkable (no collision). Stairs from PixelLab
     (res/sprites/steps/), set into the original cliff tiles 170 (over sand)
     and 90 (over grass) of grass_terrains so neighbouring cliffs line up. -->
<tileset version="1.10" tiledversion="1.12.2" name="steps" tilewidth="16" tileheight="16" tilecount="2" columns="2">
 <image source="steps.png" width="32" height="16"/>
 <tile id="0" type="steps">
  <properties>
   <property name="over" value="sand"/>
  </properties>
 </tile>
 <tile id="1" type="steps">
  <properties>
   <property name="over" value="grass"/>
  </properties>
 </tile>
</tileset>

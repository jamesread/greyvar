<?xml version="1.0" encoding="UTF-8"?>
<!-- Doors you walk through: step on one to go to its map object's
     teleport_destination (grid) at teleport_x/teleport_y (tiles). houseDoor is
     a house's front door (res/img/textures/tiles/houseDoor.png), stoneDoorway
     the way out of an interior (stoneCastleDoor.png), houseExit the way out
     of a house interior (its doormat, from house_interior). -->
<tileset version="1.10" tiledversion="1.12.2" name="doors" tilewidth="16" tileheight="16" tilecount="3" columns="3">
 <image source="../../../res/img/textures/entities/doors.png" width="48" height="16"/>
 <tile id="0" type="houseDoor">
  <properties>
   <property name="collision_type" value="pressure"/>
  </properties>
 </tile>
 <tile id="1" type="stoneDoorway">
  <properties>
   <property name="collision_type" value="pressure"/>
  </properties>
 </tile>
 <tile id="2" type="houseExit">
  <properties>
   <property name="collision_type" value="pressure"/>
  </properties>
 </tile>
</tileset>

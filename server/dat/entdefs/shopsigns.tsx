<?xml version="1.0" encoding="UTF-8"?>
<!-- Hanging shop signs over house doors (art: res/sprites/shopsigns). Walk-
     through; within touch_range px the map object's msg (the shop's name)
     shows above it, like a signpost. -->
<tileset version="1.10" tiledversion="1.12.2" name="shopsigns" tilewidth="16" tileheight="16" tilecount="5" columns="5">
 <image source="../../../res/img/textures/entities/shopsigns.png" width="80" height="16"/>
 <tile id="0" type="signHair">
  <properties>
   <property name="collision_type" value="touch"/>
   <property name="touch_range" value="40"/>
  </properties>
 </tile>
 <tile id="1" type="signTailor">
  <properties>
   <property name="collision_type" value="touch"/>
   <property name="touch_range" value="40"/>
  </properties>
 </tile>
 <tile id="2" type="signScience">
  <properties>
   <property name="collision_type" value="touch"/>
   <property name="touch_range" value="40"/>
  </properties>
 </tile>
 <tile id="3" type="signFarm">
  <properties>
   <property name="collision_type" value="touch"/>
   <property name="touch_range" value="40"/>
  </properties>
 </tile>
 <tile id="4" type="signBank">
  <properties>
   <property name="collision_type" value="touch"/>
   <property name="touch_range" value="40"/>
  </properties>
 </tile>
</tileset>

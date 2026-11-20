<?xml version="1.0" encoding="UTF-8"?>
<!-- Rift coin pickup: a small gold coin spinning on the spot (8-frame turn). -->
<tileset version="1.10" tiledversion="1.12.2" name="riftCoin" tilewidth="16" tileheight="16" tilecount="8" columns="8">
 <image source="../../../res/img/textures/entities/riftCoin.png" width="128" height="16"/>
 <tile id="0" type="riftCoin">
  <properties>
   <property name="collision_type" value="pickup"/>
   <property name="item_id" value="rift_coin"/>
  </properties>
  <animation>
   <frame tileid="0" duration="200"/>
   <frame tileid="1" duration="90"/>
   <frame tileid="2" duration="90"/>
   <frame tileid="3" duration="90"/>
   <frame tileid="4" duration="90"/>
   <frame tileid="5" duration="90"/>
   <frame tileid="6" duration="90"/>
   <frame tileid="7" duration="90"/>
  </animation>
 </tile>
</tileset>

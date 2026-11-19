<?xml version="1.0" encoding="UTF-8"?>
<tileset version="1.10" tiledversion="1.12.2" name="water" tilewidth="16" tileheight="16" tilecount="3" columns="3">
 <image source="../fluids/water.png" width="48" height="16"/>
 <tile id="0" type="water">
  <properties>
   <property name="water" type="bool" value="true"/>
  </properties>
  <objectgroup draworder="index" id="2">
   <object id="1" x="0" y="0" width="16" height="16"/>
  </objectgroup>
  <animation>
   <frame tileid="0" duration="500"/>
   <frame tileid="1" duration="500"/>
   <frame tileid="2" duration="500"/>
  </animation>
 </tile>
</tileset>

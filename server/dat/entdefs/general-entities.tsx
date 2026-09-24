<?xml version="1.0" encoding="UTF-8"?>
<tileset version="1.10" tiledversion="1.12.2" name="general-entities" tilewidth="15" tileheight="15" tilecount="30" columns="5">
 <image source="../../../res/img/textures/entities/general-entities.png" width="75" height="90"/>
 <tile id="0" type="key">
  <properties>
   <property name="collision_type" value="pickup"/>
  </properties>
 </tile>
 <tile id="2" type="bottle">
  <properties>
   <property name="collision_type" value="pickup"/>
  </properties>
 </tile>
 <tile id="3" type="sign_low"/>
 <tile id="5" type="tv"/>
 <tile id="8" type="sign_high"/>
 <tile id="9" type="construct_entity"/>
 <tile id="10" type="chest">
  <properties>
   <property name="state_closed" type="int" value="0"/>
  </properties>
 </tile>
 <tile id="13" type="signpost">
  <properties>
   <property name="collision_type" value=""/>
   <property name="msg" value=""/>
  </properties>
  <objectgroup draworder="index" id="2">
   <object id="1" x="0.181818" y="0.0909091" width="15" height="14.9091"/>
  </objectgroup>
 </tile>
 <tile id="15" type="chest_open">
  <animation>
   <frame tileid="10" duration="250"/>
   <frame tileid="15" duration="100"/>
   <frame tileid="16" duration="100"/>
   <frame tileid="17" duration="100"/>
  </animation>
 </tile>
 <tile id="20" type="pressureButton">
  <properties>
   <property name="collision_transition" value="pressed"/>
   <property name="collision_type" value="pressure"/>
  </properties>
 </tile>
 <tile id="21" type="pressureButton_pressed">
  <properties>
   <property name="collision_type" value="pressure"/>
  </properties>
 </tile>
 <tile id="23" type="pot">
  <objectgroup draworder="index" id="2">
   <object id="2" x="-0.181818" y="0.181818" width="15.3636" height="15.2727"/>
  </objectgroup>
 </tile>
 <tile id="25" type="teleportPad">
  <properties>
   <property name="collision_type" value="pressure"/>
   <property name="teleport_destination" value=""/>
   <property name="teleport_world" value=""/>
  </properties>
  <animation>
   <frame tileid="25" duration="180"/>
   <frame tileid="26" duration="180"/>
   <frame tileid="27" duration="180"/>
   <frame tileid="28" duration="180"/>
   <frame tileid="29" duration="180"/>
  </animation>
 </tile>
</tileset>

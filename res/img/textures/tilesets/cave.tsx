<?xml version="1.0" encoding="UTF-8"?>
<!-- Cave terrain (PixelLab top-down Wang set): floor = packed earth, rock =
     raised rock. Rock corners collide (8x8 per corner). Used by
     isleOfStarting 1.1.debugCave.tmj. -->
<tileset version="1.10" tiledversion="1.12.2" name="cave" tilewidth="16" tileheight="16" tilecount="16" columns="4">
 <image source="cave.png" width="64" height="64"/>
 <tile id="0">
  <objectgroup draworder="index" id="2">
   <object id="1" x="0" y="0" width="8" height="8"/>
   <object id="2" x="8" y="0" width="8" height="8"/>
   <object id="3" x="8" y="8" width="8" height="8"/>
  </objectgroup>
 </tile>
 <tile id="1">
  <objectgroup draworder="index" id="2">
   <object id="1" x="0" y="0" width="8" height="8"/>
   <object id="2" x="0" y="8" width="8" height="8"/>
  </objectgroup>
 </tile>
 <tile id="2">
  <objectgroup draworder="index" id="2">
   <object id="1" x="8" y="0" width="8" height="8"/>
  </objectgroup>
 </tile>
 <tile id="3">
  <objectgroup draworder="index" id="2">
   <object id="1" x="0" y="0" width="8" height="8"/>
   <object id="2" x="8" y="0" width="8" height="8"/>
  </objectgroup>
 </tile>
 <tile id="4">
  <objectgroup draworder="index" id="2">
   <object id="1" x="8" y="0" width="8" height="8"/>
   <object id="2" x="0" y="8" width="8" height="8"/>
  </objectgroup>
 </tile>
 <tile id="5">
  <objectgroup draworder="index" id="2">
   <object id="1" x="0" y="0" width="8" height="8"/>
  </objectgroup>
 </tile>
 <tile id="7">
  <objectgroup draworder="index" id="2">
   <object id="1" x="8" y="8" width="8" height="8"/>
  </objectgroup>
 </tile>
 <tile id="8">
  <objectgroup draworder="index" id="2">
   <object id="1" x="0" y="0" width="8" height="8"/>
   <object id="2" x="0" y="8" width="8" height="8"/>
   <object id="3" x="8" y="8" width="8" height="8"/>
  </objectgroup>
 </tile>
 <tile id="9">
  <objectgroup draworder="index" id="2">
   <object id="1" x="0" y="8" width="8" height="8"/>
   <object id="2" x="8" y="8" width="8" height="8"/>
  </objectgroup>
 </tile>
 <tile id="10">
  <objectgroup draworder="index" id="2">
   <object id="1" x="0" y="8" width="8" height="8"/>
  </objectgroup>
 </tile>
 <tile id="11">
  <objectgroup draworder="index" id="2">
   <object id="1" x="8" y="0" width="8" height="8"/>
   <object id="2" x="8" y="8" width="8" height="8"/>
  </objectgroup>
 </tile>
 <tile id="12">
  <objectgroup draworder="index" id="2">
   <object id="1" x="0" y="0" width="8" height="8"/>
   <object id="2" x="8" y="0" width="8" height="8"/>
   <object id="3" x="0" y="8" width="8" height="8"/>
   <object id="4" x="8" y="8" width="8" height="8"/>
  </objectgroup>
 </tile>
 <tile id="13">
  <objectgroup draworder="index" id="2">
   <object id="1" x="0" y="0" width="8" height="8"/>
   <object id="2" x="8" y="0" width="8" height="8"/>
   <object id="3" x="0" y="8" width="8" height="8"/>
  </objectgroup>
 </tile>
 <tile id="14">
  <objectgroup draworder="index" id="2">
   <object id="1" x="0" y="0" width="8" height="8"/>
   <object id="2" x="8" y="8" width="8" height="8"/>
  </objectgroup>
 </tile>
 <tile id="15">
  <objectgroup draworder="index" id="2">
   <object id="1" x="8" y="0" width="8" height="8"/>
   <object id="2" x="0" y="8" width="8" height="8"/>
   <object id="3" x="8" y="8" width="8" height="8"/>
  </objectgroup>
 </tile>
 <wangsets>
  <wangset name="cave" type="corner" tile="-1">
   <wangcolor name="floor" color="#7a4a1e" tile="-1" probability="1"/>
   <wangcolor name="rock" color="#3a2a1a" tile="-1" probability="1"/>
   <wangtile tileid="0" wangid="0,2,0,2,0,1,0,2"/>
   <wangtile tileid="1" wangid="0,1,0,1,0,2,0,2"/>
   <wangtile tileid="2" wangid="0,2,0,1,0,1,0,1"/>
   <wangtile tileid="3" wangid="0,2,0,1,0,1,0,2"/>
   <wangtile tileid="4" wangid="0,2,0,1,0,2,0,1"/>
   <wangtile tileid="5" wangid="0,1,0,1,0,1,0,2"/>
   <wangtile tileid="6" wangid="0,1,0,1,0,1,0,1"/>
   <wangtile tileid="7" wangid="0,1,0,2,0,1,0,1"/>
   <wangtile tileid="8" wangid="0,1,0,2,0,2,0,2"/>
   <wangtile tileid="9" wangid="0,1,0,2,0,2,0,1"/>
   <wangtile tileid="10" wangid="0,1,0,1,0,2,0,1"/>
   <wangtile tileid="11" wangid="0,2,0,2,0,1,0,1"/>
   <wangtile tileid="12" wangid="0,2,0,2,0,2,0,2"/>
   <wangtile tileid="13" wangid="0,2,0,1,0,2,0,2"/>
   <wangtile tileid="14" wangid="0,1,0,2,0,1,0,2"/>
   <wangtile tileid="15" wangid="0,2,0,2,0,2,0,1"/>
  </wangset>
 </wangsets>
</tileset>

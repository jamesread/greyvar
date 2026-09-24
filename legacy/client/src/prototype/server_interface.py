import socket
import threading
import logging
import player
import clgame
import tile
import random
import select
import time
import json

class server_interface:
  sock = None
  thread = None

  def __init__(self, frame, renderer, clgame):
    self.frame = frame
    self.renderer = renderer
    self.connected = False
    self.clgame = clgame

  def connect(self):
    try:
      self.sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM);
      self.sock.connect(("localhost", 1337));
      self.sock.setblocking(True);
      self.thread = threading.Thread(target=self.read)

      logging.debug("Started server listener");
      self.renderer.serverHostname = "CONNECTED"
#     self.renderer.serverHostname = str(self.sock.getpeername())[0]

      self.thread.start();

      self.connected = True
    except:
      self.renderer.serverHostname = "N"
      self.connected = False

  def read(self):
    ETB = chr(0xc0)
    buf = ""

    while self.frame.running:
        char = self.sock.recv(1)
        if not char: break

        if char == ETB:
          self.read_block(buf);
          buf = "";
        else:
          buf += char;
     
    logging.debug("Server reading thread is finished.");

  def read_block(self, block):
    logging.debug("rcvd from server: " + block);

    cmd = block[0:4]
    pkt = block[4:]

    try: 
      pkt = json.loads(pkt)

    except:
      logging.debug("invalid packet from server: " + pkt)
      return

    self.handle_block(cmd, pkt)

  def handle_block(self, cmd, pkt):
    if cmd == "SPWN":
      pass
#     self.renderer.playerMove();
    elif cmd == "PLRJ":
      self.read_player_join(pkt);
    elif cmd == "MOVE":
      self.read_move(pkt)
    elif cmd == "TILE":
      self.read_tile(pkt)
    elif cmd == "PLRU":
      self.read_player_you(pkt)
    elif cmd == "MESG":
      self.read_mesg(pkt);
    else: 
      logging.debug("Unhandled command from server: " + line)
      
  def read_mesg(self, pkg):
    self.renderer.addMessage(pkg['message']);

  def read_player_you(self, plru):
    self.localPlayer = player.player(None, None)

    self.localPlayer.id = plru['id']
    self.localPlayer.nickname = plru['nickname']
    self.localPlayer.tex = plru['skin']

  def read_tile(self, tyle):
    self.clgame.m.setTile(
      tyle['row'],
      tyle['col'], 
      tile.tile(
        tyle['tex'], 
        False, 
        tyle['rot'], 
        tyle['flipV'], 
        tyle['flipH']
        )
      )

  def read_move(self, move):
    print(("players", self.clgame.players))

    plr = self.clgame.getPlayer(move['playerId']);
    plr.lastX = plr.x
    plr.lastY = plr.y
    plr.lastMoved = time.time()
    plr.x = move['posX'];
    plr.y = move['posY'];
    plr.walkState = move['walkState']
       
    self.renderer.draw()

  def read_player_join(self, nplr):
    if self.localPlayer.id == nplr['id']:
      p = self.localPlayer
    else:
      p = player.player(nplr['id'], nplr['nickname'])
      p.tex = nplr['skin']

    logging.debug("client doing player join: "  + ":" + str(nplr['id']) + ":" + str(nplr['nickname']))

    self.renderer.addPlayer(p)
    self.clgame.addPlayer(p, nplr['id'])
    logging.debug(self.clgame.players)

  def send_helo(self):
    helo = {
      "username": random.choice(["Alice", "Bob", "Charles", "Dave"])
    }

    self.send("HELO", helo);

  def disconnect(self):
    if self.sock != None:
      self.send("BYEE")
      self.sock.shutdown(socket.SHUT_RDWR)
      self.sock.close();

  def send_halt(self):
    self.send("HALT");

  def send_move(self, x, y):
    movr = {
      "playerId": self.localPlayer.id,
      "x": x,
      "y": y
    }

    self.send("MOVR", movr);

  def send(self, messageType, payload = None):
    if not self.connected: 
      logging.warning("Cannot move player. Not connected to server.")
      return

    if payload == None:
      payload = {}

    message = messageType + " " + json.dumps(payload, indent=4)
    message = message.encode("utf-8")
    
    logging.debug("sending: " +  message) 

    self.sock.send(message + chr(0xc0));


